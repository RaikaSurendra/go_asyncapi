package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"github.com/jmoiron/sqlx"
)

type RefreshTokenStore struct {
	db *sqlx.DB
}

//define type of refresh token table structure

type RefreshToken struct {
	UserId      uuid.UUID `db:"user_id"`
	HashedToken string    `db:"hashed_token"`
	CreatedAt   time.Time `db:"created_at"`
	ExpiresAt   time.Time `db:"expires_at"`
}

// NewRefreshTokenStore initializes a new RefreshTokenStore.
func NewRefreshTokenStore(db *sql.DB) *RefreshTokenStore {
	return &RefreshTokenStore{
		db: sqlx.NewDb(db, "postgres"),
	}
}

func (s *RefreshTokenStore) getBase64HashFromToken(token *jwt.Token) (string, error) {
	h := sha256.New()
	h.Write([]byte(token.Raw))
	hashedBytes := h.Sum(nil)
	hashedTokenB64 := base64.StdEncoding.EncodeToString(hashedBytes)
	return hashedTokenB64, nil
}

func (s *RefreshTokenStore) Create(ctx context.Context, userId uuid.UUID, token *jwt.Token) (*RefreshToken, error) {
	const insert = `INSERT INTO refresh_tokens( user_id, hashed_token, expires_at) VALUES ($1, $2, $3) RETURNING *;`

	hashedTokenB64, err := s.getBase64HashFromToken(token)
	if err != nil {
		return nil, fmt.Errorf("getBase64HashFromToken: %w", err)
	}
	//fetch expriesAt
	expiresAt, err := token.Claims.GetExpirationTime()
	if err != nil {
		return nil, fmt.Errorf("failed to extract expiration time: %w", err)
	}
	var refreshToken RefreshToken
	if err = s.db.GetContext(ctx, &refreshToken, insert, userId, hashedTokenB64, expiresAt.Time); err != nil {
		return nil, fmt.Errorf("db.GetContext: %w", err)
	}

	return &refreshToken, nil
}

func (s *RefreshTokenStore) ByPrimaryKey(ctx context.Context, userId uuid.UUID, token *jwt.Token) (*RefreshToken, error) {
	const query = `SELECT * FROM refresh_tokens WHERE user_id = $1 AND hashed_token= $2;`
	hashedTokenB64, err := s.getBase64HashFromToken(token)
	if err != nil {
		return nil, fmt.Errorf("getBase64HashFromToken: %w", err)
	}

	var refreshToken RefreshToken
	if err = s.db.GetContext(ctx, &refreshToken, query, userId, hashedTokenB64); err != nil {
		return nil, fmt.Errorf("failed to fetch hashed_token %s record for user %s: %w", hashedTokenB64, userId, err)
	}
	return &refreshToken, nil
}

func (s *RefreshTokenStore) DeleteUserTokens(ctx context.Context, userId uuid.UUID) (sql.Result, error) {
	const deleteDDL = `DELETE FROM refresh_tokens WHERE user_id = $1;`
	result, err := s.db.ExecContext(ctx, deleteDDL, userId)
	if err != nil {
		return result, fmt.Errorf("failed to delete refresh_tokens record: %w", err)
	}
	return result, nil
}
