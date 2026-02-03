package consume

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/gzydong/go-chat/internal/entity"
	"github.com/gzydong/go-chat/internal/pkg/logger"
)

// 通话事件消息
func (h *Handler) onConsumeTalkCall(ctx context.Context, body []byte, event string) {
	var in entity.SubEventImCallPayload

	if err := json.Unmarshal(body, &in); err != nil {
		logger.Errorf("[ChatSubscribe] onConsumeTalkCall Unmarshal err: %s", err.Error())
		return
	}

	pushEvent := ""
	switch event {
	case entity.SubEventImCallInvite:
		pushEvent = entity.PushEventImCallInvite
	case entity.SubEventImCallAccept:
		pushEvent = entity.PushEventImCallAccept
	case entity.SubEventImCallReject:
		pushEvent = entity.PushEventImCallReject
	case entity.SubEventImCallHangup:
		pushEvent = entity.PushEventImCallHangup
	}

	data := Message(pushEvent, entity.ImCallPayload{
		FromId: in.FromId,
		ToId:   in.ToId,
		RoomId: in.RoomId,
		Type:   in.Type,
	})

	for _, session := range h.serv.SessionManager().GetSessions(int64(in.ToId)) {
		if err := session.Write(data); err != nil {
			slog.Error("session write call message error", "error", err)
		}
	}
}
