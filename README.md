# Expense Sharing API

A RESTful API for managing shared expenses between groups of users. Built with Go and SQLite.

## Features

- **User Management**
  - User registration and login
  - JWT-based authentication
  - Secure password hashing
  
- **Group Management**
  - Create groups for shared expenses
  - Add/view group members
  - Multiple groups per user
  
- **Expense Management**
  - Add expenses with multiple split types:
    - Equal splits
    - Exact amount splits
    - Percentage-based splits
  - Track payments and settlements
  - View expense history
  
- **Balance Sheet**
  - View individual balances
  - Track who owes what to whom
  - Download balance reports

## Prerequisites

- Go 1.21 or higher
- SQLite3

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/expense-sharing-api
cd expense-sharing-api
```

2. Install dependencies:
```bash
go mod download
```

3. Run the application:
```bash
go run cmd/api/main.go
```

## API Endpoints

### Authentication
```bash
| Method | Path          | Description     |
|--------|---------------|-----------------|
| POST   | /api/register | Register user   |
| POST   | /api/login    | Login user      |
```
### Groups
```bash
| Method | Path             | Description         |
|--------|---------------   |---------------------|
| POST   | /api/groups      | Create group        |
| GET    | /api/groups      | Get user\'s groups  |
| GET    | /api/groups/{id} | Get group details   |
```
### Expenses
```bash
| Method | Path                      | Description            |
|--------|---------------------------|------------------------|
| POST   | /api/expenses             | Create expense         |
| GET    | /api/groups/{id}/expenses | Get group expenses     |
| GET    | /api/groups/{id}/balance  | Get balance sheet      |
```
## Usage Examples

### Register User
```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"email":"user@example.com", "full_name":"John Doe", "password":"password123"}' \
  http://localhost:8080/api/register
```

### Create Group
```bash
curl -X POST -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"name":"Roommates", "description":"Apartment expenses", "members":[1,2,3]}' \
  http://localhost:8080/api/groups
```

### Add Expense (Equal Split)
```bash
curl -X POST -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "group_id": 1,
    "description": "Dinner",
    "amount": 3000,
    "split_type": "EQUAL",
    "shares": [
      {"user_id": 1, "share_amount": 1000},
      {"user_id": 2, "share_amount": 1000},
      {"user_id": 3, "share_amount": 1000}
    ]
  }' \
  http://localhost:8080/api/expenses
```

### Get Balance Sheet
```bash
curl -H "Authorization: Bearer <token>" \
  http://localhost:8080/api/groups/1/balance
```

## Database Schema

```sql
-- Users table
CREATE TABLE users (
    user_id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    full_name TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Groups table
CREATE TABLE groups (
    group_id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    created_by INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (created_by) REFERENCES users(user_id)
);

-- Group members junction table
CREATE TABLE group_members (
    group_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (group_id, user_id),
    FOREIGN KEY (group_id) REFERENCES groups(group_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id)
);

-- Expenses table
CREATE TABLE expenses (
    expense_id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER NOT NULL,
    description TEXT NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    created_by INTEGER NOT NULL,
    split_type TEXT NOT NULL CHECK (split_type IN ('EQUAL', 'EXACT', 'PERCENTAGE')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES groups(group_id),
    FOREIGN KEY (created_by) REFERENCES users(user_id)
);

-- Expense shares table
CREATE TABLE expense_shares (
    expense_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    share_amount DECIMAL(10,2) NOT NULL,
    share_percentage DECIMAL(5,2),
    paid_amount DECIMAL(10,2) DEFAULT 0,
    PRIMARY KEY (expense_id, user_id),
    FOREIGN KEY (expense_id) REFERENCES expenses(expense_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id)
);

-- Settlements table
CREATE TABLE settlements (
    settlement_id INTEGER PRIMARY KEY AUTOINCREMENT,
    payer_id INTEGER NOT NULL,
    payee_id INTEGER NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    group_id INTEGER NOT NULL,
    settled_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    notes TEXT,
    FOREIGN KEY (payer_id) REFERENCES users(user_id),
    FOREIGN KEY (payee_id) REFERENCES users(user_id),
    FOREIGN KEY (group_id) REFERENCES groups(group_id)
);
```

## Error Handling

The API returns errors in the following format:
```json
{
    "success": false,
    "error": "Error message"
}
```

## Success Response

Successful responses are returned in the following format:
```json
{
    "success": true,
    "data": {
        // Response data
    }
}
```

## Security

- Passwords are hashed using bcrypt
- Authentication uses JWT tokens
- Protected routes require valid JWT token
- SQL injection prevention using prepared statements
- Input validation for all requests

## Future Improvements

- Add email verification
- Implement password reset functionality
- Add support for recurring expenses
- Add expense categories and tags
- Implement expense analytics and reports
- Add support for different currencies
- Implement push notifications
- Add support for expense attachments (receipts)
