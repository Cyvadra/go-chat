package service

import (
	"context"
	"time"
)

var _ IWalletService = (*MockWalletService)(nil)

// IWalletService 钱包服务接口（对接外部钱包服务）
type IWalletService interface {
	// GetBalance 获取余额
	GetBalance(ctx context.Context, userId int) (float64, error)

	// Recharge 充值
	Recharge(ctx context.Context, userId int, amount float64, payMethod string) (*RechargeResult, error)

	// Transfer 转账
	Transfer(ctx context.Context, fromUserId int, toUserId int, amount float64, remark string) (*TransferResult, error)

	// GetTransactionHistory 获取交易记录
	GetTransactionHistory(ctx context.Context, userId int, startDate, endDate time.Time, page, pageSize int) (*TransactionHistoryResult, error)

	// SendRedEnvelope 发送红包
	SendRedEnvelope(ctx context.Context, req *SendRedEnvelopeRequest) (*RedEnvelopeResult, error)

	// ReceiveRedEnvelope 领取红包
	ReceiveRedEnvelope(ctx context.Context, envelopeId string, userId int) (*ReceiveRedEnvelopeResult, error)

	// GetRedEnvelopeDetail 获取红包详情
	GetRedEnvelopeDetail(ctx context.Context, envelopeId string) (*RedEnvelopeDetail, error)
}

// MockWalletService 钱包服务的Mock实现（用于开发测试）
type MockWalletService struct{}

// RechargeResult 充值结果
type RechargeResult struct {
	OrderId   string    `json:"order_id"`
	Amount    float64   `json:"amount"`
	Balance   float64   `json:"balance"`
	Status    string    `json:"status"` // success, pending, failed
	CreatedAt time.Time `json:"created_at"`
}

// TransferResult 转账结果
type TransferResult struct {
	TransferId string    `json:"transfer_id"`
	FromUserId int       `json:"from_user_id"`
	ToUserId   int       `json:"to_user_id"`
	Amount     float64   `json:"amount"`
	Fee        float64   `json:"fee"`
	Status     string    `json:"status"` // success, pending, failed
	CreatedAt  time.Time `json:"created_at"`
}

// TransactionHistoryResult 交易记录结果
type TransactionHistoryResult struct {
	Items      []*TransactionItem `json:"items"`
	Total      int                `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}

// TransactionItem 交易记录项
type TransactionItem struct {
	Id          string    `json:"id"`
	Type        string    `json:"type"` // recharge, transfer, red_envelope, receive
	Amount      float64   `json:"amount"`
	Balance     float64   `json:"balance"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// SendRedEnvelopeRequest 发送红包请求
type SendRedEnvelopeRequest struct {
	SenderId    int     `json:"sender_id"`
	ChatType    int     `json:"chat_type"` // 1:私聊 2:群聊
	ChatId      int     `json:"chat_id"`
	Amount      float64 `json:"amount"`
	Count       int     `json:"count"`        // 红包个数
	Type        string  `json:"type"`         // normal:普通红包 lucky:拼手气红包
	Greeting    string  `json:"greeting"`     // 祝福语
	Password    string  `json:"password"`     // 支付密码
}

// RedEnvelopeResult 红包结果
type RedEnvelopeResult struct {
	EnvelopeId string    `json:"envelope_id"`
	SenderId   int       `json:"sender_id"`
	Amount     float64   `json:"amount"`
	Count      int       `json:"count"`
	Type       string    `json:"type"`
	Status     string    `json:"status"` // available, finished, expired
	CreatedAt  time.Time `json:"created_at"`
}

// ReceiveRedEnvelopeResult 领取红包结果
type ReceiveRedEnvelopeResult struct {
	EnvelopeId string    `json:"envelope_id"`
	UserId     int       `json:"user_id"`
	Amount     float64   `json:"amount"`
	Status     string    `json:"status"` // success, finished, expired
	ReceivedAt time.Time `json:"received_at"`
}

// RedEnvelopeDetail 红包详情
type RedEnvelopeDetail struct {
	EnvelopeId   string                  `json:"envelope_id"`
	SenderId     int                     `json:"sender_id"`
	SenderName   string                  `json:"sender_name"`
	Amount       float64                 `json:"amount"`
	Count        int                     `json:"count"`
	Type         string                  `json:"type"`
	Greeting     string                  `json:"greeting"`
	Status       string                  `json:"status"`
	ReceivedCount int                    `json:"received_count"`
	ReceivedList []*RedEnvelopeReceiver  `json:"received_list"`
	CreatedAt    time.Time               `json:"created_at"`
}

// RedEnvelopeReceiver 红包领取记录
type RedEnvelopeReceiver struct {
	UserId     int       `json:"user_id"`
	UserName   string    `json:"user_name"`
	Amount     float64   `json:"amount"`
	ReceivedAt time.Time `json:"received_at"`
}

// Mock implementation

func (m *MockWalletService) GetBalance(ctx context.Context, userId int) (float64, error) {
	// Mock data - 实际应调用外部钱包服务
	return 1250.50, nil
}

func (m *MockWalletService) Recharge(ctx context.Context, userId int, amount float64, payMethod string) (*RechargeResult, error) {
	// Mock data
	return &RechargeResult{
		OrderId:   "RCH" + time.Now().Format("20060102150405"),
		Amount:    amount,
		Balance:   1250.50 + amount,
		Status:    "success",
		CreatedAt: time.Now(),
	}, nil
}

func (m *MockWalletService) Transfer(ctx context.Context, fromUserId int, toUserId int, amount float64, remark string) (*TransferResult, error) {
	// Mock data
	return &TransferResult{
		TransferId: "TRF" + time.Now().Format("20060102150405"),
		FromUserId: fromUserId,
		ToUserId:   toUserId,
		Amount:     amount,
		Fee:        0, // 免手续费
		Status:     "success",
		CreatedAt:  time.Now(),
	}, nil
}

func (m *MockWalletService) GetTransactionHistory(ctx context.Context, userId int, startDate, endDate time.Time, page, pageSize int) (*TransactionHistoryResult, error) {
	// Mock data
	items := []*TransactionItem{
		{
			Id:          "TXN001",
			Type:        "recharge",
			Amount:      100.00,
			Balance:     1250.50,
			Description: "账户充值",
			Status:      "success",
			CreatedAt:   time.Now().Add(-24 * time.Hour),
		},
		{
			Id:          "TXN002",
			Type:        "transfer",
			Amount:      -50.00,
			Balance:     1200.50,
			Description: "转账给好友",
			Status:      "success",
			CreatedAt:   time.Now().Add(-12 * time.Hour),
		},
	}

	return &TransactionHistoryResult{
		Items:      items,
		Total:      2,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: 1,
	}, nil
}

func (m *MockWalletService) SendRedEnvelope(ctx context.Context, req *SendRedEnvelopeRequest) (*RedEnvelopeResult, error) {
	// Mock data
	return &RedEnvelopeResult{
		EnvelopeId: "RED" + time.Now().Format("20060102150405"),
		SenderId:   req.SenderId,
		Amount:     req.Amount,
		Count:      req.Count,
		Type:       req.Type,
		Status:     "available",
		CreatedAt:  time.Now(),
	}, nil
}

func (m *MockWalletService) ReceiveRedEnvelope(ctx context.Context, envelopeId string, userId int) (*ReceiveRedEnvelopeResult, error) {
	// Mock data - 随机金额（拼手气红包）
	amount := 10.00 // 简化版本，实际应根据红包类型计算

	return &ReceiveRedEnvelopeResult{
		EnvelopeId: envelopeId,
		UserId:     userId,
		Amount:     amount,
		Status:     "success",
		ReceivedAt: time.Now(),
	}, nil
}

func (m *MockWalletService) GetRedEnvelopeDetail(ctx context.Context, envelopeId string) (*RedEnvelopeDetail, error) {
	// Mock data
	return &RedEnvelopeDetail{
		EnvelopeId:    envelopeId,
		SenderId:      1001,
		SenderName:    "张三",
		Amount:        100.00,
		Count:         10,
		Type:          "lucky",
		Greeting:      "恭喜发财，大吉大利！",
		Status:        "available",
		ReceivedCount: 3,
		ReceivedList: []*RedEnvelopeReceiver{
			{
				UserId:     1002,
				UserName:   "李四",
				Amount:     15.50,
				ReceivedAt: time.Now().Add(-5 * time.Minute),
			},
			{
				UserId:     1003,
				UserName:   "王五",
				Amount:     20.30,
				ReceivedAt: time.Now().Add(-3 * time.Minute),
			},
			{
				UserId:     1004,
				UserName:   "赵六",
				Amount:     8.80,
				ReceivedAt: time.Now().Add(-1 * time.Minute),
			},
		},
		CreatedAt: time.Now().Add(-10 * time.Minute),
	}, nil
}
