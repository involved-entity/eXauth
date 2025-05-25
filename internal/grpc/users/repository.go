package users

import (
	utils "auth/internal/grpc"
	"auth/internal/pkg/database"

	"gorm.io/gorm"
)

type Repository struct {
	db     *gorm.DB
	UserID int
}

func (r Repository) GetUser(userInfo utils.UserInfo) (database.User, error) {
	return utils.GetUser(userInfo, r.db)
}
