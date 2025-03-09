package handler

import "context"

type AuthService interface {
	RegisterUser(ctx context.Context, email, password string) error
	LoginUser(ctx context.Context, email, password string) (string, error)
}
