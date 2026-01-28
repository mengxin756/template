# Classic Go Project

A classic Go project based on Clean Architecture, utilizing a modern technology stack and best practices

## 🏗️ Project Architecture

### Directory Structure
```
template/
├── cmd/                    # Main application entry points
│   └── api/              # API service entry
├── internal/              # Core business logic (cannot be imported externally)
│   ├── config/           # Configuration management
│   ├── domain/           # Domain objects, entities, interfaces
│   ├── service/          # Business use case implementations
│   ├── repository/       # Data access layer
│   ├── handler/          # HTTP handlers
│   ├── data/             # Data layer
│   │   └── ent/         # Ent ORM
│   └── server/           # Server configuration
├── pkg/                   # Public utility libraries
│   ├── errors/           # Unified error handling
│   ├── logger/           # Structured logging
│   └── response/         # HTTP response formatting
├── config/                # Configuration files
├── api/                   # API definitions
└── test/                  # Test files
```

### Technology Stack
- **Web Framework**: Gin
- **ORM**: Ent
- **Dependency Injection**: Google Wire
- **Logging**: Zerolog
- **Configuration Management**: Viper
- **Database**: SQLite (supports MySQL/PostgreSQL)
- **Task Queue**: Asynq (planned)
- **Message Queue**: Kafka (planned)

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- SQLite (for development)


### Install Dependencies
```bash
go mod tidy
```

### Run the Project
```bash
go run ./cmd/api
```

### Build the Project
```bash
go build ./cmd/api
```

## 📋 Features

### User Management
- ✅ User registration
- ✅ User query
- ✅ User update
- ✅ User deletion
- ✅ User status management
- ✅ Paginated queries

### Technical Features
- ✅ Clean Architecture layered design
- ✅ Dependency Injection (Wire)
- ✅ Structured Logging (Zerolog)
- ✅ Unified error handling
- ✅ Unified response formatting
- ✅ Request tracing (Trace ID)
- ✅ Middleware support
- ✅ Configuration management
- ✅ Unit testing

## 🔧 Configuration

### Environment Variables
Override configurations via environment variables using `SECTION_KEY` format, e.g.:
- `HTTP_ADDRESS` → `http.address`
- `DB_DRIVER` → `db.driver`

### Configuration File
Primary config: `config/config.yaml`, including:
- HTTP service settings
- Logging configuration
- Database configuration
- Redis configuration
- Asynq configuration
- Kafka configuration

## 🧪 Testing

### Run Tests
```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./internal/service

# Run tests with coverage report
go test -cover ./...
```

### Test Coverage Target
 ≥ 60%

## 📚 API Documentation

### User Management APIs

#### Register User
```http
POST /api/v1/users
Content-Type: application/json

{
  "name": "username",
  "email": "user@example.com",
  "password": "password123"
}
```

#### Get User List
```http
GET /api/v1/users?page=1&page_size=20&status=active
```

#### Get User Details
```http
GET /api/v1/users/{id}
```

#### Update User
```http
PUT /api/v1/users/{id}
Content-Type: application/json

{
  "name": "新用户名",
  "status": "inactive"
}
```

#### Delete User
```http
DELETE /api/v1/users/{id}
```

#### Change User Status
```http
PATCH /api/v1/users/{id}/status
Content-Type: application/json

{
  "status": "banned"
}
```

## 🔍 Current Status

### ✅ Completed
- Clean Architecture implementation
- User domain models and interfaces
- User repository layer
- User service layer
- HTTP handlers
- Unified error handling
- Unified response formatting
- Configuration management
- Middleware implementation
- Unit testing framework

### ⚠️ Known Issues
- Logger package (`pkg/logger`) has zerolog API usage issues
- Ent code generation needs regeneration
- Partial dependency injection configuration requires improvement

### 🚧 In Progress
- Project refactoring and architecture optimization
- Code quality improvements

### 📋 Planned
- Redis integration
- Asynq task queue
- Kafka message queue
- Monitoring and metrics
- Integration testing
- Docker support
- CI/CD configuration

## 🤝 Contribution Guide

1. Fork the project
2. Create a feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

## 📞 Contact

For questions or suggestions:
- Open an Issue
- Send an email
- Join discussions

---

**Note**: This is an under-refactoring project. Some features may be unstable. Recommended for development environments only.

## 🔗 Featured Link

**Tools.Beer** is a free online toolkit for developers, designers, and general users.  
No installation required – open your browser to access data processing, encryption, image editing, and document conversion tools.

### 🔧 Key Features
- 🛠 Developer Tools: [JSON Formatter](https://tools.beer/en/json), [Regex Tester](https://tools.beer/en/regex), [Base64 Encoder/Decoder](https://tools.beer/en/base64), [UUID/Password Generator](https://tools.beer/en/password)
- 🔐 Security & Encryption: [JWT Decoder](https://tools.beer/en/jwt), [Hash Calculator](https://tools.beer/en/hash)
- 📊 Data Conversion: [CSV ↔ Parquet](https://tools.beer/en/parquet), [YAML ↔ JSON](https://tools.beer/en/yaml), [URL Encoder/Decoder](https://tools.beer/en/url), [Timestamp Converter](https://tools.beer/en/timestamp)
- 🖼 Image Tools: [Image Compression](https://tools.beer/en/imgcompress), [Format Conversion](https://tools.beer/en/imgconvert), [Cropping](https://tools.beer/en/imgcrop), [Watermarking](https://tools.beer/en/imgwatermark), [Rotation](https://tools.beer/en/imgrotate)
- 📄 Files & Documents: [PDF Tools](https://tools.beer/en/pdf), [Smart Tools](https://tools.beer/en/smart)
- 🎨 Design Utilities: [Color Picker](https://tools.beer/en/colorpicker), [QR Code Generator](https://tools.beer/en/qrcode), [Barcode Generator](https://tools.beer/en/barcode)

✨ Fast, minimalistic, and secure. Supports multiple languages (English & 中文). Forever free.

