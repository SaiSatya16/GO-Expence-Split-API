package models

import (
	"errors"
	"time"
)

type SplitType string

const (
	SplitEqual      SplitType = "EQUAL"
	SplitExact      SplitType = "EXACT"
	SplitPercentage SplitType = "PERCENTAGE"
)

type Expense struct {
	ExpenseID   int       `json:"expense_id" db:"expense_id"`
	GroupID     int       `json:"group_id" db:"group_id"`
	Description string    `json:"description" db:"description"`
	Amount      float64   `json:"amount" db:"amount"`
	CreatedBy   int       `json:"created_by" db:"created_by"`
	SplitType   SplitType `json:"split_type" db:"split_type"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	Shares      []Share   `json:"shares,omitempty"`
}

type Share struct {
	ExpenseID       int     `json:"expense_id" db:"expense_id"`
	UserID          int     `json:"user_id" db:"user_id"`
	ShareAmount     float64 `json:"share_amount" db:"share_amount"`
	SharePercentage float64 `json:"share_percentage,omitempty" db:"share_percentage"`
	PaidAmount      float64 `json:"paid_amount" db:"paid_amount"`
}

type ExpenseCreate struct {
	GroupID     int           `json:"group_id"`
	Description string        `json:"description"`
	Amount      float64       `json:"amount"`
	SplitType   SplitType     `json:"split_type"`
	Shares      []ShareCreate `json:"shares"`
}

type ShareCreate struct {
	UserID          int     `json:"user_id"`
	ShareAmount     float64 `json:"share_amount,omitempty"`
	SharePercentage float64 `json:"share_percentage,omitempty"`
	PaidAmount      float64 `json:"paid_amount"`
}

func (e *ExpenseCreate) Validate() error {
	if e.GroupID == 0 {
		return errors.New("group ID is required")
	}
	if e.Description == "" {
		return errors.New("description is required")
	}
	if e.Amount <= 0 {
		return errors.New("amount must be greater than 0")
	}
	if len(e.Shares) == 0 {
		return errors.New("at least one share is required")
	}

	switch e.SplitType {
	case SplitEqual:
		return e.validateEqualSplit()
	case SplitExact:
		return e.validateExactSplit()
	case SplitPercentage:
		return e.validatePercentageSplit()
	default:
		return errors.New("invalid split type")
	}
}

func (e *ExpenseCreate) validateEqualSplit() error {
	shareAmount := e.Amount / float64(len(e.Shares))
	for _, share := range e.Shares {
		if share.ShareAmount != shareAmount {
			return errors.New("all shares must be equal")
		}
	}
	return nil
}

func (e *ExpenseCreate) validateExactSplit() error {
	var total float64
	for _, share := range e.Shares {
		total += share.ShareAmount
	}
	if total != e.Amount {
		return errors.New("sum of shares must equal total amount")
	}
	return nil
}

func (e *ExpenseCreate) validatePercentageSplit() error {
	var total float64
	for _, share := range e.Shares {
		total += share.SharePercentage
	}
	if total != 100 {
		return errors.New("sum of percentages must equal 100")
	}
	return nil
}
