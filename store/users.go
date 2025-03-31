package store

import (
	//import sqlx
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // Postgres driver
	"golang.org/x/crypto/bcrypt"
)

type UserStore struct {
	// DB is the database connection
	db *sqlx.DB
}

type User struct {
	// ID is the unique identifier for the user
	Id                   uuid.UUID `db:"id"`
	Email                string    `db:"email"`
	HashedPasswordBase64 string    `db:"hashed_password"`
	CreatedAt            time.Time `db:"created_at"`
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{
		db: sqlx.NewDb(db, "postgres"),
	}
}

func (u *User) ComparePassword(password string) error {
	// Decode the base64 hashed password
	hashedPassword, err := base64.StdEncoding.DecodeString(u.HashedPasswordBase64)
	if err != nil {
		return fmt.Errorf("failed to decode hashed password: %w", err)
	}

	// Compare the provided password with the hashed password
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		return fmt.Errorf("invalid password: %w", err)
	}
	return nil
}

// CreateUser creates a new user in the database.
// It hashes the password and stores it in base64 format.
// It returns the created user or an error if the creation fails.
// It uses the context to support cancellation and deadlines.
func (s *UserStore) CreateUser(ctx context.Context, email, password string) (*User, error) {
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	hashedPasswordBase64 := base64.StdEncoding.EncodeToString(hashedPassword)

	// Insert a new user into the database
	const dml = `INSERT INTO users (email, hashed_password) VALUES ($1, $2) RETURNING id, email, hashed_password, created_at`
	var user User
	if err := s.db.GetContext(ctx, &user, dml, email, hashedPasswordBase64); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	// Return the created user
	return &user, nil
}

func (s *UserStore) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	// Query the user by email
	const query = `SELECT id, email, hashed_password AS hashed_password, created_at FROM users WHERE email = $1`
	var user User
	if err := s.db.GetContext(ctx, &user, query, email); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	// Return the found user
	return &user, nil
}

// GetUserByID retrieves a user by their ID from the database.
// It returns the user or an error if the retrieval fails.
// It uses the context to support cancellation and deadlines.
// The ID is a UUID that uniquely identifies the user.
// The user is returned as a pointer to a User struct.
// The User struct contains the user's ID, email, hashed password, and creation timestamp.
// The hashed password is stored in base64 format for security.
// The function uses the sqlx package to execute the SQL query and map the result to the User struct.
// The function handles errors by returning a wrapped error message if the retrieval fails.
// The function also checks if the user exists and returns a specific error message if not.
// The function uses the context to support cancellation and deadlines, allowing for better control over long-running operations.
// It uses the sql package to handle SQL errors and the fmt package for error formatting.
// It uses the time package to handle timestamps and the uuid package to generate and handle UUIDs.
func (s *UserStore) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	// Query the user by ID
	const query = `SELECT * FROM users WHERE id = $1`
	var user User
	if err := s.db.GetContext(ctx, &user, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	// Return the found user
	return &user, nil
}
