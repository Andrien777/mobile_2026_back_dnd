package server

import (
	"context"
	"dnd_back/api"
	"dnd_back/auth"
	"dnd_back/model"
	"errors"
	"log"

	"github.com/induzo/gocom/http/middleware/writablecontext"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

var _ api.StrictServerInterface = (*Server)(nil)

type Server struct{}

func NewServer() Server {
	return Server{}

}

var internalErrorString = "Internal Error"
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

	model.GetDB().Table("internal_users").Create(&temp)

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
	store := writablecontext.FromContext(ctx)
	username, _ := store.Get(auth.JWTClaimsContextKey)
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
		obj, err := model.MapInternalCharacterToObject(char)
		if err != nil {
			return api.GetApiListCharacters500JSONResponse{Message: err.Error()}, nil
		}
		res = append(res, struct {
			Class   string `json:"class"`
			Id      int    `json:"id"`
			Level   uint   `json:"level"`
			Name    string `json:"name"`
			Picture string `json:"picture"`
			Race    string `json:"race"`
		}{Class: *obj.Class[0].Name, Id: int(char.ID), Level: obj.Level, Name: obj.Name, Picture: obj.Picture, Race: obj.Race})
	}
	return res, nil
}

func (*Server) GetApiGetCharacter(ctx context.Context, request api.GetApiGetCharacterRequestObject) (api.GetApiGetCharacterResponseObject, error) {
	store := writablecontext.FromContext(ctx)
	username, _ := store.Get(auth.JWTClaimsContextKey)
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
	res := api.CharacterObject{}
	res, err = model.MapInternalCharacterToObject(*temp)
	if err != nil {
		return api.GetApiGetCharacter500JSONResponse{Message: err.Error()}, nil
	}
	return api.GetApiGetCharacter200JSONResponse(res), nil
}

func (*Server) PostApiNewCharacter(ctx context.Context, request api.PostApiNewCharacterRequestObject) (api.PostApiNewCharacterResponseObject, error) {
	store := writablecontext.FromContext(ctx)
	username, _ := store.Get(auth.JWTClaimsContextKey)
	temp, err := model.MapObjectToInternalCharacter(*request.Body)
	if err != nil {
		return api.PostApiNewCharacter500JSONResponse{Message: err.Error()}, nil
	}
	temp.Owner = username.(string)

	err = model.GetDB().Table("internal_characters").Create(&temp).Error
	if err != nil {
		return api.PostApiNewCharacter500JSONResponse{Message: internalErrorString}, nil
	}

	res := api.CharacterObject{}
	res, err = model.MapInternalCharacterToObject(temp)
	if err != nil {
		return api.PostApiNewCharacter500JSONResponse{Message: err.Error()}, nil
	}
	return api.PostApiNewCharacter200JSONResponse{
		Character: res,
		Id:        int(temp.ID),
	}, nil
}

func (*Server) PostApiUpdateCharacter(ctx context.Context, request api.PostApiUpdateCharacterRequestObject) (api.PostApiUpdateCharacterResponseObject, error) {
	store := writablecontext.FromContext(ctx)
	username, _ := store.Get(auth.JWTClaimsContextKey)
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
	res := api.CharacterObject{}
	res, err = model.MapInternalCharacterToObject(*temp)
	if err != nil {
		return api.PostApiUpdateCharacter500JSONResponse{Message: err.Error()}, nil
	}
	return api.PostApiUpdateCharacter200JSONResponse(res), nil
}

func (*Server) PostApiDeleteCharacter(ctx context.Context, request api.PostApiDeleteCharacterRequestObject) (api.PostApiDeleteCharacterResponseObject, error) {
	store := writablecontext.FromContext(ctx)
	username, _ := store.Get(auth.JWTClaimsContextKey)
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
	res := api.CharacterObject{}
	res, err = model.MapInternalCharacterToObject(*temp)
	if err != nil {
		return api.PostApiDeleteCharacter500JSONResponse{Message: err.Error()}, nil
	}
	return api.PostApiDeleteCharacter200JSONResponse{
		Character: res,
		Id:        request.Body.Id,
	}, nil
}

func (*Server) PostApiNewSpell(ctx context.Context, request api.PostApiNewSpellRequestObject) (api.PostApiNewSpellResponseObject, error) {
	temp := model.MapObjectToInternalSpell(*request.Body)
	err := model.GetDB().Table("internal_spells").Create(&temp).Error
	if err != nil {
		return api.PostApiNewSpell500JSONResponse{Message: internalErrorString}, nil
	}
	return api.PostApiNewSpell200JSONResponse{
		Id:    int(temp.ID),
		Spell: model.MapInternalSpellToObject(temp),
	}, nil
}

func (*Server) GetApiGetAllSpells(ctx context.Context, request api.GetApiGetAllSpellsRequestObject) (api.GetApiGetAllSpellsResponseObject, error) {
	var results []model.InternalSpell
	result := model.GetDB().Table("internal_spells").Find(&results)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return api.GetApiGetAllSpells200JSONResponse{}, nil
		}
		return api.GetApiGetAllSpells500JSONResponse{Message: internalErrorString}, nil
	}
	res := api.GetApiGetAllSpells200JSONResponse{}
	for _, spell := range results {
		res = append(res, model.MapInternalSpellToObject(spell))
	}
	return res, nil
}

func (*Server) PostApiDeleteSpell(ctx context.Context, request api.PostApiDeleteSpellRequestObject) (api.PostApiDeleteSpellResponseObject, error) {
	temp := &model.InternalSpell{}
	err := model.GetDB().Table("internal_spells").Where("id = ?", request.Body.Id).First(temp).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return api.PostApiDeleteSpell400JSONResponse{Message: "Spell not found", Id: request.Body.Id}, nil
		}
		return api.PostApiDeleteSpell500JSONResponse{Message: internalErrorString}, nil
	}
	err = model.GetDB().Table("internal_spells").Delete(temp).Error
	if err != nil {
		return api.PostApiDeleteSpell500JSONResponse{Message: internalErrorString}, nil
	}
	return api.PostApiDeleteSpell200JSONResponse{
		Id:    request.Body.Id,
		Spell: model.MapInternalSpellToObject(*temp),
	}, nil
}

func (*Server) PostApiNewItem(ctx context.Context, request api.PostApiNewItemRequestObject) (api.PostApiNewItemResponseObject, error) {
	temp, err := model.MapObjectToInternalItem(*request.Body)
	if err != nil {
		return api.PostApiNewItem500JSONResponse{Message: err.Error()}, nil
	}
	err = model.GetDB().Table("internal_items").Create(&temp).Error
	if err != nil {
		return api.PostApiNewItem500JSONResponse{Message: internalErrorString}, nil
	}
	res := api.ItemObject{}
	res, err = model.MapInternalItemToObject(temp)
	if err != nil {
		return api.PostApiNewItem500JSONResponse{Message: err.Error()}, nil
	}
	return api.PostApiNewItem200JSONResponse{
		Id:   int(temp.ID),
		Item: res,
	}, nil
}

func (*Server) GetApiGetAllItems(ctx context.Context, request api.GetApiGetAllItemsRequestObject) (api.GetApiGetAllItemsResponseObject, error) {
	var results []model.InternalItem
	result := model.GetDB().Table("internal_items").Find(&results)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return api.GetApiGetAllItems200JSONResponse{}, nil
		}
		return api.GetApiGetAllItems500JSONResponse{Message: internalErrorString}, nil
	}
	res := api.GetApiGetAllItems200JSONResponse{}
	for _, item := range results {
		temp, err := model.MapInternalItemToObject(item)
		if err != nil {
			return api.GetApiGetAllItems500JSONResponse{Message: err.Error()}, nil
		}
		res = append(res, temp)
	}
	return res, nil
}

func (*Server) PostApiDeleteItem(ctx context.Context, request api.PostApiDeleteItemRequestObject) (api.PostApiDeleteItemResponseObject, error) {
	temp := &model.InternalItem{}
	err := model.GetDB().Table("internal_items").Where("id = ?", request.Body.Id).First(temp).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return api.PostApiDeleteItem400JSONResponse{Message: "Item not found", Id: request.Body.Id}, nil
		}
		return api.PostApiDeleteItem500JSONResponse{Message: internalErrorString}, nil
	}
	err = model.GetDB().Table("internal_items").Delete(temp).Error
	if err != nil {
		return api.PostApiDeleteItem500JSONResponse{Message: internalErrorString}, nil
	}
	res := api.ItemObject{}
	res, err = model.MapInternalItemToObject(*temp)
	if err != nil {
		return api.PostApiDeleteItem500JSONResponse{Message: err.Error()}, nil
	}
	return api.PostApiDeleteItem200JSONResponse{
		Id:   request.Body.Id,
		Item: res,
	}, nil
}
