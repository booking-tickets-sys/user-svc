# User Service

A production-ready gRPC-based user authentication service with comprehensive testing, graceful shutdown, and modern Go practices.

## 🚀 Features

- **User Authentication**: Registration and login with email/password
- **Token Management**: JWT and PASETO token support with refresh tokens
- **Database**: PostgreSQL with transaction support and migrations
- **gRPC API**: Full gRPC implementation with reflection enabled
- **Testing**: Comprehensive unit tests with mocked dependencies
- **Configuration**: Flexible configuration with environment variable support
- **Logging**: Structured logging with multiple output formats
- **Graceful Shutdown**: Proper service shutdown handling
- **Error Handling**: Standardized gRPC error responses
- **Security**: Password hashing, token validation, and secure defaults

## 📋 Prerequisites

- **Go 1.24.4** or later
- **PostgreSQL** database
- **Protocol Buffers** compiler (protoc)
- **Make** (for build automation)

## 🛠️ Installation & Setup

### 1. Clone and Setup

```bash
git clone <repository-url>
cd user-svc
```

### 2. Install Dependencies

```bash
make deps
```

### 3. Setup Protocol Buffers

```bash
# Install protoc (macOS)
brew install protobuf

# Install Go protobuf plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Setup proto submodule and generate files
make proto-setup
```

### 4. Database Setup

```bash
# Create PostgreSQL database
createdb users

# Run migrations (manually or using a migration tool)
# Migrations are located in db/migrations/
```

## ⚙️ Configuration

Update `config.yaml` with your settings:

```yaml
app:
  name: "user-svc"
  version: "1.0.0"
  environment: "development"

server:
  grpc:
    port: 9090
    host: "localhost"
    graceful_shutdown_timeout: 30s

database:
  host: "localhost"
  port: 5432
  user: "your_db_user"
  password: "your_db_password"
  db_name: "users"
  ssl_mode: "disable"
  max_open_conns: 10
  max_idle_conns: 5
  conn_max_lifetime: 5m

security:
  jwt:
    secret_key: "your-base64-encoded-jwt-secret"
    token_duration: 15m
    issuer: "user-svc"
  
  paseto:
    secret_key: "your-paseto-secret-key"
    token_duration: 15m

logging:
  level: "info"
  format: "json"
  output: "stdout"
```

### Environment Variables

For production, use environment variables:

```bash
export JWT_SECRET_KEY="your-jwt-secret"
export PASETO_SECRET_KEY="your-paseto-secret"
export DB_PASSWORD="your-db-password"
```

## 🏃‍♂️ Running the Service

### Development

```bash
# Build and run
make run

# Or build and run separately
make build
./bin/user-svc
```

### Production

```bash
# Build for production
make build

# Run with production config
./bin/user-svc -config=config.prod.yaml
```

The gRPC server will start on `localhost:9090`.

## 📚 API Documentation

### User Service

#### Register User

```protobuf
rpc Register(RegisterRequest) returns (RegisterResponse)
```

**Request:**
```json
{
  "email": "user@example.com",
  "username": "username",
  "password": "securepassword"
}
```

**Response:**
```json
{
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "username": "username",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "access_token": "jwt_token_here"
}
```

#### Login User

```protobuf
rpc Login(LoginRequest) returns (LoginResponse)
```

**Request:**
```json
{
  "email": "user@example.com",
  "password": "securepassword"
}
```

**Response:**
```json
{
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "username": "username",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "access_token": "jwt_token_here"
}
```

## 🧪 Testing

### Run Tests

```bash
# Run all tests
make test

# Run tests with coverage
go test -v -cover ./...

# Run specific test
go test -v ./internal/app/service
```

### Mock Generation

```bash
# Generate mocks
make mock

# Clean mocks
make mock-clean
```

### Test Structure

```
internal/app/service/
├── user_test.go          # Service unit tests with mocks
└── user.go              # Service implementation

internal/domain/
├── user_test.go         # Domain model tests
├── refresh_token_test.go # Refresh token tests
└── password_test.go     # Password validation tests

token/
├── jwt_maker_test.go    # JWT token maker tests
└── paesto_maker_test.go # PASETO token maker tests

db/
└── connection_test.go   # Database connection tests
```

## 🔧 Development

### Using gRPC Tools

The server has gRPC reflection enabled for development:

```bash
# List services
grpcurl -plaintext localhost:9090 list

# List methods
grpcurl -plaintext localhost:9090 list user.UserService

# Call register method
grpcurl -plaintext -d '{
  "email": "test@example.com", 
  "username": "testuser", 
  "password": "password123"
}' localhost:9090 user.UserService/Register
```

### Protocol Buffer Development

```bash
# Update proto submodule
make proto-update

# Generate protobuf files
make proto-gen

# Clean generated files
make proto-clean
```

## 📁 Project Structure

```
user-svc/
├── cmd/                    # Application entry points (future)
├── config/                 # Configuration management
│   ├── config.go          # Configuration structs
│   └── config_test.go     # Configuration tests
├── db/                     # Database layer
│   ├── connection.go      # Database connection
│   ├── connection_test.go # Connection tests
│   └── migrations/        # Database migrations
├── docs/                   # Documentation
├── internal/               # Private application code
│   ├── app/               # Application layer
│   │   ├── grpc/          # gRPC server implementation
│   │   ├── repository/    # Data access layer
│   │   └── service/       # Business logic layer
│   └── domain/            # Domain models and business rules
│       ├── dto/           # Data transfer objects
│       └── errs/          # Domain errors
├── logger/                 # Logging utilities
├── mocks/                  # Generated mock files
├── pb/                     # Generated protobuf code
├── submodules/             # Git submodules
│   └── proto/             # Protocol buffer definitions
├── token/                  # Token management (JWT/PASETO)
├── utils/                  # Utility functions
├── config.yaml            # Configuration file
├── Dockerfile             # Container configuration
├── go.mod                 # Go module definition
├── go.sum                 # Dependency checksums
├── main.go               # Application entry point
└── Makefile              # Build automation
```

## 🐳 Docker

### Build Image

```bash
docker build -t user-svc .
```

### Run Container

```bash
docker run -p 9090:9090 \
  -e DB_HOST=host.docker.internal \
  -e DB_PASSWORD=your_password \
  user-svc
```

## 📋 Available Commands

```bash
make help          # Show all available commands
make build         # Build the application
make clean         # Clean build artifacts
make run           # Build and run the application
make test          # Run all tests
make deps          # Install dependencies
make dev-setup     # Setup development environment
make mock          # Generate mocks for testing
make mock-clean    # Clean generated mocks
make proto-update  # Update proto submodule
make proto-gen     # Generate protobuf files
make proto-clean   # Clean protobuf files
make proto-setup   # Setup proto submodule and generate files
```

## 🔒 Security Features

- **Password Hashing**: Bcrypt with configurable cost
- **Token Security**: JWT and PASETO token support
- **Input Validation**: Comprehensive validation for all inputs
- **Error Handling**: Secure error responses without information leakage
- **Configuration Security**: Environment variable support for secrets

## 🚨 Error Handling

The service returns standardized gRPC status codes:

- `INVALID_ARGUMENT`: Missing required fields or invalid input
- `INTERNAL`: Server errors
- `ALREADY_EXISTS`: User already exists (registration)
- `UNAUTHENTICATED`: Invalid credentials (login)
- `NOT_FOUND`: User not found

## 📊 Monitoring & Logging

### Logging Configuration

```yaml
logging:
  level: "info"           # debug, info, warn, error, fatal, panic
  format: "json"          # json or text
  output: "stdout"        # stdout, stderr, or file path
  file:
    enabled: false
    path: "logs/app.log"
    max_size: 100         # MB
    max_age: 30           # days
    max_backups: 10
```

### Graceful Shutdown

The service implements graceful shutdown with configurable timeout:

```yaml
server:
  grpc:
    graceful_shutdown_timeout: 30s
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the test suite
6. Submit a pull request

## 📄 License

[Add your license information here]

## 🆘 Support

For issues and questions:
- Create an issue in the repository
- Check the documentation in the `docs/` directory
- Review the test files for usage examples 