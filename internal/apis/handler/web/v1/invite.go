package v1

import (
	"context"

	"github.com/gzydong/go-chat/internal/service"
)

type Invite struct {
	InviteCodeService service.IInviteCodeService
}

// GenerateInviteCode 生成邀请码
func (i *Invite) GenerateInviteCode(ctx context.Context, req *InviteGenerateRequest) (*InviteGenerateResponse, error) {
	userId := GetContextUserId(ctx)

	expireDays := req.ExpireDays
	if expireDays <= 0 {
		expireDays = 365 // 默认1年
	}

	maxUsage := req.MaxUsage
	if maxUsage <= 0 {
		maxUsage = 1 // 默认1次
	}

	inviteCode, err := i.InviteCodeService.GenerateInviteCode(ctx, int(userId), int(expireDays), int(maxUsage))
	if err != nil {
		return nil, err
	}

	return &InviteGenerateResponse{
		Code:          inviteCode.Code,
		ExpireAt:      inviteCode.ExpireAt.Format("2006-01-02 15:04:05"),
		MaxUsageCount: int32(inviteCode.MaxUsageCount),
	}, nil
}

// GetMyInviteCodes 获取我的邀请码列表
func (i *Invite) GetMyInviteCodes(ctx context.Context, req *InviteListRequest) (*InviteListResponse, error) {
	userId := GetContextUserId(ctx)

	codes, err := i.InviteCodeService.GetUserInviteCodes(ctx, int(userId))
	if err != nil {
		return nil, err
	}

	items := make([]*InviteCodeItem, 0, len(codes))
	for _, code := range codes {
		items = append(items, &InviteCodeItem{
			Id:            int32(code.Id),
			Code:          code.Code,
			Status:        int32(code.Status),
			ExpireAt:      code.ExpireAt.Format("2006-01-02 15:04:05"),
			MaxUsageCount: int32(code.MaxUsageCount),
			UsageCount:    int32(code.UsageCount),
			CreatedAt:     code.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &InviteListResponse{
		Items: items,
	}, nil
}

// GetInviteStats 获取邀请统计
func (i *Invite) GetInviteStats(ctx context.Context, req *InviteStatsRequest) (*InviteStatsResponse, error) {
	userId := GetContextUserId(ctx)

	stats, err := i.InviteCodeService.GetInviteStats(ctx, int(userId))
	if err != nil {
		return nil, err
	}

	return &InviteStatsResponse{
		TotalCodes:       int32(stats["total_codes"]),
		AvailableCodes:   int32(stats["available_codes"]),
		UsedCodes:        int32(stats["used_codes"]),
		TotalInvitations: int32(stats["total_invitations"]),
	}, nil
}

// DisableInviteCode 禁用邀请码
func (i *Invite) DisableInviteCode(ctx context.Context, req *InviteDisableRequest) (*InviteDisableResponse, error) {
	userId := GetContextUserId(ctx)

	if err := i.InviteCodeService.DisableInviteCode(ctx, req.Code, int(userId)); err != nil {
		return nil, err
	}

	return &InviteDisableResponse{}, nil
}

// Request and Response types (these would normally be generated from proto)
type InviteGenerateRequest struct {
	ExpireDays int32 `json:"expire_days"`
	MaxUsage   int32 `json:"max_usage"`
}

type InviteGenerateResponse struct {
	Code          string `json:"code"`
	ExpireAt      string `json:"expire_at"`
	MaxUsageCount int32  `json:"max_usage_count"`
}

type InviteListRequest struct{}

type InviteListResponse struct {
	Items []*InviteCodeItem `json:"items"`
}

type InviteCodeItem struct {
	Id            int32  `json:"id"`
	Code          string `json:"code"`
	Status        int32  `json:"status"`
	ExpireAt      string `json:"expire_at"`
	MaxUsageCount int32  `json:"max_usage_count"`
	UsageCount    int32  `json:"usage_count"`
	CreatedAt     string `json:"created_at"`
}

type InviteStatsRequest struct{}

type InviteStatsResponse struct {
	TotalCodes       int32 `json:"total_codes"`
	AvailableCodes   int32 `json:"available_codes"`
	UsedCodes        int32 `json:"used_codes"`
	TotalInvitations int32 `json:"total_invitations"`
}

type InviteDisableRequest struct {
	Code string `json:"code"`
}

type InviteDisableResponse struct{}

// GetContextUserId 从上下文获取用户ID (helper function - should exist in the codebase)
func GetContextUserId(ctx context.Context) int32 {
	// This is a simplified version - the actual implementation should extract from JWT
	// For now, we'll use a placeholder
	return 0
}
