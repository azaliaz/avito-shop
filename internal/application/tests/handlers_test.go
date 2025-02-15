package tests

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/azaliaz/avito-shop/internal/application"
	"github.com/azaliaz/avito-shop/internal/storage"
	"github.com/azaliaz/avito-shop/internal/storage/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func TestAuth(t *testing.T) {
	ctrl := gomock.NewController(t)

	tests := []struct {
		name   string
		secret string
		req    *application.AuthRequest
		want   func(storage *mocks.MockShopStorage) (*application.AuthResponse, error)
	}{
		{
			name:   "success",
			secret: "secret",
			req: &application.AuthRequest{
				Password: "pass",
				Username: "log",
			},
			want: func(mockStorage *mocks.MockShopStorage) (*application.AuthResponse, error) {
				mockStorage.EXPECT().Auth(gomock.Any(), gomock.Any()).Return(&storage.AuthResponse{
					UserId:   1,
					UserName: "log",
					PassHash: "$2a$10$BaZWnWzCru2yy64fHEFC5e0TB4eDbCzkPFzXjIOkAxcuVMR8FFraW",
				}, nil)
				token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
					"user_id": 1,
				}).SignedString([]byte("secret"))
				require.NoError(t, err)
				return &application.AuthResponse{
					Token: token,
				}, nil
			},
		},
		{
			name:   "invalid creds",
			secret: "secret",
			req: &application.AuthRequest{
				Password: "pass",
				Username: "log",
			},
			want: func(mockStorage *mocks.MockShopStorage) (*application.AuthResponse, error) {
				mockStorage.EXPECT().Auth(gomock.Any(), gomock.Any()).Return(&storage.AuthResponse{
					UserId:   1,
					UserName: "log",
					PassHash: "",
				}, nil)
				return nil, fmt.Errorf("invalid creds")
			},
		},
		{
			name:   "error auth in storage",
			secret: "",
			req: &application.AuthRequest{
				Password: "pass",
				Username: "log",
			},
			want: func(mockStorage *mocks.MockShopStorage) (*application.AuthResponse, error) {
				mockStorage.EXPECT().Auth(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("storage error"))
				return nil, fmt.Errorf("error auth in db: %w", fmt.Errorf("storage error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := mocks.NewMockShopStorage(ctrl)
			want, wantErr := tt.want(mockStorage)

			app := application.NewService(nil, &application.Config{Secret: "secret"}, mockStorage)
			got, err := app.Auth(context.Background(), tt.req)

			assert.Equal(t, want, got)
			assert.Equal(t, wantErr, err)
		})
	}
}

func TestGetInfo(t *testing.T) {
	ctrl := gomock.NewController(t)

	tests := []struct {
		name   string
		secret string
		req    *application.GetInfoRequest
		want   func(storage *mocks.MockShopStorage) (*application.GetInfoResponse, error)
	}{
		{
			name:   "success",
			secret: "secret",
			req: &application.GetInfoRequest{
				Token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I",
			},
			want: func(mockStorage *mocks.MockShopStorage) (*application.GetInfoResponse, error) {
				mockStorage.EXPECT().GetInventory(gomock.Any(), uint64(1)).Return([]*storage.ProductStock{
					{
						Type:     "cup",
						Quantity: 1,
					},
				}, nil)
				mockStorage.EXPECT().GetBalance(gomock.Any(), uint64(1)).Return(800, nil)
				mockStorage.EXPECT().GetCoinHistory(gomock.Any(), uint64(1)).Return(&storage.CoinHistory{
					Received: []*storage.Transaction{
						{
							Amount:    200,
							FromUser:  "user1",
							ToUser:    "user2",
							CreatedAt: time.Time{},
						},
					},
					Sent: []*storage.Transaction{
						{
							Amount:    300,
							FromUser:  "user2",
							ToUser:    "user1",
							CreatedAt: time.Time{},
						},
					},
				}, nil)
				return &application.GetInfoResponse{
					CoinHistory: &application.CoinHistory{
						Received: []*application.Transaction{
							{
								Amount:    200,
								FromUser:  "user1",
								ToUser:    "user2",
								CreatedAt: time.Time{},
							},
						},
						Sent: []*application.Transaction{
							{
								Amount:    300,
								FromUser:  "user2",
								ToUser:    "user1",
								CreatedAt: time.Time{},
							},
						},
					},
					Coins: 800,
					Inventory: []*application.ProductStock{
						{
							Type:     "cup",
							Quantity: 1,
						},
					},
				}, nil
			},
		},
		{
			name:   "error get coin history from db",
			secret: "secret",
			req: &application.GetInfoRequest{
				Token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I",
			},
			want: func(mockStorage *mocks.MockShopStorage) (*application.GetInfoResponse, error) {
				mockStorage.EXPECT().GetInventory(gomock.Any(), uint64(1)).Return([]*storage.ProductStock{
					{
						Type:     "cup",
						Quantity: 1,
					},
				}, nil)
				mockStorage.EXPECT().GetBalance(gomock.Any(), uint64(1)).Return(800, nil)
				mockStorage.EXPECT().GetCoinHistory(gomock.Any(), uint64(1)).Return(nil, fmt.Errorf("error"))
				return nil, fmt.Errorf("error get coin history from db: %w", fmt.Errorf("error"))
			},
		},
		{
			name:   "error get balance from db",
			secret: "secret",
			req: &application.GetInfoRequest{
				Token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I",
			},
			want: func(mockStorage *mocks.MockShopStorage) (*application.GetInfoResponse, error) {
				mockStorage.EXPECT().GetInventory(gomock.Any(), uint64(1)).Return([]*storage.ProductStock{
					{
						Type:     "cup",
						Quantity: 1,
					},
				}, nil)
				mockStorage.EXPECT().GetBalance(gomock.Any(), uint64(1)).Return(0, fmt.Errorf("error"))
				return nil, fmt.Errorf("error get balance from db: %w", fmt.Errorf("error"))
			},
		},
		{
			name:   "error get inventory from db",
			secret: "secret",
			req: &application.GetInfoRequest{
				Token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I",
			},
			want: func(mockStorage *mocks.MockShopStorage) (*application.GetInfoResponse, error) {
				mockStorage.EXPECT().GetInventory(gomock.Any(), uint64(1)).Return(nil, fmt.Errorf("error"))
				return nil, fmt.Errorf("error get inventory from db: %w", fmt.Errorf("error"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := mocks.NewMockShopStorage(ctrl)
			want, wantErr := tt.want(mockStorage)

			app := application.NewService(nil, &application.Config{Secret: "secret"}, mockStorage)
			got, err := app.GetInfo(context.Background(), tt.req)

			assert.Equal(t, want, got)
			assert.Equal(t, wantErr, err)
		})
	}
}

func TestSendCoin(t *testing.T) {
	ctrl := gomock.NewController(t)

	tests := []struct {
		name   string
		secret string
		req    *application.SendCoinRequest
		want   func(storage *mocks.MockShopStorage) (*application.SendCoinResponse, error)
	}{
		{
			name:   "success",
			secret: "secret",
			req: &application.SendCoinRequest{
				Token:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I",
				Amount: 10,
				ToUser: "username2",
			},
			want: func(mockStorage *mocks.MockShopStorage) (*application.SendCoinResponse, error) {
				mockStorage.EXPECT().SendCoin(gomock.Any(), &storage.SendCoinRequest{
					UserId: 1,
					Amount: 10,
					ToUser: "username2",
				}).Return(&storage.SendCoinResponse{}, nil)
				return &application.SendCoinResponse{}, nil
			},
		},
		{
			name:   "error sending coin in db",
			secret: "secret",
			req: &application.SendCoinRequest{
				Token:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I",
				Amount: 10,
				ToUser: "username2",
			},
			want: func(mockStorage *mocks.MockShopStorage) (*application.SendCoinResponse, error) {
				mockStorage.EXPECT().SendCoin(gomock.Any(), &storage.SendCoinRequest{
					UserId: 1,
					Amount: 10,
					ToUser: "username2",
				}).Return(nil, fmt.Errorf("error"))
				return nil, fmt.Errorf("error send coin in db: %w", fmt.Errorf("error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := mocks.NewMockShopStorage(ctrl)
			want, wantErr := tt.want(mockStorage)

			app := application.NewService(nil, &application.Config{Secret: "secret"}, mockStorage)
			got, err := app.SendCoin(context.Background(), tt.req)

			assert.Equal(t, want, got)
			assert.Equal(t, wantErr, err)
		})
	}
}

func TestBuyItem(t *testing.T) {
	ctrl := gomock.NewController(t)

	tests := []struct {
		name   string
		secret string
		req    *application.BuyItemRequest
		want   func(storage *mocks.MockShopStorage) (*application.BuyItemResponse, error)
	}{
		{
			name:   "success",
			secret: "secret",
			req: &application.BuyItemRequest{
				Token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I",
				Item:  "cup",
			},
			want: func(mockStorage *mocks.MockShopStorage) (*application.BuyItemResponse, error) {
				mockStorage.EXPECT().BuyItem(gomock.Any(), &storage.BuyItemRequest{
					UserId: 1,
					Item:   "cup",
				}).Return(&storage.BuyItemResponse{}, nil)
				return &application.BuyItemResponse{}, nil
			},
		},
		{
			name:   "error buy item in db",
			secret: "secret",
			req: &application.BuyItemRequest{
				Token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I",
				Item:  "cup",
			},
			want: func(mockStorage *mocks.MockShopStorage) (*application.BuyItemResponse, error) {
				mockStorage.EXPECT().BuyItem(gomock.Any(), &storage.BuyItemRequest{
					UserId: 1,
					Item:   "cup",
				}).Return(nil, fmt.Errorf("error"))
				return nil, fmt.Errorf("error buy item in db: %w", fmt.Errorf("error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := mocks.NewMockShopStorage(ctrl)
			want, wantErr := tt.want(mockStorage)

			app := application.NewService(nil, &application.Config{Secret: "secret"}, mockStorage)
			got, err := app.BuyItem(context.Background(), tt.req)

			assert.Equal(t, want, got)
			assert.Equal(t, wantErr, err)
		})
	}
}
