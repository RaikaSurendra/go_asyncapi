package apiserver

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"asyncapi/reports"

	"github.com/google/uuid"
)

type SigninRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SigninResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ApiResponse[T any] struct {
	Data    *T     `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}

// Validate checks the SignupRequest fields for required values.
// It returns an error if the Email or Password fields are empty,
// indicating that both fields are mandatory for a valid signup request.

func (r SignupRequest) Validate() error {
	if r.Email == "" {
		return errors.New("email is required")
	}
	if r.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

// signupHandler handles user signup requests. It decodes the incoming request
// into a SignupRequest, checks for existing users with the same email, and

// creates a new user if none exists. If the email is already taken, it returns
// a conflict error. Upon successful creation, it sends a success response with
// HTTP status 201. In case of any errors during decoding, user lookup, or
// creation, appropriate error responses with corresponding HTTP status codes
// are returned.
func (s *ApiServer) signupHandler() http.HandlerFunc {

	return handler(func(w http.ResponseWriter, r *http.Request) error {

		req, err := decode[SignupRequest](r)
		if err != nil {
			return NewErrWithStatus(http.StatusBadRequest, fmt.Errorf("failed to decode signup request: %w", err))
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
			return NewErrWithStatus(http.StatusInternalServerError, fmt.Errorf("failed to create the user: %w", err))
		}

		if err := encode[ApiResponse[struct{}]](ApiResponse[struct{}]{
			Message: "successfully signed up user",
		}, http.StatusCreated, w); err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}
		return nil
	})

}
func (r SigninRequest) Validate() error {
	if r.Email == "" {
		return errors.New("email is required")
	}
	if r.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

// function signin handler
func (s *ApiServer) signinHandler() http.HandlerFunc {
	return handler(func(w http.ResponseWriter, r *http.Request) error {
		req, err := decode[SigninRequest](r)
		if err != nil {
			return NewErrWithStatus(http.StatusBadRequest, err)
		}

		user, err := s.store.Users.GetUserByEmail(r.Context(), req.Email)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}
		//copare password
		if err := user.ComparePassword(req.Password); err != nil {
			return NewErrWithStatus(http.StatusUnauthorized, err)
		}

		//issue a token
		tokenPair, err := s.jwtManager.GenerateTokenPair(user.Id)
		if err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}
		//delete before creation of user token
		_, err = s.store.RefreshTokenStore.DeleteUserTokens(r.Context(), user.Id)
		if err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}

		//create user tokens
		_, err = s.store.RefreshTokenStore.Create(r.Context(), user.Id, tokenPair.RefreshToken)
		if err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}

		if err := encode(ApiResponse[SigninResponse]{
			Data: &SigninResponse{
				AccessToken:  tokenPair.AccessToken.Raw,
				RefreshToken: tokenPair.RefreshToken.Raw,
			},
		}, http.StatusOK, w); err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}

		//persists token

		return nil

	})
}

type TokenRefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type TokenRefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (r TokenRefreshRequest) Validate() error {
	if r.RefreshToken == "" {
		return errors.New("refresh token is required")
	}
	return nil
}

// Validate checks the CreateReportRequest fields for required values.
// It returns an error if the ReportType field is empty, indicating
// that this field is mandatory for a valid report creation request.

func (r CreateReportRequest) Validate() error {
	if r.ReportType == "" {
		return errors.New("report_type is required")
	}
	return nil
}

func (s *ApiServer) tokenRefreshHandler() http.HandlerFunc {
	return handler(func(w http.ResponseWriter, r *http.Request) error {

		req, err := decode[TokenRefreshRequest](r)
		if err != nil {
			return NewErrWithStatus(http.StatusBadRequest, fmt.Errorf("error decoding request body: %w", err))
		}

		// Parse the token
		currentRefreshToken, err := s.jwtManager.Parse(req.RefreshToken)
		if err != nil {
			return NewErrWithStatus(http.StatusUnauthorized, fmt.Errorf("invalid refresh token: %w", err))
		}

		userIdStr, err := currentRefreshToken.Claims.GetSubject()
		if err != nil {
			return NewErrWithStatus(http.StatusUnauthorized, fmt.Errorf("failed to get subject from token: %w", err))
		}

		userId, err := uuid.Parse(userIdStr)
		if err != nil {
			return NewErrWithStatus(http.StatusUnauthorized, fmt.Errorf("invalid user ID in token: %w", err))
		}

		currentRefreshTokenRecord, err := s.store.RefreshTokenStore.ByPrimaryKey(r.Context(), userId, currentRefreshToken)
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, sql.ErrNoRows) {
				status = http.StatusUnauthorized
			}
			return NewErrWithStatus(status, fmt.Errorf("failed to fetch refresh token record: %w", err))
		}

		// Verify the current refresh token is not expired
		if currentRefreshTokenRecord.ExpiresAt.Before(time.Now()) {
			return NewErrWithStatus(http.StatusUnauthorized, errors.New("refresh token expired"))
		}

		// Generate a new token pair, delete old tokens, and persist the new ones
		tokenPair, err := s.jwtManager.GenerateTokenPair(userId)
		if err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, fmt.Errorf("failed to generate token pair: %w", err))
		}

		if _, err := s.store.RefreshTokenStore.DeleteUserTokens(r.Context(), userId); err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, fmt.Errorf("failed to delete old tokens: %w", err))
		}

		if _, err := s.store.RefreshTokenStore.Create(r.Context(), userId, tokenPair.RefreshToken); err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, fmt.Errorf("failed to persist new refresh token: %w", err))
		}

		if err := encode(ApiResponse[TokenRefreshResponse]{
			Data: &TokenRefreshResponse{
				AccessToken:  tokenPair.AccessToken.Raw,
				RefreshToken: tokenPair.RefreshToken.Raw,
			},
		}, http.StatusOK, w); err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, fmt.Errorf("failed to encode response: %w", err))
		}

		return nil
	})
}

type CreateReportRequest struct {
	ReportType string `json:"report_type"`
}
type ApiReport struct {
	// The ID of the user who owns the report.
	Id                   uuid.UUID  `json:"id"`                                // The unique ID of the report.
	ReportType           string     `json:"report_type,omitempty"`             // The type of the report (e.g., "summary", "detailed").
	OutputFilePath       *string    `json:"output_file_path,omitempty"`        // The file path where the report is stored.
	DownloadUrl          *string    `json:"download_url,omitempty"`            // The URL to download the report.
	DownloadUrlExpiresAt *time.Time `json:"download_url_expires_at,omitempty"` // The expiration time of the download URL.
	ErrorMessage         *string    `json:"error_message,omitempty"`           // Any error message associated with the report generation.
	CreatedAt            time.Time  `json:"created_at,omitempty"`              // The timestamp when the report was created.
	StartedAt            *time.Time `json:"started_at,omitempty"`              // The timestamp when the report generation started.
	CompletedAt          *time.Time `json:"completed_at,omitempty"`            // The timestamp when the report generation completed.
	FailedAt             *time.Time `json:"failed_at,omitempty"`
	Status               string     `json:"status,omitempty"`
}

// createReportHandler is the HTTP handler to create a new report.
//
// Parameters:
// - req: The request containing the report type.
// - w: The response writer to write the response to.
// - r: The request object.
//
// Returns:
// - An error if the operation fails.
//
// This function will first decode the request to a CreateReportRequest struct.
// It will then create a new report with the given report type in the database.
// After that, it will send an SQS message to the report generation queue.
// Finally, it will return the created report as a JSON response with a 201 status code.
func (s *ApiServer) createReportHandler() http.HandlerFunc {
	return handler(func(w http.ResponseWriter, r *http.Request) error {
		req, err := decode[CreateReportRequest](r)
		if err != nil {
			return NewErrWithStatus(http.StatusBadRequest, err)
		}
		user, ok := UserFromContext(r.Context())
		if !ok {
			return NewErrWithStatus(http.StatusUnauthorized, fmt.Errorf("user not found in context"))
		}
		report, err := s.store.ReportStore.Create(r.Context(), user.Id, req.ReportType)
		if err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}
		//send sqs message for report generation
		//send as json
		sqsMessage := reports.SqsMessage{
			UserId:   report.UserId,
			ReportId: report.Id,
		}
		//get sqs queue url
		queueUrlOutput, err := s.sqsClient.GetQueueUrl(r.Context(), &sqs.GetQueueUrlInput{
			QueueName: aws.String(s.config.SqsQueue),
		})
		if err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}
		_, err = s.sqsClient.SendMessage(r.Context(), &sqs.SendMessageInput{
			QueueUrl:    queueUrlOutput.QueueUrl, // Replace with your SQS queue URL
			MessageBody: aws.String(fmt.Sprintf(`{"user_id":"%s","report_id":"%s"}`, sqsMessage.UserId, sqsMessage.ReportId)),
		})
		if err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, fmt.Errorf("failed to send SQS message: %w", err))
		}
		if err := encode(ApiResponse[ApiReport]{
			Data: &ApiReport{
				Id:                   report.Id,
				ReportType:           report.ReportType,
				OutputFilePath:       report.OutputFilePath,
				DownloadUrl:          report.DownloadUrl,
				DownloadUrlExpiresAt: report.DownloadUrlExpiresAt,
				ErrorMessage:         report.ErrorMessage,
				CreatedAt:            report.CreatedAt,
				StartedAt:            report.StartedAt,
				CompletedAt:          report.CompletedAt,
				FailedAt:             report.FailedAt,
				Status:               report.Status(),
			},
		}, int(http.StatusCreated), w); err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}

		return nil
	})
}

// getReportHandler is the HTTP handler to retrieve a report.
//
// Parameters:
// - w: The response writer to write the response to.
// - r: The request object.
//
// Returns:
// - An error if the operation fails.
//
// This function will first decode the request to a report ID.
// It will then retrieve the report with the given ID in the database.
// After that, it will check if the report generation is completed and if the download URL is expired.
// If the report is completed and the download URL is expired, it will generate a new signed URL and update the report in the database.
// Finally, it will return the report as a JSON response with a 200 status code.
func (s *ApiServer) getReportHandler() http.HandlerFunc {
	return handler(func(w http.ResponseWriter, r *http.Request) error {
		reportIdStr := r.PathValue("id")
		reportId, err := uuid.Parse(reportIdStr)
		if err != nil {
			return NewErrWithStatus(http.StatusBadRequest, err)
		}

		user, ok := UserFromContext(r.Context())
		if !ok {
			return NewErrWithStatus(http.StatusUnauthorized, fmt.Errorf("user not found in context"))
		}

		report, err := s.store.ReportStore.GetByPrimaryKey(r.Context(), user.Id, reportId)
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, sql.ErrNoRows) {
				status = http.StatusNotFound
			}
			return NewErrWithStatus(status, err)
		}
		//hasExpiration := report.DownloadUrlExpiresAt != nil && report.DownloadUrlExpiresAt.Before(time.Now())
		if report.CompletedAt != nil {
			needsRefesh := report.DownloadUrlExpiresAt != nil && report.DownloadUrlExpiresAt.Before(time.Now())
			if report.DownloadUrl == nil || needsRefesh {
				expiresAt := time.Now().Add(time.Second * 40)
				signedUrl, err := s.presignClient.PresignGetObject(r.Context(), &s3.GetObjectInput{
					Bucket: aws.String(s.config.S3Bucket),
					Key:    report.OutputFilePath,
				}, func(options *s3.PresignOptions) {
					options.Expires = time.Second * 40
				})
				if err != nil {
					return NewErrWithStatus(http.StatusInternalServerError, err)
				}
				//update the report
				report.DownloadUrl = aws.String(signedUrl.URL)
				report.DownloadUrlExpiresAt = &expiresAt
				//update the report in db

				report, err = s.store.ReportStore.Update(r.Context(), report)
				if err != nil {
					return NewErrWithStatus(http.StatusInternalServerError, err)
				}
			}
		}
		if err := encode(ApiResponse[ApiReport]{
			Data: &ApiReport{
				Id:                   report.Id,
				ReportType:           report.ReportType,
				OutputFilePath:       report.OutputFilePath,
				DownloadUrl:          report.DownloadUrl,
				DownloadUrlExpiresAt: report.DownloadUrlExpiresAt,
				ErrorMessage:         report.ErrorMessage,
				CreatedAt:            report.CreatedAt,
				StartedAt:            report.StartedAt,
				CompletedAt:          report.CompletedAt,
				FailedAt:             report.FailedAt,
				Status:               report.Status(),
			},
		}, http.StatusOK, w); err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}

		return nil
	})
}
