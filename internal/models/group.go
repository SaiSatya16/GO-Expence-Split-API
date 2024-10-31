package models

import (
	"errors"
	"time"
)

type Group struct {
	GroupID     int       `json:"group_id" db:"group_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	CreatedBy   int       `json:"created_by" db:"created_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	Members     []User    `json:"members,omitempty"`
}

type GroupCreate struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Members     []int  `json:"members"` // User IDs
}

func (g *GroupCreate) Validate() error {
	if g.Name == "" {
		return errors.New("group name is required")
	}
	if len(g.Members) == 0 {
		return errors.New("group must have at least one member")
	}
	return nil
}
