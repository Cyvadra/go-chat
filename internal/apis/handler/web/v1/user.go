package v1

import (
	"context"
	"strings"

	"github.com/gzydong/go-chat/api/pb/web/v1"
	"github.com/gzydong/go-chat/internal/entity"
	"github.com/gzydong/go-chat/internal/pkg/core/errorx"
	"github.com/gzydong/go-chat/internal/pkg/core/middleware"
	"github.com/gzydong/go-chat/internal/pkg/encrypt"
	"github.com/gzydong/go-chat/internal/pkg/encrypt/rsautil"
	"github.com/gzydong/go-chat/internal/pkg/timeutil"
	"github.com/gzydong/go-chat/internal/repository/repo"
	"github.com/gzydong/go-chat/internal/service"
	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
)

var _ web.IUserHandler = (*User)(nil)

type User struct {
	Redis        *redis.Client
	UsersRepo    *repo.Users
	OrganizeRepo *repo.Organize
	UserService  service.IUserService
	SmsService   service.ISmsService
	Rsa          rsautil.IRsa
}

// Detail 获取登录用户详情接口
//
//	@Summary		User Detail
//	@Description	Get current logged in user details
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			request	body		web.UserDetailRequest	true	"User Detail request"
//	@Success		200		{object}	web.UserDetailResponse
//	@Router			/api/v1/user/detail [post]
//	@Security		Bearer
func (u *User) Detail(ctx context.Context, _ *web.UserDetailRequest) (*web.UserDetailResponse, error) {
	session, _ := middleware.FormContext[entity.WebClaims](ctx)

	user, err := u.UsersRepo.FindByIdWithCache(ctx, int(session.UserId))
	if err != nil {
		return nil, err
	}

	return &web.UserDetailResponse{
		Mobile:   lo.FromPtr(user.Mobile),
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Gender:   int32(user.Gender),
		Motto:    user.Motto,
		Email:    user.Email,
		Birthday: user.Birthday,
	}, nil
}

// Setting 获取用户配置信息接口
//
//	@Summary		User Setting
//	@Description	Get user configuration and profile settings
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			request	body		web.UserSettingRequest	true	"User Setting request"
//	@Success		200		{object}	web.UserSettingResponse
//	@Router			/api/v1/user/setting [post]
//	@Security		Bearer
func (u *User) Setting(ctx context.Context, req *web.UserSettingRequest) (*web.UserSettingResponse, error) {
	session, err := middleware.FormContext[entity.WebClaims](ctx)
	if err != nil {
		return nil, err
	}

	user, err := u.UsersRepo.FindByIdWithCache(ctx, int(session.UserId))
	if err != nil {
		return nil, err
	}

	isOk, err := u.OrganizeRepo.IsQiyeMember(ctx, int(session.UserId))
	if err != nil {
		return nil, err
	}

	return &web.UserSettingResponse{
		UserInfo: &web.UserSettingResponse_UserInfo{
			Uid:      int32(user.Id),
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
			Motto:    user.Motto,
			Gender:   int32(user.Gender),
			IsQiye:   isOk,
			Mobile:   lo.FromPtr(user.Mobile),
			Email:    user.Email,
		},
		Setting: &web.UserSettingResponse_ConfigInfo{},
	}, nil
}

// DetailUpdate 更新用户信息接口
//
//	@Summary		Update User Detail
//	@Description	Update user profile information like nickname, avatar, gender, etc.
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			request	body		web.UserDetailUpdateRequest	true	"Update Detail request"
//	@Success		200		{object}	web.UserDetailUpdateResponse
//	@Router			/api/v1/user/detail-update [post]
//	@Security		Bearer
func (u *User) DetailUpdate(ctx context.Context, req *web.UserDetailUpdateRequest) (*web.UserDetailUpdateResponse, error) {
	session, _ := middleware.FormContext[entity.WebClaims](ctx)

	if req.Birthday != "" {
		if !timeutil.IsDate(req.Birthday) {
			return nil, errorx.New(400, "birthday 错误")
		}
	}

	uid := session.UserId

	_, err := u.UsersRepo.UpdateById(ctx, uid, map[string]any{
		"nickname": strings.TrimSpace(strings.ReplaceAll(req.Nickname, " ", "")),
		"avatar":   req.Avatar,
		"gender":   req.Gender,
		"motto":    req.Motto,
		"birthday": req.Birthday,
	})

	if err != nil {
		return nil, err
	}

	_ = u.UsersRepo.ClearTableCache(ctx, int(uid))
	return &web.UserDetailUpdateResponse{}, nil
}

// PasswordUpdate 更新用户密码接口
//
//	@Summary		Update Password
//	@Description	Change user login password
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			request	body		web.UserPasswordUpdateRequest	true	"Update Password request"
//	@Success		200		{object}	web.UserPasswordUpdateResponse
//	@Router			/api/v1/user/password-update [post]
//	@Security		Bearer
func (u *User) PasswordUpdate(ctx context.Context, in *web.UserPasswordUpdateRequest) (*web.UserPasswordUpdateResponse, error) {
	session, _ := middleware.FormContext[entity.WebClaims](ctx)

	uid := session.UserId
	if uid == 2054 || uid == 2055 {
		return nil, entity.ErrPermissionDenied
	}

	oldPassword, err := u.Rsa.Decrypt(in.OldPassword)
	if err != nil {
		return nil, err
	}

	newPassword, err := u.Rsa.Decrypt(in.NewPassword)
	if err != nil {
		return nil, err
	}

	if err := u.UserService.UpdatePassword(ctx, int(uid), string(oldPassword), string(newPassword)); err != nil {
		return nil, err
	}

	_ = u.UsersRepo.ClearTableCache(ctx, int(uid))
	return nil, nil
}

// MobileUpdate 更新用户手机号接口
//
//	@Summary		Update Mobile
//	@Description	Change user registered mobile number
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			request	body		web.UserMobileUpdateRequest	true	"Update Mobile request"
//	@Success		200		{object}	web.UserMobileUpdateResponse
//	@Router			/api/v1/user/mobile-update [post]
//	@Security		Bearer
func (u *User) MobileUpdate(ctx context.Context, in *web.UserMobileUpdateRequest) (*web.UserMobileUpdateResponse, error) {
	session, _ := middleware.FormContext[entity.WebClaims](ctx)
	uid := session.UserId

	user, _ := u.UsersRepo.FindById(ctx, uid)
	if lo.FromPtr(user.Mobile) == in.Mobile {
		return nil, errorx.New(400, "手机号与原手机号一致无需修改")
	}

	password, err := u.Rsa.Decrypt(in.Password)
	if err != nil {
		return nil, err
	}

	if !encrypt.VerifyPassword(user.Password, string(password)) {
		return nil, entity.ErrAccountOrPasswordError
	}

	if uid == 2054 || uid == 2055 {
		return nil, entity.ErrPermissionDenied
	}

	if !u.SmsService.Verify(ctx, entity.SmsChangeAccountChannel, in.Mobile, in.SmsCode) {
		return nil, entity.ErrSmsCodeError
	}

	_, err = u.UsersRepo.UpdateById(ctx, user.Id, map[string]any{
		"mobile": in.Mobile,
	})

	if err != nil {
		return nil, err
	}

	_ = u.UsersRepo.ClearTableCache(ctx, user.Id)
	return nil, nil
}

// EmailUpdate 更新用户邮箱接口
//
//	@Summary		Update Email
//	@Description	Change user registered email address
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			request	body		web.UserEmailUpdateRequest	true	"Update Email request"
//	@Success		200		{object}	web.UserEmailUpdateResponse
//	@Router			/api/v1/user/email-update [post]
//	@Security		Bearer
func (u *User) EmailUpdate(ctx context.Context, req *web.UserEmailUpdateRequest) (*web.UserEmailUpdateResponse, error) {
	//TODO implement me
	panic("implement me")
}
