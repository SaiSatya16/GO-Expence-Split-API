package repository

import (
	"expense-sharing-api/internal/models"

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
        WITH user_expenses AS (
            SELECT 
                e.expense_id,
                e.created_by as paid_by,
                es.user_id as owed_by,
                es.share_amount,
                es.paid_amount
            FROM expenses e
            JOIN expense_shares es ON e.expense_id = es.expense_id
            WHERE e.group_id = ?
        )
        SELECT 
            owed_by as user_id,
            paid_by as owed_to,
            SUM(share_amount - paid_amount) as amount
        FROM user_expenses
        WHERE (owed_by = ? OR paid_by = ?)
        GROUP BY owed_by, paid_by
        HAVING amount > 0`

	var balances []models.Balance
	err := r.db.Select(&balances, query, groupID, userID, userID)
	return balances, err
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
