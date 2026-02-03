package comet

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gzydong/go-chat/internal/entity"
	"github.com/gzydong/go-chat/internal/logic"
	"github.com/gzydong/go-chat/internal/pkg/jsonutil"
	"github.com/gzydong/go-chat/internal/pkg/longnet"
	"github.com/gzydong/go-chat/internal/pkg/server"
	"github.com/gzydong/go-chat/internal/repository/cache"
	"github.com/tidwall/gjson"
)

var _ longnet.IHandler = (*Handler)(nil)

type Handler struct {
	UserClient  *cache.UserClient
	PushMessage *logic.PushMessage
}

// OnOpen 链接建立成功
func (h *Handler) OnOpen(smg longnet.ISessionManager, s longnet.ISession) {
	if err := h.UserClient.Bind(context.Background(), server.ID(), s.ConnId(), s.UserId()); err != nil {
		_ = s.Close()
		return
	}

	_ = s.Write([]byte(fmt.Sprintf(`{"event":"connect","payload":{"ping_interval":%d,"ping_timeout":%d}}`, smg.Options().PingInterval, smg.Options().PingTimeout)))
}

// OnMessage 接收到消息
func (h *Handler) OnMessage(smg longnet.ISessionManager, c longnet.ISession, message []byte) {
	event := gjson.GetBytes(message, "event").String()

	switch event {
	case "ping":
		_ = h.UserClient.Bind(context.Background(), server.ID(), c.ConnId(), c.UserId())
		_ = c.Write([]byte(`{"event":"pong"}`))

	case "im.message.keyboard":
		_ = h.PushMessage.Push(context.Background(), entity.ImTopicChat, &entity.SubscribeMessage{
			Event: entity.SubEventImMessageKeyboard,
			Payload: jsonutil.Encode(entity.SubEventImMessageKeyboardPayload{
				FromId:   int(c.UserId()),
				ToFromId: int(gjson.GetBytes(message, "payload.to_from_id").Int()),
			}),
		})

	case "im.call.invite", "im.call.accept", "im.call.reject", "im.call.hangup":
		subEvent := ""
		switch event {
		case "im.call.invite":
			subEvent = entity.SubEventImCallInvite
		case "im.call.accept":
			subEvent = entity.SubEventImCallAccept
		case "im.call.reject":
			subEvent = entity.SubEventImCallReject
		case "im.call.hangup":
			subEvent = entity.SubEventImCallHangup
		}

		_ = h.PushMessage.Push(context.Background(), entity.ImTopicChat, &entity.SubscribeMessage{
			Event: subEvent,
			Payload: jsonutil.Encode(entity.SubEventImCallPayload{
				FromId: int(c.UserId()),
				ToId:   int(gjson.GetBytes(message, "payload.to_id").Int()),
				RoomId: int(gjson.GetBytes(message, "payload.room_id").Int()),
				Type:   gjson.GetBytes(message, "payload.type").String(),
			}),
		})
	}
}

// OnClose 链接关闭
func (h *Handler) OnClose(cid int64, uid int64) {
	if err := h.UserClient.UnBind(context.Background(), server.ID(), cid, uid); err != nil {
		slog.Error("unbind error", "error", err)
	}
}
