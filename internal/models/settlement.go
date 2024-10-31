package models

import "time"

type Settlement struct {
	SettlementID int       `json:"settlement_id" db:"settlement_id"`
	PayerID      int       `json:"payer_id" db:"payer_id"`
	PayeeID      int       `json:"payee_id" db:"payee_id"`
	Amount       float64   `json:"amount" db:"amount"`
	GroupID      int       `json:"group_id" db:"group_id"`
	SettledAt    time.Time `json:"settled_at" db:"settled_at"`
	Notes        string    `json:"notes" db:"notes"`
}

type Balance struct {
	UserID int     `db:"user_id" json:"user_id"`
	OwedTo int     `db:"owed_to" json:"owed_to"`
	Amount float64 `db:"amount" json:"amount"`
}
