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
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var JWT string

var client auth.AuthClient

type UserData struct {
	ID       int
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

	response, err := client.Register(context.Background(), &auth.RegisterRequest{
		Email:    email,
		Username: userData.Username,
		Password: userData.Password,
	})

	if err != nil {
		log.Fatal(err)
	}

	userData.ID = int(response.Id)
}

func TestRegenerateCode(t *testing.T) {
	_, err := client.RegenerateCode(context.Background(), &auth.RegenerateCodeRequest{
		Id:    int64(userData.ID),
		Email: userData.Username,
	})

	if err != nil {
		log.Fatal(err)
	}
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

func TestIsAdmin(t *testing.T) {
	response, err := client.IsAdmin(context.Background(), &auth.IsAdminRequest{
		Token: JWT,
	})

	if err != nil {
		log.Fatal(err)
	}

	require.Equal(t, response.IsAdmin, false)
}
