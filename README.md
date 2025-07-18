# Middleware Strategy for Go Microservices

This is a high-performance Go microservice designed to calculate loan pricing rates by intermediating calls to an external decision engine. It is built to be resilient, scalable, and observable, capable of handling millions of requests. The service uses middleware layers for observability (Datadog metrics), data enrichment, and guardrails to ensure accurate pricing and prevent failures.

## Features

- **HTTP Endpoint**: Exposes a `/pricing` POST endpoint to process loan pricing requests.
- **Observability**: Logs metrics (request count, duration, errors) to Datadog.
- **Data Enrichment**: Fetches additional data (e.g., credit score) to enhance pricing decisions.
- **Guardrails**: Validates calculated rates to prevent invalid or erroneous results.
- **Resilience**: Uses a circuit breaker for calls to the external decision engine.
- **Scalability**: Stateless design for horizontal scaling.

## Project Structure

```

pricing-service/
├── api/             # HTTP handlers and routes
├── middleware/      # Middleware for observability, enrichment, and guardrails
├── service/         # Business logic (decision engine integration)
├── model/           # Data structures (request/response models)
├── config/          # Configuration loading (e.g., Datadog, timeouts)
├── test_pricing.sh  # Test script for Linux/macOS
├── test_pricing.bat # Test script for Windows
├── doc.go           # Package documentation
├── main.go          # Service entry point
└── README.md        # This file

```

## Prerequisites

- **Go**: Version 1.18 or higher.
- **Datadog**: Agent running locally or configured for remote metrics (default: `127.0.0.1:8125`).
- **curl**: For testing (included in macOS and Windows 10/11).
- **jq** (optional): For formatted JSON output in tests (install via `brew install jq` on macOS or download for Windows).
- **External Decision Engine**: Mocked in the example; replace with your service URL.

## Installation

1. Clone the repository:

   ```bash
   git clone <repository-url>
   cd pricing-service
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```
3. Ensure the Datadog agent is running or configured.

## Running the Service

1. Start the service:

   ```bash
   go run main.go
   ```

   The service will listen on `http://localhost:8080`.

2. Verify the service is running by checking logs for:
   ```
   [GIN-debug] Listening and serving HTTP on :8080
   ```

## Testing

Test the `/pricing` endpoint using the provided scripts.

### Linux/macOS

1. Make the script executable:
   ```bash
   chmod +x test_pricing.sh
   ```
2. Run the test:
   ```bash
   ./test_pricing.sh
   ```

### Windows

1. Run the test:
   ```cmd
   test_pricing.bat
   ```

**Example Payload**:

```json
{
  "loan_amount": 10000.0,
  "customer_id": "CUST12345"
}
```

**Expected Output (Success)**:

```json
{
    "rate": 5.0,
    "status": "success"
}
Requisição enviada com sucesso!
```

**Expected Output (Error)**:

```json
{
    "error": "missing enrichment data"
}
Erro ao enviar a requisição.
```

## Debugging

- **Logs**: Check the service logs for errors (e.g., `zap` logs in `middleware/data_enrichment.go`).
- **Datadog Metrics**: Monitor `pricing_service.request.count`, `pricing_service.request.duration`, and `pricing_service.request.error`.
- **Common Issues**:
  - `"missing enrichment data"`: Ensure `c.Set("enriched_data", ...)` is called in `middleware/data_enrichment.go`.
  - HTTP 400/500: Verify the payload format and decision engine connectivity.

## Configuration

- **Datadog**: Update the agent address in `main.go` if not using `127.0.0.1:8125`.
- **Timeouts**: Adjust HTTP client and circuit breaker timeouts in `service/decision_engine.go`.
- **Endpoint**: Modify the decision engine URL in `service/decision_engine.go`.

## Future Improvements

- Add unit and integration tests for middleware and handlers.
- Implement caching (e.g., Redis) for enriched data.
- Add structured logging for better traceability.
- Externalize configuration using environment variables or a config file.

## License

MIT License
