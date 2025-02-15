package tests

import (
	"context"
	"fmt"
	"github.com/azaliaz/avito-shop/internal/storage"
	"github.com/azaliaz/avito-shop/migrations"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"log/slog"
	"strconv"
	"testing"
	"time"
)

func (s *RepositoryTestSuite) TestGetBalance() {
	ctx := context.Background()
	userId := uint64(1)
	expectBalance := 800
	prepare := func() {
		conn, err := s.db.Pool().Acquire(ctx)
		require.NoError(s.T(), err)
		defer conn.Release()
		_, err = conn.Exec(ctx, `INSERT INTO users(id, username, password_hash, balance)
					SELECT @user_id, 'abc', 'abc', @balance`,
			pgx.NamedArgs{
				"user_id": userId,
				"balance": expectBalance,
			})
		require.NoError(s.T(), err)
	}
	clear := func() {
		conn, err := s.db.Pool().Acquire(ctx)
		require.NoError(s.T(), err)
		defer conn.Release()
		_, err = conn.Exec(ctx, `DELETE FROM users`)
		require.NoError(s.T(), err)
	}

	prepare()
	fmt.Println("going to get balance")
	balance, err := s.repo.GetBalance(ctx, userId)
	fmt.Println("got balance")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), expectBalance, balance)
	clear()
}

func (s *RepositoryTestSuite) TestAuth() {
	ctx := context.Background()
	username := "test_user"
	passwordHash := "hashed_password"

	prepare := func() {
		conn, err := s.db.Pool().Acquire(ctx)
		require.NoError(s.T(), err)
		defer conn.Release()

		_, err = conn.Exec(ctx, `DELETE FROM users WHERE username = $1`, username)
		require.NoError(s.T(), err)
	}
	clear := func() {
		conn, err := s.db.Pool().Acquire(ctx)
		require.NoError(s.T(), err)
		defer conn.Release()

		_, err = conn.Exec(ctx, `DELETE FROM users WHERE username = $1`, username)
		require.NoError(s.T(), err)
	}

	authRequest := &storage.AuthRequest{
		UserName: username,
		PassHash: passwordHash,
	}

	
	s.T().Run("Authenticate and create new user", func(t *testing.T) {
		authRequest := &storage.AuthRequest{
			UserName: "new_user",
			PassHash: "new_password_hash",
		}

		resp, err := s.repo.Auth(ctx, authRequest)
		require.NoError(s.T(), err)

		assert.Equal(s.T(), "new_user", resp.UserName)
		assert.NotEmpty(s.T(), resp.PassHash)
		assert.Greater(s.T(), resp.UserId, uint64(0))
	})

	
	s.T().Run("Authenticate with incorrect password", func(t *testing.T) {
		prepare()
		resp, err := s.repo.Auth(ctx, authRequest)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), username, resp.UserName)
		assert.NotEmpty(s.T(), resp.PassHash)
		assert.Greater(s.T(), resp.UserId, uint64(0))
	})

	
	s.T().Run("Authenticate with empty password", func(t *testing.T) {
		prepare()

		authRequest := &storage.AuthRequest{
			UserName: username,
			PassHash: "",
		}

		resp, err := s.repo.Auth(ctx, authRequest)
		require.Error(s.T(), err)
		assert.Nil(s.T(), resp)
	})

	
	s.T().Run("Authenticate multiple times with same credentials", func(t *testing.T) {
		prepare()

		authRequest := &storage.AuthRequest{
			UserName: username,
			PassHash: passwordHash,
		}

		resp1, err := s.repo.Auth(ctx, authRequest)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), username, resp1.UserName)

		resp2, err := s.repo.Auth(ctx, authRequest)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), username, resp2.UserName)
		assert.Equal(s.T(), resp1.UserId, resp2.UserId)
	})

	clear()
}

func (s *RepositoryTestSuite) TestGetInventory() {
	ctx := context.Background()
	userId := uint64(1)
	expectInventory := []*storage.ProductStock{
		{Type: "item1", Quantity: 10},
		{Type: "item2", Quantity: 5},
	}

	prepare := func() {
		conn, err := s.db.Pool().Acquire(ctx)
		require.NoError(s.T(), err)
		defer conn.Release()

		_, err = conn.Exec(ctx, `INSERT INTO users(id, username, password_hash, balance)
            VALUES ($1, 'user1', 'password_hash', 0)`,
			userId)
		require.NoError(s.T(), err)

		_, err = conn.Exec(ctx, `DELETE FROM inventory WHERE user_id = $1`, userId)
		require.NoError(s.T(), err)

		for _, product := range expectInventory {
			_, err = conn.Exec(ctx, `INSERT INTO inventory(user_id, item, quantity) 
                VALUES ($1, $2, $3)`,
				userId, product.Type, product.Quantity)
			require.NoError(s.T(), err)
		}
	}
	clear := func() {
		conn, err := s.db.Pool().Acquire(ctx)
		require.NoError(s.T(), err)
		defer conn.Release()

		_, err = conn.Exec(ctx, `DELETE FROM inventory WHERE user_id = $1`, userId)
		require.NoError(s.T(), err)
	}

	prepare()

	inventory, err := s.repo.GetInventory(ctx, userId)
	require.NoError(s.T(), err)

	assert.Len(s.T(), inventory, len(expectInventory))
	for i, item := range inventory {
		assert.Equal(s.T(), expectInventory[i].Type, item.Type)
		assert.Equal(s.T(), expectInventory[i].Quantity, item.Quantity)
	}

	clear()
}

func (s *RepositoryTestSuite) TestGetCoinHistory() {
	ctx := context.Background()
	userId := uint64(1)

	prepare := func() {
		conn, err := s.db.Pool().Acquire(ctx)
		require.NoError(s.T(), err)
		defer conn.Release()

		users := []struct {
			id       uint64
			username string
		}{
			{userId, "user1"},
			{uint64(2), "user2"},
			{uint64(3), "user3"},
			{uint64(4), "user4"},
		}

		for _, user := range users {
			_, err := conn.Exec(ctx, `INSERT INTO users(id, username, password_hash, balance)
            VALUES ($1, $2, 'password_hash', 100)`, user.id, user.username)
			require.NoError(s.T(), err)
		}

		
		transactions := []struct {
			fromUser uint64
			toUser   uint64
			amount   int
		}{
			{userId, uint64(2), 100},
			{userId, uint64(3), 50},
			{userId, uint64(4), 200},
		}

		for _, tx := range transactions {
			_, err := conn.Exec(ctx, `INSERT INTO transactions(from_user_id, to_user_id, amount, created_at)
            VALUES ($1, $2, $3, NOW())`, tx.fromUser, tx.toUser, tx.amount)
			require.NoError(s.T(), err)
		}
	}

	clear := func() {
		conn, err := s.db.Pool().Acquire(ctx)
		require.NoError(s.T(), err)
		defer conn.Release()

		_, err = conn.Exec(ctx, `DELETE FROM transactions WHERE from_user_id = $1 OR to_user_id = $1`, userId)
		require.NoError(s.T(), err)

		_, err = conn.Exec(ctx, `DELETE FROM users WHERE id = $1`, userId)
		require.NoError(s.T(), err)
	}

	prepare()

	coinHistory, err := s.repo.GetCoinHistory(ctx, userId)
	require.NoError(s.T(), err)

	assert.Len(s.T(), coinHistory.Sent, 3)
	assert.Equal(s.T(), "user2", coinHistory.Sent[0].ToUser)
	assert.Equal(s.T(), "user3", coinHistory.Sent[1].ToUser)
	assert.Equal(s.T(), "user4", coinHistory.Sent[2].ToUser)
	assert.Equal(s.T(), 100, coinHistory.Sent[0].Amount)
	assert.Equal(s.T(), 50, coinHistory.Sent[1].Amount)
	assert.Equal(s.T(), 200, coinHistory.Sent[2].Amount)

	assert.Len(s.T(), coinHistory.Received, 0)
	clear()
}

func (s *RepositoryTestSuite) TestSendCoin() {
	ctx := context.Background()
	prepare := func() {
		conn, err := s.db.Pool().Acquire(ctx)
		require.NoError(s.T(), err)
		defer conn.Release()

		users := []struct {
			id       uint64
			username string
		}{
			{1, "user1"},
			{2, "user2"},
		}

		for _, user := range users {
			_, err := conn.Exec(ctx, `INSERT INTO users(id, username, password_hash, balance)
            VALUES ($1, $2, 'password_hash', 100)`, user.id, user.username)
			require.NoError(s.T(), err)
		}
	}

	clear := func() {
		conn, err := s.db.Pool().Acquire(ctx)
		require.NoError(s.T(), err)
		defer conn.Release()
		_, err = conn.Exec(ctx, `DELETE FROM transactions`)
		require.NoError(s.T(), err)
		_, err = conn.Exec(ctx, `DELETE FROM users WHERE id IN (1, 2)`)
		require.NoError(s.T(), err)
	}

	prepare()

	s.T().Run("Test successful SendCoin", func(t *testing.T) {
		request := &storage.SendCoinRequest{
			UserId: 1,
			ToUser: "user2",
			Amount: 50,
		}

		_, err := s.repo.SendCoin(ctx, request)
		require.NoError(t, err)

		var balanceSender int
		conn, err := s.db.Pool().Acquire(ctx)
		require.NoError(t, err)
		defer conn.Release()

		err = conn.QueryRow(ctx, `SELECT balance FROM users WHERE id = $1`, 1).Scan(&balanceSender)
		require.NoError(t, err)
		assert.Equal(t, 50, balanceSender)

		var balanceReceiver int
		err = conn.QueryRow(ctx, `SELECT balance FROM users WHERE id = $1`, 2).Scan(&balanceReceiver)
		require.NoError(t, err)
		assert.Equal(t, 150, balanceReceiver)

		var transactionCount int
		err = conn.QueryRow(ctx, `SELECT count(*) FROM transactions WHERE from_user_id = $1 AND to_user_id = $2`, 1, 2).Scan(&transactionCount)
		require.NoError(t, err)
		assert.Equal(t, 1, transactionCount)
	})

	s.T().Run("Test SendCoin with insufficient balance", func(t *testing.T) {
		request := &storage.SendCoinRequest{
			UserId: 1,
			ToUser: "user2",
			Amount: 200,
		}

		response, err := s.repo.SendCoin(ctx, request)
		assert.Error(t, err)
		assert.Nil(t, response)
	})

	s.T().Run("Test SendCoin to non-existing user", func(t *testing.T) {
		request := &storage.SendCoinRequest{
			UserId: 1,
			ToUser: "user3",
			Amount: 50,
		}

		response, err := s.repo.SendCoin(ctx, request)
		assert.Error(t, err)
		assert.Nil(t, response)
	})

	clear()
}

func (s *RepositoryTestSuite) TestBuyItem() {
	ctx := context.Background()
	userId := uint64(1)
	item := "item1"
	prepare := func() {
		conn, err := s.db.Pool().Acquire(ctx)
		require.NoError(s.T(), err)
		defer conn.Release()
		_, err = conn.Exec(ctx, `INSERT INTO users(id, username, password_hash, balance)
			VALUES ($1, 'user1', 'password_hash', 100)`, userId)
		require.NoError(s.T(), err)

		_, err = conn.Exec(ctx, `INSERT INTO items(name, price) VALUES ($1, 50)`, item)
		require.NoError(s.T(), err)
	}

	clear := func() {
		conn, err := s.db.Pool().Acquire(ctx)
		require.NoError(s.T(), err)
		defer conn.Release()
		_, err = conn.Exec(ctx, `DELETE FROM users WHERE id = $1`, userId)
		require.NoError(s.T(), err)

		_, err = conn.Exec(ctx, `DELETE FROM items WHERE name = $1`, item)
		require.NoError(s.T(), err)

		_, err = conn.Exec(ctx, `DELETE FROM inventory WHERE user_id = $1`, userId)
		require.NoError(s.T(), err)
	}
	prepare()

	request := &storage.BuyItemRequest{
		UserId: userId,
		Item:   item,
	}

	_, err := s.repo.BuyItem(ctx, request)
	require.NoError(s.T(), err)

	conn, err := s.db.Pool().Acquire(ctx)
	require.NoError(s.T(), err)
	defer conn.Release()

	var balance int
	err = conn.QueryRow(ctx, `SELECT balance FROM users WHERE id = $1`, userId).Scan(&balance)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 50, balance)

	var quantity int
	err = conn.QueryRow(ctx, `SELECT quantity FROM inventory WHERE user_id = $1 AND item = $2`, userId, item).Scan(&quantity)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 1, quantity)
	clear()
}

func TestRepositorySuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}

type RepositoryTestSuite struct {
	container *postgres.PostgresContainer
	suite.Suite

	dbConfig storage.Config

	db   *storage.DB
	repo storage.ShopStorage
}

func (s *RepositoryTestSuite) SetupSuite() {
	ctx := context.Background()
	s.dbConfig = s.setupPostgres(ctx)

	logger := slog.Default()
	db := storage.NewDB(&s.dbConfig, logger)
	if err := db.Init(); err != nil {
		require.NoError(s.T(), err)
	}
	s.db = db
	s.repo = storage.NewService(db, logger)
}

func (s *RepositoryTestSuite) SetupTest() {
	ctx := context.Background()

	conn, err := s.db.Pool().Acquire(ctx)
	require.NoError(s.T(), err)
	defer conn.Release()

	_, err = conn.Exec(ctx, `DROP SCHEMA public CASCADE; CREATE SCHEMA public;`)
	require.NoError(s.T(), err)

	require.NoError(s.T(), migrations.PostgresMigrate(s.dbConfig.UrlPostgres()))
}

func (s *RepositoryTestSuite) TearDownTest() {
	require.NoError(s.T(), migrations.PostgresMigrateDown(s.dbConfig.UrlPostgres()))
}
func (s *RepositoryTestSuite) setupPostgres(ctx context.Context) storage.Config {
	cfg := storage.Config{
		Host:             "",
		DbName:           "test-db",
		User:             "user",
		Password:         "1",
		MaxOpenConns:     10,
		ConnIdleLifetime: 60 * time.Second,
		ConnMaxLifetime:  60 * time.Minute,
	}
	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:14-alpine"),
		postgres.WithDatabase(cfg.DbName),
		postgres.WithUsername(cfg.User),
		postgres.WithPassword(cfg.Password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	require.NoError(s.T(), err)
	s.container = pgContainer

	host, err := pgContainer.Host(ctx)
	require.NoError(s.T(), err)
	cfg.Host = host
	ports, err := pgContainer.MappedPort(ctx, "5432")
	require.NoError(s.T(), err)
	cfg.Host += ":" + strconv.Itoa(ports.Int())

	s.dbConfig = cfg
	return cfg
}

func (s *RepositoryTestSuite) TearDownSuite() {
	s.db.Stop()
	s.container.Stop(context.Background(), nil)
}
