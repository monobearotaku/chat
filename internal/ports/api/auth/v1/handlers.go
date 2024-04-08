package v1

import (
	"context"

	"github.com/monobearotaku/online-chat-api/internal/domain/credentials"
	authv1 "github.com/monobearotaku/online-chat-api/proto/auth/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (a *AuthV1) SignIn(ctx context.Context, request *authv1.SignInRequest) (*authv1.SignInResponse, error) {
	cred := credentials.NewCredentials(request.Login, request.Password)

	if err := cred.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	token, err := a.authService.SignIn(ctx, cred)
	if err != nil {
		return nil, err
	}

	return &authv1.SignInResponse{
		Token: token.String(),
	}, nil
}

func (a *AuthV1) SignUp(ctx context.Context, request *authv1.SignUpRequest) (*authv1.SignUpResponse, error) {
	cred := credentials.NewCredentials(request.Login, request.Password)

	if err := cred.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	token, err := a.authService.SignUp(ctx, cred)
	if err != nil {
		return nil, err
	}

	return &authv1.SignUpResponse{
		Token: token.String(),
	}, nil
}
