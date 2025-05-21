package grpc

import (
	conf "auth/internal/config"
	"auth/internal/machinery"
	"auth/internal/redis"
	"context"
	"crypto/rand"
	"errors"
	"log"
	"math/big"
	"net/url"
	"strconv"
	"time"

	machineryTasks "github.com/RichardKnop/machinery/v2/tasks"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func generateSecureToken(elements string, length int) (string, error) {
	token := make([]byte, length)

	for i := range token {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(elements))))
		if err != nil {
			log.Printf("failed to generate token: %v", err)
			return "", errors.New("failed to generate token")
		}
		token[i] = elements[num.Int64()]
	}

	return string(token), nil
}

func CreateAndSendToken(id uint, email string) error {
	tokenOTP, err := generateSecureToken("0123456789", 5)
	if err != nil {
		return status.Error(codes.Internal, "failed to generate token")
	}
	redisClient := redis.GetClient()
	config := conf.GetConfig()
	redisClient.Set(
		context.Background(),
		config.OTP.RedisName+":"+strconv.Itoa(int(id)),
		tokenOTP,
		time.Minute*time.Duration(config.OTP.OTP_TTL),
	)

	machineryServer := machinery.GetServer()
	signature := &machineryTasks.Signature{
		Name: "send_email",
		Args: []machineryTasks.Arg{
			{Name: "email", Type: "string", Value: email},
			{Name: "code", Type: "string", Value: tokenOTP},
		},
	}
	machineryServer.SendTaskWithContext(context.Background(), signature)

	return nil
}

func CheckRedisToken(id int, token string, name string) error {
	redisClient := redis.GetClient()
	otp, err := redisClient.Get(context.Background(), name+":"+strconv.Itoa(id)).Result()
	if err != nil {
		return status.Error(codes.Internal, "failed to get token")
	}

	if otp != token {
		return status.Error(codes.InvalidArgument, "invalid token")
	}
	redisClient.Del(context.Background(), name+":"+strconv.Itoa(id))
	return nil
}

func CreateAndSendResetPasswordLink(id uint, email string) error {
	token, err := generateSecureToken("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", 64)
	if err != nil {
		return status.Error(codes.Internal, "failed to generate token")
	}

	redisClient := redis.GetClient()
	config := conf.GetConfig()
	redisClient.Set(
		context.Background(),
		config.ResetToken.RedisName+":"+strconv.Itoa(int(id)),
		token,
		time.Minute*time.Duration(config.ResetToken.RT_TTL),
	)

	baseURL, err := url.Parse(config.ResetToken.FrontendUrl)
	if err != nil {
		return status.Error(codes.Internal, "failed to parse frontend url")
	}

	query := url.Values{
		"token": {token},
		"id":    {strconv.Itoa(int(id))},
	}

	baseURL.RawQuery = query.Encode()

	machineryServer := machinery.GetServer()
	signature := &machineryTasks.Signature{
		Name: "reset_password",
		Args: []machineryTasks.Arg{
			{Name: "email", Type: "string", Value: email},
			{Name: "link", Type: "string", Value: baseURL.String()},
		},
	}
	machineryServer.SendTaskWithContext(context.Background(), signature)

	return nil
}
