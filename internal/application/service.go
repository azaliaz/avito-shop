package application

import (
	"context"
	"github.com/azaliaz/avito-shop/internal/storage"
	"log/slog"
	"time"
)

//go:generate mockgen -source=service.go -destination=./mocks/service_mock.go -package=mocks

type ShopService interface {
	Auth(ctx context.Context, request *AuthRequest) (*AuthResponse, error)
	GetInfo(ctx context.Context, request *GetInfoRequest) (*GetInfoResponse, error)
	SendCoin(ctx context.Context, request *SendCoinRequest) (*SendCoinResponse, error)
	BuyItem(ctx context.Context, request *BuyItemRequest) (*BuyItemResponse, error)
}

type AuthRequest struct {
	Password string
	Username string
}

type AuthResponse struct {
	Token string
}

type GetInfoRequest struct {
	Token string
}

type GetInfoResponse struct {
	CoinHistory *CoinHistory
	Coins       int
	Inventory   []*ProductStock
}

type CoinHistory struct {
	Received []*Transaction
	Sent     []*Transaction
}

type Transaction struct {
	Amount    int
	FromUser  string
	ToUser    string
	CreatedAt time.Time
}

type ProductStock struct {
	Type     string
	Quantity int
}

type SendCoinRequest struct {
	Token  string
	Amount int
	ToUser string
}

type SendCoinResponse struct{}

type BuyItemRequest struct {
	Token string
	Item  string
}

type BuyItemResponse struct {
}

type Service struct {
	log    *slog.Logger
	config *Config
	db     storage.ShopStorage
}

func NewService(
	logger *slog.Logger,
	config *Config,
	db storage.ShopStorage,
) *Service {
	return &Service{
		log:    logger,
		config: config,
		db:     db,
	}
}

func (s *Service) Init() error {
	return nil
}

func (s *Service) Run(ctx context.Context) {

}

func (s *Service) Stop() {

}
