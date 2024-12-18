# Memes as a Service (MaaS)

A high-performance microservice for delivering memes on demand. This service provides a RESTful API that allows clients to fetch memes based on location and query parameters, with a token-based billing system.

## Prerequisites

- Go 1.19 or higher
- SQLite3

## Installation

1. Clone the repository:

```bash
git clone https://github.com/jeanffc/maas-memes-service.git
cd maas-memes-service
```

2. Install dependencies:

```bash
go mod download
```

## Running the Service

Start the service:

```bash
go run main.go
```

## API Endpoints

### Get a Meme

```
GET /memes?lat={latitude}&lon={longitude}&query={search_term}
```

Headers:

- `X-Client-ID`: Required for authentication and token billing

Parameters:

- `lat`: Latitude (float)
- `lon`: Longitude (float)
- `query`: Search term (string)

Request Example:

```bash
curl -H "X-Client-ID: client123" "http://localhost:8080/memes?lat=40.73061&lon=-73.935242&query=food"
```

Response:

```json
{
  "id": "1234567890",
  "url": "https://example.com/meme.jpg",
  "caption": "A meme about food",
  "query": "food",
  "latitude": 40.73061,
  "longitude": -73.935242,
  "created_at": "2024-12-17T12:00:00Z"
}
```

### Check Token Balance

```
GET /balance
```

Headers:

- `X-Client-ID`: Required for authentication

Request Example:

```bash
curl -H "X-Client-ID: client123" http://localhost:8080/balance
```

Response:

```json
{
  "client_id": "client123",
  "balance": 100
}
```

### Add Tokens

```
POST /tokens
```

Request Body:

```json
{
  "client_id": "client123",
  "balance": 50
}
```

Request Example:

```bash
curl -X POST -H "Content-Type: application/json" -d '{"client_id":"client123","balance":50}' http://localhost:8080/tokens
```

Response:

```json
{
  "status": "success"
}
```

## Error Handling

The service returns appropriate HTTP status codes and error messages:

- 400: Bad Request (invalid parameters)
- 401: Unauthorized (missing client ID)
- 402: Payment Required (insufficient tokens)
- 429: Too Many Requests (rate limit exceeded)
- 500: Internal Server Error

Example error response:

```json
{
  "error": "invalid latitude"
}
```

## Rate Limiting

The service implements rate limiting of 100 requests per second with a burst capacity of 200 requests. When the rate limit is exceeded, the service returns a `429` status code.

## Development

### Run Tests

```bash
go test ./...
```

### Format Code

```bash
go fmt ./...
go vet ./...
```
