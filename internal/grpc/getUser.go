package grpc

import (
	"auth/internal/pkg/database"
	"log"

	"gorm.io/gorm"
)

type UserInfo struct {
	Username              string
	ID                    int
	VerificateNotRequired bool
}

func GetUser(userInfo UserInfo, db *gorm.DB) (database.User, error) {
	var user database.User

	var tx *gorm.DB
	if userInfo.ID != 0 {
		tx = db.Where(ternary(userInfo.VerificateNotRequired, "id = ?", "id = ? AND is_verified = true"), userInfo.ID)
	} else {
		tx = db.Where(ternary(userInfo.VerificateNotRequired, "username = ?", "username = ? AND is_verified = true"), userInfo.Username)
	}

	err := tx.First(&user).Error
	if err != nil {
		log.Println("Error when get a user", userInfo, err)
		return database.User{}, err
	}

	return user, nil
}

func ternary(a bool, b, c any) any {
	if a {
		return b
	}
	return c
}
