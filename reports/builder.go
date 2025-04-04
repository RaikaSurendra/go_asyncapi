package reports

import (
	config "asyncapi/config"
	"asyncapi/store"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/csv"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type ReportBuilder struct {
	reportStore *store.ReportStore
	lozClient   *LozClient
	s3Client    *s3.Client
	config      *config.Config
	logger      *slog.Logger
}

// NewReportBuilder initializes a new ReportBuilder instance.
//
// Parameters:
// - reportStore: The store for interacting with reports in the database.
// - lozClient: The client for fetching data from the Loz API.
// - s3Client: The AWS S3 client for uploading files.
//
// Returns:
// - A pointer to a new ReportBuilder instance.
func NewReportBuilder(config *config.Config, reportStore *store.ReportStore, lozClient *LozClient, s3Client *s3.Client, logger *slog.Logger) *ReportBuilder {
	return &ReportBuilder{
		reportStore: reportStore,
		lozClient:   lozClient,
		s3Client:    s3Client,
		config:      config,
		logger:      logger,
	}
}

// Build generates a report for the given user and report ID.
//
// Parameters:
// - ctx: The context for managing request lifetimes and cancellations.
// - userId: The ID of the user who owns the report.
// - reportId: The unique ID of the report.
//
// Returns:
// - A pointer to the updated Report instance.
// - An error if the operation fails.
func (b *ReportBuilder) Build(ctx context.Context, userId uuid.UUID, reportId uuid.UUID) (report *store.Report, err error) {
	// Fetch the report from the database
	report, err = b.reportStore.GetByPrimaryKey(ctx, userId, reportId)
	if err != nil {
		return nil, fmt.Errorf("failed to get the report %s for user %s: %w", reportId, userId, err)
	}

	// Ensure the report is not already being processed
	if report.StartedAt != nil {
		return report, nil
	}

	//defer funtion to catch all errors and updat ethe report
	defer func() {
		if err != nil {
			if report != nil {
				// update report status
			} else {
				b.logger.Error("report is nil, cannot update status", "error", err.Error())
			}

			now := aws.Time(time.Now())
			errMsg := err.Error()
			report.FailedAt = now
			report.ErrorMessage = &errMsg
			if _, updateErr := b.reportStore.Update(ctx, report); updateErr != nil {
				b.logger.Error("failed to update the report", "error", err.Error())
			}

		}
	}()

	// Update the report timestamps
	now := time.Now()
	report.StartedAt = &now
	report.CompletedAt = nil
	report.FailedAt = nil
	report.ErrorMessage = nil
	report.DownloadUrl = nil
	report.DownloadUrlExpiresAt = nil
	report.OutputFilePath = nil

	report, err = b.reportStore.Update(ctx, report)
	if err != nil {
		return nil, fmt.Errorf("failed to update the report %s for user %s: %w", reportId, userId, err)
	}

	// Fetch monsters data from LozClient
	resp, err := b.lozClient.GetMonsters()
	if err != nil {
		return nil, fmt.Errorf("failed to get monsters data from LozClient: %w", err)
	}
	if len(resp) == 0 {
		return nil, fmt.Errorf("no monsters data found")
	}

	// Create a buffer for the CSV and gzip writers
	var buffer bytes.Buffer

	// Initialize gzip writer
	gzipWriter := gzip.NewWriter(&buffer)
	//defer gzipWriter.Close()

	// Initialize CSV writer
	csvWriter := csv.NewWriter(gzipWriter)
	//defer csvWriter.Flush()

	// Define CSV headers
	headers := []string{"id", "name", "description", "common_locations", "drops", "category", "image", "dlc"}

	// Write headers to the CSV
	if err := csvWriter.Write(headers); err != nil {
		return nil, fmt.Errorf("failed to write CSV headers: %w", err)
	}

	// Write monster data to the CSV
	for _, monster := range resp {
		row := []string{
			strconv.Itoa(monster.Id),
			monster.Name,
			monster.Description,
			fmt.Sprintf("%v", monster.Location),
			fmt.Sprintf("%v", monster.Drops),
			monster.Category,
			monster.Image,
			fmt.Sprintf("%t", monster.Dlc),
		}
		if err := csvWriter.Write(row); err != nil {
			return nil, fmt.Errorf("failed to write monster data to CSV: %w", err)
		}
	}

	// Flush the CSV writer
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return nil, fmt.Errorf("failed to flush CSV writer: %w", err)
	}

	// Close the gzip writer to flush data to the buffer
	if err := gzipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	// Prepare the S3 path
	key := fmt.Sprintf("/users/%s/%s.csv.gz", userId.String(), reportId.String())

	// Upload the file to S3
	_, err = b.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Key:    aws.String(key),
		Bucket: aws.String(b.config.S3Bucket),
		Body:   bytes.NewReader(buffer.Bytes()),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload report to S3: %w", err)
	}

	// Update the report with the S3 path and completion timestamp
	report.OutputFilePath = &key
	report.CompletedAt = aws.Time(time.Now())

	report, err = b.reportStore.Update(ctx, report)
	if err != nil {
		return nil, fmt.Errorf("failed to update report with S3 path: %w", err)
	}
	b.logger.Info("successfully generated report", "report id", report.Id, "for user id", userId.String(), "path", key)
	return report, nil
}
