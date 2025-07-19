# Record Query Microservice

This is a Go-based microservice for querying records from a MySQL database via an HTTP API. It uses the Gin framework for routing, Datadog for observability (metrics and tracing), and a middleware chain for request processing. The service is designed to be lightweight, scalable, and observable.

## Features

- **HTTP API**: Exposes a POST `/consulta` endpoint to query records by user ID.
- **Database Integration**: Connects to a MySQL database to retrieve records.
- **Observability**: Integrates with Datadog for metrics and tracing, and uses structured logging with `slog`.
- **Middleware**: Supports extensible middleware for observability, validation, and business logic.
- **Containerized Services**: Uses Docker Compose to manage dependencies (MySQL, Datadog).

## Prerequisites

- **Go**: Version 1.18 or higher.
- **Docker** and **Docker Compose**: For running MySQL and Datadog services.
- **Datadog API Key**: Required for Datadog observability (set as `DD_API_KEY` environment variable).
- **Make**: For running build and service management commands.

## Setup

1. **Clone the Repository**:

   ```bash
   git clone <repository-url>
   cd <repository-directory>
   ```

2. **Install Dependencies**:
   Ensure Go dependencies are installed and code is formatted:

   ```bash
   make deps
   ```

3. **Set Up Datadog API Key**:
   Export your Datadog API key as an environment variable:

   ```bash
   export DD_API_KEY=your-datadog-api-key
   ```

4. **Start Services**:
   Start the MySQL and Datadog containers using Docker Compose:

   ```bash
   make services
   ```

   This command runs the `config/run-services.sh` script to initialize the services defined in `config/docker-compose.yml`.

5. **Build the Application**:
   Compile the Go code to generate the binary:
   ```bash
   make build
   ```

## Usage

1. **Run the Application**:
   Start the microservice, which will listen on port 8080:

   ```bash
   make run
   ```

2. **Query the API**:
   Send a POST request to the `/consulta` endpoint with a JSON payload containing the `user_id` field. Example using `curl`:

   ```bash
   curl -X POST http://localhost:8080/consulta -H "Content-Type: application/json" -d '{"user_id": "12345", "name": "optional-name"}'
   ```

   **Example Response (Success)**:

   ```json
   [
     {
       "id": "12345",
       "name": "John Doe"
     }
   ]
   ```

   **Example Response (Error)**:

   ```json
   {
     "error": "Payload inválido: missing user_id"
   }
   ```

3. **View Logs**:
   Monitor Datadog agent logs to inspect service activity. Use the following commands to filter logs:
   - General service logs:
     ```bash
     tail -f /var/log/datadog/agent.log | grep "sample_service"
     ```
   - Logs for a specific user ID (e.g., `abc123`):
     ```bash
     tail -f /var/log/datadog/agent.log | grep "user_id:abc123"
     ```
   - Logs for a specific action (e.g., `login`, though not implemented in this service):
     ```bash
     tail -f /var/log/datadog/agent.log | grep "action:login"
     ```
   - Metrics logs for request counts:
     ```bash
     tail -f /var/log/datadog/agent.log | grep "sample_service.registros.requests_total"
     ```

## Makefile Commands

- `make services`: Starts MySQL and Datadog containers.
- `make run`: Runs the Go application with Datadog observability enabled.
- `make build`: Compiles the Go code into a binary.
- `make stop`: Stops and removes Docker containers.
- `make status`: Checks the status of Docker containers.
- `make test`: Runs all tests with verbose output.
- `make bench`: Runs benchmarks with memory profiling.
- `make clean`: Removes the binary and stops Docker containers.
- `make deps`: Installs dependencies and formats code.

## Configuration

- **Database**: The MySQL connection string is set to `dbuser:senha@tcp(localhost:3306)/db?parseTime=true`. Update the credentials and host in `database.go` if needed.
- **Datadog**: The Datadog agent is expected at `127.0.0.1:8125`. Ensure the agent is running via `make services`.
- **Port**: The HTTP server listens on port `8080`.

## Project Structure

- `main.go`: Configures the HTTP server, middleware, and routes.
- `database.go`: Manages MySQL connections and queries.
- `metrics.go`: Handles Datadog metrics collection.
- `models.go`: Defines data structures for API and database interactions.
- `doc.go`: Package-level documentation for the microservice.
- `config/docker-compose.yml`: Docker Compose configuration for MySQL and Datadog.
- `config/run-services.sh`: Script to start services with the Datadog API key.

## Error Handling

The service logs errors using `slog` and returns appropriate HTTP status codes:

- `400 Bad Request`: For invalid payloads.
- `500 Internal Server Error`: For database or processing errors.

Metrics are recorded for each request to the `/consulta` endpoint using the Datadog statsd client.

## Troubleshooting

- **Datadog API Key Error**: Ensure `DD_API_KEY` is set before running `make services`.
- **Database Connection Failure**: Verify MySQL is running and accessible at `localhost:3306`.
- **Log Inspection**: Use the `tail -f` commands above to debug issues.

For further details, refer to the [package documentation](doc.go).
