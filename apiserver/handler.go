package apiserver

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
)

type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ApiResponse[T any] struct {
	Data    *T     `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}

func (r SignupRequest) Validate() error {
	if r.Email == "" {
		return errors.New("Email is required")
	}
	if r.Password == "" {
		return errors.New("Password is required")
	}
	return nil
}

func (s *ApiServer) signupHandler() http.HandlerFunc {
	return handler(func(w http.ResponseWriter, r *http.Request) error {

		req, err := decode[SignupRequest](r)
		if err != nil {
			return NewErrWithStatus(http.StatusBadRequest, err)
		}
		existingUser, err := s.store.Users.GetUserByEmail(r.Context(), req.Email)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}
		if existingUser != nil {
			return NewErrWithStatus(http.StatusConflict, fmt.Errorf("user exists: %v", existingUser))
		}

		_, err = s.store.Users.CreateUser(r.Context(), req.Email, req.Password)
		if err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, fmt.Errorf("Failed to create the user: %w", err))
		}

		if err := encode[ApiResponse[struct{}]](ApiResponse[struct{}]{
			Message: "successfully signed up user",
		}, http.StatusCreated, w); err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}
		return nil
	})

}
