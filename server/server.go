package server

import (
	"context"
	"dnd_back/api"
	"dnd_back/model"

	"golang.org/x/crypto/bcrypt"
)

var _ api.StrictServerInterface = (*Server)(nil)

type Server struct{}

func NewServer() Server {
	return Server{}

}
func (*Server) PostApiRegister(ctx context.Context, request api.PostApiRegisterRequestObject) (api.PostApiRegisterResponseObject, error) {
	temp := model.InternalUser{
		Password: *request.Body.Password,
		Username: *request.Body.Username,
	}

	if !model.Check_account(temp) {
		return api.PostApiRegister400JSONResponse{Password: request.Body.Password, Username: request.Body.Username}, nil
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(account.Password), bcrypt.DefaultCost)
	account.Password = string(hashedPassword)

	GetDB().Create(account)

	if account.ID <= 0 {
		return u.Message(false, "Failed to create account, connection error.")
	}

}
