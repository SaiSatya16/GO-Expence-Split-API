package repository

import (
	"expense-sharing-api/internal/models"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
)

type ExpenseRepository struct {
	db *sqlx.DB
}

func NewExpenseRepository(db *sqlx.DB) *ExpenseRepository {
	return &ExpenseRepository{db: db}
}

func (r *ExpenseRepository) Create(expense *models.ExpenseCreate, createdBy int) (*models.Expense, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Create expense
	expenseQuery := `
        INSERT INTO expenses (group_id, description, amount, created_by, split_type)
        VALUES (?, ?, ?, ?, ?)
        RETURNING expense_id, group_id, description, amount, created_by, split_type, created_at`

	var created models.Expense
	err = tx.QueryRowx(expenseQuery,
		expense.GroupID,
		expense.Description,
		expense.Amount,
		createdBy,
		expense.SplitType,
	).StructScan(&created)
	if err != nil {
		return nil, err
	}

	// Add shares
	shareQuery := `
        INSERT INTO expense_shares (expense_id, user_id, share_amount, share_percentage, paid_amount)
        VALUES (?, ?, ?, ?, ?)`

	for _, share := range expense.Shares {
		_, err = tx.Exec(shareQuery,
			created.ExpenseID,
			share.UserID,
			share.ShareAmount,
			share.SharePercentage,
			share.PaidAmount,
		)
		if err != nil {
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &created, nil
}

func (r *ExpenseRepository) GetByID(expenseID int) (*models.Expense, error) {
	var expense models.Expense
	query := `SELECT * FROM expenses WHERE expense_id = ?`
	err := r.db.Get(&expense, query, expenseID)
	if err != nil {
		return nil, err
	}

	// Get shares
	sharesQuery := `SELECT * FROM expense_shares WHERE expense_id = ?`
	err = r.db.Select(&expense.Shares, sharesQuery, expenseID)
	if err != nil {
		return nil, err
	}

	return &expense, nil
}

func (r *ExpenseRepository) GetGroupExpenses(groupID int) ([]models.Expense, error) {
	query := `SELECT * FROM expenses WHERE group_id = ? ORDER BY created_at DESC`
	var expenses []models.Expense
	err := r.db.Select(&expenses, query, groupID)
	if err != nil {
		return nil, err
	}

	// Get shares for each expense
	sharesQuery := `SELECT * FROM expense_shares WHERE expense_id = ?`
	for i := range expenses {
		err = r.db.Select(&expenses[i].Shares, sharesQuery, expenses[i].ExpenseID)
		if err != nil {
			return nil, err
		}
	}

	return expenses, nil
}

func (r *ExpenseRepository) GetUserBalance(userID, groupID int) ([]models.Balance, error) {
	query := `
        WITH user_balances AS (
            SELECT 
                es.user_id,
                e.created_by as owed_to,
                SUM(es.share_amount) as total_share,
                SUM(es.paid_amount) as total_paid
            FROM expenses e
            JOIN expense_shares es ON e.expense_id = es.expense_id
            WHERE e.group_id = ?
            GROUP BY es.user_id, e.created_by
        )
        SELECT 
            user_id,    -- matches Balance.UserID
            owed_to,    -- matches Balance.OwedTo
            (total_share - total_paid) as amount  -- matches Balance.Amount
        FROM user_balances
        WHERE (user_id = ? OR owed_to = ?)
            AND (total_share - total_paid) > 0
        ORDER BY amount DESC`

	var balances []models.Balance
	err := r.db.Select(&balances, query, groupID, userID, userID)
	if err != nil {
		log.Printf("Error fetching balance sheet: %v", err)
		return nil, fmt.Errorf("error fetching balance sheet: %v", err)
	}

	// Debug logging
	log.Printf("Found %d balance records", len(balances))
	for _, b := range balances {
		log.Printf("Balance: UserID=%d, OwedTo=%d, Amount=%.2f", b.UserID, b.OwedTo, b.Amount)
	}

	return balances, nil
}

func (r *ExpenseRepository) Settle(settlement *models.Settlement) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Create settlement record
	_, err = tx.Exec(`
        INSERT INTO settlements (payer_id, payee_id, amount, group_id, notes)
        VALUES (?, ?, ?, ?, ?)`,
		settlement.PayerID,
		settlement.PayeeID,
		settlement.Amount,
		settlement.GroupID,
		settlement.Notes,
	)
	if err != nil {
		return err
	}

	// Update paid amounts in expense_shares
	// This is a simplified version - in a real app, you'd need a more sophisticated
	// algorithm to determine which expenses to mark as paid
	_, err = tx.Exec(`
        UPDATE expense_shares
        SET paid_amount = share_amount
        WHERE user_id = ?
        AND expense_id IN (
            SELECT expense_id
            FROM expenses
            WHERE group_id = ?
            AND created_by = ?
        )`,
		settlement.PayerID,
		settlement.GroupID,
		settlement.PayeeID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}
