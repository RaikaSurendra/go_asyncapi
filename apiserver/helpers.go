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

// Error implements the error interface for *ErrWithStatus. It returns the
// error message of the embedded error.
func (e *ErrWithStatus) Error() string {
	return e.err.Error()
}

// NewErrWithStatus creates a new *ErrWithStatus. The status code and error
// message are used to construct a new *ErrWithStatus.
//
// status is the HTTP status code that should be used to respond to the
// request, and err is the error that caused the request to fail.
func NewErrWithStatus(status int, err error) *ErrWithStatus {
	return &ErrWithStatus{status: status, err: err}
}

// handler takes a function that returns an error and returns an http.HandlerFunc.
// If the function returns an error, it is logged and the http.ResponseWriter is
// written with the status code and an error message. If the error is an
// *ErrWithStatus the status code and error message are taken from it. If the
// error is not an *ErrWithStatus, the status code is set to
// http.StatusInternalServerError.
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

// encode writes the given value to the http.ResponseWriter as JSON with the given status code.
// If writing the value fails, an error is returned.
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

// decode decodes the given http.Request.Body into the given value of type T
// and validates it using the Validate method on T. If decoding or validation
// fail, an error is returned.
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
