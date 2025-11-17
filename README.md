# ReportingM8 - Reporting Mate

<div align="center">

**Production-grade Go microservice for automated vulnerability reporting and distribution.**

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker)](dockerfile)
[![Status](https://img.shields.io/badge/status-active%20development-yellow)](https://github.com/yourusername/reportingm8)
[![License](https://img.shields.io/badge/license-Apache%202.0-green.svg)](LICENSE)

[Features](#key-features) • [Quick Start](#quick-start) • [Documentation](#documentation) • [Architecture](#architecture) • [API](#api-reference)

</div>

---

## Overview

ReportingM8 (Reporting Mate) is a production-grade Go microservice designed for automated vulnerability report generation and distribution. It provides a robust REST API for managing scheduled reports, email notifications, and multi-channel alert distribution with AWS S3 integration for report storage.

**Built for:**
- Security operations teams managing vulnerability reporting
- Compliance teams requiring periodic security assessments
- Penetration testing teams distributing findings
- Security consultants delivering client reports

### Key Features

- **Automated Report Generation**: Scheduled vulnerability reports with configurable frequencies (weekly, monthly, quarterly)
- **Multi-Channel Notifications**: Email, SMS, and in-app notifications for security alerts
- **Cloud Storage Integration**: AWS S3 integration for report archival and distribution
- **Asynchronous Processing**: RabbitMQ-based message queuing with advanced connection pooling
- **Reliable Message Processing**: Manual acknowledgment support with intelligent error handling
- **Flexible Scheduling**: Timezone-aware job scheduling with go-co-op/gocron/v2
- **Production Ready**: Health checks, graceful shutdown, and comprehensive logging

---

## Quick Start

### Prerequisites

- **Go** 1.24 or higher
- **PostgreSQL** 12+ (for data persistence)
- **RabbitMQ** 3.8+ (for message queuing)
- **AWS S3** account (for report storage)
- **SMTP server** (for email notifications)
- **Docker** (optional, for containerized deployment)

### Installation

#### Option 1: Build from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/reportingm8.git
cd reportingm8

# Install Go dependencies
go mod download

# Build the binary
go build -o reportingm8

# Run the service
./reportingm8 launch
```

#### Option 2: Docker

```bash
# Build the Docker image
docker build -t reportingm8:latest .

# Run the container
docker run -d \
  -p 8004:8004 \
  -e DATABASE_HOST=your-db-host \
  -e RABBITMQ_HOST=your-rabbitmq-host \
  -e AWS_ACCESS_KEY=your-aws-key \
  --name reportingm8 \
  reportingm8:latest
```

### Configuration

1. Copy the example configuration:
```bash
cp configs/configuration.yaml.example configs/configuration.yaml
```

2. Edit `configs/configuration.yaml` with your settings:
```yaml
APP_ENV: PROD
LOG_LEVEL: "1"  # 0=debug, 1=info, 2=warn, 3=error

Database:
  location: localhost
  port: 5432
  database: reportingm8
  username: report_dbuser
  password: ${DB_PASSWORD}  # Use environment variables for secrets

RabbitMQ:
  location: localhost
  port: 5672
  username: ${RABBITMQ_USER}
  password: ${RABBITMQ_PASS}
  pool:
    max_connections: 10
    min_connections: 2
    health_check_period: "30s"

Cloud:
  provider: AWS
  AWS:
    region: us-east-1
    key: ${AWS_ACCESS_KEY}
    secret: ${AWS_SECRET_KEY}
    bucket: your-reports-bucket

REPORTINGM8:
  SMTP:
    server: smtp.example.com
    port: 587
    username: ${SMTP_USER}
    password: ${SMTP_PASS}
    emailsender: reports@example.com
  Schedule:
    report: monthly  # weekly, monthly, quarterly
    email: weekly
  Timezone: "America/New_York"
```

3. Set environment variables for sensitive data:
```bash
export DB_PASSWORD="your-secure-password"
export RABBITMQ_USER="your-rabbitmq-user"
export RABBITMQ_PASS="your-rabbitmq-password"
export AWS_ACCESS_KEY="your-aws-access-key"
export AWS_SECRET_KEY="your-aws-secret-key"
export SMTP_USER="your-smtp-user"
export SMTP_PASS="your-smtp-password"
```

### Database Setup

```sql
-- Create database
CREATE DATABASE reportingm8;

-- Create tables (example schema)
CREATE TABLE cptm8report (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    company_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    file_path VARCHAR(500),
    status VARCHAR(50)
);

CREATE TABLE cptm8vulnerability (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id UUID REFERENCES cptm8report(id),
    title VARCHAR(255) NOT NULL,
    severity VARCHAR(50),
    description TEXT,
    impact TEXT,
    discovered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_vulnerability_report ON cptm8vulnerability(report_id);
CREATE INDEX idx_vulnerability_severity ON cptm8vulnerability(severity);
```

---

## Usage

### Basic Workflow

1. **Check service health:**
```bash
curl http://localhost:8004/health
```

2. **Check readiness (verifies DB + RabbitMQ):**
```bash
curl http://localhost:8004/ready
```

3. **Get current scheduler configuration:**
```bash
curl http://localhost:8004/details
```

**Example Response:**
```json
{
  "report_frequency": "monthly",
  "email_frequency": "weekly",
  "timezone": "America/New_York",
  "next_report_run": "2025-12-01T00:00:00Z",
  "next_email_run": "2025-11-24T09:00:00Z"
}
```

4. **Update scheduler settings:**
```bash
curl -X POST http://localhost:8004/update \
  -H "Content-Type: application/json" \
  -d '{
    "report_frequency": "quarterly",
    "email_frequency": "monthly"
  }'
```

### Scheduling Frequencies

#### Report Generation
```yaml
REPORTINGM8:
  Schedule:
    report: monthly  # weekly, monthly, or quarterly
```
- **Weekly**: Reports generated every Monday
- **Monthly**: Reports generated on the 1st of each month
- **Quarterly**: Reports generated on the 1st of Jan/Apr/Jul/Oct

#### Email Distribution
```yaml
REPORTINGM8:
  Schedule:
    email: weekly  # weekly or monthly
```
- **Weekly**: Email summaries sent every Monday
- **Monthly**: Email summaries sent on the 1st of each month

---

## API Reference

### Health Check Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Liveness probe (always 200 OK) |
| `GET` | `/ready` | Readiness probe (checks DB + RabbitMQ) |

**Readiness Response:**
```json
{
  "status": "ready",
  "database": "connected",
  "rabbitmq": "connected",
  "timestamp": "2025-11-17T12:34:56Z"
}
```

### Scheduler Management Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/details` | Retrieve current scheduler configuration |
| `POST` | `/update` | Update scheduler settings |

**Update Request:**
```json
{
  "report_frequency": "monthly",
  "email_frequency": "weekly"
}
```

## Architecture

### High-Level Overview

```
┌─────────────┐
│   Client    │───┐
│   (HTTP)    │   │
└─────────────┘   │
                  │    ┌─────────────┐    ┌─────────────┐
┌─────────────┐   ├───▶│ ReportingM8 │───▶│  Database   │
│  RabbitMQ   │   │    │  API(Gin/Go)│    │ (PostgreSQL)│
│  (Message   │───┘    └─────────────┘    └─────────────┘
│   Queue)    │                 │
└─────────────┘                 │
                 ┌──────────────┼──────────────┐
                 ▼              ▼              ▼
         ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
         │   Report    │ │    Email    │ │  AWS S3     │
         │ Generation  │ │Notification │ │   Storage   │
         │(reporting8) │ │  (email8)   │ │  (cloud8)   │
         └─────────────┘ └─────────────┘ └─────────────┘
```

### Report Generation Workflow

**Complete Report Pipeline:**

```
1. SCHEDULER TRIGGER
   ├─→ Cron job (gocron8) triggers at configured time
   └─→ Publish message to RabbitMQ queue

2. DATA AGGREGATION
   ├─→ Query vulnerabilities from database
   ├─→ Calculate security metrics and statistics
   ├─→ Group by severity, hostname, time period
   └─→ Prepare template data structures

3. REPORT GENERATION
   ├─→ Render HTML templates with vulnerability data
   ├─→ Generate PDF/HTML report files
   ├─→ Store metadata in database
   └─→ Upload to AWS S3 storage

4. NOTIFICATION DISTRIBUTION
   ├─→ Email summary to stakeholders
   ├─→ Multi-channel notifications (SMS, app push)
   └─→ Update delivery status in database

5. CLEANUP
   └─→ Remove temporary files based on age
       └─→ Log completion and error notifications
```

### Package Structure

```
reportingm8/
├── cmd/                    # CLI commands (Cobra)
│   ├── root.go            # Base command setup
│   ├── launch.go          # API service launcher
│   └── version.go         # Version information
├── pkg/                    # 14 packages, 73 Go files
│   ├── amqpM8/            # RabbitMQ connection pooling (5 files)
│   ├── api8/              # HTTP API routes and initialization
│   ├── cleanup8/          # Temporary file cleanup utilities
│   ├── cloud8/            # AWS S3 integration
│   ├── configparser/      # Configuration management (Viper)
│   ├── controller8/       # Business logic controllers
│   ├── db8/               # Database access layer (23 files, 8 interfaces)
│   ├── email8/            # SMTP email service
│   ├── gocron8/           # Job scheduler wrapper
│   ├── log8/              # Structured logging (zerolog)
│   ├── model8/            # Data models and domain entities (11 files)
│   ├── notification8/     # Notification system (multi-channel)
│   ├── orchestrator8/     # Service orchestration
│   ├── reporting8/        # Report generation logic
│   └── utils/             # Utility functions
├── configs/               # Configuration files
├── docs/                  # Comprehensive documentation
├── assets/                # HTML email and report templates
└── main.go                # Application entry point
```

### Key Components

- **[API Layer](pkg/api8/)** - Gin-based REST API
- **[Controllers](pkg/controller8/)** - Scheduler management logic
- **[Database Layer](pkg/db8/)** - PostgreSQL repository pattern (8 domain interfaces)
- **[Message Queue](pkg/amqpM8/)** - RabbitMQ with connection pooling (manual ACK support)
- **[Orchestrator](pkg/orchestrator8/)** - Service coordination and message routing
- **[Report Generator](pkg/reporting8/)** - Core report generation and template rendering
- **[Email Service](pkg/email8/)** - SMTP integration for notifications
- **[Cloud Storage](pkg/cloud8/)** - AWS S3 upload and management

For detailed architecture documentation, see [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

---

## Advanced Features

### RabbitMQ Integration

**Advanced Connection Pooling:**
- Configurable pool size (default: 2-10 connections)
- Automatic connection recovery on failures
- Periodic health checks (30-second intervals)
- Manual message acknowledgment with smart ACK/NACK logic
- Delivery tag tracking for message lifecycle management

**Message Flow:**
```
RabbitMQ → Consumer → Extract deliveryTag → HTTP Request
    ↓
ReportGeneration() → Extract from header → GenerateReport()
    ↓
Defer ACK/NACK → Success: ACK | Failure: NACK+requeue
```

### Error Handling

**Resilient Design:**
- Panic recovery with defer blocks
- Automatic requeue on report generation failures
- Error notifications via RabbitMQ and email
- Graceful degradation on service failures
- Centralized error logging with Zerolog

### Scheduled Tasks

- Configurable cron expressions for custom schedules
- Timezone-aware scheduling (configured per deployment)
- Event listeners for job completion and failure
- Dynamic job creation and deletion via API
- Singleton scheduler pattern for thread safety

---

## Documentation

Comprehensive documentation is available in the [docs/](docs/) directory:

| Document | Description |
|----------|-------------|
| [ARCHITECTURE.md](docs/ARCHITECTURE.md) | Detailed system architecture (20 KB) |
| [DEVELOPMENT.md](docs/DEVELOPMENT.md) | Development setup and guidelines |
| [SECURITY.md](docs/SECURITY.md) | Security best practices and hardening |
| [PERFORMANCE.md](docs/PERFORMANCE.md) | Performance optimization guide |
| [TODO.md](docs/TODO.md) | Known issues and production roadmap |
| [CODE_REVIEW.md](docs/CODE_REVIEW.md) | Code quality analysis and recommendations |

---

## Technology Stack

### Core Technologies
- **Language**: Go 1.24+
- **Web Framework**: Gin
- **Database**: PostgreSQL with `lib/pq` driver
- **Message Broker**: RabbitMQ with `amqp091-go`
- **Cloud Storage**: AWS S3 SDK v2
- **Scheduler**: go-co-op/gocron/v2
- **Configuration**: Viper with hot-reload support
- **Logging**: Zerolog with Lumberjack rotation
- **CLI**: Cobra

### External Dependencies
- **UUID Generation**: gofrs/uuid/v5, google/uuid
- **Template Engine**: Go's html/template and text/template
- **File Watching**: fsnotify for configuration hot-reload
- **Log Rotation**: natefinch/lumberjack with custom permissions

---

## Performance

**Typical Performance Metrics:**

- **Report Generation**: 2-10 minutes (depends on vulnerability count and template complexity)
- **Email Delivery**: 5-30 seconds per recipient batch
- **Database Queries**: Optimized with prepared statements and batch operations
- **S3 Upload**: Varies by file size and network conditions

**Resource Requirements:**

- **CPU**: 2+ cores recommended
- **Memory**: 2 GB minimum, 4 GB recommended
- **Storage**: 20 GB for application + logs + temporary report files
- **Network**: Stable connection for SMTP, AWS S3, and RabbitMQ

For optimization tips, see [docs/PERFORMANCE.md](docs/PERFORMANCE.md).

---

## Security Considerations

### Current Limitations

- **No authentication** on API endpoints (planned for v2.0)
- **Database credentials** in configuration file (use environment variables)
- **Limited input validation** (basic Gin binding only)
- **No rate limiting** on API endpoints
- **Plaintext SMTP credentials** in configuration

### Recommendations

1. **Use environment variables** for all secrets (DB, RabbitMQ, AWS, SMTP)
2. **Deploy behind API gateway** with authentication (Kong, Traefik, AWS API Gateway)
3. **Enable TLS/SSL** for all production connections
4. **Implement rate limiting** to prevent abuse
5. **Run as non-root user** in containers
6. **Encrypt data at rest** (database encryption, encrypted S3 buckets)
7. **Use secure credential management** (HashiCorp Vault, AWS Secrets Manager)

See [docs/SECURITY.md](docs/SECURITY.md) for comprehensive security guidelines.

---

## Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Follow Go best practices and existing code style
4. Add tests for new functionality (target: 80% coverage)
5. Update documentation as needed
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) for detailed development guidelines.

---

## Roadmap

### Version 1.x (Current)
- [x] Core report generation functionality
- [x] PostgreSQL persistence
- [x] RabbitMQ message queuing with connection pooling
- [x] AWS S3 integration
- [x] SMTP email notifications
- [x] Scheduled job management
- [x] Health check endpoints
- [x] Manual acknowledgment support

### Version 2.0 (Planned)
- [ ] JWT-based authentication
- [ ] Rate limiting and request throttling
- [ ] Unit and integration tests (target: 80% coverage)
- [ ] Docker containerization
- [ ] Kubernetes deployment manifests
- [ ] Prometheus metrics integration
- [ ] Database connection pooling
- [ ] Template caching for improved performance
- [ ] GraphQL API option
- [ ] Web dashboard for report management

See [docs/TODO.md](docs/TODO.md) for the complete roadmap and known issues.

---

## Troubleshooting

### Common Issues

**1. Database connection failures**
```bash
# Check PostgreSQL is running
systemctl status postgresql

# Verify connection settings in configs/configuration.yaml
# Ensure database and tables are created
```

**2. RabbitMQ connection errors**
```bash
# Check RabbitMQ status
systemctl status rabbitmq-server

# Verify credentials and port in configuration
# Check exchange and queue creation
```

**3. AWS S3 upload failures**
```bash
# Verify AWS credentials are set correctly
echo $AWS_ACCESS_KEY
echo $AWS_SECRET_KEY

# Check bucket permissions and region settings
# Ensure IAM user has s3:PutObject permission
```

**4. SMTP email delivery issues**
```bash
# Test SMTP connection
telnet smtp.example.com 587

# Verify SMTP credentials in configuration
# Check firewall rules for outbound port 587/465
```

**5. Permission errors**
```bash
# Ensure log directory is writable
chmod 755 log/

# Check file permissions for config files
chmod 640 configs/configuration.yaml

# Verify umask settings (should be 0027)
```

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## Acknowledgments

- [Go Team](https://go.dev/) for the excellent programming language
- [Gin Web Framework](https://github.com/gin-gonic/gin) for the HTTP router
- [RabbitMQ](https://www.rabbitmq.com/) for reliable message queuing
- [PostgreSQL](https://www.postgresql.org/) for robust data persistence
- [go-co-op/gocron](https://github.com/go-co-op/gocron) for flexible job scheduling

---

## Contact

For questions, issues, or feature requests, please open an issue on GitHub.

**Project Link:** [https://github.com/yourusername/reportingm8](https://github.com/yourusername/reportingm8)

---

<div align="center">

**Built with ❤️ for the security community**

[⬆ Back to Top](#reportingm8---reporting-mate)

</div>
