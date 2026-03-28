package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// User 對應 Users Table
type User struct {
	ID          uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name        string         `gorm:"type:varchar(255);not null"` // 加入 not null 確保資料完整性
	Preferences datatypes.JSON `gorm:"type:jsonb;default:'{}'"`

	// 追蹤時間與軟刪除機制
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// 一對多關聯 (Has Many)：一個使用者可以有多筆通話紀錄
	Calls []Call `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
}

// Store 對應 Stores Table
type Store struct {
	ID            string         `gorm:"type:varchar(255);primaryKey"` // 使用 Google Place ID
	Name          string         `gorm:"type:varchar(255);not null"`
	PhoneNumber   string         `gorm:"type:varchar(50);not null;index"` // 加入 index 提升未來若需以電話反查店家的效能
	KnowledgeBase datatypes.JSON `gorm:"type:jsonb;default:'{}'"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// 一對多關聯 (Has Many)：一個店家可以有多筆通話紀錄
	Calls []Call `gorm:"foreignKey:StoreID;constraint:OnDelete:CASCADE;"`
}

// Call 對應 Calls Table (通話紀錄)
type Call struct {
	ID            uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID        uuid.UUID      `gorm:"type:uuid;not null;index"`         // 加上 index 加速跨表查詢
	StoreID       string         `gorm:"type:varchar(255);not null;index"` // 加上 index 加速跨表查詢
	Status        string         `gorm:"type:varchar(50);not null"`
	Transcript    string         `gorm:"type:text"`
	ExtractedData datatypes.JSON `gorm:"type:jsonb;default:'{}'"`

	CreatedAt time.Time
	UpdatedAt time.Time

	// 屬於關聯 (Belongs To)
	User  User  `gorm:"foreignKey:UserID"`
	Store Store `gorm:"foreignKey:StoreID"`
}

// ==========================================
// GORM Hooks (物件生命週期掛鉤)
// ==========================================

// BeforeCreate 是 GORM 的 Hook，會在新增 Call 紀錄到資料庫「前」自動觸發
func (c *Call) BeforeCreate(tx *gorm.DB) (err error) {
	// 防呆邏輯：如果在寫入時沒有指定狀態，預設為 "pending"
	if c.Status == "" {
		c.Status = "pending"
	}
	return nil
}
