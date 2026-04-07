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

func usernameFromContext(ctx context.Context) (string, error) {
	store := writablecontext.FromContext(ctx)
	username, _ := store.Get(auth.JWTClaimsContextKey)
	usernameStr, ok := username.(string)
	if !ok || usernameStr == "" {
		return "", errors.New("missing username in token context")
	}
	return usernameStr, nil
}

func canAccessCharacter(username string, character model.InternalCharacter) bool {
	if character.Owner == username {
		return true
	}
	share := &model.InternalCharacterShare{}
	err := model.GetDB().Table("internal_character_shares").
		Where("character_id = ? AND shared_with_username = ?", character.ID, username).
		First(share).Error
	return err == nil
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
	username, err := usernameFromContext(ctx)
	if err != nil {
		return api.GetApiListCharacters500JSONResponse{Message: internalErrorString}, nil
	}

	owned := []model.InternalCharacter{}
	if err := model.GetDB().Table("internal_characters").Where("owner = ?", username).Find(&owned).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return api.GetApiListCharacters500JSONResponse{Message: internalErrorString}, nil
		}
	}

	res := api.GetApiListCharacters200JSONResponse{}
	for _, char := range owned {
		obj, err := model.MapInternalCharacterToObject(char)
		if err != nil {
			return api.GetApiListCharacters500JSONResponse{Message: err.Error()}, nil
		}
		res = append(res, api.CharacterSummaryObject{
			Class:         *obj.Class[0].Name,
			Id:            int(char.ID),
			IsShared:      false,
			Level:         obj.Level,
			Name:          obj.Name,
			OwnerUsername: username,
			Picture:       obj.Picture,
			Race:          obj.Race,
		})
	}

	shared := []model.InternalCharacterShare{}
	if err := model.GetDB().Table("internal_character_shares").Where("shared_with_username = ?", username).Find(&shared).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return api.GetApiListCharacters500JSONResponse{Message: internalErrorString}, nil
		}
	}
	for _, share := range shared {
		char := &model.InternalCharacter{}
		if err := model.GetDB().Table("internal_characters").Where("id = ?", share.CharacterID).First(char).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return api.GetApiListCharacters500JSONResponse{Message: internalErrorString}, nil
		}
		obj, err := model.MapInternalCharacterToObject(*char)
		if err != nil {
			return api.GetApiListCharacters500JSONResponse{Message: err.Error()}, nil
		}
		res = append(res, api.CharacterSummaryObject{
			Class:         *obj.Class[0].Name,
			Id:            int(char.ID),
			IsShared:      true,
			Level:         obj.Level,
			Name:          obj.Name,
			OwnerUsername: char.Owner,
			Picture:       obj.Picture,
			Race:          obj.Race,
		})
	}

	return res, nil
}

func (*Server) GetApiGetCharacter(ctx context.Context, request api.GetApiGetCharacterRequestObject) (api.GetApiGetCharacterResponseObject, error) {
	username, err := usernameFromContext(ctx)
	if err != nil {
		return api.GetApiGetCharacter500JSONResponse{Message: internalErrorString}, nil
	}
	temp := &model.InternalCharacter{}
	err = model.GetDB().Table("internal_characters").Where("id = ?", request.Params.Id).First(temp).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return api.GetApiGetCharacter400JSONResponse{Message: "Character not found", Id: request.Params.Id}, nil
		}
		return api.GetApiGetCharacter500JSONResponse{Message: internalErrorString}, nil
	}
	if !canAccessCharacter(username, *temp) {
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
	username, err := usernameFromContext(ctx)
	if err != nil {
		return api.PostApiNewCharacter500JSONResponse{Message: internalErrorString}, nil
	}
	temp, err := model.MapObjectToInternalCharacter(*request.Body)
	if err != nil {
		return api.PostApiNewCharacter500JSONResponse{Message: err.Error()}, nil
	}
	temp.Owner = username
	if temp.Version == 0 {
		temp.Version = 1
	}

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
	username, err := usernameFromContext(ctx)
	if err != nil {
		return api.PostApiUpdateCharacter500JSONResponse{Message: internalErrorString}, nil
	}
	current := &model.InternalCharacter{}
	err = model.GetDB().Table("internal_characters").Where("id = ?", request.Body.Id).First(current).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return api.PostApiUpdateCharacter400JSONResponse{Message: "Character not found", Id: request.Body.Id}, nil
		}
		return api.PostApiUpdateCharacter500JSONResponse{Message: internalErrorString}, nil
	}
	if current.Owner != username {
		return api.PostApiUpdateCharacter400JSONResponse{Message: "Character not public", Id: request.Body.Id}, nil
	}
	if request.Body.Character.Version != current.Version {
		serverCharacter, mapErr := model.MapInternalCharacterToObject(*current)
		if mapErr != nil {
			return api.PostApiUpdateCharacter500JSONResponse{Message: mapErr.Error()}, nil
		}
		return api.PostApiUpdateCharacter409JSONResponse{
			Character: serverCharacter,
			Id:        request.Body.Id,
			Message:   "Character version conflict",
		}, nil
	}

	updated, err := model.MapObjectToInternalCharacter(request.Body.Character)
	if err != nil {
		return api.PostApiUpdateCharacter500JSONResponse{Message: err.Error()}, nil
	}
	updated.ID = current.ID
	updated.Owner = current.Owner
	updated.Version = current.Version + 1

	err = model.GetDB().Table("internal_characters").Save(&updated).Error
	if err != nil {
		return api.PostApiUpdateCharacter500JSONResponse{Message: internalErrorString}, nil
	}
	res := api.CharacterObject{}
	res, err = model.MapInternalCharacterToObject(updated)
	if err != nil {
		return api.PostApiUpdateCharacter500JSONResponse{Message: err.Error()}, nil
	}
	return api.PostApiUpdateCharacter200JSONResponse(res), nil
}

func (*Server) PostApiDeleteCharacter(ctx context.Context, request api.PostApiDeleteCharacterRequestObject) (api.PostApiDeleteCharacterResponseObject, error) {
	username, err := usernameFromContext(ctx)
	if err != nil {
		return api.PostApiDeleteCharacter500JSONResponse{Message: internalErrorString}, nil
	}
	temp := &model.InternalCharacter{}
	err = model.GetDB().Table("internal_characters").Where("id = ?", request.Body.Id).First(temp).Error
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

func (*Server) PostApiShareCharacter(ctx context.Context, request api.PostApiShareCharacterRequestObject) (api.PostApiShareCharacterResponseObject, error) {
	if request.Body == nil || request.Body.Username == "" {
		return api.PostApiShareCharacter400JSONResponse{Message: "Invalid request body"}, nil
	}
	username, err := usernameFromContext(ctx)
	if err != nil {
		return api.PostApiShareCharacter500JSONResponse{Message: internalErrorString}, nil
	}
	if request.Body.Username == username {
		return api.PostApiShareCharacter400JSONResponse{Message: "Cannot share character with yourself"}, nil
	}

	character := &model.InternalCharacter{}
	err = model.GetDB().Table("internal_characters").Where("id = ?", request.Body.Id).First(character).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return api.PostApiShareCharacter400JSONResponse{Message: "Character not found"}, nil
		}
		return api.PostApiShareCharacter500JSONResponse{Message: internalErrorString}, nil
	}
	if character.Owner != username {
		return api.PostApiShareCharacter400JSONResponse{Message: "Character not public"}, nil
	}

	recipient := &model.InternalUser{}
	err = model.GetDB().Table("internal_users").Where("username = ?", request.Body.Username).First(recipient).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return api.PostApiShareCharacter400JSONResponse{Message: "Target user not found"}, nil
		}
		return api.PostApiShareCharacter500JSONResponse{Message: internalErrorString}, nil
	}

	share := &model.InternalCharacterShare{}
	err = model.GetDB().Table("internal_character_shares").
		Where("character_id = ? AND shared_with_username = ?", character.ID, request.Body.Username).
		First(share).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return api.PostApiShareCharacter500JSONResponse{Message: internalErrorString}, nil
		}
		share.CharacterID = character.ID
		share.SharedWithUsername = request.Body.Username
		if err := model.GetDB().Table("internal_character_shares").Create(share).Error; err != nil {
			return api.PostApiShareCharacter500JSONResponse{Message: internalErrorString}, nil
		}
	}

	return api.PostApiShareCharacter200JSONResponse{
		Id:       request.Body.Id,
		Username: request.Body.Username,
	}, nil
}

func (*Server) PostApiUnshareCharacter(ctx context.Context, request api.PostApiUnshareCharacterRequestObject) (api.PostApiUnshareCharacterResponseObject, error) {
	if request.Body == nil {
		return api.PostApiUnshareCharacter400JSONResponse{Message: "Invalid request body"}, nil
	}
	username, err := usernameFromContext(ctx)
	if err != nil {
		return api.PostApiUnshareCharacter500JSONResponse{Message: internalErrorString}, nil
	}

	character := &model.InternalCharacter{}
	err = model.GetDB().Table("internal_characters").Where("id = ?", request.Body.Id).First(character).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return api.PostApiUnshareCharacter400JSONResponse{Message: "Character not found"}, nil
		}
		return api.PostApiUnshareCharacter500JSONResponse{Message: internalErrorString}, nil
	}

	var targetUser string
	if character.Owner == username {
		if request.Body.Username == nil || *request.Body.Username == "" {
			return api.PostApiUnshareCharacter400JSONResponse{Message: "username is required for owner unshare"}, nil
		}
		targetUser = *request.Body.Username
	} else {
		targetUser = username
	}

	if err := model.GetDB().Table("internal_character_shares").
		Where("character_id = ? AND shared_with_username = ?", character.ID, targetUser).
		Delete(&model.InternalCharacterShare{}).Error; err != nil {
		return api.PostApiUnshareCharacter500JSONResponse{Message: internalErrorString}, nil
	}

	response := api.PostApiUnshareCharacter200JSONResponse{
		Id: request.Body.Id,
	}
	if request.Body.Username != nil {
		response.Username = request.Body.Username
	}
	return response, nil
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

func (*Server) PostApiNewFeat(ctx context.Context, request api.PostApiNewFeatRequestObject) (api.PostApiNewFeatResponseObject, error) {
	temp := &model.InternalFeat{}
	temp.FeatObject = *request.Body
	err := model.GetDB().Table("internal_feats").Create(temp).Error
	if err != nil {
		return api.PostApiNewFeat500JSONResponse{Message: internalErrorString}, nil
	}
	return api.PostApiNewFeat200JSONResponse{
		Id:   int(temp.ID),
		Feat: temp.FeatObject,
	}, nil
}

func (*Server) GetApiGetAllFeats(ctx context.Context, request api.GetApiGetAllFeatsRequestObject) (api.GetApiGetAllFeatsResponseObject, error) {
	var results []model.InternalFeat
	result := model.GetDB().Table("internal_feats").Find(&results)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return api.GetApiGetAllFeats200JSONResponse{}, nil
		}
		return api.GetApiGetAllFeats500JSONResponse{Message: internalErrorString}, nil
	}
	res := api.GetApiGetAllFeats200JSONResponse{}
	for _, item := range results {
		res = append(res, item.FeatObject)
	}
	return res, nil
}

func (*Server) PostApiDeleteFeat(ctx context.Context, request api.PostApiDeleteFeatRequestObject) (api.PostApiDeleteFeatResponseObject, error) {
	temp := &model.InternalFeat{}
	err := model.GetDB().Table("internal_feats").Where("id = ?", request.Body.Id).First(temp).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return api.PostApiDeleteFeat400JSONResponse{Message: "Feat not found", Id: request.Body.Id}, nil
		}
		return api.PostApiDeleteFeat500JSONResponse{Message: internalErrorString}, nil
	}
	err = model.GetDB().Table("internal_feats").Delete(temp).Error
	if err != nil {
		return api.PostApiDeleteFeat500JSONResponse{Message: internalErrorString}, nil
	}
	return api.PostApiDeleteFeat200JSONResponse{
		Id:   request.Body.Id,
		Feat: temp.FeatObject,
	}, nil
}
