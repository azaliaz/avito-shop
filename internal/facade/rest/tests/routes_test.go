package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/azaliaz/avito-shop/internal/application"
	"github.com/azaliaz/avito-shop/internal/application/mocks"
	"github.com/azaliaz/avito-shop/internal/facade/rest"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAuth_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockApp := mocks.NewMockShopService(ctrl)
	mockApp.EXPECT().Auth(gomock.Any(), &application.AuthRequest{
		Password: "testpass",
		Username: "testuser",
	}).Return(&application.AuthResponse{
		Token: "abc",
	}, nil)

	api := rest.NewAPI(nil, nil, mockApp)
	app := fiber.New()
	app.Add("POST", "/api/auth", api.Auth)
	requestBody, _ := json.Marshal(map[string]string{
		"username": "testuser",
		"password": "testpass",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestAuth_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockApp := mocks.NewMockShopService(ctrl)
	api := rest.NewAPI(nil, nil, mockApp)
	app := fiber.New()
	app.Add("POST", "/api/auth", api.Auth)

	req := httptest.NewRequest(http.MethodPost, "/api/auth", bytes.NewReader([]byte("{invalid_json}")))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestAuth_InvalidCredentials(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockApp := mocks.NewMockShopService(ctrl)
	mockApp.EXPECT().Auth(gomock.Any(), &application.AuthRequest{
		Password: "wrongpass",
		Username: "wronguser",
	}).Return(nil, errors.New("invalid credentials"))

	api := rest.NewAPI(nil, nil, mockApp)
	app := fiber.New()
	app.Add("POST", "/api/auth", api.Auth)

	requestBody, _ := json.Marshal(map[string]string{
		"username": "wronguser",
		"password": "wrongpass",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestBuyItem_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockApp := mocks.NewMockShopService(ctrl)
	mockApp.EXPECT().BuyItem(gomock.Any(), &application.BuyItemRequest{
		Token: "token",
		Item:  "cup",
	}).Return(&application.BuyItemResponse{}, nil)

	api := rest.NewAPI(nil, nil, mockApp)
	app := fiber.New()
	app.Add("GET", "/api/buy/:item", api.BuyItem)
	req := httptest.NewRequest(http.MethodGet, "/api/buy/cup", nil)
	req.Header.Set("Authorization", "Bearer token")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestBuyItem_BadRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockApp := mocks.NewMockShopService(ctrl)
	mockApp.EXPECT().BuyItem(gomock.Any(), &application.BuyItemRequest{
		Token: "token",
		Item:  "cup",
	}).Return(nil, fmt.Errorf("error"))

	api := rest.NewAPI(nil, nil, mockApp)
	app := fiber.New()
	app.Add("GET", "/api/buy/:item", api.BuyItem)
	req := httptest.NewRequest(http.MethodGet, "/api/buy/cup", nil)
	req.Header.Set("Authorization", "Bearer token")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestInfo_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockApp := mocks.NewMockShopService(ctrl)
	mockApp.EXPECT().GetInfo(gomock.Any(), &application.GetInfoRequest{
		Token: "token",
	}).Return(&application.GetInfoResponse{
		CoinHistory: &application.CoinHistory{
			Received: []*application.Transaction{
				{
					Amount:    100,
					FromUser:  "user1",
					ToUser:    "user2",
					CreatedAt: time.Time{},
				},
			},
			Sent: []*application.Transaction{
				{
					Amount:    200,
					FromUser:  "user2",
					ToUser:    "user1",
					CreatedAt: time.Time{},
				},
			},
		},
		Coins: 0,
		Inventory: []*application.ProductStock{
			{
				Type:     "cup",
				Quantity: 1,
			},
		},
	}, nil)

	api := rest.NewAPI(nil, nil, mockApp)
	app := fiber.New()
	app.Add("GET", "/api/info", api.Info)
	req := httptest.NewRequest(http.MethodGet, "/api/info", nil)
	req.Header.Set("Authorization", "Bearer token")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestInfo_BadRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockApp := mocks.NewMockShopService(ctrl)
	mockApp.EXPECT().GetInfo(gomock.Any(), &application.GetInfoRequest{
		Token: "token",
	}).Return(nil, fmt.Errorf("error"))

	api := rest.NewAPI(nil, nil, mockApp)
	app := fiber.New()
	app.Add("GET", "/api/info", api.Info)
	req := httptest.NewRequest(http.MethodGet, "/api/info", nil)
	req.Header.Set("Authorization", "Bearer token")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestSendCoin_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockApp := mocks.NewMockShopService(ctrl)
	mockApp.EXPECT().SendCoin(gomock.Any(), &application.SendCoinRequest{
		Token:  "token",
		Amount: 100,
		ToUser: "username1",
	}).Return(&application.SendCoinResponse{}, nil)

	api := rest.NewAPI(nil, nil, mockApp)
	app := fiber.New()
	app.Add("POST", "/api/sendCoin", api.SendCoin)
	requestBody, _ := json.Marshal(map[string]string{
		"amount": "100",
		"toUser": "username1",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/sendCoin", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestSendCoin_BadRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockApp := mocks.NewMockShopService(ctrl)
	mockApp.EXPECT().SendCoin(gomock.Any(), &application.SendCoinRequest{
		Token:  "token",
		Amount: 100,
		ToUser: "username2",
	}).Return(nil, fmt.Errorf("error"))

	api := rest.NewAPI(nil, nil, mockApp)
	app := fiber.New()
	app.Add("POST", "/api/sendCoin", api.SendCoin)
	requestBody, _ := json.Marshal(map[string]string{
		"amount": "100",
		"toUser": "username2",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/sendCoin", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestSendCoin_InvalidAmount(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockApp := mocks.NewMockShopService(ctrl)

	api := rest.NewAPI(nil, nil, mockApp)
	app := fiber.New()
	app.Add("POST", "/api/sendCoin", api.SendCoin)
	requestBody, _ := json.Marshal(map[string]string{
		"amount": "s",
		"toUser": "1",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/sendCoin", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestSendCoin_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockApp := mocks.NewMockShopService(ctrl)

	api := rest.NewAPI(nil, nil, mockApp)
	app := fiber.New()
	app.Add("POST", "/api/sendCoin", api.SendCoin)
	req := httptest.NewRequest(http.MethodPost, "/api/sendCoin", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
