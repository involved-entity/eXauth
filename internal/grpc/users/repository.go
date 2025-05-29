package users

import (
	utils "auth/internal/grpc"
	"auth/internal/pkg/database"
	"log"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	Db     *gorm.DB
	UserID int
}

func (r Repository) GetUser(userInfo utils.UserInfo) (database.User, error) {
	return utils.GetUser(userInfo, r.Db)
}

func (r Repository) UpdateAccount(email string) (database.User, error) {
	var user database.User
	err := r.Db.Where("id = ?", r.UserID).First(&user).Error
	if err != nil {
		log.Println("Error when get user", r.UserID, err)
		return database.User{}, err
	}
	user.Email = email
	if err = r.Db.Clauses(clause.Returning{}).Save(&user).Error; err != nil {
		log.Println("Error when update user", r.UserID, err)
		return database.User{}, err
	}
	return user, nil
}
