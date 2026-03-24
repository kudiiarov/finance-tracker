# Finance Tracker - Product Requirements Document

## Overview
A production-grade personal finance management application with multi-user support, 
RESTful API, and comprehensive financial tracking capabilities.

## Tech Stack
- **Language**: Go 1.22+
- **Database**: PostgreSQL 15+
- **API**: RESTful HTTP with Chi router
- **Authentication**: JWT (JSON Web Tokens)
- **Migrations**: golang-migrate
- **Configuration**: Environment variables + YAML
- **Logging**: Structured logging (slog)
- **Testing**: Unit + Integration tests
- **Documentation**: OpenAPI/Swagger

## Architecture
Following **Clean Architecture** / **Hexagonal Architecture**:
```
┌─────────────────┐
│   HTTP/API      │ ← Chi router, middleware, handlers
│   (delivery)    │
├─────────────────┤
│   Use Cases     │ ← Business logic, services
│   (domain)      │
├─────────────────┤
│   Repository    │ ← Data access layer
│   (storage)     │
├─────────────────┤
│   Database      │ ← PostgreSQL, migrations
│   (infrastructure)│
└─────────────────┘
```

## Core Features

### 1. User Management
- **Registration**: Email, password (hashed with bcrypt)
- **Login**: JWT token generation (access + refresh tokens)
- **Profile**: View/update profile, change password
- **Security**: Password validation, rate limiting

### 2. Accounts
- **Account Types**: Checking, Savings, Credit Card, Cash, Investment, Loan, Other
- **CRUD Operations**: Create, read, update, delete accounts
- **Balance Tracking**: Current balance, initial balance
- **Currency Support**: Multi-currency with exchange rates
- **Account Status**: Active, archived

### 3. Categories
- **Category Types**: Income, Expense
- **Hierarchical**: Parent categories with sub-categories
- **System Categories**: Pre-defined common categories
- **User Categories**: Custom user-defined categories
- **Budget Association**: Link categories to budgets

### 4. Transactions
- **Transaction Types**: Income, Expense, Transfer
- **Fields**: Date, amount, description, category, account, tags, notes
- **Transfer Support**: Between accounts with automatic dual-entry
- **Recurring**: Schedule recurring transactions
- **Import**: CSV/Excel import with mapping
- **Search & Filter**: By date range, category, account, amount, tags
- **Pagination**: Cursor-based pagination
- **Bulk Operations**: Bulk edit, bulk delete, bulk categorization

### 5. Budgets
- **Budget Periods**: Monthly, yearly, custom date ranges
- **Budget Types**: By category, by account, overall
- **Allocation**: Set limits per category/account
- **Progress Tracking**: Actual vs budgeted amounts
- **Alerts**: Threshold warnings (75%, 90%, 100%)

### 6. Reports & Analytics
- **Summary Dashboard**: Total balance, monthly income/expense, net worth
- **Trend Analysis**: Income/expense over time (line charts)
- **Category Breakdown**: Pie charts for spending by category
- **Cash Flow**: Monthly cash flow analysis
- **Budget vs Actual**: Comparison reports
- **Export**: PDF, CSV, Excel export

### 7. Tags & Search
- **Tags**: Multiple tags per transaction
- **Full-text Search**: Search in descriptions and notes
- **Filters**: Combined filtering with AND/OR logic

## API Endpoints

### Authentication
```
POST   /api/v1/auth/register       # Register new user
POST   /api/v1/auth/login          # Login, get tokens
POST   /api/v1/auth/refresh        # Refresh access token
POST   /api/v1/auth/logout         # Logout (invalidate token)
GET    /api/v1/auth/me             # Get current user
PUT    /api/v1/auth/me             # Update profile
PUT    /api/v1/auth/password       # Change password
```

### Accounts
```
GET    /api/v1/accounts            # List accounts (with balance)
POST   /api/v1/accounts            # Create account
GET    /api/v1/accounts/{id}       # Get account details
PUT    /api/v1/accounts/{id}       # Update account
DELETE /api/v1/accounts/{id}       # Delete (soft delete)
GET    /api/v1/accounts/{id}/transactions  # Account transactions
```

### Categories
```
GET    /api/v1/categories          # List categories (tree structure)
POST   /api/v1/categories          # Create category
GET    /api/v1/categories/{id}     # Get category
PUT    /api/v1/categories/{id}     # Update category
DELETE /api/v1/categories/{id}     # Delete category
```

### Transactions
```
GET    /api/v1/transactions        # List transactions (paginated, filterable)
POST   /api/v1/transactions        # Create transaction
GET    /api/v1/transactions/{id}   # Get transaction
PUT    /api/v1/transactions/{id}   # Update transaction
DELETE /api/v1/transactions/{id}   # Delete transaction
POST   /api/v1/transactions/bulk   # Bulk operations
POST   /api/v1/transactions/import # Import from CSV/Excel
```

### Budgets
```
GET    /api/v1/budgets             # List budgets
POST   /api/v1/budgets             # Create budget
GET    /api/v1/budgets/{id}        # Get budget with progress
PUT    /api/v1/budgets/{id}        # Update budget
DELETE /api/v1/budgets/{id}        # Delete budget
GET    /api/v1/budgets/{id}/status # Budget status/progress
```

### Reports
```
GET    /api/v1/reports/summary              # Dashboard summary
GET    /api/v1/reports/income-vs-expense    # Income vs expense over time
GET    /api/v1/reports/category-breakdown   # Spending by category
GET    /api/v1/reports/cash-flow            # Cash flow analysis
GET    /api/v1/reports/budget-comparison    # Budget vs actual
POST   /api/v1/reports/export               # Export report
```

## Database Schema (High-level)

### Tables
- `users` - User accounts
- `accounts` - Financial accounts
- `categories` - Transaction categories (hierarchical)
- `transactions` - Financial transactions
- `transfers` - Transfer records linking transactions
- `budgets` - Budget definitions
- `budget_allocations` - Category/account allocations for budgets
- `tags` - User-defined tags
- `transaction_tags` - Many-to-many relationship
- `refresh_tokens` - JWT refresh token storage
- `migrations` - Database migration tracking

## Non-Functional Requirements

### Performance
- API response time < 200ms (p95)
- Support 1000+ concurrent users
- Handle 10,000+ transactions per user efficiently

### Security
- HTTPS only
- Password hashing with bcrypt (cost 12)
- JWT with secure signing (RS256)
- SQL injection prevention (parameterized queries)
- Input validation and sanitization
- Rate limiting on auth endpoints

### Reliability
- Graceful error handling
- Structured logging with request IDs
- Database connection pooling
- Health check endpoints
- Graceful shutdown

### Maintainability
- Clean code with clear separation of concerns
- Comprehensive unit tests (>80% coverage)
- Integration tests for API endpoints
- API documentation (OpenAPI 3.0)
- Code comments and documentation

## Development Phases

### Phase 1: Foundation
1. Project setup with Clean Architecture
2. Database setup with migrations
3. Configuration management
4. Logging and error handling
5. Health check endpoints

### Phase 2: Core Features
1. User authentication (JWT)
2. Account management
3. Category management
4. Transaction CRUD with filtering

### Phase 3: Advanced Features
1. Transfer between accounts
2. Budget management
3. Reports and analytics
4. Export functionality

### Phase 4: Polish
1. Testing (unit + integration)
2. Docker setup
3. Documentation
4. Performance optimization

## Success Criteria
- All API endpoints documented and functional
- Clean, maintainable code following Go best practices
- Comprehensive test coverage
- Production-ready configuration
- Complete user flow from registration to reports
