Here is a breakdown of the engineering details of the asynchronous API project built in Go, based on the video series by the YouTube channel "devtool":

**Video 00 - Intro**

*   The project aims to build an **asynchronous API** in Golang, contrasting it with typical synchronous RESTful APIs.
*   The use case is for **long-running background processes**, where clients need to periodically check the status.
*   This asynchronous pattern is presented as an alternative to using **webhooks**.
*   The high-level architecture includes a **RESTful server in Go** with endpoints for authorization (sign-in, sign-up, token refresh using JWT), creating a report, and checking report status.
*   An **SQS queue (reports-sqs)** will be used to handle report creation requests.
*   A separate **worker** application will consume messages from the SQS queue and perform the background processing: fetching data from an external API, generating a CSV file, and uploading it to an **S3 Bucket**.
*   The worker will also **update the status of the report in a database**.
*   The project intends to cover various technologies and concepts, including testing, **JSON Web Tokens (JWT)**, **Amazon Simple Queue Service (SQS)**, **Terraform**, and **LocalStack**.

**Video 01 - Project Scaffolding & Postgres Setup with Docker**

*   The initial steps involve **project scaffolding** and setting up the development environment.
*   An **`.envrc` file** is used to manage project-specific environment variables within a shell session.
*   The `github.com/joho/godotenv` package (referred to as a package for loading environment variables into a struct) is used to load configuration from environment variables into a Go struct within a `config` package.
*   **PostgreSQL** is chosen as the database and is set up using **Docker**.
*   The standard Go **`database/sql` package** along with the **`lib/pq` PostgreSQL driver** are used for database interaction.
*   A function is implemented to establish a database connection and verify it by sending a **ping request with a timeout**.
*   The video demonstrates connecting to the PostgreSQL database from the command line using the `psql` client.
*   **`go-migrate`** is selected as the database migration tool.
*   A **`Makefile`** is created to define commands for common database tasks such as logging in, creating new migrations, and running migrations.
*   The database schema is defined in SQL migration files located in a `migrations` directory. The schema includes three main tables:
    *   `users`: Stores user information, including a unique `email`, a `hashed_password`, and `created_at` timestamp.
    *   `refresh_tokens`: Stores refresh tokens associated with users, including a foreign key `user_id` referencing the `users` table, a unique `hashed_token`, and has **`ON DELETE CASCADE`** configured.
    *   `reports`: Stores information about generated reports, including a foreign key `user_id`, a unique `id` (UUID), `report_type`, `output_s3_path`, `download_url`, `download_url_expires_at`, and various status-related fields (`created_at`, `started_at`, `completed_at`, `failed_at`, `error_message`).
*   The `go-migrate` tool also creates a `schema_migrations` table to track applied migrations.

**Video 02 - User Data Access Layer w/ SQLx + Testing**

*   The implementation of the **data access layer (DAL) for the `users` table** begins.
*   The **`github.com/jmoiron/sqlx`** library is introduced to reduce boilerplate in mapping database rows to Go structs while still allowing raw SQL queries.
*   A `store` package is created, containing a `users.go` file with a `UserStore` struct that holds a reference to `sqlx.DB`.
*   A `NewUserStore` constructor is provided to initialize the `UserStore` by taking a standard `sql.DB` and converting it to a `sqlx.DB`.
*   The video focuses on implementing the `CreateUser` operation.
*   The `github.com/google/uuid` library is used to generate unique IDs for users.
*   Struct tags (`db:`) are used to map database column names to the fields of the `User` struct.
*   The **`golang.org/x/crypto/bcrypt`** package is used to securely hash user passwords before storing them in the database. A default cost for hashing is used.
*   The hashed password is then encoded using **Base64** for storage as a string in the database.
*   A `ComparePassword` method is added to the `User` struct, which decodes the Base64 hashed password from the database and uses `bcrypt.CompareHashAndPassword` to verify a plain text password against it.
*   **Unit and integration tests** for the `UserStore` are implemented using the **`github.com/stretchr/testify/require`** assertion library.
*   A testing environment is set up using **Docker Compose** to run a PostgreSQL instance specifically for tests. Test fixtures are used to ensure a clean database state before and after each test.
*   Tests cover the `CreateUser` functionality and retrieving a user by their email address.
*   The importance of using `_test` packages in Go to avoid circular import dependencies is highlighted.

**Video 03 - API Server Setup**

*   The basic structure of the **API server in Go** is established. The approach is influenced by an article by Matt Ryer.
*   An `api_server` package with a `server.go` file is created.
*   A `Server` struct is defined to hold dependencies such as the `config`.
*   A `New` constructor and a `Start` method are implemented for the `Server` struct. The `Start` method takes a `context.Context` to handle shutdown signals.
*   The standard `net/http` package is used for building the HTTP server. A basic **/ping** endpoint is set up for health checks.
*   In `cmd/main.go`, a `run` function is implemented following a pattern by Matt Ryer to encapsulate the server's initialization and startup logic, returning an error that is then handled in the `main` function.
*   **Graceful shutdown** of the server is implemented using `os/signal` to listen for interrupt signals (like Ctrl+C) and a `context.Context` to manage the shutdown process.
*   A **structured logger** using the **`golang.org/x/exp/slog`** package is configured to log messages in JSON format to `os.Stdout`. The logger is added as a dependency to the `Server`.
*   Environment variables for the API server's host and port are added to the `.envrc` file and loaded via the `config` package.
*   The concept of **middleware** is introduced. A `middleware.go` file is created with a `Logger` middleware that logs basic information about each incoming HTTP request (method and path). This middleware wraps the main `http.ServeMux`.
*   An authentication middleware is planned but will be implemented later.
*   The implementation of a **signup handler** begins in a `handler.go` file within the `api_server` package.
*   A `SignUpRequest` struct is defined to hold the expected JSON payload for the signup endpoint (email and password), along with a `Validate` method for basic input validation.
*   A generic `Store` struct is introduced to hold instances of different data access layer stores (currently just `UserStore`) as dependencies for the API server.
*   The signup handler retrieves the `UserStore` from the `Store` dependency and checks if a user with the provided email already exists in the database, returning a 409 Conflict status if so.
*   A generic `APIResponse` struct is defined to provide a consistent structure for API responses, allowing for optional data (of a generic type `T`) and a message.
*   Helper functions `encode` and `decode` are created to streamline the process of JSON encoding responses and decoding/validating request bodies in the handlers. The `decode` function also leverages a `Validator` interface with a `Validate` method to ensure request payloads can be validated.
*   A custom `ErrorWithStatus` type is implemented to allow handlers to return errors that also specify the desired HTTP status code for the response.
*   The signup handler is refactored to use the `Handler` wrapper (which takes a function returning an error and converts it to an `http.HandlerFunc`), as well as the `decode` and `encode` helper functions for cleaner error handling and request processing.

**Video 04 - JWT Manager**

*   The implementation of a **JWT (JSON Web Token) manager** for handling authentication is the focus.
*   The project adopts a JWT authentication strategy using both **short-lived access tokens and long-lived refresh tokens**, as described in a video by "Bite By Go".
*   The **`github.com/golang-jwt/jwt/v5`** library is installed and used for JWT operations.
*   A `JWTManager` struct is created, holding a reference to the `config` to access the JWT signing secret.
*   A `JWT_SECRET` environment variable is added to the `.envrc` file and loaded into the `config`.
*   The `JWTManager` is designed to have methods for `GenerateTokenPair` (creating both access and refresh tokens) and `Parse` (parsing and validating tokens).
*   A `TokenPair` struct is defined to hold the generated access token and refresh token strings.
*   Custom JWT claims are defined, including a `TokenType` field (to distinguish between access and refresh tokens) and embedding the `jwt.RegisteredClaims` struct for standard claims like Subject, Issuer, Expiration Time, and Issued At Time.
*   The **issuer** claim is set to the API server's host and port from the configuration.
*   Default expiration times are set: **access tokens to expire in 15 minutes**, and **refresh tokens to expire in one month**.
*   Tokens are signed using the **HMAC-SHA256 (HS256)** signing algorithm.
*   The `GenerateTokenPair` method creates both the access and refresh tokens with their respective claims and expiration times, signs them using the JWT secret, and returns them as a `TokenPair`.
*   The `Parse` method takes a raw JWT string and a set of claims, parses the token, and validates its signature using the secret key. It also verifies that the signing method matches the expected algorithm (HS256).
*   A helper method `IsAccessToken` is added to the `JWTManager` to easily check if a given parsed token is an access token by inspecting its `TokenType` claim.
*   Unit tests for the `JWTManager` are implemented to ensure that token pairs are generated correctly, tokens can be parsed and validated without errors, the subject claim matches the user ID, and the `IsAccessToken` method functions as expected.

**Video 05 - User Signin**

*   The **sign-in handler** implementation is completed, focusing on user authentication and issuing JWT token pairs upon successful login.
*   A `SignInRequest` struct is defined (if not fully done previously) to hold the user's email and password from the sign-in request, along with a `Validate` method.
*   The sign-in handler decodes the `SignInRequest` from the incoming HTTP request body and performs validation.
*   It retrieves the user from the database based on the provided email using the `UserStore`'s `ByEmail` method.
*   The provided plain text password is then compared against the hashed password retrieved from the database using the `user.ComparePassword` method.
*   If the password comparison is successful (authentication successful), a **new JWT token pair (access and refresh tokens)** is generated for the authenticated user using the `JWTManager`'s `GenerateTokenPair` method, passing the user's ID as the subject.
*   A `SignInResponse` struct is defined to hold the raw string values of the generated access token and refresh token.
*   The `encode` helper function is used to marshal an `APIResponse` containing the `SignInResponse` data and return it to the client with a 200 OK status.
*   The video begins the implementation of storing the **refresh token in the database** (in the `refresh_tokens` table). This is crucial for allowing users to refresh their access tokens later and for potential token revocation.
*   A `RefreshTokensStore` struct is created in `store/refresh_tokens.go` with a `CreateRefreshToken` method. This store interacts with the `refresh_tokens` database table.
*   The `CreateRefreshToken` method takes the full JWT token object (specifically the refresh token generated by the `JWTManager`) and the user ID. It hashes the raw refresh token string using the `getBase64HashFromToken` helper, and then stores the hashed token, associated user ID, and the token's expiration time in the `refresh_tokens` table.
*   Unit tests are written for the `RefreshTokensStore`, specifically for the `CreateRefreshToken` method, to ensure that refresh tokens are stored correctly in the database.

**Video 08 - AWS Infra with Localstack, Docker, and Terraform**

*   The focus shifts to setting up the necessary **AWS infrastructure locally** using **LocalStack** and **Terraform** to support the asynchronous report generation process.
*   A **Docker Compose configuration** for running LocalStack is added to the project, allowing for easy startup of simulated AWS services.
*   Environment variables related to AWS (access key, secret key, region - dummy values are sufficient for LocalStack) and the names for the SQS queue and S3 bucket are added to the `.envrc` file.
*   **Terraform** is initialized within a `terraform` directory.
*   A `main.tf` file is created in the `terraform` directory to define the desired AWS resources: an **S3 bucket** to store the generated reports and an **SQS queue** to handle report creation requests.
*   Terraform `variable` blocks are defined to correspond to the environment variables set in `.envrc`, allowing Terraform to use these values.
*   The **AWS provider for Terraform is configured to override the default AWS API endpoints** to point to the LocalStack URLs, ensuring that the resources are created within the local simulation.
*   The S3 bucket's name is configured to use the value from the `S3_BUCKET_NAME` environment variable. Similarly, the SQS queue's name is set using the `SQS_QUEUE_NAME` ("reports-sqs") environment variable.
*   Terraform commands (`terraform init`, `terraform plan`, `terraform apply`) are used to provision the S3 bucket and SQS queue in the LocalStack environment. The `auto-approve` flag is used for the `apply` command for convenience.
*   The creation of the S3 bucket and SQS queue in LocalStack is verified using the **AWS Command Line Interface (CLI)**, specifically using the `--endpoint-url` flag to target the LocalStack service endpoints.
*   A specific configuration option, **`s3ForcePathStyle = true`**, is added to the AWS provider configuration for S3 to ensure that pre-signed URLs generated by LocalStack work correctly.
*   AWS SDK clients for S3 and SQS are created within the Go application, configured to use the LocalStack endpoints. This involves using the `aws-sdk-go-v2/config` and specific service clients (`aws-sdk-go-v2/service/s3`, `aws-sdk-go-v2/service/sqs`).

**Video 09 - Report Data Access Layer**

*   The implementation of the **data access layer for the `reports` table** is detailed.
*   A `ReportStore` struct is created in `store/reports.go`, holding a `sqlx.DB` connection. A `NewReportStore` constructor is provided.
*   The `Report` struct in Go is defined to map to the `reports` database table, with `db:` tags specifying the column names. Pointer types are used for fields that can be `NULL` in the database (e.g., `download_url`, `download_url_expires_at`, `started_at`, `completed_at`, `failed_at`, `error_message`).
*   The following CRUD (Create, Read, Update, Delete - though delete is not explicitly shown in this video) operations are implemented in the `ReportStore`:
    *   `Create`: Inserts a new report record into the `reports` table, taking the `user_id` and `report_type` as input. It uses `sqlx` to execute the `INSERT` statement and retrieve the generated `id` and `created_at` values.
    *   `Update`: Updates an existing report record in the `reports` table. It takes a `Report` struct as input, and updates fields like `output_file_path`, `download_url`, `download_url_expires_at`, and the status timestamps (`started_at`, `completed_at`, `failed_at`, `error_message`) based on the `user_id` and `id` of the report. The `sqlx.Named` function is used to build the `UPDATE` query based on the non-zero fields of the provided `Report` struct.
    *   `ByPrimaryKey`: Retrieves a specific report record from the `reports` table based on its `user_id` and `id` using a `SELECT` query and `sqlx.Get`.
*   Unit tests for the `ReportStore` are implemented, utilizing the test environment set up in previous videos.
*   The tests cover:
    *   Successfully creating a new report record and verifying its initial state (user ID, report type, `created_at`).
    *   Updating various fields of an existing report record (output file path, download URL, expiration, status timestamps) and ensuring the updates are reflected in the database. A test also verifies that non-updatable fields (like `report_type` in the current implementation of `Update`) are not modified.
    *   Retrieving a report record from the database using its primary key (`user_id` and `id`) and verifying that the retrieved data matches the expected values.
*   The tests highlight the dependency on having a user record exist in the `users` table (due to the foreign key constraint on `reports.user_id`) before creating report records.
*   An issue encountered during testing with reusing the same `Report` struct for both input and output of the `Update` operation is resolved by using a separate struct for scanning the updated values.

**Video 10 - CSV Report Builder**

*   The logic for **building the CSV report** is implemented within a `reports` package.
*   A `LegendOfZeldaClient` is created (by reusing code from a previous project) to interact with the Highroll Compendium API and fetch data about monsters. It has a `GetMonsters` method that retrieves and unmarshals the JSON response.
*   A `ReportBuilder` struct is defined, serving as a facade that encapsulates the report generation process. It holds dependencies on `ReportStore`, `LegendOfZeldaClient`, `S3Client`, and `slog.Logger`. A constructor `NewReportBuilder` is provided.
*   A central `Build` method is implemented on the `ReportBuilder`. This method takes a `context.Context`, `user_id`, and `report_id` as input (representing the information received from an SQS message).
*   The `Build` method first retrieves the corresponding `Report` record from the database using the `ReportStore`.
*   It checks if the report has already been started (by examining the `started_at` timestamp). If so, it returns without processing further to prevent duplicate work.
*   If the report hasn't started, the `started_at` timestamp is updated to the current time, and other status-related fields (`completed_at`, `failed_at`, `error_message`, `download_url`, `download_url_expires_at`, `output_file_path`) are reset to `nil` to ensure a clean state in case of retries. The updated `Report` record is then persisted in the database.
*   The `GetMonsters` method of the `LegendOfZeldaClient` is called to fetch the monster data.
*   The fetched monster data is then converted into a **gzipped CSV (Comma Separated Values) file**.
    *   A `bytes.Buffer` is created to hold the CSV data in memory.
    *   A `gzip.Writer` is created, writing to the `bytes.Buffer` to compress the data.
    *   A standard `csv.Writer` is created, writing to the `gzip.Writer`.
    *   A header row is written to the CSV writer, defining the columns (Name, ID, Category, Description, Image, Common Locations, Drops, DLC) based on the fields of the `Monster` struct.
    *   The method iterates through the slice of `Monster` data. For each monster, it creates a row of string values, converting integer IDs and boolean DLC flags to strings, and joining slices of strings (Common Locations, Drops) with a comma.
    *   Each row is written to the CSV writer.
    *   The `csv.Writer` is flushed to ensure all buffered data is written to the `gzip.Writer`.
    *   The `gzip.Writer` is closed, which also flushes any remaining data and writes the gzip footer to the `bytes.Buffer`.
*   The gzipped CSV data (now in the `bytes.Buffer`) is then **uploaded to the S3 bucket**.
    *   An S3 key (file path) is constructed using the `user_id` and `report_id` to organize the reports in S3 (e.g., `users/{user_id}/reports/{report_id}.csv.gz`).
    *   The `S3Client`'s `PutObject` method is used to upload the data. It takes the bucket name (from the config), the constructed key, and a `bytes.Reader` created from the `bytes.Buffer` as the body.
*   Upon successful S3 upload, the `Report` record in the database is updated with the `output_file_path` (the S3 key) and the `completed_at` timestamp is set to the current time.
*   A **deferred anonymous function** is used for error handling within the `Build` method. If any error occurs during the process (fetching data, CSV generation, S3 upload), this function checks if an error was returned. If so, it updates the `Report` record with the `failed_at` timestamp and the error message. It also logs the error using the `slog.Logger` dependency.

**Video 11 - Report API Endpoints**

*   Two API endpoints related to reports are implemented: one for creating a report and another for retrieving its status and a download URL.
*   A **POST endpoint at `/reports`** is created to initiate the report generation process.
    *   It expects a JSON request body with a `report_type` field. A `CreateReportRequest` struct with validation is defined.
    *   The request is decoded and validated.
    *   The **user ID of the authenticated user is retrieved from the request context** (this assumes the authentication middleware from earlier videos is in place and populates the context). A `userFromContext` helper function is introduced to extract the user from the context.
    *   A new `Report` record is created in the database using the `ReportStore`, with the extracted `user_id` and the `report_type` from the request.
    *   An **SQS message** is constructed, containing the `user_id` and the `id` of the newly created report. An `SQSMessage` struct is defined in `reports/sqs.go` for this purpose.
    *   The **URL of the SQS queue** (named in the config) is retrieved using the `SQSClient`.
    *   The SQS message (containing the user and report IDs, marshaled as JSON) is then sent to the queue using the `SQSClient`'s `SendMessage` method.
    *   An `APIReport` struct (a user-facing representation of the `Report` database model) is created and populated with the newly created report's information. This includes a `Status` field derived from the report's state (using a `Status()` method on `APIReport`).
    *   The `APIReport` is encoded as a JSON response with a 201 Created status.
*   A **GET endpoint at `/reports/{id}`** is created to retrieve the status of a specific report by its ID.
    *   The `id` (report ID) is extracted from the URL path parameters.
    *   The **user ID is again retrieved from the request context** for authorization and to ensure users can only access their own reports.
    *   The `ReportStore`'s `ByPrimaryKey` method is used to fetch the report record from the database using the user ID and the report ID.
    *   If no report is found with the given ID for the authenticated user, a 404 Not Found status is returned.
    *   If the report's `completed_at` timestamp is not `nil` (meaning the report generation is complete), a **pre-signed S3 URL** for downloading the report is generated.
        *   An `s3.PresignClient` is added as a dependency to the `Server`.
        *   The `PresignClient`'s `PresignGetObject` method is used to generate a temporary, signed URL that allows downloading the S3 object without needing AWS credentials. This method takes the bucket name (from config), the object key (from the report record's `output_file_path`), and an expiration duration (e.g., 10 seconds) as parameters.
        *   The generated pre-signed URL and its expiration time are then updated and stored in the `download_url` and `download_url_expires_at` fields of the `Report` record in the database using the `ReportStore`'s `Update` method.
    *   The `APIReport` (containing the report details and the pre-signed download URL if the report is complete) is encoded and returned as a JSON response with a 200 OK status.
    *   Logic is added to **generate a new pre-signed URL only if the existing one has expired or if it hasn't been generated yet**.

**Video 12 - SQS Worker**

*   The implementation of the **SQS worker** application is detailed. This application is responsible for consuming messages from the SQS queue and triggering the report generation process.
*   A `Worker` struct is defined in the `reports` package, holding dependencies on `config`, `ReportBuilder`, `slog.Logger`, and `sqs.Client`. It also includes a channel (`messages`) to receive SQS messages.
*   A `NewWorker` constructor is provided to initialize the `Worker` struct with its dependencies and a `MaxConcurrency` parameter to control the number of concurrent worker goroutines.
*   A `Start` method is implemented on the `Worker`. This method launches a specified number of **goroutines**. Each goroutine continuously reads messages from the `messages` channel.
*   Inside each worker goroutine:
    *   An SQS `types.Message` is received from the channel.
    *   The **body of the SQS message (which is a JSON string containing the user ID and report ID)** is unmarshaled into a custom `sqs.Message` struct.
    *   Error handling is implemented for potential issues during JSON unmarshaling. If the message body is invalid, the error is logged, and the goroutine continues to the next message.
    *   A `context.Context` with a timeout (e.g., 10 seconds) is created for the report building process.
    *   The `Build` method of the `ReportBuilder` is called, passing the context, the `user_id`, and the `report_id` extracted from the SQS message.
    *   If the `Build` method completes without error (meaning the report was generated and uploaded successfully), the **SQS message is deleted from the queue** using the `sqs.Client`'s `DeleteMessage` method and the message's receipt handle.
    *   If the `Build` method returns an error, the error is logged. The deferred error handling within the `ReportBuilder` is expected to have updated the report status in the database to "failed". The message is not deleted from the queue in case of an error, allowing for potential retries (though explicit retry logic is not implemented in this video).
*   A `cmd/worker/main.go` file is created to serve as the entry point for the SQS worker application.
    *   It loads the application configuration and AWS configuration.
    *   It creates instances of the `ReportStore`, `LegendOfZeldaClient`, `S3Client`, `ReportBuilder`, and `Worker`, injecting the necessary dependencies.
    *   It starts the worker by calling the `Start` method with the configured concurrency level and a context that handles shutdown signals.
*   The video demonstrates running both the API server and the worker application. When a new report creation request is made via the API, an SQS message is published. The worker consumes this message, the report is built and uploaded to S3, and the worker deletes the message from the queue. Subsequently, the client can poll the API's report status endpoint and, once the report is complete, download the generated CSV file using the pre-signed URL.
*   A bug in the pre-signed URL generation logic (related to checking if a URL needs to be generated) is identified and addressed.
*   The video concludes with a discussion on the scenarios where asynchronous processing with SQS workers would be beneficial compared to synchronous request-response patterns.