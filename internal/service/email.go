package service

import (
	"context"
	"fmt"
	"time"

	"github.com/gzydong/go-chat/internal/pkg/strutil"
	"github.com/gzydong/go-chat/internal/repository/cache"
)

var _ IEmailService = (*EmailService)(nil)

type IEmailService interface {
	// Verify 验证邮箱验证码
	Verify(ctx context.Context, channel string, email string, code string) bool
	// Delete 删除邮箱验证码记录
	Delete(ctx context.Context, channel string, email string)
	// Send 发送邮箱验证码
	Send(ctx context.Context, channel string, email string) (string, error)
}

type EmailService struct {
	Storage *cache.EmailStorage
}

// Verify 验证邮箱验证码是否正确
func (e *EmailService) Verify(ctx context.Context, channel string, email string, code string) bool {
	return e.Storage.Verify(ctx, channel, email, code)
}

// Delete 删除邮箱验证码记录
func (e *EmailService) Delete(ctx context.Context, channel string, email string) {
	_ = e.Storage.Del(ctx, channel, email)
}

// Send 发送邮箱验证码
func (e *EmailService) Send(ctx context.Context, channel string, email string) (string, error) {
	code := strutil.GenValidateCode(6)

	// 添加发送记录
	if err := e.Storage.Set(ctx, channel, email, code, 15*time.Minute); err != nil {
		return "", err
	}

	// Send email via SMTP or third-party email service
	// Integration options:
	// 1. SMTP (e.g., Gmail, Office 365)
	// 2. SendGrid: https://sendgrid.com/
	// 3. Mailgun: https://www.mailgun.com/
	// 4. Amazon SES: https://aws.amazon.com/ses/
	// 5. Aliyun DirectMail: https://www.aliyun.com/product/directmail
	//
	// Example implementation with SMTP:
	// import "net/smtp"
	//
	// auth := smtp.PlainAuth("", "sender@example.com", "password", "smtp.example.com")
	// to := []string{email}
	// msg := []byte(fmt.Sprintf("To: %s\r\n" +
	//     "Subject: 验证码\r\n" +
	//     "\r\n" +
	//     "您的验证码是: %s\r\n", email, code))
	// err := smtp.SendMail("smtp.example.com:587", auth, "sender@example.com", to, msg)
	// if err != nil {
	//     return "", fmt.Errorf("failed to send email: %v", err)
	// }

	// For development/testing: log the code instead of sending
	fmt.Printf("Email verification code for %s (channel: %s): %s\n", email, channel, code)

	return code, nil
}
