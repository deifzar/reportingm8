# Development Guide

## Quick Start

### Prerequisites
- Go 1.24+ 
- PostgreSQL 12+
- RabbitMQ 3.8+
- AWS S3 access (for report storage)
- SMTP server access (for email notifications)

### Initial Setup

1. **Clone and Build**
```bash
git clone <repository-url>
cd reportingm8
go mod download
go build -o reportingm8 main.go
```

2. **Configuration Setup**
```bash
# Copy example configuration
cp configs/examples/configuration.yaml configs/configuration.yaml
# Edit configs/configuration.yaml with your environment settings
# Note: The application will create the configs/ directory if it doesn't exist
```

3. **Database Setup**
```bash
# Create database and tables (schema not provided in repo)
# TODO: Add database migration scripts
```

4. **Run the Application**
```bash
./reportingm8 launch
# OR for development
go run main.go launch
```

## Development Commands

### Building and Running
```bash
# Build the application
go build -o reportingm8

# Run directly with Go
go run main.go launch

# Show available commands
./reportingm8 --help

# Show version information  
./reportingm8 version
```

### Code Quality and Maintenance
```bash
# Format code
go fmt ./...

# Lint and check for issues
go vet ./...

# Update dependencies
go mod tidy

# Download dependencies
go mod download

# Check for security vulnerabilities
go list -json -deps ./... | nancy sleuth
```

## Project Structure

### Core Packages

```
pkg/
├── amqpM8/              # RabbitMQ connection pooling (5 files)
│   ├── connection_pool.go        # Core connection pool
│   ├── pool_manager.go           # Singleton pool manager
│   ├── pooled_amqp.go            # Pooled operations
│   ├── initialization.go         # Pool initialization
│   └── shared_state.go           # Global state
├── api8/                # HTTP API endpoints (Gin framework)
├── cleanup8/            # Temporary file cleanup service
├── cloud8/              # AWS S3 integration for report storage
├── configparser/        # Viper-based YAML configuration parser
├── controller8/         # HTTP controllers and scheduler management
├── db8/                 # Data access layer (23 files, 8 interfaces)
│   ├── db8_*.go                  # Domain-specific DAOs
│   └── db8_*_interface.go        # Interface definitions
├── email8/              # SMTP email service integration
├── gocron8/             # Scheduled task management (singleton wrapper)
├── log8/                # Structured logging with Zerolog + Lumberjack
├── model8/              # Data models (11 domain types)
├── notification8/       # Multi-channel notification publishing
├── orchestrator8/       # RabbitMQ orchestration and routing
├── reporting8/          # Core report generation business logic
└── utils/               # Utility functions (security posture, IP validation)
```

### Configuration Files
- `configs/configuration.yaml` - Main application configuration (auto-created if missing)
- `configs/examples/configuration.yaml` - Example configuration template
- `go.mod` / `go.sum` - Go module dependencies
- `CLAUDE.md` - Project documentation and AI assistant instructions
- `main.go` - Application entry point with umask 0027 security setting

### Asset Files
- `assets/` - HTML templates for reports and emails
- `tmp/` - Temporary files (runtime, auto-cleaned)
- `log/` - Log files (runtime, with 0640 permissions)

## Development Workflow

### Interface-First Development
The codebase follows an interface-first approach:

1. **Define Interface** - Create `*_interface.go` file with method signatures
2. **Implement Structure** - Create concrete implementation
3. **Dependency Injection** - Use interfaces for all dependencies
4. **Testing** - Mock interfaces for unit testing

**Example Pattern**:
```go
// service_interface.go
type ServiceInterface interface {
    DoSomething() error
}

// service.go  
type Service struct {
    db database.Interface
}

func NewService(db database.Interface) ServiceInterface {
    return &Service{db: db}
}
```

### Manual Acknowledgment Workflow

The system supports manual message acknowledgment for reliable message processing. This pattern ensures messages are only removed from the queue after successful completion of long-running operations.

#### Acknowledgment Flow
```
1. RabbitMQ delivers message to consumer (manual ACK mode: autoack=false)
2. Orchestrator receives message and extracts delivery tag
3. HTTP request forwarded to service endpoint with delivery tag in headers
   - Header: X-RabbitMQ-Delivery-Tag
   - Header: X-RabbitMQ-Consumer-Tag
4. Service processes request and performs long-running operation (scan)
5. Service calls orchestrator.AckScanCompletion() with delivery tag
   - scanCompleted=true: Sends ACK to RabbitMQ (message removed)
   - scanCompleted=false: Sends NACK with requeue (message retried)
6. On handler error: Automatic NACK with requeue
7. On missing handler: Automatic NACK without requeue (permanent failure)
```

#### Manual ACK Best Practices
- **Use manual ACK for**: Long-running operations, critical scans, external API calls
- **Use auto-ACK for**: Fast operations, non-critical tasks, idempotent operations
- **Error Handling**:
  - Handler errors → NACK with requeue (temporary failure, retry)
  - Missing handler → NACK without requeue (permanent failure, no retry)
  - Panic/crash → Message redelivered automatically (not ACKed)
- **Delivery Tag Tracking**: Store delivery tags with operation context for deferred ACK
- **Timeout Handling**: Implement operation timeouts to prevent message locks

### Adding New Features

#### 1. Message Queue Integration
```go
// 1. Add routing configuration to configs/configuration.yaml
ORCHESTRATORM8:
  newservice:
    Queue: [exchange, queue_name, prefetch_count]
    Routing-keys: [routing.pattern.#]
    Consumer: [queue_name, consumer_name, autoack]  # autoack: true/false

// 2. Use the connection pool for message operations
poolManager := amqpM8.GetGlobalPoolManager()
pool := poolManager.GetPool()

// 3. Implement message consumer with manual acknowledgment (recommended)
func (s *NewService) ProcessMessage(msg amqp.Delivery) error {
    // IMPORTANT: In manual ACK mode, the handler should NOT call msg.Ack() directly
    // The orchestrator will handle ACK after scan completion

    // Process message
    if err := processLogic(msg.Body); err != nil {
        log8.BaseLogger.Error().Err(err).Msg("Processing failed")
        // Return error - the consumer will NACK with requeue automatically
        return err
    }

    // Success - return nil
    // The orchestrator will ACK via AckScanCompletion() after scan completes
    log8.BaseLogger.Debug().Msgf("Handler succeeded, deliveryTag: %d", msg.DeliveryTag)
    return nil
}

// 4. Acknowledge scan completion (called by scan endpoint after completion)
func (s *NewService) CompleteScan(deliveryTag uint64, scanCompleted bool) error {
    orchestrator, _ := orchestrator8.NewOrchestrator8()
    return orchestrator.AckScanCompletion(deliveryTag, scanCompleted)
}

// 5. For auto-ACK mode (legacy, not recommended for critical operations)
// Consumer automatically ACKs messages upon receipt - no manual ACK needed
```

#### 2. Database Operations
```go
// 1. Create interface file
type NewDAOInterface interface {
    GetAll() ([]model8.NewModel, error)
    Create(item model8.NewModel) error
}

// 2. Implement with proper resource management
func (dao *NewDAO) GetAll() ([]model8.NewModel, error) {
    query, err := dao.Db.Query("SELECT ...")
    if err != nil {
        return nil, fmt.Errorf("query failed: %w", err)
    }
    defer query.Close() // IMPORTANT: Always close resources
    
    // Process results...
}
```

#### 3. Configuration Changes
```yaml
# Add new service configuration section in configs/configuration.yaml
NEWSERVICE:
  setting1: "value1"
  setting2: 123

# Update orchestrator routing if needed
ORCHESTRATORM8:
  Services:
    newservice: "http://127.0.0.1:8005"

  # Configure service queue with manual ACK
  newservice:
    Queue: ["scheduler", "qnewservice", 1]  # [exchange, queue_name, prefetch_count]
    Routing-keys: ["scheduler.qnewservice.#"]
    Consumer: ["qnewservice", "cnewservice", "false"]  # [queue, consumer, autoack]
    # IMPORTANT: Set autoack to "false" for manual acknowledgment mode
    # Set autoack to "true" for automatic acknowledgment (legacy mode)

# RabbitMQ pool settings (if needed)
RabbitMQ:
  pool:
    max_connections: 10
    min_connections: 2
    max_idle_time: "5m"
    health_check_period: "30s"
```

### Logging Standards

#### Log Levels
```go
import "deifzar/reportingm8/pkg/log8"

// Use appropriate log levels
log8.BaseLogger.Debug().Msg("Detailed debugging information")
log8.BaseLogger.Info().Msg("General information")
log8.BaseLogger.Warn().Msg("Warning conditions")  
log8.BaseLogger.Error().Msg("Error conditions")
log8.BaseLogger.Fatal().Msg("Fatal conditions - exits program")
```

#### Structured Logging
```go
// Add context to logs
log8.BaseLogger.Info().
    Str("user_id", userID).
    Int("count", count).
    Msg("User operation completed")

// Log errors with stack traces
log8.BaseLogger.Error().
    Stack().
    Err(err).
    Msg("Operation failed")
```

### Error Handling Patterns

#### Standard Error Handling
```go
// Wrap errors with context
func (s *Service) DoOperation() error {
    result, err := s.dependency.SomeCall()
    if err != nil {
        return fmt.Errorf("operation failed: %w", err)
    }
    
    // Continue processing...
    return nil
}
```

#### Error Notification Integration
```go
import "deifzar/reportingm8/pkg/notification8"

// Critical errors should trigger notifications
func (s *Service) CriticalOperation() error {
    err := s.performOperation()
    if err != nil {
        log8.BaseLogger.Error().Err(err).Msg("Critical operation failed")
        notification8.Helper.PublishSysErrorNotification(
            "Critical operation failed", 
            "urgent", 
            "service-name",
        )
        return err
    }
    return nil
}
```

## Testing Guidelines

### Setting Up Tests (TODO - Not Currently Implemented)

The project currently has no tests. Here's the recommended testing approach:

#### 1. Unit Testing Structure
```bash
# Create test files alongside source files
pkg/
├── service/
│   ├── service.go
│   ├── service_test.go          # Unit tests
│   ├── service_interface.go
│   └── mocks/
│       └── service_mock.go      # Generated mocks
```

#### 2. Testing with Interfaces
```go
// Use interfaces for easy mocking
func TestService_DoSomething(t *testing.T) {
    // Create mock dependency
    mockDB := &mocks.DatabaseInterface{}
    mockDB.On("Query", mock.AnythingOfType("string")).Return(mockResult, nil)
    
    // Test the service
    service := NewService(mockDB)
    err := service.DoSomething()
    
    assert.NoError(t, err)
    mockDB.AssertExpectations(t)
}
```

#### 3. Integration Testing
```go
// Test with real dependencies in controlled environment
func TestIntegration_ReportGeneration(t *testing.T) {
    // Setup test database
    // Setup test message queue
    // Run end-to-end workflow
    // Verify results
}
```

### Recommended Testing Libraries
```bash
go get github.com/stretchr/testify/assert
go get github.com/stretchr/testify/mock
go get github.com/DATA-DOG/go-sqlmock
```

## Development Environment

### Required Services
```bash
# PostgreSQL (Docker example)
docker run --name reportingm8-db -p 5432:5432 \
  -e POSTGRES_DB=cptm8 \
  -e POSTGRES_USER=cpt_dbuser \
  -e POSTGRES_PASSWORD=secretpassword \
  postgres:12

# RabbitMQ (Docker example)
docker run --name reportingm8-mq -p 5672:5672 -p 15672:15672 \
  rabbitmq:3-management
```

### Environment Variables
```bash
# Override configuration values (recommended for production)
export APP_ENV=DEV
export LOG_LEVEL=0
export DATABASE_PASSWORD=dev_password
export RABBITMQ_PASSWORD=dev_password
export AWS_ACCESS_KEY_ID=your_key
export AWS_SECRET_ACCESS_KEY=your_secret
export SMTP_PASSWORD=smtp_password

# Note: Viper can load these automatically if configured
# Currently the system reads from configs/configuration.yaml
```

### IDE Configuration

#### VS Code Settings
```json
// .vscode/settings.json
{
    "go.lintTool": "golangci-lint",
    "go.formatTool": "goimports",
    "editor.formatOnSave": true,
    "go.useLanguageServer": true
}
```

#### Recommended Extensions
- Go (Google)
- golangci-lint
- Go Test Explorer

## Debugging

### Application Debugging
```bash
# Run with debug logging
APP_ENV=DEV LOG_LEVEL=0 go run main.go launch

# Debug specific package
go run main.go launch --verbose

# Check configuration parsing
go run main.go --config-check
```

### Database Debugging
```sql
-- Check table structure
\d+ cptm8vulnerability
\d+ "user"

-- Monitor query performance  
EXPLAIN ANALYZE SELECT * FROM cptm8vulnerability;
```

### Message Queue Debugging
```bash
# Check RabbitMQ management interface
# http://localhost:15672 (guest/guest)

# Monitor queue status
rabbitmqctl list_queues
rabbitmqctl list_exchanges
```

## Common Development Tasks

### Adding a New Database Table
1. Create migration script (TODO: add migration system)
2. Add model to `pkg/model8/`
3. Create DAO interface in `pkg/db8/`
4. Implement DAO with proper error handling
5. Add tests for DAO operations

### Modifying Configuration
1. Update `configs/configuration.yaml` structure
2. Update `configs/examples/configuration.yaml` template
3. Modify parsing code in `pkg/configparser/` if needed
4. Test with different APP_ENV values (DEV, TEST, PROD)
5. Note: Viper supports hot-reloading via fsnotify

### Adding New Message Types
1. Define message structure in `pkg/model8/`
2. Add routing configuration
3. Implement message handler
4. Add error handling and logging
5. Test message flow

## Performance Considerations

### Database Optimization
- Use prepared statements for repeated queries
- Implement connection pooling
- Add query performance monitoring
- Consider database indexing strategies

### Memory Management  
- Always close database query results
- Use context.Context for timeout management
- Monitor goroutine leaks
- Profile memory usage regularly

### Concurrency
- Use sync.Pool for object reuse
- Implement proper goroutine lifecycle management
- Add circuit breakers for external services
- Monitor blocking operations

## Production Deployment

### Build Process
```bash
# Build for production
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o reportingm8 main.go

# Create deployment package
tar -czf reportingm8-v1.0.0.tar.gz reportingm8 configuration.yaml assets/
```

### Environment Setup
1. Configure secure credential management
2. Set up proper logging infrastructure
3. Configure monitoring and alerting
4. Set up database backups
5. Configure load balancing if needed

### Health Checks (Partially Implemented)

**Current Endpoints** (port 8004):
```
GET /health - Kubernetes health check (basic)
GET /ready - Kubernetes readiness check (basic)
GET /details - Scheduler configuration details
POST /update - Update scheduler settings
```

**Recommended Additions**:
```
GET /health/db - Database connectivity check
GET /health/mq - RabbitMQ pool health check
GET /health/aws - AWS S3 connectivity check
GET /metrics - Prometheus metrics endpoint
```

**Using Health Checks**:
```bash
# Basic health check
curl http://localhost:8004/health

# Readiness check
curl http://localhost:8004/ready

# Scheduler details
curl http://localhost:8004/details
```

## Troubleshooting

### Common Issues

#### "Connection refused" errors
- Check PostgreSQL/RabbitMQ service status
- Verify configuration.yaml settings
- Check network connectivity and firewall rules

#### Template not found errors  
- Verify asset file paths in configuration
- Check file permissions
- Ensure assets directory is accessible

#### Message queue connection issues
- Verify RabbitMQ credentials and permissions
- Check queue and exchange configuration
- Monitor RabbitMQ logs for errors

### Log Analysis
```bash
# Monitor application logs
tail -f log/reportingm8.log

# Filter error messages
grep "ERROR" log/reportingm8.log

# Monitor specific component
grep "orchestrator8" log/reportingm8.log
```

## Contributing Guidelines

### Code Style
- Follow standard Go formatting (gofmt)
- Use meaningful variable and function names
- Add comprehensive error handling
- Include proper documentation comments
- Maintain interface-based design patterns

### Pull Request Process
1. Create feature branch from main
2. Implement changes with proper error handling
3. Add tests for new functionality
4. Update documentation
5. Ensure all quality checks pass
6. Submit PR with detailed description

### Code Review Checklist
- [ ] Proper error handling with context
- [ ] Resource cleanup (defer statements)
- [ ] Appropriate logging levels
- [ ] Interface usage maintained  
- [ ] Configuration changes documented
- [ ] Tests included for new features