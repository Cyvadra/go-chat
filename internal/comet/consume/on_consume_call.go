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

	slog.Info("[CallEvent] Received call event", "event", event, "from_id", in.FromId, "to_id", in.ToId, "room_id", in.RoomId, "call_type", in.CallType)

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
		FromUserId:     in.FromId,
		ToUserId:       in.ToId,
		RoomId:         in.RoomId,
		CallType:       in.CallType,
		FromUserName:   in.FromUserName,
		FromUserAvatar: in.FromUserAvatar,
	})

	sessions := h.serv.SessionManager().GetSessions(int64(in.ToId))
	slog.Info("[CallEvent] Found sessions for target user", "to_id", in.ToId, "session_count", len(sessions))

	if len(sessions) == 0 {
		slog.Warn("[CallEvent] No active sessions found for target user", "to_id", in.ToId, "event", event)
	}

	for _, session := range sessions {
		slog.Info("[CallEvent] Writing to session", "to_id", in.ToId, "conn_id", session.ConnId(), "event", pushEvent)
		if err := session.Write(data); err != nil {
			slog.Error("[CallEvent] session write call message error", "error", err, "to_id", in.ToId, "conn_id", session.ConnId())
		} else {
			slog.Info("[CallEvent] Successfully wrote call event to session", "to_id", in.ToId, "conn_id", session.ConnId(), "event", pushEvent)
		}
	}
}
