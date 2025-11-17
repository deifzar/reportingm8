# System Architecture

## Overview

ReportingM8 is a Go-based security vulnerability reporting system that follows a message-driven, microservice-oriented architecture. The system generates periodic vulnerability reports, manages user notifications, and integrates with external services like AWS S3 for storage and SMTP for email delivery.

## Architectural Patterns

### Interface-Based Design
The system heavily uses Go interfaces to define contracts between components:
- Every major service has a corresponding `*_interface.go` file
- Enables dependency injection and testability
- Promotes loose coupling between layers

### Message-Driven Architecture
- **RabbitMQ**: Central message broker for asynchronous communication
- **Topic Exchanges**: Pattern-based message routing
- **Service Orchestration**: Central orchestrator coordinates message flow
- **Consumer Pattern**: Background workers process queued tasks

### Layered Architecture
```
┌─────────────────────────────────────────────────┐
│                    API Layer                    │
│                   (api8/)                       │
├─────────────────────────────────────────────────┤
│                Business Logic                   │
│            (reporting8/, orchestrator8/)        │
├─────────────────────────────────────────────────┤
│               Service Layer                     │
│     (email8/, cloud8/, notification8/)         │
├─────────────────────────────────────────────────┤
│                Data Access                      │
│                   (db8/)                        │
├─────────────────────────────────────────────────┤
│              Infrastructure                     │
│         (amqpM8/, log8/, gocron8/)             │
└─────────────────────────────────────────────────┘
```

## Core Components

### orchestrator8/
**Purpose**: Central service coordinator and message queue manager
- Initializes RabbitMQ exchanges and queues
- Routes messages between services
- Manages service health and error notifications
- **Manual Acknowledgment Support**:
  - `AckScanCompletion()` - Acknowledges messages after scan completion verification
  - `NackScanMessage()` - Rejects messages with configurable requeue behavior
  - Delivery tag tracking via HTTP headers (`X-RabbitMQ-Delivery-Tag`, `X-RabbitMQ-Consumer-Tag`)
  - Deferred acknowledgment pattern for long-running operations
- **Key Files**: `orchestrator8.go`, `orchestrator8_interface.go`

### reporting8/
**Purpose**: Core business logic for report generation
- Generates vulnerability reports from database data
- Creates email summaries and notifications
- Integrates with cloud storage and email services
- **Key Files**: `reportingm8.go`, `reportingm8_interface.go`

### db8/
**Purpose**: Data access layer with domain-specific interfaces
- **Domain-specific DAOs**: Reports, vulnerabilities, users, hosts, domains
- **Database Models**: Clean separation of data models in `model8/`
- **Query Patterns**: Prepared statements and parameterized queries
- **Key Files**: `db8_*.go` and corresponding interfaces

### amqpM8/
**Purpose**: RabbitMQ connection pooling and abstraction layer
- **Connection Pool Management**: Thread-safe pooling with configurable min/max connections
- **Health Monitoring**: Per-consumer health tracking with error counts and restart tracking
- **Auto-Reconnection**: Context-aware consumer reconnection logic
- **Metrics Tracking**: Active, idle, healthy connection counts; total borrowed/returned statistics
- **Shared State**: Global AMQP configuration state across connections
- **Manual Acknowledgment Support**:
  - Configurable auto-ACK vs. manual ACK mode per consumer
  - Intelligent NACK handling: handler errors trigger requeue, missing handlers reject permanently
  - Channel access via `GetChannel()` for direct ACK/NACK operations
  - Deferred acknowledgment after scan completion in manual mode
- **Key Files**:
  - `connection_pool.go` - Core connection pool implementation
  - `pool_manager.go` - Singleton pool manager
  - `pooled_amqp.go` - High-level pooled operations with ACK/NACK logic
  - `initialization.go` - Pool initialization
  - `shared_state.go` - Global state management

### gocron8/
**Purpose**: Scheduled task management
- Uses `go-co-op/gocron/v2` for job scheduling
- Timezone-aware scheduling
- Integration with orchestrator for error handling
- **Key Files**: `gocron8.go`

### email8/, cloud8/, notification8/, cleanup8/, controller8/
**Purpose**: Supporting services and external integrations
- **email8**: SMTP integration for email delivery
- **cloud8**: AWS S3 integration for report storage with upload capabilities
- **notification8**: Multi-channel notification publishing via RabbitMQ (app, email, SMS)
- **cleanup8**: Temporary file management and cleanup based on age
- **controller8**: HTTP controllers for scheduler management and health checks

## Data Flow

### Report Generation Flow
```
┌─────────────┐    ┌──────────────┐    ┌─────────────┐
│   Scheduler │────│ Orchestrator │────│  Reporting  │
│  (gocron8)  │    │ (orchestr8)  │    │ (reporting8)│
└─────────────┘    └──────────────┘    └─────────────┘
                           │                    │
                           ▼                    ▼
┌─────────────┐    ┌──────────────┐    ┌─────────────┐
│   RabbitMQ  │    │   Database   │    │  AWS S3     │
│  (amqpM8)   │    │    (db8)     │    │  (cloud8)   │
└─────────────┘    └──────────────┘    └─────────────┘
                           │                    │
                           ▼                    ▼
                   ┌──────────────┐    ┌─────────────┐
                   │    Users     │    │   Email     │
                   │   Models     │    │  (email8)   │
                   └──────────────┘    └─────────────┘
```

### Message Queue Architecture
- **Exchange**: `scheduler` (topic type)
- **Queue**: `qreportingm8`
- **Routing Key Pattern**: `scheduler.qreportingm8.#`
- **Consumer**: `creportingm8` (configurable auto/manual acknowledgment)
- **Acknowledgment Modes**:
  - **Auto-ACK**: Messages are acknowledged immediately upon receipt (legacy mode)
  - **Manual ACK**: Messages are acknowledged only after successful processing and scan completion
  - **NACK with Requeue**: Failed messages are rejected and requeued for retry
  - **NACK without Requeue**: Permanently failed messages are rejected (sent to DLQ if configured)

## Configuration Management

### Centralized Configuration
- **File**: `configuration.yaml` (located in configs/ directory)
- **Parser**: Viper-based configuration management with hot-reloading support
- **Environment Support**: DEV, TEST, PROD environments
- **Service-specific sections**: Database, RabbitMQ, AWS, SMTP, Templates, Scheduling
- **Auto-creation**: Creates configs directory if it doesn't exist

### Configuration Structure
```yaml
APP_ENV: DEV|TEST|PROD
LOG_LEVEL: "0"  # 0=DEBUG, 1=INFO, 2=WARN, 3=ERROR, etc.

ORCHESTRATORM8:
  Services: # Service endpoint URLs
  Exchanges:  # RabbitMQ exchange definitions (type: topic)
    scheduler: "topic"
  reportingm8: # Service-specific configuration
    Queue: [exchange_name, queue_name, prefetch_count]
    Routing-keys: [routing.pattern.#]
    Consumer: [queue_name, consumer_name, auto_ack]

REPORTINGM8:
  Template:
    emailsummary: "path/to/template.html"
    emailnotificationreport: "path/to/notification.html"
    report: "path/to/report.html"
    clientorganisation: "Company Name"
    dashboard: "https://dashboard.url"
  SMTP:
    server: "smtp.example.com"
    port: 587
    username: "smtp_user"
    password: "smtp_pass"
    emailsender: "sender@example.com"
  Schedule:
    report: monthly|quarterly
    email: weekly|monthly
  Timezone: "timezone_string"

Database:
  location: "localhost"
  port: 5432
  schema: "public"
  database: "dbname"
  username: "db_user"
  password: "db_pass"

RabbitMQ:
  location: "localhost"
  port: 5672
  username: "mq_user"
  password: "mq_pass"
  pool:  # Connection pool configuration
    max_connections: 10
    min_connections: 2
    max_idle_time: "5m"
    max_lifetime: "30m"
    health_check_period: "30s"
    connection_timeout: "10s"
    retry_attempts: 3
    retry_delay: "2s"

Cloud:
  provider: AWS  # AWS, Azure, GCP (only AWS implemented)
  AWS:
    region: "us-east-1"
    key: "aws_access_key"
    secret: "aws_secret_key"
    bucket: "bucket-name"
```

## Security Architecture

### Authentication & Authorization
- User role-based access (`user`, `admin`)
- Database-stored user credentials
- Role-based report access controls

### Data Security
- Parameterized database queries prevent SQL injection
- Configuration contains sensitive credentials (requires secure management)
- AWS IAM-based cloud access control

## Deployment Architecture

### Service Dependencies
```
ReportingM8 Service
├── PostgreSQL Database
├── RabbitMQ Message Broker
├── AWS S3 (Report Storage)
└── SMTP Server (Email Notifications)
```

### Logging and Monitoring
- **Logger**: Zerolog-based structured logging with Lumberjack rotation
- **Log Files**:
  - Maximum 100MB per file
  - 3 backup files retained
  - Custom 0640 permissions (overrides Lumberjack default 0600)
  - Dual output in DEV (console + file), file-only in PROD
- **Log Levels**: Configurable (0=DEBUG to 7=PANIC)
- **Git Revision**: Embedded in log metadata for version tracking
- **Error Handling**: Centralized error notifications through message queue
- **Health Monitoring**:
  - Service orchestrator tracks component health
  - RabbitMQ connection pool health checks
  - Per-consumer health tracking (error counts, restarts, last-seen)

## Scalability Considerations

### Horizontal Scaling
- Message-driven design supports multiple service instances
- Database connection pooling
- Stateless service design

### Performance Patterns
- Asynchronous processing through message queues
- Background job processing for report generation
- Database query optimization opportunities identified

## Technology Stack

### Core Technologies
- **Language**: Go 1.24+
- **Web Framework**: Gin
- **Database**: PostgreSQL with `lib/pq` driver
- **Message Broker**: RabbitMQ with `amqp091-go`
- **Cloud Storage**: AWS S3 SDK v2
- **Scheduler**: go-co-op/gocron/v2
- **Configuration**: Viper
- **Logging**: Zerolog
- **CLI**: Cobra

### External Dependencies
- **UUID Generation**: gofrs/uuid/v5
- **Template Engine**: Go's html/template and text/template
- **Log Rotation**: natefinch/lumberjack

## Current API Endpoints

The system exposes the following HTTP endpoints (default port 8004):

- `GET /health` - Kubernetes health check endpoint
- `GET /ready` - Kubernetes readiness check endpoint
- `GET /details` - Retrieve current scheduler configuration
- `POST /update` - Update scheduler settings (document and email frequencies)

## File System Security

The application implements security hardening at the file system level:

- **Umask 0027**: Set at application startup for secure file creation
  - Files created with 0640 permissions (owner read/write, group read)
  - Directories created with 0750 permissions (owner rwx, group rx)
- **Log File Permissions**: Custom 0640 permissions for log rotation
- **Purpose**: Allows monitoring tools (vector, fluentd) to read logs while maintaining security

## Singleton Patterns

The system uses singleton patterns for shared resources:

1. **gocron8.BaseReportScheduler**: Global job scheduler instance
2. **log8.BaseLogger**: Global logger instance
3. **amqpM8 Pool Manager**: Global RabbitMQ connection pool manager via `GetGlobalPoolManager()`
4. **notification8.NotificationPoolHelper**: Static notification helper

All singletons use `sync.Once` for thread-safe initialization.

## Future Architecture Enhancements

### Recommended Improvements
1. **Metrics**: Implement Prometheus metrics collection for performance monitoring
2. **Circuit Breakers**: Add resilience patterns for external services
3. **API Gateway**: Consider API gateway for service routing
4. **Container Deployment**: Docker/Kubernetes manifests and Helm charts
5. **Configuration Management**: External configuration management (Consul, etcd) or Kubernetes ConfigMaps
6. **Service Discovery**: Dynamic service registration and discovery
7. **Database Migration System**: Versioned schema migrations
8. **API Authentication**: JWT-based authentication and authorization
9. **Rate Limiting**: API endpoint rate limiting
10. **Caching Layer**: Redis for frequently accessed data