package user

import (
	"time"

	"github.com/TeluTrix/tarc/api/role"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SchemaUser struct {
	ID        uuid.UUID `gorm:"primaryKey"`
	Email     string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Role      role.SchemaRole
}
