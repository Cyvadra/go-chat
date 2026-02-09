package talk

import (
	"context"
	"net/http"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gzydong/go-chat/api/pb/web/v1"
	"github.com/gzydong/go-chat/internal/entity"
	"github.com/gzydong/go-chat/internal/pkg/core/errorx"
	"github.com/gzydong/go-chat/internal/pkg/core/middleware"
	"github.com/gzydong/go-chat/internal/pkg/filesystem"
	"github.com/gzydong/go-chat/internal/pkg/jsonutil"
	"github.com/gzydong/go-chat/internal/pkg/strutil"
	"github.com/gzydong/go-chat/internal/pkg/timeutil"
	"github.com/gzydong/go-chat/internal/repository/model"
	"github.com/gzydong/go-chat/internal/repository/repo"
	"github.com/gzydong/go-chat/internal/service"
	"github.com/samber/lo"
)

var _ web.IMessageHandler = (*Message)(nil)

type Message struct {
	TalkService          service.ITalkService
	AuthService          service.IAuthService
	Filesystem           filesystem.IFilesystem
	GroupMemberRepo      *repo.GroupMember
	TalkRecordFriendRepo *repo.TalkUserMessage
	TalkRecordGroupRepo  *repo.TalkGroupMessage
	TalkRecordsService   service.ITalkRecordService
	GroupMemberService   service.IGroupMemberService
}

// Records 获取会话消息记录
//
//	@Summary		获取消息记录
//	@Description	获取会话的近期消息历史
//	@Tags			消息
//	@Accept			json
//	@Produce		json
//	@Param			request	body		web.MessageRecordsRequest	true	"消息记录请求"
//	@Success		200		{object}	web.MessageRecordsResponse
//	@Router			/api/v1/message/records [post]
//	@Security		Bearer
func (m *Message) Records(ctx context.Context, in *web.MessageRecordsRequest) (*web.MessageRecordsResponse, error) {
	uid := middleware.FormContextAuthId[entity.WebClaims](ctx)

	if in.TalkMode == entity.ChatGroupMode {
		err := m.AuthService.IsAuth(ctx, &service.AuthOption{
			TalkType: int(in.TalkMode),
			UserId:   uid,
			ToFromId: int(in.ToFromId),
		})

		if err != nil {
			return &web.MessageRecordsResponse{
				Items: []*web.MessageRecord{
					{
						MsgId:     strutil.NewMsgId(),
						Sequence:  1,
						MsgType:   entity.ChatMsgSysText,
						FromId:    0,
						IsRevoked: model.No,
						SendTime:  timeutil.DateTime(),
						Extra: jsonutil.Encode(model.TalkRecordExtraText{
							Content: "暂无权限查看群消息",
						}),
						Quote: "{}",
					},
				},
				Cursor: 1,
			}, nil
		}
	}

	records, err := m.TalkRecordsService.FindAllTalkRecords(ctx, &service.FindAllTalkRecordsOpt{
		TalkType:   int(in.TalkMode),
		UserId:     uid,
		ReceiverId: int(in.ToFromId),
		Cursor:     int(in.Cursor),
		Limit:      int(in.Limit),
	})

	if err != nil {
		return nil, err
	}

	cursor := 0
	if length := len(records); length > 0 {
		cursor = records[length-1].Sequence
	}

	return &web.MessageRecordsResponse{
		Items: lo.Map(records, func(item *model.TalkMessageRecord, _ int) *web.MessageRecord {
			return &web.MessageRecord{
				FromId:    int32(item.FromId),
				MsgId:     item.MsgId,
				Sequence:  int32(item.Sequence),
				MsgType:   int32(item.MsgType),
				Nickname:  item.Nickname,
				Avatar:    item.Avatar,
				IsRevoked: int32(item.IsRevoked),
				SendTime:  item.SendTime.Format(time.DateTime),
				Extra:     lo.Ternary(item.IsRevoked == model.Yes, "{}", item.Extra),
				Quote:     item.Quote,
			}
		}),
		Cursor: int32(cursor),
	}, nil
}

// HistoryRecords 获取会话历史消息记录
//
//	@Summary		获取历史消息记录
//	@Description	搜索和筛选会话的历史消息记录
//	@Tags			消息
//	@Accept			json
//	@Produce		json
//	@Param			request	body		web.MessageHistoryRecordsRequest	true	"历史消息请求"
//	@Success		200		{object}	web.MessageHistoryRecordsResponse
//	@Router			/api/v1/message/history-records [post]
//	@Security		Bearer
func (m *Message) HistoryRecords(ctx context.Context, in *web.MessageHistoryRecordsRequest) (*web.MessageHistoryRecordsResponse, error) {
	uid := middleware.FormContextAuthId[entity.WebClaims](ctx)

	if in.TalkMode == entity.ChatGroupMode {
		err := m.AuthService.IsAuth(ctx, &service.AuthOption{
			TalkType: int(in.TalkMode),
			UserId:   uid,
			ToFromId: int(in.ToFromId),
		})

		if err != nil {
			return &web.MessageHistoryRecordsResponse{}, nil
		}
	}

	msgTypes := []int{
		entity.ChatMsgTypeText,
		entity.ChatMsgTypeMixed,
		entity.ChatMsgTypeCode,
		entity.ChatMsgTypeImage,
		entity.ChatMsgTypeVideo,
		entity.ChatMsgTypeAudio,
		entity.ChatMsgTypeFile,
		entity.ChatMsgTypeLocation,
		entity.ChatMsgTypeForward,
		entity.ChatMsgTypeVote,
		entity.ChatMsgTypeRedEnvelope,
		entity.ChatMsgTypeTransfer,
	}

	if slices.Contains(msgTypes, int(in.MsgType)) {
		msgTypes = []int{int(in.MsgType)}
	}

	records, err := m.TalkRecordsService.FindAllTalkRecords(ctx, &service.FindAllTalkRecordsOpt{
		TalkType:   int(in.TalkMode),
		MsgType:    msgTypes,
		UserId:     uid,
		ReceiverId: int(in.ToFromId),
		Cursor:     int(in.Cursor),
		Limit:      int(in.Limit),
	})

	if err != nil {
		return nil, err
	}

	cursor := 0
	if length := len(records); length > 0 {
		cursor = records[length-1].Sequence
	}

	// 补充红包消息的状态信息
	items := m.enrichRedEnvelopeStatus(ctx, uid, records)

	return &web.MessageHistoryRecordsResponse{
		Items:  items,
		Cursor: int32(cursor),
	}, nil
}

// ForwardRecords 转发消息记录
//
//	@Summary		转发消息记录
//	@Description	获取待转发的消息列表
//	@Tags			消息
//	@Accept			json
//	@Produce		json
//	@Param			request	body		web.MessageForwardRecordsRequest	true	"转发消息请求"
//	@Success		200		{object}	web.MessageRecordsClearResponse
//	@Router			/api/v1/message/forward-records [post]
//	@Security		Bearer
func (m *Message) ForwardRecords(ctx context.Context, in *web.MessageForwardRecordsRequest) (*web.MessageRecordsClearResponse, error) {
	uid := middleware.FormContextAuthId[entity.WebClaims](ctx)

	records, err := m.TalkRecordsService.FindForwardRecords(ctx, uid, in.MsgIds, int(in.TalkMode))
	if err != nil {
		return nil, err
	}

	// 补充红包消息的状态信息
	items := m.enrichRedEnvelopeStatus(ctx, uid, records)

	return &web.MessageRecordsClearResponse{
		Items: items,
	}, nil
}

// enrichRedEnvelopeStatus 补充红包消息的状态信息
func (m *Message) enrichRedEnvelopeStatus(ctx context.Context, userId int, records []*model.TalkMessageRecord) []*web.MessageRecord {
	return lo.Map(records, func(item *model.TalkMessageRecord, _ int) *web.MessageRecord {
		extra := lo.Ternary(item.IsRevoked == model.Yes, "{}", item.Extra)

		// 如果是红包消息且未撤回，补充状态信息
		if item.MsgType == entity.ChatMsgTypeRedEnvelope && item.IsRevoked == model.No {
			var redEnvelopeData model.TalkRecordExtraRedEnvelope
			if err := jsonutil.Unmarshal(item.Extra, &redEnvelopeData); err == nil && redEnvelopeData.EnvelopeId != "" {
				// 获取红包状态
				if status, err := m.RedEnvelopeService.GetStatus(ctx, redEnvelopeData.EnvelopeId, userId); err == nil {
					// 合并状态信息到原有的红包数据
					enrichedData := map[string]interface{}{
						"envelope_id":    redEnvelopeData.EnvelopeId,
						"amount":         redEnvelopeData.Amount,
						"count":          redEnvelopeData.Count,
						"type":           redEnvelopeData.Type,
						"greeting":       redEnvelopeData.Greeting,
						"status":         status.Status,
						"status_text":    status.StatusText,
						"has_received":   status.HasReceived,
						"received_amt":   status.ReceivedAmt,
						"is_best":        status.IsBest,
						"best_user_id":   status.BestUserId,
						"best_user_name": status.BestUserName,
						"best_amount":    status.BestAmount,
					}
					extra = jsonutil.Encode(enrichedData)
				}
			}
		}

		return &web.MessageRecord{
			FromId:    int32(item.FromId),
			MsgId:     item.MsgId,
			Sequence:  int32(item.Sequence),
			MsgType:   int32(item.MsgType),
			Nickname:  item.Nickname,
			Avatar:    item.Avatar,
			IsRevoked: int32(item.IsRevoked),
			SendTime:  item.SendTime.Format(time.DateTime),
			Extra:     extra,
			Quote:     item.Quote,
		}
	})
}

// Revoke 撤回消息接口
//
//	@Summary		撤回消息
//	@Description	撤回之前发送的消息
//	@Tags			消息
//	@Accept			json
//	@Produce		json
//	@Param			request	body		web.MessageRevokeRequest	true	"撤回请求"
//	@Success		200		{object}	web.MessageRevokeResponse
//	@Router			/api/v1/message/revoke [post]
//	@Security		Bearer
func (m *Message) Revoke(ctx context.Context, in *web.MessageRevokeRequest) (*web.MessageRevokeResponse, error) {
	uid := middleware.FormContextAuthId[entity.WebClaims](ctx)

	if err := m.TalkService.Revoke(ctx, &service.TalkRevokeOption{
		UserId:   uid,
		TalkMode: int(in.TalkMode),
		MsgId:    in.MsgId,
	}); err != nil {
		return nil, err
	}

	return &web.MessageRevokeResponse{}, nil
}

// Delete 删除消息记录
//
//	@Summary		删除消息
//	@Description	从历史记录中永久移除消息
//	@Tags			消息
//	@Accept			json
//	@Produce		json
//	@Param			request	body		web.MessageDeleteRequest	true	"删除请求"
//	@Success		200		{object}	web.MessageDeleteResponse
//	@Router			/api/v1/message/delete [post]
//	@Security		Bearer
func (m *Message) Delete(ctx context.Context, in *web.MessageDeleteRequest) (*web.MessageDeleteResponse, error) {
	uid := middleware.FormContextAuthId[entity.WebClaims](ctx)

	if err := m.TalkService.DeleteRecord(ctx, &service.TalkDeleteRecordOption{
		UserId:   uid,
		TalkMode: int(in.TalkMode),
		ToFromId: int(in.ToFromId),
		MsgIds:   in.MsgIds,
	}); err != nil {
		return nil, err
	}

	return &web.MessageDeleteResponse{}, nil
}

type DownloadChatFileRequest struct {
	TalkMode int    `form:"talk_mode" json:"talk_mode" binding:"required,oneof=1 2"`
	MsgId    string `form:"msg_id" json:"msg_id" binding:"required"`
}

// Download 聊天文件下载
//
//	@Summary		下载聊天文件
//	@Description	下载聊天会话中分享的文件
//	@Tags			消息
//	@Accept			json
//	@Produce		octet-stream
//	@Param			talk_mode	query		int		true	"对话模式 (1:私聊, 2:群聊)"
//	@Param			msg_id		query		string	true	"消息 ID"
//	@Success		200			{file}		binary
//	@Router			/api/v1/talk/file-download [get]
//	@Security		Bearer
func (m *Message) Download(ctx *gin.Context) error {
	params := &DownloadChatFileRequest{}
	if err := ctx.ShouldBind(params); err != nil {
		return errorx.New(400, err.Error())
	}

	uid := middleware.FormContextAuthId[entity.WebClaims](ctx.Request.Context())

	var fileInfo model.TalkRecordExtraFile
	if params.TalkMode == entity.ChatGroupMode {
		record, err := m.TalkRecordGroupRepo.FindByWhere(ctx, "msg_id = ?", params.MsgId)
		if err != nil {
			return ctx.Error(err)
		}

		if !m.GroupMemberRepo.IsMember(ctx, record.GroupId, uid, false) {
			return entity.ErrPermissionDenied
		}

		if err := jsonutil.Unmarshal(record.Extra, &fileInfo); err != nil {
			return err
		}
	} else {
		record, err := m.TalkRecordFriendRepo.FindByWhere(ctx, "user_id = ? and msg_id = ?", uid, params.MsgId)
		if err != nil {
			return errorx.New(400, "文件不存在")
		}

		if err := jsonutil.Unmarshal(record.Extra, &fileInfo); err != nil {
			return err
		}
	}

	switch m.Filesystem.Driver() {
	case filesystem.LocalDriver:
		filePath := m.Filesystem.(*filesystem.LocalFilesystem).Path(m.Filesystem.BucketPrivateName(), fileInfo.Path)
		ctx.FileAttachment(filePath, fileInfo.Name)
	case filesystem.MinioDriver:
		ctx.Redirect(http.StatusFound, m.Filesystem.PrivateUrl(m.Filesystem.BucketPrivateName(), fileInfo.Path, fileInfo.Name, 60*time.Second))
	default:
		return errorx.New(400, "未知文件驱动类型")
	}

	return nil
}
