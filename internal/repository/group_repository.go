package repository

import (
	"expense-sharing-api/internal/models"

	"github.com/jmoiron/sqlx"
)

type GroupRepository struct {
	db *sqlx.DB
}

func NewGroupRepository(db *sqlx.DB) *GroupRepository {
	return &GroupRepository{db: db}
}

func (r *GroupRepository) Create(group *models.GroupCreate, createdBy int) (*models.Group, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Create group
	query := `
        INSERT INTO groups (name, description, created_by)
        VALUES (?, ?, ?)
        RETURNING group_id, name, description, created_by, created_at`

	var created models.Group
	err = tx.QueryRowx(query, group.Name, group.Description, createdBy).StructScan(&created)
	if err != nil {
		return nil, err
	}

	// Add members
	memberQuery := `INSERT INTO group_members (group_id, user_id) VALUES (?, ?)`
	for _, memberID := range group.Members {
		_, err = tx.Exec(memberQuery, created.GroupID, memberID)
		if err != nil {
			return nil, err
		}
	}

	// Add creator as member if not already included
	creatorIncluded := false
	for _, memberID := range group.Members {
		if memberID == createdBy {
			creatorIncluded = true
			break
		}
	}
	if !creatorIncluded {
		_, err = tx.Exec(memberQuery, created.GroupID, createdBy)
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

func (r *GroupRepository) GetByID(groupID int) (*models.Group, error) {
	var group models.Group
	query := `SELECT * FROM groups WHERE group_id = ?`
	err := r.db.Get(&group, query, groupID)
	if err != nil {
		return nil, err
	}

	// Get members
	membersQuery := `
        SELECT u.*
        FROM users u
        JOIN group_members gm ON u.user_id = gm.user_id
        WHERE gm.group_id = ?`

	err = r.db.Select(&group.Members, membersQuery, groupID)
	if err != nil {
		return nil, err
	}

	return &group, nil
}

func (r *GroupRepository) GetUserGroups(userID int) ([]models.Group, error) {
	query := `
        SELECT g.*
        FROM groups g
        JOIN group_members gm ON g.group_id = gm.group_id
        WHERE gm.user_id = ?`

	var groups []models.Group
	err := r.db.Select(&groups, query, userID)
	return groups, err
}
