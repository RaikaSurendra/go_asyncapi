package apiserver

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

type ErrWithStatus struct {
	status int
	err    error
}

func (e *ErrWithStatus) Error() string {
	return e.err.Error()
}

func NewErrWithStatus(status int, err error) *ErrWithStatus {
	return &ErrWithStatus{status: status, err: err}
}

func handler(f func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			status := http.StatusInternalServerError
			msg := http.StatusText(status)
			if e, ok := err.(*ErrWithStatus); ok {
				status = e.status
				msg = http.StatusText(e.status)
				if status == http.StatusBadRequest || status == http.StatusNotFound || status == http.StatusConflict {
					msg = e.err.Error()
				}
			}
			//log the error slog message
			slog.Error("HTTP handler error", "status", status, "error", err, "message", msg)
			w.WriteHeader(status)
			if err := json.NewEncoder(w).Encode(ApiResponse[struct{}]{
				Message: msg,
			}); err != nil {
				slog.Error("HTTP handler encoding response error", "status", status, "error", err, "message", msg)
			}

		}
	}
}

func encode[T any](v T, status int, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json; chartset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("Error encoding reposne: %w", err)
	}
	return nil
}

type Validator interface {
	Validate() error
}

func decode[T Validator](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("error decoding request body: %w", err)
	}
	if err := v.Validate(); err != nil {
		return v, fmt.Errorf("validation error: %w", err)
	}
	return v, nil
}
