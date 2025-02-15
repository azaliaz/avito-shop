package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"log/slog"
)

func (r *Service) Auth(ctx context.Context, request *AuthRequest) (*AuthResponse, error) {
	if request.PassHash == "" {
		return nil, errors.New("password cannot be empty")
	}
	conn, err := r.Pool().Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			r.logger.Error("rollback error", slog.String("err", err.Error()))
		}
	}()
	_, err = tx.Exec(ctx,
		`INSERT INTO users(username, password_hash, balance)
					SELECT @username, @password_hash, 1000
					WHERE NOT EXISTS(SELECT 1 FROM users WHERE username = @username)`,
		pgx.NamedArgs{
			"username":      request.UserName,
			"password_hash": request.PassHash,
		},
	)
	if err != nil {
		return nil, errors.New("error creating user")
	}

	var userId uint64
	var userName string
	var passwordHash string
	err = tx.QueryRow(ctx,
		`SELECT id, username, password_hash
				FROM users
				WHERE username = @username`,
		pgx.NamedArgs{
			"username": request.UserName,
		},
	).Scan(&userId, &userName, &passwordHash)
	if err != nil {
		return nil, errors.New("error get user")
	}

	return &AuthResponse{
		UserId:   userId,
		UserName: userName,
		PassHash: passwordHash,
	}, tx.Commit(ctx)
}

func (r *Service) GetInventory(ctx context.Context, userId uint64) ([]*ProductStock, error) {
	conn, err := r.Pool().Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()
	rows, err := conn.Query(ctx,
		`SELECT item, quantity 
			FROM inventory
			WHERE user_id = @user_id`,
		pgx.NamedArgs{
			"user_id": userId,
		},
	)
	if err != nil {
		return nil, err
	}

	var inventory []*ProductStock
	for rows.Next() {
		var productStock ProductStock
		err := rows.Scan(&productStock.Type, &productStock.Quantity)
		if err != nil {
			return nil, err
		}
		inventory = append(inventory, &productStock)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("cannot fetch user inventory", "err", err)
	}
	return inventory, nil
}

func (r *Service) GetBalance(ctx context.Context, userId uint64) (int, error) {
	conn, err := r.Pool().Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()
	var balance int
	err = conn.QueryRow(ctx,
		`SELECT balance
					FROM users
					WHERE id = @user_id`,
		pgx.NamedArgs{
			"user_id": userId,
		},
	).Scan(&balance)
	if err != nil {
		return 0, err
	}
	return balance, err
}

func (r *Service) GetCoinHistory(ctx context.Context, userId uint64) (*CoinHistory, error) {
	conn, err := r.Pool().Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()
	rows, err := conn.Query(ctx,
		`SELECT 
    			users1.username as from_user, 
       			users2.username as to_user, 
       			transactions.amount, 
       			transactions.created_at
			FROM transactions
			LEFT JOIN users as users1 ON transactions.from_user_id = users1.id
			LEFT JOIN users as users2 ON transactions.to_user_id = users2.id
			WHERE transactions.from_user_id = @from_user_id`,
		pgx.NamedArgs{
			"from_user_id": userId,
		},
	)


	if err != nil {
		return nil, err
	}
	var sent []*Transaction
	for rows.Next() {
		var transaction Transaction
		err := rows.Scan(&transaction.FromUser, &transaction.ToUser, &transaction.Amount, &transaction.CreatedAt)
		if err != nil {
			return nil, err
		}
		sent = append(sent, &transaction)
	}
	if err := rows.Err(); err != nil {
		r.logger.Error("cannot fetch user inventory", "err", err)
	}

	rows, err = conn.Query(ctx,
		`SELECT 
    			users1.username as from_user, 
       			users2.username as to_user, 
       			transactions.amount, 
       			transactions.created_at
			FROM transactions
			LEFT JOIN users as users1 ON transactions.from_user_id = users1.id
			LEFT JOIN users as users2 ON transactions.to_user_id = users2.id
			WHERE transactions.to_user_id = @to_user_id`,
		pgx.NamedArgs{
			"to_user_id": userId,
		},
	)
	if err != nil {
		return nil, err
	}
	var received []*Transaction
	for rows.Next() {
		var transaction Transaction
		err := rows.Scan(&transaction.FromUser, &transaction.ToUser, &transaction.Amount, &transaction.CreatedAt)
		if err != nil {
			return nil, err
		}
		received = append(received, &transaction)
	}
	if err := rows.Err(); err != nil {
		r.logger.Error("cannot fetch user inventory", "err", err)
	}

	return &CoinHistory{
		Sent:     sent,
		Received: received,
	}, nil
}

func (r *Service) SendCoin(ctx context.Context, request *SendCoinRequest) (*SendCoinResponse, error) {
	conn, err := r.Pool().Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			r.logger.Error("rollback error", slog.String("err", err.Error()))
		}
	}()

	var targetUserId uint64
	err = tx.QueryRow(ctx,
		`SELECT id
					FROM users
					WHERE username = @username`,
		&pgx.NamedArgs{
			"username": request.ToUser,
		},
	).Scan(&targetUserId)
	if err != nil {
		return nil, errors.New("target user not found")
	}

	res, err := tx.Exec(ctx,
		`UPDATE users
				SET balance = balance - @amount
				WHERE id = @user_id`,
		pgx.NamedArgs{
			"amount":  request.Amount,
			"user_id": request.UserId,
		},
	)
	if err != nil {
		return nil, err
	}
	if res.RowsAffected() == 0 {
		return nil, errors.New("current user not found")
	}
	res, err = tx.Exec(ctx,
		`UPDATE users
					SET balance = balance + @amount
					WHERE id = @user_id`,
		pgx.NamedArgs{
			"amount":  request.Amount,
			"user_id": targetUserId,
		},
	)
	if err != nil {
		return nil, err
	}
	if res.RowsAffected() == 0 {
		return nil, errors.New("target user not found")
	}
	_, err = tx.Exec(ctx,
		`INSERT INTO transactions (from_user_id, to_user_id, amount)
					VALUES (@from_user_id, @to_user_id, @amount)`,
		pgx.NamedArgs{
			"from_user_id": request.UserId,
			"to_user_id":   targetUserId,
			"amount":       request.Amount,
		},
	)
	if err != nil {
		return nil, err
	}
	return nil, tx.Commit(ctx)
}

func (r *Service) BuyItem(ctx context.Context, request *BuyItemRequest) (*BuyItemResponse, error) {
	conn, err := r.Pool().Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			r.logger.Error("rollback error", slog.String("err", err.Error()))
		}
	}()
	var price int
	err = tx.QueryRow(ctx,
		`SELECT price
				FROM items
				WHERE name = @item`,
		pgx.NamedArgs{
			"item": request.Item,
		},
	).Scan(&price)
	if err != nil {
		return nil, fmt.Errorf("error get item price from db: %w", err)
	}
	_, err = tx.Exec(ctx,
		`UPDATE users SET balance = balance - @amount WHERE id = @user_id`,
		pgx.NamedArgs{
			"amount":  price,
			"user_id": request.UserId,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error update user balance: %w", err)
	}
	_, err = tx.Exec(ctx,
		`UPDATE inventory SET quantity = quantity + 1
					WHERE user_id = @user_id AND item = @item`,
		pgx.NamedArgs{
			"user_id": request.UserId,
			"item":    request.Item,
		},
	)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("error update user inventory: %w", err)
	}
	_, err = tx.Exec(ctx,
		`INSERT INTO inventory(user_id, item, quantity)
				SELECT @user_id, @item, 1
				WHERE NOT EXISTS (SELECT 1 FROM inventory WHERE user_id = @user_id AND item = @item)`,
		pgx.NamedArgs{
			"user_id": request.UserId,
			"item":    request.Item,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error create user inventory: %w", err)
	}
	return nil, tx.Commit(ctx)
}
