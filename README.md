# Grafto - Full-Stack Web Development in Go

A modern, production-ready Go web application starter template that provides everything you need to build robust web applications quickly. Built with battle-tested technologies and modern development practices.

Created by MBV, creator of the [Golang Blog Course](https://golangblogcourse.com?utm_source=github&utm_campaign=grafto).

## What is Grafto?

Grafto is a comprehensive starter template designed to help solo developers and small teams bootstrap new web projects quickly. Taking inspiration from Laravel's approach to web development, Grafto provides opinionated but idiomatic Go solutions for common web application needs.

**Key Philosophy:**
- No ORM - Direct SQL with type safety via sqlc
- Server-side rendering with modern templating
- Production-ready from day one
- Comprehensive tooling and observability
- Security best practices built-in

## Quick Start

Get up and running in under 5 minutes:

### Prerequisites

- Go 1.24+ installed
- PostgreSQL running locally
- Node.js (for CSS building)
- [Just](https://github.com/casey/just) task runner

### Setup Steps

1. **Clone and customize the project:**
   ```bash
   git clone https://github.com/mbvlabs/grafto.git your-project-name
   cd your-project-name
   ./rename.sh  # Follow prompts to customize project name
   ```

2. **Configure environment:**
   ```bash
   cp .env.example .env
   # Edit .env with your database credentials and other settings
   ```

3. **Set up database:**
   ```bash
   # Create your PostgreSQL database first, then:
   just up-migrations  # Apply database migrations
   just seed          # Add sample data (optional)
   ```

4. **Install dependencies and run:**
   ```bash
   npm install        # Install frontend dependencies
   just run          # Start with live reload
   ```

Visit `http://localhost:8080` to see your application running.

## Technical Stack

### Backend
- **Framework:** Echo (high-performance HTTP router)
- **Database:** PostgreSQL with sqlc for type-safe queries
- **Background Jobs:** River queue system
- **Authentication:** Session-based with secure defaults
- **Configuration:** Environment-based with validation

### Frontend
- **Templates:** templ (type-safe Go templates)
- **Styling:** Tailwind CSS v4
- **JavaScript:** Alpine.js and HTMX
- **Build System:** Vite for CSS processing

### Infrastructure
- **Observability:** OpenTelemetry with Prometheus, Grafana, Loki
- **Docker:** Complete containerization setup
- **Development:** Live reload with comprehensive tooling
- **Testing:** Unit, integration, and e2e test support

## Project Structure

```
grafto/
├── assets/           # Static assets (CSS, JS, images)
├── cmd/             # Application entry points
│   ├── app/         # Main web application
│   ├── worker/      # Background job worker
│   ├── email/       # Email service
│   └── seed/        # Database seeding
├── config/          # Configuration management
├── handlers/        # HTTP request handlers
│   └── middleware/  # Custom middleware
├── models/          # Domain models and validation
│   ├── internal/db/ # Generated sqlc code
│   └── seeds/       # Seed data definitions
├── psql/            # Database related code
│   ├── migrations/  # Database migrations
│   └── queue/       # River job definitions
├── router/          # Route definitions and contexts
├── services/        # Business logic layer
├── views/           # templ templates
│   ├── internal/    # Reusable components and layouts
│   └── dashboard/   # Dashboard-specific views
├── emails/          # Email templates
├── telemetry/       # Observability setup
└── docker/          # Docker configuration
```

### Key Directories Explained

**`cmd/`** - Contains all application entry points. Each subdirectory is a separate executable:
- `app/` - Main web server
- `worker/` - Background job processor
- `email/` - Email service for development
- `seed/` - Database seeding utility

**`handlers/`** - HTTP request handlers that process incoming requests and return responses
- Organized by feature area
- Includes comprehensive middleware stack
- Handles authentication, rate limiting, logging

**`models/`** - Domain models containing business logic and validation
- `internal/db/` contains sqlc-generated type-safe database code
- Each model includes validation rules and business methods

**`services/`** - Business logic layer that orchestrates between handlers and models
- User management, authentication, registration flows
- Keeps handlers thin by encapsulating complex business rules

**`views/`** - templ templates for server-side rendering
- `internal/components/` - Reusable UI components
- `internal/layouts/` - Page layouts and structure
- Feature-specific view directories (dashboard, sessions, etc.)

## Available Development Tools

Grafto uses [Just](https://github.com/casey/just) as a task runner. All commands are defined in the `justfile`.

### Application Commands
```bash
just run           # Run with live reload (recommended for development)
just run-app       # Run without live reload
just run-worker    # Start background job worker
just run-email     # Email service with auto-reload
```

### Database Management
```bash
just create-migration name    # Create new migration file
just migration-status        # Show migration status
just up-migrations          # Apply all pending migrations
just down-migrations        # Rollback all migrations
just reset-db              # Reset database (danger!)
just generate-db-functions # Regenerate sqlc code
just seed                  # Populate database with test data
```

### Code Quality
```bash
just golangci    # Run golangci-lint (alias: just ci)
just vet         # Run go vet
```

### Testing
```bash
just test-units         # Unit tests only (alias: just tu)
just test-integrations  # Integration tests (alias: just ti)
just test-e2e          # End-to-end tests
just test-all          # All tests
```

### Templates and Assets
```bash
just compile-templates  # Generate Go code from templ files
just fmt-templates     # Format templ templates
```

### Aliases
Common commands have short aliases:
- `just r` → `just run`
- `just ci` → `just golangci`
- `just um` → `just up-migrations`
- `just cm name` → `just create-migration name`

## Architecture Overview

### Key Components

**Session Management**
- Secure session handling with gorilla/sessions
- Automatic CSRF protection
- Flash message support for user feedback

**Background Jobs**
- River queue system for reliable job processing
- Email sending, cleanup tasks, and custom jobs
- Retry logic and job monitoring

**Type-Safe Database**
- sqlc generates type-safe Go code from SQL
- No ORM - direct SQL with full type safety
- Migration system using goose

**Observability**
- OpenTelemetry integration for metrics, traces, and logs
- Prometheus metrics collection
- Grafana dashboards included
- Request tracing and performance monitoring

## Database Management

### Migrations

Grafto uses [goose](https://github.com/pressly/goose) for database migrations:

```bash
# Create a new migration
just create-migration add_user_profile

# Check migration status
just migration-status

# Apply all pending migrations
just up-migrations

# Rollback one migration
just down-migrations
```

Migration files are stored in `psql/migrations/` and follow the naming pattern:
`YYYYMMDDHHMMSS_description.sql`

### Type-Safe Queries

Database queries are written in SQL and stored in `psql/*.sql` files. sqlc generates type-safe Go code:

```sql
-- psql/users.sql
-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (email, password_hash, created_at)
VALUES ($1, $2, $3)
RETURNING *;
```

After writing queries, regenerate the Go code:
```bash
just generate-db-functions
```

### Seeding Data

Seed data is defined in `models/seeds/` and can be loaded with:
```bash
just seed
```

Seeds are useful for:
- Development data
- Testing scenarios
- Demo content

## Development Workflow

### Daily Development

1. **Start development server:**
   ```bash
   just run  # Starts with live reload
   ```

2. **Make changes** to Go files, templates, or CSS - the server automatically reloads

3. **Database changes:**
   ```bash
   just create-migration your_change_name
   # Edit the migration file
   just up-migrations
   just generate-db-functions  # If you modified queries
   ```

4. **Run quality checks:**
   ```bash
   just ci    # golangci-lint
   just vet   # go vet
   ```

### Testing Strategy

Grafto includes comprehensive testing support:

- **Unit Tests** (`_test.go` files) - Test individual functions and methods
- **Integration Tests** - Test handler/service integration with real database
- **E2E Tests** - Test complete user workflows

Run tests with build tags:
```bash
just test-units         # -tags=unit
just test-integrations  # -tags=integration  
just test-e2e          # -tags=e2e
```

### Observability Stack

The included monitoring stack provides:

- **Prometheus** - Metrics collection
- **Grafana** - Dashboards and visualization  
- **Loki** - Log aggregation
- **Tempo** - Distributed tracing
- **Alertmanager** - Alert management

### Environment Configuration

Production environments require these key variables:

```bash
# Database
DB_HOST=your-db-host
DB_NAME=your-db-name
DB_USER=your-db-user
DB_PASSWORD=your-secure-password

# Security
PASSWORD_SALT=your-random-salt
SESSION_KEY=your-session-key
SESSION_ENCRYPTION_KEY=your-encryption-key
TOKEN_SIGNING_KEY=your-signing-key

# Application
APP_DOMAIN=yourdomain.com
APP_PROTOCOL=https
DEFAULT_SENDER_SIGNATURE=noreply@yourdomain.com
```


### Security Considerations

- Change all default keys and secrets
- Use HTTPS in production
- Enable proper firewall rules
- Regular security updates
- Monitor logs for suspicious activity

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and quality checks
5. Submit a pull request

## License

MIT License - see LICENSE file for details.

## Support

- **Issues:** [GitHub Issues](https://github.com/mbvlabs/grafto/issues)
- **Discussions:** [GitHub Discussions](https://github.com/mbvlabs/grafto/discussions)
- **Blog Course:** [Golang Blog Course](https://golangblogcourse.com)
