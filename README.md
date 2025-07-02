# Mateo Service

A Go service template with clean architecture, PostgreSQL, and Echo framework.

## Project Structure

```
.
├── cmd/
│   └── http/              # Main application entry point
├── internal/
│   ├── config/           # Configuration loading and validation
│   ├── domain/            # Core business logic and interfaces
│   ├── service/           # Application service layer
│   ├── store/             # Data access layer
│   │   └── pg/            # PostgreSQL implementation
│   └── transport/         # HTTP handlers and routing
│       └── http/           # Echo web server setup
├── .env.example           # Example environment variables
└── README.md              # This file
```

## Getting Started

### Prerequisites

- Go 1.16+
- PostgreSQL 13+
- Make (optional, but recommended)

### Installation

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd mateo
   ```

2. Copy the example environment file and update the values:
   ```bash
   cp .env.example .env
   ```

3. Install dependencies:
   ```bash
   go mod download
   ```

4. Set up the database:
   ```sql
   CREATE DATABASE mateo_db;
   ```

5. Run database migrations (if any):
   ```bash
   # Add migration commands here when you have them
   ```

### Running the Application

```bash
# Start the server
go run cmd/http/main.go
```

The server will start on `http://localhost:8080` by default.

## Development

### Environment Variables

| Variable     | Default   | Description                |
|--------------|-----------|----------------------------|
| HTTP_PORT    | 8080      | Port for the HTTP server   |
| DB_HOST      | localhost | Database host             |
| DB_PORT      | 5432      | Database port             |
| DB_USER      | postgres  | Database user             |
| DB_PASSWORD  | postgres  | Database password         |
| DB_NAME      | mateo_db  | Database name             |
| DB_SSLMODE   | disable   | SSL mode for database      |


### Testing

```bash
# Run tests
go test ./...

```

## License

MIT
