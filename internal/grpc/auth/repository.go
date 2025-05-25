package auth

import (
	utils "auth/internal/grpc"
	"auth/internal/pkg/database"
	"log"

	"gorm.io/gorm"
)

type Repository struct {
	db     *gorm.DB
	UserID int
}

func (r Repository) SaveUser(username string, email string, password string) (database.User, error) {
	user := database.User{Username: username, Email: email, Password: password}
	if err := r.db.Create(&user).Error; err != nil {
		log.Println("Error when saving user", user)
		return database.User{}, err
	}
	return user, nil
}

func (r Repository) GetUser(userInfo utils.UserInfo) (database.User, error) {
	return utils.GetUser(userInfo, r.db)
}

func (r Repository) VerificateUser() error {
	var user database.User
	if err := r.db.Where("id = ?", r.UserID).First(&user).Error; err != nil {
		log.Println("Error when get a user", r.UserID, err)
		return err
	}
	user.IsVerified = true
	if err := r.db.Save(&user).Error; err != nil {
		log.Println("Error when save user verified status", err)
		return err
	}
	return nil
}

func (r Repository) ChangeUserPassword(hashedPassword string) error {
	if err := r.db.Model(&database.User{}).Where("id = ?", r.UserID).Update("password", hashedPassword).Error; err != nil {
		log.Println("Error when set new password for user", r.UserID, err)
		return err
	}
	return nil
}

// func (r Repository) UpdateAccount(email string) (database.User, error) {
// 	var user database.User
// 	err := r.db.Where("id = ?", r.UserID).First(&user).Error
// 	if err != nil {
// 		log.Println("Error when get user", r.UserID, err)
// 		return database.User{}, err
// 	}
// 	user.Email = email
// 	if err = r.db.Clauses(clause.Returning{}).Save(&user).Error; err != nil {
// 		log.Println("Error when update user", r.UserID, err)
// 		return database.User{}, err
// 	}
// 	return user, nil
// }
