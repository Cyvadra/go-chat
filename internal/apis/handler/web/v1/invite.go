package v1

import (
	"context"

	pb "github.com/gzydong/go-chat/api/pb/web/v1"
	"github.com/gzydong/go-chat/internal/entity"
	"github.com/gzydong/go-chat/internal/pkg/core/middleware"
	"github.com/gzydong/go-chat/internal/service"
)

type Invite struct {
	InviteCodeService service.IInviteCodeService
}

// GenerateInviteCode 生成邀请码
func (i *Invite) GenerateInviteCode(ctx context.Context, req *pb.InviteGenerateRequest) (*pb.InviteGenerateResponse, error) {
	session, _ := middleware.FormContext[entity.WebClaims](ctx)

	expireDays := req.ExpireDays
	if expireDays <= 0 {
		expireDays = 365 // 默认1年
	}

	maxUsage := req.MaxUsage
	if maxUsage <= 0 {
		maxUsage = 1 // 默认1次
	}

	inviteCode, err := i.InviteCodeService.GenerateInviteCode(ctx, int(session.UserId), int(expireDays), int(maxUsage))
	if err != nil {
		return nil, err
	}

	return &pb.InviteGenerateResponse{
		Code:          inviteCode.Code,
		ExpireAt:      inviteCode.ExpireAt.Format("2006-01-02 15:04:05"),
		MaxUsageCount: int32(inviteCode.MaxUsageCount),
	}, nil
}

// GetMyInviteCodes 获取我的邀请码列表
func (i *Invite) GetMyInviteCodes(ctx context.Context, req *pb.InviteListRequest) (*pb.InviteListResponse, error) {
	session, _ := middleware.FormContext[entity.WebClaims](ctx)

	codes, err := i.InviteCodeService.GetUserInviteCodes(ctx, int(session.UserId))
	if err != nil {
		return nil, err
	}

	items := make([]*pb.InviteCodeItem, 0, len(codes))
	for _, code := range codes {
		items = append(items, &pb.InviteCodeItem{
			Id:            int32(code.Id),
			Code:          code.Code,
			Status:        int32(code.Status),
			ExpireAt:      code.ExpireAt.Format("2006-01-02 15:04:05"),
			MaxUsageCount: int32(code.MaxUsageCount),
			UsageCount:    int32(code.UsageCount),
			CreatedAt:     code.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &pb.InviteListResponse{
		Items: items,
	}, nil
}

// GetInviteStats 获取邀请统计
func (i *Invite) GetInviteStats(ctx context.Context, req *pb.InviteStatsRequest) (*pb.InviteStatsResponse, error) {
	session, _ := middleware.FormContext[entity.WebClaims](ctx)

	stats, err := i.InviteCodeService.GetInviteStats(ctx, int(session.UserId))
	if err != nil {
		return nil, err
	}

	return &pb.InviteStatsResponse{
		TotalCodes:       int32(stats["total_codes"]),
		AvailableCodes:   int32(stats["available_codes"]),
		UsedCodes:        int32(stats["used_codes"]),
		TotalInvitations: int32(stats["total_invitations"]),
	}, nil
}

// DisableInviteCode 禁用邀请码
func (i *Invite) DisableInviteCode(ctx context.Context, req *pb.InviteDisableRequest) (*pb.InviteDisableResponse, error) {
	session, _ := middleware.FormContext[entity.WebClaims](ctx)

	if err := i.InviteCodeService.DisableInviteCode(ctx, req.Code, int(session.UserId)); err != nil {
		return nil, err
	}

	return &pb.InviteDisableResponse{}, nil
}
