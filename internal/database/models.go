package database

import (
	"time"

	_ "ariga.io/atlas-provider-gorm/gormschema"
)

type User struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	JoinedAt   time.Time `gorm:"autoCreateTime" json:"joinedAt"`
	Username   string    `gorm:"uniqueIndex;not null" json:"username"`
	Password   string    `gorm:"not null" json:"-"`
	Email      string    `gorm:"uniqueIndex;not null" json:"email"`
	IsVerified bool      `gorm:"default:false" json:"-"`
	IsAdmin    bool      `gorm:"default:false" json:"-"`
}
