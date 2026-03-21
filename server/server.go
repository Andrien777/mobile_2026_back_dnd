package server

import (
	"context"
	"dnd_back/api"
	"dnd_back/auth"
	"dnd_back/model"
	"errors"
	"log"

	"github.com/jinzhu/gorm"
	"github.com/lestrrat-go/jwx/jwt"
	"golang.org/x/crypto/bcrypt"
)

var _ api.StrictServerInterface = (*Server)(nil)

type Server struct{}

func NewServer() Server {
	return Server{}

}

var internalErrorString = "Internal Error"
var unauthString = "Unauthorised"
var emptyString = ""
var fa *auth.FakeAuthenticator
var err error

func init() {
	fa, err = auth.NewFakeAuthenticator()
	if err != nil {
		log.Fatalln("error creating authenticator:", err)
	}
}

func (*Server) PostApiRegister(ctx context.Context, request api.PostApiRegisterRequestObject) (api.PostApiRegisterResponseObject, error) {
	temp := model.InternalUser{
		Password: request.Body.Password,
		Username: request.Body.Username,
	}

	if !model.CheckAccount(temp) {
		return api.PostApiRegister400JSONResponse{Password: request.Body.Password, Username: request.Body.Username}, nil
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(temp.Password), bcrypt.DefaultCost)
	temp.Password = string(hashedPassword)

	model.GetDB().Table("internal_users").Create(temp)

	if temp.ID <= 0 {
		return api.PostApiRegister500JSONResponse{Message: internalErrorString}, nil
	}

	token, err := fa.CreateJWSWithClaims([]string{"auth"}, request.Body.Username)
	if err != nil {
		return api.PostApiRegister500JSONResponse{Message: internalErrorString}, nil
	}

	return api.PostApiRegister200JSONResponse{
		Token:    string(token),
		Username: request.Body.Username,
	}, nil
}

func (*Server) PostApiLogin(ctx context.Context, request api.PostApiLoginRequestObject) (api.PostApiLoginResponseObject, error) {
	temp := &model.InternalUser{}
	err := model.GetDB().Table("internal_users").Where("username = ?", request.Body.Username).First(temp).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return api.PostApiLogin401JSONResponse{
				Password: "",
				Username: request.Body.Username,
			}, nil
		}

		return api.PostApiLogin500JSONResponse{Message: internalErrorString}, nil
	}

	err = bcrypt.CompareHashAndPassword([]byte(temp.Password), []byte(request.Body.Password))
	if err != nil && errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return api.PostApiLogin401JSONResponse{
			Password: "",
			Username: request.Body.Username,
		}, nil
	}

	token, err := fa.CreateJWSWithClaims([]string{"auth"}, request.Body.Username)
	if err != nil {
		return api.PostApiLogin500JSONResponse{Message: internalErrorString}, nil
	}

	return api.PostApiLogin200JSONResponse{
		Token:    string(token),
		Username: request.Body.Username,
	}, nil
}

func (*Server) GetApiListCharacters(ctx context.Context, request api.GetApiListCharactersRequestObject) (api.GetApiListCharactersResponseObject, error) {
	username := ctx.Value(auth.JWTClaimsContextKey).(map[string]interface{})[jwt.SubjectKey].(string)
	var results []model.InternalCharacter
	result := model.GetDB().Table("internal_characters").Where("owner = ?", username).Find(&results)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return api.GetApiListCharacters200JSONResponse{}, nil
		}
		return api.GetApiListCharacters500JSONResponse{Message: internalErrorString}, nil
	}
	var res api.GetApiListCharacters200JSONResponse
	for _, char := range results {
		res = append(res, struct {
			Class   string `json:"class"`
			Id      int    `json:"id"`
			Level   uint   `json:"level"`
			Name    string `json:"name"`
			Picture string `json:"picture"`
			Race    string `json:"race"`
		}{Class: *char.Class[0].Name, Id: int(char.ID), Level: *char.Class[0].Level, Name: char.Name, Picture: char.Picture, Race: char.Race})
	}
	return res, nil
}

func (*Server) GetApiGetCharacter(ctx context.Context, request api.GetApiGetCharacterRequestObject) (api.GetApiGetCharacterResponseObject, error) {
	username := ctx.Value(auth.JWTClaimsContextKey).(map[string]interface{})[jwt.SubjectKey].(string)
	temp := &model.InternalCharacter{}
	err := model.GetDB().Table("internal_characters").Where("id = ?", request.Params.Id).First(temp).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return api.GetApiGetCharacter400JSONResponse{Message: "Character not found", Id: request.Params.Id}, nil
		}
		return api.GetApiGetCharacter500JSONResponse{Message: internalErrorString}, nil
	}
	if temp.Owner != username {
		return api.GetApiGetCharacter400JSONResponse{Message: "Character not public", Id: request.Params.Id}, nil
	}
	return api.GetApiGetCharacter200JSONResponse(temp.CharacterObject), nil
}

func (*Server) PostApiNewCharacter(ctx context.Context, request api.PostApiNewCharacterRequestObject) (api.PostApiNewCharacterResponseObject, error) {
	temp := &model.InternalCharacter{}
	temp.CharacterObject = *request.Body

	err := model.GetDB().Table("internal_characters").Create(temp).Error
	if err != nil {
		return api.PostApiNewCharacter500JSONResponse{Message: internalErrorString}, nil
	}

	return api.PostApiNewCharacter200JSONResponse{
		Character: temp.CharacterObject,
		Id:        int(temp.ID),
	}, nil
}

func (*Server) PostApiUpdateCharacter(ctx context.Context, request api.PostApiUpdateCharacterRequestObject) (api.PostApiUpdateCharacterResponseObject, error) {
	username := ctx.Value(auth.JWTClaimsContextKey).(map[string]interface{})[jwt.SubjectKey].(string)
	temp := &model.InternalCharacter{}
	err := model.GetDB().Table("internal_characters").Where("id = ?", request.Body.Id).First(temp).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return api.PostApiUpdateCharacter400JSONResponse{Message: "Character not found", Id: request.Body.Id}, nil
		}
		return api.PostApiUpdateCharacter500JSONResponse{Message: internalErrorString}, nil
	}
	if temp.Owner != username {
		return api.PostApiUpdateCharacter400JSONResponse{Message: "Character not public", Id: request.Body.Id}, nil
	}
	err = model.GetDB().Table("internal_characters").Save(temp).Error
	if err != nil {
		return api.PostApiUpdateCharacter500JSONResponse{Message: internalErrorString}, nil
	}
	return api.PostApiUpdateCharacter200JSONResponse(temp.CharacterObject), nil
}

func (*Server) PostApiDeleteCharacter(ctx context.Context, request api.PostApiDeleteCharacterRequestObject) (api.PostApiDeleteCharacterResponseObject, error) {
	username := ctx.Value(auth.JWTClaimsContextKey).(map[string]interface{})[jwt.SubjectKey].(string)
	temp := &model.InternalCharacter{}
	err := model.GetDB().Table("internal_characters").Where("id = ?", request.Body.Id).First(temp).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return api.PostApiDeleteCharacter400JSONResponse{Message: "Character not found", Id: request.Body.Id}, nil
		}
		return api.PostApiDeleteCharacter500JSONResponse{Message: internalErrorString}, nil
	}
	if temp.Owner != username {
		return api.PostApiDeleteCharacter400JSONResponse{Message: "Character not public", Id: request.Body.Id}, nil
	}
	err = model.GetDB().Table("internal_characters").Delete(temp).Error
	if err != nil {
		return api.PostApiDeleteCharacter500JSONResponse{Message: internalErrorString}, nil
	}
	return api.PostApiDeleteCharacter200JSONResponse{
		Character: temp.CharacterObject,
		Id:        request.Body.Id,
	}, nil
}
