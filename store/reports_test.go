package store_test

import (
	"context"
	"testing"
	"time"

	"asyncapi/fixtures"
	"asyncapi/store"

	"github.com/stretchr/testify/require"
)

// TestReportStore tests the functionality of the ReportStore, including
// creating, updating, and retrieving reports. It ensures that the methods
// behave as expected and validates the integrity of the data.
//
// The test performs the following steps:
//  1. Initializes the test environment and sets up the database.
//  2. Creates a new ReportStore instance.
//  3. Tests the Create method by creating a report for a user and verifying
//     the report's properties, such as UserId, ReportType, and CreatedAt.
//  4. Tests the Update method by updating various fields of the report,
//     including OutputFilePath, DownloadUrl, DownloadUrlExpiresAt, ErrorMessage,
//     StartedAt, and CompletedAt. It ensures that the updated fields are
//     correctly persisted.
//  5. Tests the GetByPrimaryKey method by retrieving the report using its
//     primary key (UserId and ReportId) and verifying that the retrieved
//     report matches the expected values.
//
// This test uses the require package to assert conditions and fail the test
// immediately if any assertion fails.
func TestReportStore(t *testing.T) {
	// Initialize the test environment
	env := fixtures.NewTestEnv(t)
	cleanup := env.SetupDb(t)
	t.Cleanup(func() {
		cleanup(t)
	})

	// Create a new ReportStore instance
	reportStore := store.NewReportStore(env.Db)

	// Test the Create method
	ctx := context.Background()

	//time
	now := time.Now()

	// Create a user
	userStore := store.NewUserStore(env.Db)
	user, err := userStore.CreateUser(ctx, "sample@test.com", "samplepassword")
	require.NoError(t, err)
	//fetch userId for furtehr testing
	userId := user.Id
	reportType := "test_report"
	report, err := reportStore.Create(ctx, userId, reportType)
	require.NoError(t, err)
	require.NotNil(t, report)
	require.Equal(t, userId, report.UserId)
	require.Equal(t, reportType, report.ReportType)
	require.Less(t, now.UnixNano(), report.CreatedAt.UnixNano())

	// Test the Update method
	updatedFilePath := "/path/to/report.pdf"
	downloadUrl := "http://example.com/report.pdf"
	downloadUrlExpiresAt := time.Now().Add(24 * time.Hour)
	errorMessage := "No errors"
	startedAt := time.Now()
	completedAt := time.Now()
	report.OutputFilePath = &updatedFilePath
	report.DownloadUrl = &downloadUrl
	report.DownloadUrlExpiresAt = &downloadUrlExpiresAt
	report.ErrorMessage = &errorMessage
	report.StartedAt = &startedAt
	report.CompletedAt = &completedAt
	reportType = "test_report1"

	updatedReport, err := reportStore.Update(ctx, report)
	require.NoError(t, err)
	require.NotNil(t, updatedReport)
	require.Equal(t, updatedFilePath, *updatedReport.OutputFilePath)
	require.Equal(t, downloadUrl, *updatedReport.DownloadUrl)
	require.Equal(t, errorMessage, *updatedReport.ErrorMessage)
	require.Equal(t, "test_report", updatedReport.ReportType)
	require.NotEqual(t, reportType, updatedReport.ReportType)

	// Test the GetByPrimaryKey method
	retrievedReport, err := reportStore.GetByPrimaryKey(ctx, userId, report.Id)
	require.NoError(t, err)
	require.NotNil(t, retrievedReport)
	require.Equal(t, report.Id, retrievedReport.Id)
	require.Equal(t, userId, retrievedReport.UserId)
	require.Equal(t, updatedFilePath, *retrievedReport.OutputFilePath)
	require.Equal(t, downloadUrl, *retrievedReport.DownloadUrl)
	require.WithinDuration(t, downloadUrlExpiresAt, *retrievedReport.DownloadUrlExpiresAt, time.Second)
	require.Equal(t, errorMessage, *retrievedReport.ErrorMessage)
	require.WithinDuration(t, startedAt, *retrievedReport.StartedAt, time.Second)
	require.WithinDuration(t, completedAt, *retrievedReport.CompletedAt, time.Second)
	require.Equal(t, "test_report", retrievedReport.ReportType)
}
