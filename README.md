# Grafto - Full-Stack Web Development in Go

A modern, production-ready Go web application starter template that provides everything you need to build robust web applications quickly. Built with battle-tested technologies and modern development practices.

Created by MBV, creator of the [Golang Blog Course](https://golangblogcourse.com?utm_source=github&utm_campaign=grafto).

## Aim

The aim of Grafto is to be starter template that provides most of what you'll need to get a new web project off the 
ground, taking inspiration from Laravel equivalent. There are some opinionated choices made (like no ORM, old fashioned 
HTML rendered on the server) but tries to be as idiomatic as possible.

The target audience for the starter is mostly going to be solo-devs building side-projects or trying to bootstrap a 
new business.

## Features

- **Modern Go Web Stack**: Echo framework, PostgreSQL, templ templates
- **Complete Authentication**: Registration, login, email verification, password reset
- **Admin Dashboard**: User management with pagination and CRUD operations
- **Type-Safe Templates**: templ for compile-time template safety
- **Hot Reload Development**: Automatic recompilation and restart during development
- **Background Jobs**: River queue for async processing (email sending, etc.)
- **Full Observability**: OpenTelemetry integration with Grafana stack
- **Production Ready**: Docker containerization, security best practices

### Technology Stack

**Backend:**
- **Framework**: Echo v4 (high-performance HTTP framework)
- **Database**: PostgreSQL with pgx driver
- **Query Builder**: sqlc for type-safe SQL operations
- **Migrations**: goose for database schema management
- **Background Jobs**: River (PostgreSQL-backed job queue)
- **Sessions**: Gorilla Sessions with encryption

**Frontend:**
- **Templates**: templ (type-safe Go templates)
- **CSS**: Tailwind CSS v4 with Vite
- **JavaScript**: Alpine.js + HTMX for interactivity
- **Build Tools**: Vite for asset compilation

**Development & Operations:**
- **Task Runner**: Just (modern alternative to Make)
- **Hot Reload**: wgo for live development
- **Linting**: golangci-lint with 37+ enabled linters
- **Observability**: OpenTelemetry with Grafana, Prometheus, Loki, Tempo
- **Containerization**: Docker with multi-stage builds

### Project Structure

```
grafto/
├── cmd/                    # Application entry points
│   ├── app/               # Main web application
│   ├── worker/            # Background job processor
│   ├── email/             # Email service development
│   └── migration/         # Database migration tools
├── config/                # Configuration management
├── handlers/              # HTTP request handlers
│   └── middleware/        # Custom middleware
├── models/                # Domain models and database layer
│   ├── internal/db/       # sqlc generated code
│   └── seeds/             # Test data generation
├── psql/                  # Database layer
│   ├── migrations/        # SQL migrations
│   └── queue/             # Background job definitions
├── router/                # HTTP routing and contexts
├── services/              # Business logic layer
├── views/                 # HTML templates
│   ├── internal/
│   │   ├── components/    # Reusable UI components
│   │   └── layouts/       # Page layouts
│   └── dashboard/         # Admin dashboard pages
├── emails/                # Email templates
├── assets/                # Static assets (CSS, JS, images)
└── docker/               # Docker configurations and monitoring
```

## Quick Start

### Prerequisites

- **Go 1.24+**
- **PostgreSQL 12+**
- **Node.js 18+** (for CSS building)
- **Just** command runner ([installation guide](https://just.systems/man/en/chapter_4.html))

### Installation

1. **Clone the repository:**
```bash
git clone https://github.com/mbvlabs/grafto.git
cd grafto
```

2. **Install dependencies:**
```bash
# Install Go dependencies
go mod download

# Install Node.js dependencies
npm install

# Install required Go tools
just install-tools
```

3. **Set up environment:**
```bash
# Copy environment template
cp .env.example .env

# Edit .env with your database credentials
# At minimum, configure:
# - Database connection (DB_HOST, DB_USER, DB_PASSWORD, DB_NAME)
# - Session keys (generate secure random strings)
# - Email settings (for AWS SES integration)
```

4. **Set up database:**
```bash
# Run database migrations
just up-migrations

# (Optional) Seed with test data
just seed
```

5. **Start development server:**
```bash
# Start with hot reload
just run

# Or start without hot reload
just run-app
```

6. **Access the application:**
- **Web Interface**: http://localhost:8080
- **Register**: http://localhost:8080/registrations/new
- **Login**: http://localhost:8080/sessions/new

## Development Workflow

### Daily Development Commands

```bash
# Primary development with hot reload
just run

# Run specific services
just run-worker          # Background job processor
just run-email          # Email template development
just run-components     # UI component development

# Database operations
just create-migration name    # Create new migration
just up-migrations           # Apply pending migrations
just migration-status        # Check migration status
just reset-db               # Reset database (destroys data)

# Code quality
just golangci               # Run linting
just test-units            # Run unit tests
just test-integrations     # Run integration tests
just test-all              # Run all tests

# Template and asset compilation
just compile-templates     # Compile templ templates
just fmt-templates        # Format templates
npm run build-css         # Build Tailwind CSS
```

### Hot Reload Development

The `just run` command provides sophisticated hot reloading:
- Watches `.go` and `.templ` files for changes
- Automatically recompiles templates and CSS
- Restarts the application instantly
- Excludes generated files from watch

### Testing Strategy

- **Unit Tests**: Fast tests without database (`just test-units`)
- **Integration Tests**: Database-backed tests with embedded PostgreSQL (`just test-integrations`)
- **E2E Tests**: Full application flow testing (`just test-e2e`)

## Configuration

### Environment Variables

Create a `.env` file with the following required variables:

```bash
# Application
ENVIRONMENT=development
SERVER_HOST=localhost
SERVER_PORT=8080
PROJECT_NAME=YourProjectName

# Database
DB_HOST=127.0.0.1
DB_PORT=5432
DB_NAME=grafto_dev
DB_USER=your_db_user
DB_PASSWORD=your_db_password

# Security (generate secure random strings)
PASSWORD_SALT=your-secure-salt-here
SESSION_KEY=your-session-key-32-chars-long
SESSION_ENCRYPTION_KEY=your-encryption-key-32-chars
TOKEN_SIGNING_KEY=your-token-signing-key-here

# Email (AWS SES)
AWS_ACCESS_KEY_ID=your-aws-access-key
AWS_SECRET_ACCESS_KEY=your-aws-secret-key
DEFAULT_SENDER_SIGNATURE=your-email@example.com
```

For complete configuration documentation, see [docs/configuration.md](docs/configuration.md).

## Production Deployment

### Docker Deployment

```bash
# Build the application
docker build -t grafto .

# Run with environment variables
docker run -d \
  --name grafto \
  -p 8080:8080 \
  --env-file .env.production \
  grafto
```

### Monitoring Stack

The project includes a complete observability stack:

```bash
# Start monitoring services
cd docker
docker-compose up -d

# Access services
# Grafana: http://localhost:3001 (admin/admin)
# Prometheus: http://localhost:9091
```

For complete deployment documentation, see [PRODUCTION_DEPLOYMENT.md](PRODUCTION_DEPLOYMENT.md).

## Key Features Explained

### Authentication System

- **Registration** with email verification
- **Login/Logout** with session management
- **Password reset** flow with secure tokens
- **Admin user** capabilities
- **Session security** with encryption and CSRF protection

### Admin Dashboard

- **User management** with pagination
- **CRUD operations** for user accounts
- **Admin privilege** management
- **Background job monitoring** with River UI

### Background Jobs

- **Email sending** (welcome emails, password resets)
- **Cleanup tasks** (expired tokens, old sessions)
- **PostgreSQL-backed** job queue with River
- **Retry logic** and error handling
- **Web UI** for job monitoring

### Type-Safe Templates

- **templ** for compile-time template safety
- **Component-based** UI architecture
- **Layout system** for consistent design
- **HTMX integration** for dynamic interactions

### Database Layer

- **PostgreSQL** with connection pooling
- **sqlc** for type-safe SQL operations
- **goose** for database migrations
- **Embedded PostgreSQL** for testing
- **Transaction support** for data consistency

## Testing

### Running Tests

```bash
# Run all tests
just test-all

# Run specific test suites
just test-units        # Unit tests only
just test-integrations # Integration tests only
just test-e2e         # End-to-end tests only
```

### Test Structure

Tests are organized by type using build tags:
- `//go:build unit` - Fast unit tests
- `//go:build integration` - Database integration tests
- `//go:build e2e` - Full application tests

### Integration Testing

Uses embedded PostgreSQL for consistent, isolated testing:
- No external database required
- Automatic test database creation/cleanup
- Fast parallel test execution

## 🔍 Observability

### Telemetry Stack

The application includes comprehensive observability:

- **Tracing**: Distributed tracing with Tempo
- **Metrics**: Prometheus metrics collection
- **Logging**: Structured logging with Loki
- **Dashboards**: Pre-built Grafana dashboards

### Monitoring Development

```bash
# Start telemetry stack
cd docker
docker-compose up -d

# View application metrics
open http://localhost:3001  # Grafana
open http://localhost:9091  # Prometheus
```

## Contributing

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Set up development environment**: `just run`
4. **Make your changes** using hot reload
5. **Write tests**: `just test-all`
6. **Run quality checks**: `just golangci`
7. **Commit your changes**: `git commit -m 'Add amazing feature'`
8. **Push to branch**: `git push origin feature/amazing-feature`
9. **Create Pull Request**

### Development Guidelines

- Follow existing code patterns and architecture
- Write tests for new features
- Use `just golangci` for code quality
- No comments in code (code should be self-documenting)
- Prefer editing existing files over creating new ones

## Alternatives

If you're looking for alternatives, consider:
- **[Pagoda](https://github.com/mikestefanello/pagoda)** - More feature-complete Go web starter
- **[Buffalo](https://gobuffalo.io/)** - Full-stack Go web framework

Grafto focuses on simplicity, type safety, and modern development practices while staying close to Go's standard library and idioms.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- **[Echo](https://echo.labstack.com/)** - High performance HTTP framework
- **[templ](https://templ.guide/)** - Type-safe Go templates
- **[sqlc](https://sqlc.dev/)** - Generate type-safe code from SQL
- **[River](https://riverqueue.com/)** - Fast and reliable background jobs
- **[Tailwind CSS](https://tailwindcss.com/)** - Utility-first CSS framework

For questions, feedback, or support, visit the [GitHub Issues](https://github.com/mbvlabs/grafto/issues).
