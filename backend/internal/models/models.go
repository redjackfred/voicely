package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type User struct {
	ID          uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name        string         `gorm:"type:varchar(255);not null"`
	Preferences datatypes.JSON `gorm:"type:jsonb;default:'{}'"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// One-to-Many relationship: A user can have multiple call records
	Calls []Call `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
}

type Store struct {
	ID            string         `gorm:"type:varchar(255);primaryKey"` // Google Place ID
	Name          string         `gorm:"type:varchar(255);not null"`
	PhoneNumber   string         `gorm:"type:varchar(50);not null;index"` // Add index for faster lookups by phone number
	KnowledgeBase datatypes.JSON `gorm:"type:jsonb;default:'{}'"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// One-to-Many relationship: A store can have multiple call records
	Calls []Call `gorm:"foreignKey:StoreID;constraint:OnDelete:CASCADE;"`
}

type Call struct {
	ID            uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID        uuid.UUID      `gorm:"type:uuid;not null;index"`         // Add index for faster lookups by user ID
	StoreID       string         `gorm:"type:varchar(255);not null;index"` // Add index for faster lookups by store ID
	Status        string         `gorm:"type:varchar(50);not null"`
	Transcript    string         `gorm:"type:text"`
	ExtractedData datatypes.JSON `gorm:"type:jsonb;default:'{}'"`

	CreatedAt time.Time
	UpdatedAt time.Time

	// Belongs To Relationships: Each call record belongs to one user and one store
	// No need to get user and store details every time we query call records,
	// so we can omit the "preload" of these associations in most cases to optimize performance.
	// We can always fetch the related user and store details separately when needed.
	User  User  `gorm:"foreignKey:UserID"`
	Store Store `gorm:"foreignKey:StoreID"`
}

// ==========================================
// GORM Hooks
// ==========================================

// BeforeCreate is a hook in GORM that is triggered before creating a Call record
func (c *Call) BeforeCreate(tx *gorm.DB) (err error) {
	// Defensive programming logic: If no status is specified when writing, default to "pending"
	if c.Status == "" {
		c.Status = "pending"
	}
	return nil
}
