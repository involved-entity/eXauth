package auth

import (
	utils "auth/internal/grpc"
	"auth/internal/pkg/database"
	"log"

	"gorm.io/gorm"
)

type Repository struct {
	Db     *gorm.DB
	UserID int
}

func (r Repository) SaveUser(username string, email string, password string) (database.User, error) {
	user := database.User{Username: username, Email: email, Password: password}
	if err := r.Db.Create(&user).Error; err != nil {
		log.Println("Error when saving user", user)
		return database.User{}, err
	}
	return user, nil
}

func (r Repository) GetUser(userInfo utils.UserInfo) (database.User, error) {
	return utils.GetUser(userInfo, r.Db)
}

func (r Repository) VerificateUser() error {
	var user database.User
	if err := r.Db.Where("id = ?", r.UserID).First(&user).Error; err != nil {
		log.Println("Error when get a user", r.UserID, err)
		return err
	}
	user.IsVerified = true
	if err := r.Db.Save(&user).Error; err != nil {
		log.Println("Error when save user verified status", err)
		return err
	}
	return nil
}

func (r Repository) ChangeUserPassword(hashedPassword string) error {
	if err := r.Db.Model(&database.User{}).Where("id = ?", r.UserID).Update("password", hashedPassword).Error; err != nil {
		log.Println("Error when set new password for user", r.UserID, err)
		return err
	}
	return nil
}
