package application

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/azaliaz/avito-shop/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

const hashCost = 10

func (s *Service) Auth(ctx context.Context, request *AuthRequest) (*AuthResponse, error) {
	if request.Password == "" {
		return nil, errors.New("password cannot be empty")
	}
	hashedRequestPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), hashCost)
	if err != nil {
		return nil, fmt.Errorf("error generate password hash: %w", err)
	}
	res, err := s.db.Auth(ctx, &storage.AuthRequest{
		UserName: request.Username,
		PassHash: string(hashedRequestPassword),
	})
	if err != nil {
		return nil, fmt.Errorf("error auth in db: %w", err)
	}
	err = bcrypt.CompareHashAndPassword([]byte(res.PassHash), []byte(request.Password))
	if err != nil {
		return nil, errors.New("invalid creds")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": res.UserId,
	})
	t, err := token.SignedString([]byte(s.config.Secret))
	if err != nil {
		return nil, fmt.Errorf("error sign token: %w", err)
	}
	return &AuthResponse{
		Token: t,
	}, nil
}

func (s *Service) GetInfo(ctx context.Context, request *GetInfoRequest) (*GetInfoResponse, error) {
	userId, err := s.userIdFromToken(request.Token)
	if err != nil {
		return nil, fmt.Errorf("error get user id from token: %w", err)
	}

	inventory, err := s.db.GetInventory(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("error get inventory from db: %w", err)
	}
	resInventory := make([]*ProductStock, 0, len(inventory))
	for _, productStock := range inventory {
		resInventory = append(resInventory, &ProductStock{
			Type:     productStock.Type,
			Quantity: productStock.Quantity,
		})
	}
	balance, err := s.db.GetBalance(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("error get balance from db: %w", err)
	}
	coinHistory, err := s.db.GetCoinHistory(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("error get coin history from db: %w", err)
	}
	received := make([]*Transaction, 0, len(coinHistory.Received))
	for _, transaction := range coinHistory.Received {
		received = append(received, &Transaction{
			Amount:    transaction.Amount,
			FromUser:  transaction.FromUser,
			ToUser:    transaction.ToUser,
			CreatedAt: transaction.CreatedAt,
		})
	}
	sent := make([]*Transaction, 0, len(coinHistory.Sent))
	for _, transaction := range coinHistory.Sent {
		sent = append(sent, &Transaction{
			Amount:    transaction.Amount,
			FromUser:  transaction.FromUser,
			ToUser:    transaction.ToUser,
			CreatedAt: transaction.CreatedAt,
		})
	}
	return &GetInfoResponse{
		CoinHistory: &CoinHistory{
			Received: received,
			Sent:     sent,
		},
		Coins:     balance,
		Inventory: resInventory,
	}, nil
}
func (s *Service) SendCoin(ctx context.Context, request *SendCoinRequest) (*SendCoinResponse, error) {
	userId, err := s.userIdFromToken(request.Token)
	if err != nil {
		return nil, fmt.Errorf("error get user id from token: %w", err)
	}

	_, err = s.db.SendCoin(ctx, &storage.SendCoinRequest{
		UserId: userId,
		Amount: request.Amount,
		ToUser: request.ToUser,
	})
	if err != nil {
		return nil, fmt.Errorf("error send coin in db: %w", err)
	}
	return &SendCoinResponse{}, nil
}
func (s *Service) BuyItem(ctx context.Context, request *BuyItemRequest) (*BuyItemResponse, error) {
	userId, err := s.userIdFromToken(request.Token)
	if err != nil {
		return nil, fmt.Errorf("error get user id from token: %w", err)
	}
	_, err = s.db.BuyItem(ctx, &storage.BuyItemRequest{
		UserId: userId,
		Item:   request.Item,
	})
	if err != nil {
		return nil, fmt.Errorf("error buy item in db: %w", err)
	}
	return &BuyItemResponse{}, nil
}

func (s *Service) userIdFromToken(token string) (uint64, error) {
	claims := jwt.MapClaims{}
	t, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.Secret), nil
	})
	if err != nil {
		return 0, fmt.Errorf("error parse token: %w", err)
	}
	if !t.Valid {
		return 0, errors.New("invalid token")
	}
	floatUserId, ok := claims["user_id"].(float64)
	if !ok {
		return 0, errors.New("invalid user_id in token")
	}
	userId := uint64(floatUserId)
	return userId, nil
}
