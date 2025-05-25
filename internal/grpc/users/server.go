package users

import (
	"auth/api/users"
	"context"

	"google.golang.org/grpc"
)

type usersAPI struct {
	users.UnimplementedUsersServer
}

func Register(gRPC *grpc.Server) {
	users.RegisterUsersServer(gRPC, &usersAPI{})
}

func (s *usersAPI) GetMe(c context.Context, r *users.GetMeRequest) (*users.GetMeResponse, error) {
	return nil, nil
}
