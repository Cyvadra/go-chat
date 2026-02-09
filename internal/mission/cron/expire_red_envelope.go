package cron

import (
	"context"
	"log/slog"

	"github.com/gzydong/go-chat/internal/pkg/core/crontab"
	"github.com/gzydong/go-chat/internal/service"
)

var _ crontab.ICrontab = (*ExpireRedEnvelope)(nil)

type ExpireRedEnvelope struct {
	WalletService service.IWalletService
}

func (c *ExpireRedEnvelope) Name() string {
	return "red_envelope.expire"
}

// Spec 配置定时任务规则
// 每小时执行一次，检查并处理过期红包
// Cron表达式: "0 * * * *" - 在每小时的第0分钟执行 (例如: 01:00, 02:00, 03:00...)
func (c *ExpireRedEnvelope) Spec() string {
	return "0 * * * *"
}

func (c *ExpireRedEnvelope) Enable() bool {
	return true
}

func (c *ExpireRedEnvelope) Do(ctx context.Context) error {
	// 在实际应用中，这里应该：
	// 1. 查询数据库中所有已过期但状态仍为 available 的红包
	// 2. 将这些红包状态更新为 expired
	// 3. 计算未领取的金额
	// 4. 将未领取的金额退回给发红包的用户
	// 5. 记录退款日志
	
	// 由于目前使用的是 MockWalletService，这里只做日志记录
	slog.InfoContext(ctx, "执行红包过期检查和自动退款任务")
	
	// 实际实现示例（伪代码）：
	// expiredEnvelopes := findExpiredRedEnvelopes()
	// for _, envelope := range expiredEnvelopes {
	//     remainingAmount := calculateRemainingAmount(envelope)
	//     if remainingAmount > 0 {
	//         refundToSender(envelope.SenderId, remainingAmount)
	//         updateEnvelopeStatus(envelope.Id, "expired")
	//         logRefund(envelope.Id, remainingAmount)
	//     }
	// }

	return nil
}
