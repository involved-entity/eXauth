package utils

import (
	"auth/internal/grpc/auth"
	"auth/internal/pkg/database"

	"github.com/brianvoe/gofakeit/v7"
)

func GetAuthorizedUser() (UserData, string) {
	var userData UserData

	userData.Email = gofakeit.Email()
	userData.Username = gofakeit.Username()
	userData.Password = gofakeit.Password(true, true, true, false, false, 8)

	r := auth.Repository{Db: database.GetDB()}
	hashedPassword, _ := auth.GenerateHashedPassword(userData.Password)
	u, _ := r.SaveUser(userData.Username, userData.Email, hashedPassword)
	r.UserID = int(u.ID)
	r.VerificateUser()

	userData.ID = int(u.ID)
	tokenString, _ := auth.CreateJWTToken(u)

	return userData, tokenString
}
