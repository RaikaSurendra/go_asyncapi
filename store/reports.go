package store

// Package store provides functionality to interact with the database for storing and retrieving data.
// This file, reports.go, contains functions related to generating and managing reports in the system.
//
// Functions in this file:
// - NewReportStore: Initializes a new ReportStore with a database connection.
// - Create: Creates a new report in the database.
// - Update: Updates an existing report in the database.
// - GetByPrimaryKey: Retrieves a report by its unique primary key (userId and id).
//
// Dependencies:
// - github.com/google/uuid: Used for generating unique identifiers.
// - github.com/jmoiron/sqlx: Provides extensions to the standard database/sql library for easier database interactions.
// - github.com/lib/pq: PostgreSQL driver for Go.
//
// Note:
// - Ensure the database connection is properly configured before using the functions in this file.
// - Context is used for managing request lifetimes and cancellations.

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"github.com/jmoiron/sqlx"
)

// ReportStore provides methods to interact with the reports table in the database.
type ReportStore struct {
	db *sqlx.DB
}

// Report represents a report entity in the database.
type Report struct {
	UserId               uuid.UUID  `db:"user_id"`                 // The ID of the user who owns the report.
	Id                   uuid.UUID  `db:"id"`                      // The unique ID of the report.
	ReportType           string     `db:"report_type"`             // The type of the report (e.g., "summary", "detailed").
	OutputFilePath       *string    `db:"output_file_path"`        // The file path where the report is stored.
	DownloadUrl          *string    `db:"download_url"`            // The URL to download the report.
	DownloadUrlExpiresAt *time.Time `db:"download_url_expires_at"` // The expiration time of the download URL.
	ErrorMessage         *string    `db:"error_message"`           // Any error message associated with the report generation.
	CreatedAt            time.Time  `db:"created_at"`              // The timestamp when the report was created.
	StartedAt            *time.Time `db:"started_at"`              // The timestamp when the report generation started.
	CompletedAt          *time.Time `db:"completed_at"`            // The timestamp when the report generation completed.
	FailedAt             *time.Time `db:"failed_at"`               // The timestamp when the report generation failed.
	UpdatedAt            *time.Time `db:"updated_at"`              // The timestamp when the report was last updated.
}

// NewReportStore initializes a new ReportStore with the given database connection.
//
// Parameters:
// - db: A pointer to an sql.DB instance representing the database connection.
//
// Returns:
// - A pointer to a new ReportStore instance.
func NewReportStore(db *sql.DB) *ReportStore {
	return &ReportStore{
		db: sqlx.NewDb(db, "postgres"),
	}
}

// Create inserts a new report into the database.
//
// Parameters:
// - ctx: The context for managing request lifetimes and cancellations.
// - userId: The ID of the user who owns the report.
// - reportType: The type of the report (e.g., "summary", "detailed").
//
// Returns:
// - A pointer to the created Report instance.
// - An error if the operation fails.
func (s *ReportStore) Create(ctx context.Context, userId uuid.UUID, reportType string) (*Report, error) {
	const insert = `INSERT INTO reports(user_id, report_type) VALUES ($1, $2) RETURNING *;`
	var report Report
	if err := s.db.GetContext(ctx, &report, insert, userId, reportType); err != nil {
		return nil, fmt.Errorf("failed to insert report for user %s: %w", userId, err)
	}
	return &report, nil
}

// Update modifies an existing report in the database.
//
// Parameters:
// - ctx: The context for managing request lifetimes and cancellations.
// - report: A pointer to the Report instance containing updated values.
//
// Returns:
// - A pointer to the updated Report instance.
// - An error if the operation fails.
func (s *ReportStore) Update(ctx context.Context, report *Report) (*Report, error) {
	query := `
        UPDATE reports
        SET output_file_path = $1, download_url = $2, download_url_expires_at = $3,
            error_message = $4, started_at = $5, completed_at = $6
        WHERE id = $7 AND user_id = $8
        RETURNING id, user_id, report_type, output_file_path, download_url, 
                  download_url_expires_at, error_message, started_at, completed_at, 
                  created_at
    `
	row := s.db.QueryRowContext(ctx, query,
		report.OutputFilePath, report.DownloadUrl, report.DownloadUrlExpiresAt,
		report.ErrorMessage, report.StartedAt, report.CompletedAt,
		report.Id, report.UserId,
	)
	var updatedReport Report
	err := row.Scan(&updatedReport.Id, &updatedReport.UserId, &updatedReport.ReportType,
		&updatedReport.OutputFilePath, &updatedReport.DownloadUrl, &updatedReport.DownloadUrlExpiresAt,
		&updatedReport.ErrorMessage, &updatedReport.StartedAt, &updatedReport.CompletedAt,
		&updatedReport.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update report %s for user %s: %w", report.Id, report.UserId, err)
	}
	return &updatedReport, nil
}

// GetByPrimaryKey retrieves a report from the database using its unique primary key.
//
// Parameters:
// - ctx: The context for managing request lifetimes and cancellations.
// - userId: The ID of the user who owns the report.
// - id: The unique ID of the report.
//
// Returns:
// - A pointer to the retrieved Report instance.
// - An error if the operation fails or the report is not found.
func (s *ReportStore) GetByPrimaryKey(ctx context.Context, userId uuid.UUID, id uuid.UUID) (*Report, error) {
	const query = `SELECT * FROM reports WHERE user_id = $1 AND id = $2;`
	var report Report
	if err := s.db.GetContext(ctx, &report, query, userId, id); err != nil {
		return nil, fmt.Errorf("failed to retrieve report with id %s for user %s: %w", id, userId, err)
	}
	return &report, nil
}
