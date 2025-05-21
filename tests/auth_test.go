package tests

import (
	"auth/auth/auth"
	conf "auth/internal/config"
	"auth/internal/database"
	"context"
	"log"
	"os"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	_ "github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var JWT string

var client auth.AuthClient

type UserData struct {
	Username string
	Password string
}

var userData UserData

func InitTest() *grpc.ClientConn {
	conf.MustLoad()
	config := conf.GetConfig()
	conn, err := grpc.NewClient("localhost:"+config.GRPC.Port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}

	database.Init(config.DSN)

	client = auth.NewAuthClient(conn)

	return conn
}

func ExitTest(con *grpc.ClientConn, exitCode int) {
	con.Close()

	os.Exit(exitCode)
}

func TestMain(m *testing.M) {
	conn := InitTest()

	exitCode := m.Run()

	ExitTest(conn, exitCode)
}

func TestRegister(t *testing.T) {
	if client == nil {
		t.Fatal("gRPC client is not initialized")
	}

	email := gofakeit.Email()
	userData.Username = gofakeit.Username()
	userData.Password = gofakeit.Password(true, true, true, false, false, 8)

	_, err := client.Register(context.Background(), &auth.RegisterRequest{
		Email:    email,
		Username: userData.Username,
		Password: userData.Password,
	})

	if err != nil {
		log.Fatal(err)
	}

	db := database.GetDB()
	var user database.User
	db.Where("username = ?", userData.Username).First(&user)
	user.IsVerified = true
	db.Save(&user)
}

func TestLogin(t *testing.T) {
	_, err := client.Login(context.Background(), &auth.LoginRequest{
		Username: userData.Username,
		Password: userData.Password,
	})

	if err != nil {
		log.Fatal(err)
	}
}
