package config

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type DBConfig struct {
	DBPath string
}

func NewDBConfig() *DBConfig {
	return &DBConfig{
		DBPath: "expense_sharing.db",
	}
}

func (c *DBConfig) Connect() (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite3", c.DBPath)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	// Enable foreign keys
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return nil, fmt.Errorf("error enabling foreign keys: %v", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(1) // SQLite only supports one writer at a time
	db.SetMaxIdleConns(1)

	return db, nil
}

func (c *DBConfig) InitSchema(db *sqlx.DB) error {
	// Create tables one by one to better handle errors
	schemas := []string{
		`CREATE TABLE IF NOT EXISTS users (
            user_id INTEGER PRIMARY KEY AUTOINCREMENT,
            email TEXT UNIQUE NOT NULL,
            full_name TEXT NOT NULL,
            password_hash TEXT NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );`,

		`CREATE TABLE IF NOT EXISTS groups (
            group_id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL,
            description TEXT,
            created_by INTEGER NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (created_by) REFERENCES users(user_id)
        );`,

		`CREATE TABLE IF NOT EXISTS group_members (
            group_id INTEGER NOT NULL,
            user_id INTEGER NOT NULL,
            joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            PRIMARY KEY (group_id, user_id),
            FOREIGN KEY (group_id) REFERENCES groups(group_id),
            FOREIGN KEY (user_id) REFERENCES users(user_id)
        );`,

		`CREATE TABLE IF NOT EXISTS expenses (
            expense_id INTEGER PRIMARY KEY AUTOINCREMENT,
            group_id INTEGER NOT NULL,
            description TEXT NOT NULL,
            amount DECIMAL(10,2) NOT NULL,
            created_by INTEGER NOT NULL,
            split_type TEXT NOT NULL CHECK (split_type IN ('EQUAL', 'EXACT', 'PERCENTAGE')),
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (group_id) REFERENCES groups(group_id),
            FOREIGN KEY (created_by) REFERENCES users(user_id)
        );`,

		`CREATE TABLE IF NOT EXISTS expense_shares (
            expense_id INTEGER NOT NULL,
            user_id INTEGER NOT NULL,
            share_amount DECIMAL(10,2) NOT NULL,
            share_percentage DECIMAL(5,2),
            paid_amount DECIMAL(10,2) DEFAULT 0,
            PRIMARY KEY (expense_id, user_id),
            FOREIGN KEY (expense_id) REFERENCES expenses(expense_id),
            FOREIGN KEY (user_id) REFERENCES users(user_id)
        );`,

		`CREATE TABLE IF NOT EXISTS settlements (
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
        );`,

		`CREATE INDEX IF NOT EXISTS idx_expense_shares_expense_id ON expense_shares(expense_id);`,
		`CREATE INDEX IF NOT EXISTS idx_expense_shares_user_id ON expense_shares(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_expenses_group_id ON expenses(group_id);`,
		`CREATE INDEX IF NOT EXISTS idx_settlements_payer_payee ON settlements(payer_id, payee_id);`,
	}

	// Execute each schema statement separately
	for _, schema := range schemas {
		if _, err := db.Exec(schema); err != nil {
			return fmt.Errorf("error executing schema: %v\nQuery: %s", err, schema)
		}
	}

	return nil
}
