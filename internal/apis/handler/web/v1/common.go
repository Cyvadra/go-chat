package v1

import (
	"context"
	"fmt"

	"github.com/gzydong/go-chat/api/pb/web/v1"
	"github.com/gzydong/go-chat/config"
	"github.com/gzydong/go-chat/internal/entity"
	"github.com/gzydong/go-chat/internal/pkg/email"
	"github.com/gzydong/go-chat/internal/pkg/strutil"
	"github.com/gzydong/go-chat/internal/repository/repo"
	"github.com/gzydong/go-chat/internal/service"
)

var _ web.ICommonHandler = (*Common)(nil)

type Common struct {
	Config          *config.Config
	UsersRepo       *repo.Users
	SmsService      service.ISmsService
	UserService     service.IUserService
	EmailClient     *email.Client
	TemplateService service.ITemplateService
}

// SendSms 发送短信验证码接口
//
//	@Summary		发送短信
//	@Description	发送用于登录、注册或更换账号的短信验证码
//	@Tags			公共
//	@Accept			json
//	@Produce		json
//	@Param			request	body		web.CommonSendSmsRequest	true	"发送短信请求"
//	@Success		200		{object}	web.CommonSendSmsResponse
//	@Router			/api/v1/common/send-sms [post]
func (c *Common) SendSms(ctx context.Context, in *web.CommonSendSmsRequest) (*web.CommonSendSmsResponse, error) {
	switch in.Channel {
	// 需要判断账号是否存在
	case entity.SmsLoginChannel, entity.SmsForgetAccountChannel:
		if !c.UsersRepo.IsMobileExist(ctx, in.Mobile) {
			return nil, entity.ErrAccountOrPassword
		}

	// 需要判断账号是否存在
	case entity.SmsRegisterChannel, entity.SmsChangeAccountChannel:
		if c.UsersRepo.IsMobileExist(ctx, in.Mobile) {
			return nil, entity.ErrPhoneExist
		}
	case entity.SmsOauthBindChannel:
	default:
		return nil, entity.ErrSmsChannelInvalid
	}

	// 发送短信验证码
	code, err := c.SmsService.Send(ctx, in.Channel, in.Mobile)
	if err != nil {
		return nil, err
	}

	if in.Channel == entity.SmsRegisterChannel || in.Channel == entity.SmsChangeAccountChannel || in.Channel == entity.SmsOauthBindChannel {
		return &web.CommonSendSmsResponse{
			SmsCode: code,
		}, nil
	}

	return &web.CommonSendSmsResponse{}, nil
}

// SendEmail 发送邮件验证码接口
//
//	@Summary		发送邮件
//	@Description	发送邮件验证码
//	@Tags			公共
//	@Accept			json
//	@Produce		json
//	@Param			request	body		web.CommonSendEmailRequest	true	"发送邮件请求"
//	@Success		200		{object}	web.CommonSendEmailResponse
//	@Router			/api/v1/common/send-email [post]
func (c *Common) SendEmail(ctx context.Context, req *web.CommonSendEmailRequest) (*web.CommonSendEmailResponse, error) {
	// Generate 6-digit verification code
	code := strutil.GenValidateCode(6)

	// Store verification code in cache (reusing SMS storage mechanism)
	_, err := c.SmsService.Send(ctx, "email_verify", req.Email)
	if err != nil {
		return nil, err
	}

	// Prepare email template data
	templateData := map[string]string{
		"code": code,
	}

	// Render email template
	body, err := c.TemplateService.CodeTemplate(templateData)
	if err != nil {
		return nil, err
	}

	// Send email
	if c.EmailClient != nil {
		err = c.EmailClient.SendMail(&email.Option{
			To:      []string{req.Email},
			Subject: "验证码",
			Body:    body,
		})
		if err != nil {
			return nil, err
		}
	} else {
		// If email client is not configured, just log the code (for development)
		fmt.Printf("Email verification code for %s: %s\n", req.Email, code)
	}

	return &web.CommonSendEmailResponse{}, nil
}

// Test 发送测试接口
//
//	@Summary		测试端点
//	@Description	内部测试端点
//	@Tags			公共
//	@Accept			json
//	@Produce		json
//	@Param			request	body		web.CommonSendTestRequest	true	"测试请求"
//	@Success		200		{object}	web.CommonSendTestResponse
//	@Router			/api/v1/common/send-test [post]
func (c *Common) Test(ctx context.Context, req *web.CommonSendTestRequest) (*web.CommonSendTestResponse, error) {
	// This is a test endpoint for internal testing purposes
	// Log the request for debugging
	fmt.Printf("Test endpoint called with email: %s\n", req.Email)
	
	// Return empty response indicating success
	return &web.CommonSendTestResponse{}, nil
}
