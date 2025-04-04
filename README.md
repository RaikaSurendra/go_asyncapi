


# AsyncAPI Project
=====================================

## Overview

The AsyncAPI project is a Go-based API framework for building scalable and efficient APIs.
Followed AsyncAPI project created by devtool https://www.youtube.com/watch?v=jjcIY_gNdAY&list=PLlRRjbUiNHm3fcvjiSnAa4n3zqmKF4wCl

## Technical Overview

The AsyncAPI project is built on top of the Go programming language and uses a number of open-source libraries and frameworks to provide its functionality. The project consists of several components, including:

* **API Gateway**: The API gateway is the entry point for all incoming API requests. It is responsible for routing requests to the correct handler.
* **Handler**: The handler is responsible for processing incoming API requests and returning responses. Handlers are written in Go and can be customized to handle specific business logic.
* **Database**: The database is used to store data for the API. The project uses a PostgreSQL database.

## Architecture

The AsyncAPI project uses a microservices architecture, with each component running in its own process.

## Technical Details

* **Go Version**: The project uses Go 1.14 or later.
* **Database**: The project uses PostgreSQL 12 or later.

## Running the Project

To run the project, follow these steps:

### Step 1: Start LocalStack

Run `docker-compose up` to start LocalStack, which provides a mock AWS environment for testing.

### Step 2: Start API Server

Start the API server by running `go run main.go` in the `apiserver` directory.

### Step 3: Start Worker

Start the worker process by running `go run main.go` in the `worker` directory. This process is responsible for report creation and search.

### Note

Make sure LocalStack is properly configured before running the project. This includes setting up the necessary AWS credentials and configuring the LocalStack environment.

## Usage

To use the AsyncAPI project, follow these steps:

1. Send a request to the API gateway using a tool like `curl` or a web browser.
2. The API gateway will route the request to the correct handler.
3. The handler will process the request and return a response.
4. The response will be returned to the client.
