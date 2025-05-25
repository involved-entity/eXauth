package grpc

import (
	"auth/internal/pkg/database"
	"log"

	"gorm.io/gorm"
)

type UserInfo struct {
	Username string
	ID       int
}

func GetUser(userInfo UserInfo, db *gorm.DB) (database.User, error) {
	var user database.User

	var err error
	if userInfo.ID != 0 {
		err = db.Where("id = ? AND is_verified = true", userInfo.ID).First(&user).Error
	} else {
		err = db.Where("username = ? AND is_verified = true", userInfo.Username).First(&user).Error
	}

	if err != nil {
		log.Println("Error when get a user", userInfo, err)
		return database.User{}, err
	}

	return user, nil
}
