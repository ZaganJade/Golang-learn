package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend-golang/UTS/app/model"
)

var (
	ErrNotFound  = errors.New("data tidak ditemukan")
	ErrDuplicate = errors.New("data sudah ada")
)

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

// ---------- users ----------

type UserRepository interface {
	// FindForLogin mengambil user berdasarkan email sekaligus deleted_at
	// mahasiswanya (untuk memblokir login mahasiswa yang di-soft delete).
	FindForLogin(ctx context.Context, email string) (model.User, *time.Time, error)
	FindByID(ctx context.Context, id int) (model.User, error)
	CreateWithTx(ctx context.Context, tx pgx.Tx, u model.User) (int, error)
}

type userRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepository{pool: pool}
}

func (r *userRepository) FindForLogin(
	ctx context.Context, email string,
) (model.User, *time.Time, error) {
	var u model.User
	var deletedAt *time.Time

	err := r.pool.QueryRow(ctx, `
		SELECT u.id, u.email, u.password, u.role, u.created_at, s.deleted_at
		  FROM users u
		  LEFT JOIN students s ON s.user_id = u.id
		 WHERE LOWER(u.email) = LOWER($1)`, email,
	).Scan(&u.ID, &u.Email, &u.Password, &u.Role, &u.CreatedAt, &deletedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, nil, ErrNotFound
		}
		return model.User{}, nil, fmt.Errorf("mencari user untuk login: %w", err)
	}
	return u, deletedAt, nil
}

func (r *userRepository) FindByID(
	ctx context.Context, id int,
) (model.User, error) {
	var u model.User
	err := r.pool.QueryRow(ctx, `
		SELECT id, email, password, role, created_at FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Email, &u.Password, &u.Role, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("mengambil user: %w", err)
	}
	return u, nil
}

func (r *userRepository) CreateWithTx(
	ctx context.Context, tx pgx.Tx, u model.User,
) (int, error) {
	var id int
	err := tx.QueryRow(ctx, `
		INSERT INTO users (email, password, role) VALUES ($1, $2, $3)
		RETURNING id`, u.Email, u.Password, u.Role,
	).Scan(&id)
	if err != nil {
		if isUniqueViolation(err) {
			return 0, ErrDuplicate
		}
		return 0, fmt.Errorf("menyimpan user: %w", err)
	}
	return id, nil
}
