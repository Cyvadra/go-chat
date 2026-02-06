package talk

import (
	"context"
	"fmt"

	"github.com/gzydong/go-chat/api/pb/web/v1"
	"github.com/gzydong/go-chat/internal/entity"
	"github.com/gzydong/go-chat/internal/pkg/core/middleware"
	"github.com/gzydong/go-chat/internal/pkg/timeutil"
	"github.com/gzydong/go-chat/internal/repository/cache"
	"github.com/gzydong/go-chat/internal/repository/repo"
	"github.com/gzydong/go-chat/internal/service"
)

var _ web.ITalkHandler = (*Session)(nil)

type Session struct {
	RedisLock          *cache.RedisLock
	MessageStorage     *cache.MessageStorage
	UnreadStorage      *cache.UnreadStorage
	ContactRemark      *cache.ContactRemark
	ContactRepo        *repo.Contact
	UsersRepo          *repo.Users
	GroupRepo          *repo.Group
	TalkService        service.ITalkService
	TalkSessionService service.ITalkSessionService
	UserService        service.IUserService
	GroupService       service.IGroupService
	AuthService        service.IAuthService
}

// SessionCreate 会话创建接口
//
//	@Summary		创建会话
//	@Description	与用户或群组创建一个新的聊天会话
//	@Tags			会话
//	@Accept			json
//	@Produce		json
//	@Param			request	body		web.TalkSessionCreateRequest	true	"创建会话请求"
//	@Success		200		{object}	web.TalkSessionCreateResponse
//	@Router			/api/v1/talk/session-create [post]
//	@Security		Bearer
func (s *Session) SessionCreate(ctx context.Context, in *web.TalkSessionCreateRequest) (*web.TalkSessionCreateResponse, error) {
	uid := middleware.FormContextAuthId[entity.WebClaims](ctx)

	// Agent identifier for session tracking (currently not used, reserved for future client type tracking)
	agent := ""

	// 判断对方是否是自己
	if in.TalkMode == entity.ChatPrivateMode && int(in.ToFromId) == uid {
		return nil, entity.ErrPermissionDenied
	}

	key := fmt.Sprintf("talk:list:%d-%d-%d-%s", uid, in.ToFromId, in.TalkMode, agent)
	if !s.RedisLock.Lock(ctx, key, 10) {
		return nil, entity.ErrTooFrequentOperation
	}

	if s.AuthService.IsAuth(ctx, &service.AuthOption{
		TalkType: int(in.TalkMode),
		UserId:   uid,
		ToFromId: int(in.ToFromId),
	}) != nil {
		return nil, entity.ErrPermissionDenied
	}

	result, err := s.TalkSessionService.Create(ctx, &service.TalkSessionCreateOpt{
		UserId:     uid,
		TalkType:   int(in.TalkMode),
		ReceiverId: int(in.ToFromId),
	})
	if err != nil {
		return nil, err
	}

	item := &web.TalkSessionItem{
		Id:        int32(result.Id),
		TalkMode:  int32(result.TalkMode),
		ToFromId:  int32(result.ToFromId),
		IsTop:     int32(result.IsTop),
		IsDisturb: int32(result.IsDisturb),
		IsRobot:   int32(result.IsRobot),
		Name:      "",
		Avatar:    "",
		Remark:    "",
		UnreadNum: 0,
		MsgText:   "",
		UpdatedAt: timeutil.DateTime(),
	}

	if item.TalkMode == entity.ChatPrivateMode {
		item.UnreadNum = int32(s.UnreadStorage.Get(ctx, uid, 1, int(in.ToFromId)))

		item.Remark = s.ContactRepo.GetFriendRemark(ctx, uid, int(in.ToFromId))
		if user, err := s.UsersRepo.FindById(ctx, result.ToFromId); err == nil {
			item.Name = user.Nickname
			item.Avatar = user.Avatar
		}
	} else if result.TalkMode == entity.ChatGroupMode {
		if group, err := s.GroupRepo.FindById(ctx, int(in.ToFromId)); err == nil {
			item.Name = group.Name
			item.Avatar = group.Avatar
		}
	}

	// 查询缓存消息
	if msg, err := s.MessageStorage.Get(ctx, result.TalkMode, uid, result.ToFromId); err == nil {
		item.MsgText = msg.Content
		item.UpdatedAt = msg.Datetime
	}

	return &web.TalkSessionCreateResponse{
		Id:        item.Id,
		TalkMode:  item.TalkMode,
		ToFromId:  item.ToFromId,
		IsTop:     item.IsTop,
		IsDisturb: item.IsDisturb,
		IsRobot:   item.IsRobot,
		Name:      item.Name,
		Avatar:    item.Avatar,
		Remark:    item.Remark,
		UnreadNum: item.UnreadNum,
		MsgText:   item.MsgText,
		UpdatedAt: item.UpdatedAt,
	}, nil
}

// SessionDelete 会话删除接口
//
//	@Summary		删除会话
//	@Description	从列表中移除聊天会话
//	@Tags			会话
//	@Accept			json
//	@Produce		json
//	@Param			request	body		web.TalkSessionDeleteRequest	true	"删除会话请求"
//	@Success		200		{object}	web.TalkSessionDeleteResponse
//	@Router			/api/v1/talk/session-delete [post]
//	@Security		Bearer
func (s *Session) SessionDelete(ctx context.Context, in *web.TalkSessionDeleteRequest) (*web.TalkSessionDeleteResponse, error) {
	uid := middleware.FormContextAuthId[entity.WebClaims](ctx)

	if err := s.TalkSessionService.Delete(ctx, uid, int(in.TalkMode), int(in.ToFromId)); err != nil {
		return nil, err
	}

	return &web.TalkSessionDeleteResponse{}, nil
}

// SessionTop 会话置顶接口
//
//	@Summary		置顶会话
//	@Description	将聊天会话置顶或取消置顶
//	@Tags			会话
//	@Accept			json
//	@Produce		json
//	@Param			request	body		web.TalkSessionTopRequest	true	"置顶会话请求"
//	@Success		200		{object}	web.TalkSessionTopResponse
//	@Router			/api/v1/talk/session-top [post]
//	@Security		Bearer
func (s *Session) SessionTop(ctx context.Context, in *web.TalkSessionTopRequest) (*web.TalkSessionTopResponse, error) {
	uid := middleware.FormContextAuthId[entity.WebClaims](ctx)
	isTop, err := s.TalkSessionService.Top(ctx, &service.TalkSessionTopOpt{
		UserId:   uid,
		TalkMode: int(in.TalkMode),
		ToFromId: int(in.ToFromId),
		Action:   int(in.Action),
	})
	if err != nil {
		return nil, err
	}

	return &web.TalkSessionTopResponse{
		IsTop: int32(isTop),
	}, nil
}

// SessionDisturb 会话免打扰接口
//
//	@Summary		会话免打扰
//	@Description	为聊天会话启用或禁用免打扰模式
//	@Tags			会话
//	@Accept			json
//	@Produce		json
//	@Param			request	body		web.TalkSessionDisturbRequest	true	"免打扰会话请求"
//	@Success		200		{object}	web.TalkSessionDisturbResponse
//	@Router			/api/v1/talk/session-disturb [post]
//	@Security		Bearer
func (s *Session) SessionDisturb(ctx context.Context, in *web.TalkSessionDisturbRequest) (*web.TalkSessionDisturbResponse, error) {
	uid := middleware.FormContextAuthId[entity.WebClaims](ctx)
	isDisturb, err := s.TalkSessionService.Disturb(ctx, &service.TalkSessionDisturbOpt{
		UserId:   uid,
		TalkMode: int(in.TalkMode),
		ToFromId: int(in.ToFromId),
		Action:   int(in.Action),
	})
	if err != nil {
		return nil, err
	}

	return &web.TalkSessionDisturbResponse{
		IsDisturb: int32(isDisturb),
	}, nil
}

// SessionDetail 会话详情接口
//
//	@Summary		会话详情
//	@Description	获取聊天会话的详细设置信息
//	@Tags			会话
//	@Accept			json
//	@Produce		json
//	@Param			request	body		web.TalkSessionDetailRequest	true	"会话详情请求"
//	@Success		200		{object}	web.TalkSessionDetailResponse
//	@Router			/api/v1/talk/session-detail [post]
//	@Security		Bearer
func (s *Session) SessionDetail(ctx context.Context, in *web.TalkSessionDetailRequest) (*web.TalkSessionDetailResponse, error) {
	uid := middleware.FormContextAuthId[entity.WebClaims](ctx)

	detail, err := s.TalkSessionService.SessionDetail(ctx, uid, int(in.TalkMode), int(in.ToFromId))
	if err != nil {
		return nil, err
	}

	return &web.TalkSessionDetailResponse{
		IsTop:     int32(detail.IsTop),
		IsDisturb: int32(detail.IsDisturb),
	}, nil
}

// SessionList 会话列表接口
//
//	@Summary		会话列表
//	@Description	获取用户的所有聊天会话列表
//	@Tags			会话
//	@Accept			json
//	@Produce		json
//	@Param			request	body		web.TalkSessionListRequest	true	"会话列表请求"
//	@Success		200		{object}	web.TalkSessionListResponse
//	@Router			/api/v1/talk/session-list [post]
//	@Security		Bearer
func (s *Session) SessionList(ctx context.Context, req *web.TalkSessionListRequest) (*web.TalkSessionListResponse, error) {
	uid := middleware.FormContextAuthId[entity.WebClaims](ctx)

	data, err := s.TalkSessionService.List(ctx, uid)
	if err != nil {
		return nil, err
	}

	friends := make([]int, 0)
	for _, item := range data {
		if item.TalkMode == 1 {
			friends = append(friends, item.ToFromId)
		}
	}

	// 获取好友备注
	remarks, _ := s.ContactRepo.Remarks(ctx, uid, friends)

	items := make([]*web.TalkSessionItem, 0)
	for _, item := range data {
		value := &web.TalkSessionItem{
			Id:        int32(item.Id),
			TalkMode:  int32(item.TalkMode),
			ToFromId:  int32(item.ToFromId),
			IsTop:     int32(item.IsTop),
			IsDisturb: int32(item.IsDisturb),
			IsRobot:   int32(item.IsRobot),
			Avatar:    item.Avatar,
			MsgText:   "...",
			UpdatedAt: timeutil.FormatDatetime(item.UpdatedAt),
			UnreadNum: int32(s.UnreadStorage.Get(ctx, uid, item.TalkMode, item.ToFromId)),
		}

		if item.TalkMode == entity.ChatPrivateMode {
			value.Name = item.Nickname
			value.Avatar = item.Avatar
			value.Remark = remarks[item.ToFromId]
		} else {
			value.Name = item.GroupName
			value.Avatar = item.GroupAvatar
		}

		// 查询缓存消息
		if msg, err := s.MessageStorage.Get(ctx, item.TalkMode, uid, item.ToFromId); err == nil {
			value.MsgText = msg.Content
			value.UpdatedAt = msg.Datetime
		}

		items = append(items, value)
	}

	return &web.TalkSessionListResponse{Items: items}, nil
}

// SessionClearUnreadNum 会话未读数清除接口
//
//	@Summary		清除会话未读数
//	@Description	将某个会话中的所有消息标记为已读
//	@Tags			会话
//	@Accept			json
//	@Produce		json
//	@Param			request	body		web.TalkSessionClearUnreadNumRequest	true	"清除未读数请求"
//	@Success		200		{object}	web.TalkSessionClearUnreadNumResponse
//	@Router			/api/v1/talk/session-clear-unread-num [post]
//	@Security		Bearer
func (s *Session) SessionClearUnreadNum(ctx context.Context, in *web.TalkSessionClearUnreadNumRequest) (*web.TalkSessionClearUnreadNumResponse, error) {
	uid := middleware.FormContextAuthId[entity.WebClaims](ctx)
	s.UnreadStorage.Reset(ctx, uid, int(in.TalkMode), int(in.ToFromId))
	return &web.TalkSessionClearUnreadNumResponse{}, nil
}
