package model

import (
	"dnd_back/api"
	"encoding/json"

	_ "github.com/jinzhu/gorm"
	"github.com/lib/pq"
)

type InternalCharacter struct {
	ID        uint `gorm:"primary_key"`
	Abilities struct {
		Cha uint
		Con uint
		Dex uint
		Int uint
		Str uint
		Wis uint
	} `gorm:"embedded;embedded_prefix:abilities_"`
	Ac                uint
	Age               string
	Alignment         string
	Allies            string
	Appearance        string
	ArmourProficiency pq.StringArray `gorm:"type:text[]"`
	AttacksJson       []byte
	Background        struct {
		Description string
		Name        string
	} `gorm:"embedded;embedded_prefix:background_"`
	Backstory         string
	Bonds             string
	BonusActionsJson  []byte
	Capacity          uint
	ClassJson         []byte
	ClassFeaturesJson []byte
	Conditions        pq.StringArray `gorm:"type:text[]"`
	DeathSaves        struct {
		Failure uint
		Success uint
	} `gorm:"embedded;embedded_prefix:death_saves_"`
	Defences struct {
		Custom     pq.StringArray `gorm:"type:text[]"`
		Immune     pq.StringArray `gorm:"type:text[]"`
		Resistance pq.StringArray `gorm:"type:text[]"`
		Vulnerable pq.StringArray `gorm:"type:text[]"`
	} `gorm:"embedded;embedded_prefix:defences_"`
	Enemies    string
	Expertises pq.StringArray `gorm:"type:text[]"`
	Eyes       string
	Faith      string
	FeatsJson  []byte
	Flaws      string
	Gender     string
	Hair       string
	Height     string
	Hp         struct {
		Current uint
		HitDice struct {
			Left  uint
			Max   uint
			Value uint
		} `gorm:"embedded;embedded_prefix:hp_hit_dice_"`
		Max  uint
		Temp uint
	} `gorm:"embedded;embedded_prefix:hp_"`
	Ideals        string
	InitiativeMod int
	InventoryJson []byte
	Languages     pq.StringArray `gorm:"type:text[]"`
	Level         uint
	Money         struct {
		Cp uint
		Ep uint
		Gp uint
		Pp uint
		Sp uint
	} `gorm:"embedded;embedded_prefix:money_"`
	Name            string
	Notes           string
	Organizations   string
	Personality     string
	Picture         string
	Proficiencies   pq.StringArray `gorm:"type:text[]"`
	ProficiencyMod  uint
	Race            string
	RaceTraitsJson  []byte
	SavingThrowProf pq.StringArray `gorm:"type:text[]"`
	Senses          pq.StringArray `gorm:"type:text[]"`
	Shielded        bool
	Size            string
	Skin            string
	Speed           uint
	Spells          struct {
		CurrentConcentrationJson []byte
		KnownSpellsJson          []byte
		ReadySpellsJson          []byte
		SlotsJson                []byte
	} `gorm:"embedded;embedded_prefix:spells_"`
	ToolsProficiency  pq.StringArray `gorm:"type:text[]"`
	WeaponProficiency pq.StringArray `gorm:"type:text[]"`
	Weight            uint
	Xp                uint
	Version           uint
	Owner             string
}

func MapInternalCharacterToObject(internal InternalCharacter) (api.CharacterObject, error) {
	result := api.CharacterObject{}
	result.Abilities = struct {
		Cha uint `json:"cha"`
		Con uint `json:"con"`
		Dex uint `json:"dex"`
		Int uint `json:"int"`
		Str uint `json:"str"`
		Wis uint `json:"wis"`
	}(internal.Abilities)
	result.Ac = internal.Ac
	result.Age = internal.Age
	result.Alignment = api.CharacterObjectAlignment(internal.Alignment)
	result.Allies = internal.Allies
	result.Appearance = internal.Appearance
	result.ArmourProficiency = internal.ArmourProficiency
	err := json.Unmarshal(internal.AttacksJson, &result.Attacks)
	if err != nil {
		return result, err
	}
	result.Background = struct {
		Description string `json:"description"`
		Name        string `json:"name"`
	}(internal.Background)
	result.Backstory = internal.Backstory
	result.Bonds = internal.Bonds
	err = json.Unmarshal(internal.BonusActionsJson, &result.BonusActions)
	if err != nil {
		return result, err
	}
	result.Capacity = internal.Capacity
	err = json.Unmarshal(internal.ClassJson, &result.Class)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(internal.ClassFeaturesJson, &result.ClassFeatures)
	if err != nil {
		return result, err
	}
	result.Conditions = internal.Conditions
	result.DeathSaves = struct {
		Failure uint `json:"failure"`
		Success uint `json:"success"`
	}(internal.DeathSaves)
	result.Defences.Custom = internal.Defences.Custom
	result.Defences.Immune = internal.Defences.Immune
	result.Defences.Resistance = internal.Defences.Resistance
	result.Defences.Vulnerable = internal.Defences.Vulnerable
	result.Enemies = internal.Enemies
	result.Expertises = []api.CharacterObjectExpertises{}
	for _, expert := range internal.Expertises {
		result.Expertises = append(result.Expertises, api.CharacterObjectExpertises(expert))
	}
	result.Eyes = internal.Eyes
	result.Faith = internal.Faith
	err = json.Unmarshal(internal.FeatsJson, &result.Feats)
	if err != nil {
		return result, err
	}
	result.Flaws = internal.Flaws
	result.Gender = internal.Gender
	result.Hair = internal.Hair
	result.Height = internal.Height
	result.Hp.Current = internal.Hp.Current
	result.Hp.Max = internal.Hp.Max
	result.Hp.Temp = internal.Hp.Temp
	result.Hp.HitDice.Left = internal.Hp.HitDice.Left
	result.Hp.HitDice.Max = internal.Hp.HitDice.Max
	result.Hp.HitDice.Value = api.CharacterObjectHpHitDiceValue(internal.Hp.HitDice.Value)
	result.Ideals = internal.Ideals
	result.InitiativeMod = internal.InitiativeMod
	err = json.Unmarshal(internal.InventoryJson, &result.Inventory)
	if err != nil {
		return result, err
	}
	result.Languages = internal.Languages
	result.Level = internal.Level
	result.Money = struct {
		Cp uint `json:"cp"`
		Ep uint `json:"ep"`
		Gp uint `json:"gp"`
		Pp uint `json:"pp"`
		Sp uint `json:"sp"`
	}(internal.Money)
	result.Name = internal.Name
	result.Notes = internal.Notes
	result.Organizations = internal.Organizations
	result.Personality = internal.Personality
	result.Picture = internal.Picture
	result.Proficiencies = []api.CharacterObjectProficiencies{}
	for _, prof := range internal.Proficiencies {
		result.Proficiencies = append(result.Proficiencies, api.CharacterObjectProficiencies(prof))
	}
	result.ProficiencyMod = internal.ProficiencyMod
	result.Race = internal.Race
	err = json.Unmarshal(internal.RaceTraitsJson, &result.RaceTraits)
	if err != nil {
		return result, err
	}
	result.SavingThrowProf = []api.CharacterObjectSavingThrowProf{}
	for _, prof := range internal.SavingThrowProf {
		result.SavingThrowProf = append(result.SavingThrowProf, api.CharacterObjectSavingThrowProf(prof))
	}
	result.Senses = internal.Senses
	result.Shielded = internal.Shielded
	result.Size = api.CharacterObjectSize(internal.Size)
	result.Skin = internal.Skin
	result.Speed = internal.Speed
	err = json.Unmarshal(internal.Spells.CurrentConcentrationJson, &result.Spells.CurrentConcentration)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(internal.Spells.KnownSpellsJson, &result.Spells.KnownSpells)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(internal.Spells.ReadySpellsJson, &result.Spells.ReadySpells)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(internal.Spells.SlotsJson, &result.Spells.Slots)
	if err != nil {
		return result, err
	}
	result.ToolsProficiency = internal.ToolsProficiency
	result.WeaponProficiency = internal.WeaponProficiency
	result.Weight = internal.Weight
	result.Xp = internal.Xp
	result.Version = internal.Version
	return result, nil
}

func MapObjectToInternalCharacter(object api.CharacterObject) (InternalCharacter, error) {
	result := InternalCharacter{}
	result.Abilities = struct {
		Cha uint
		Con uint
		Dex uint
		Int uint
		Str uint
		Wis uint
	}(object.Abilities)
	result.Ac = object.Ac
	result.Age = object.Age
	result.Alignment = string(object.Alignment)
	result.Allies = object.Allies
	result.Appearance = object.Appearance
	result.ArmourProficiency = object.ArmourProficiency
	var err error
	result.AttacksJson, err = json.Marshal(object.Attacks)
	if err != nil {
		return result, err
	}
	result.Background = struct {
		Description string
		Name        string
	}(object.Background)
	result.Backstory = object.Backstory
	result.Bonds = object.Bonds
	result.BonusActionsJson, err = json.Marshal(object.BonusActions)
	if err != nil {
		return result, err
	}
	result.Capacity = object.Capacity
	result.ClassJson, err = json.Marshal(object.Class)
	if err != nil {
		return result, err
	}
	result.ClassFeaturesJson, err = json.Marshal(object.ClassFeatures)
	if err != nil {
		return result, err
	}
	result.Conditions = object.Conditions
	result.DeathSaves = struct {
		Failure uint
		Success uint
	}(object.DeathSaves)
	result.Defences.Custom = object.Defences.Custom
	result.Defences.Immune = object.Defences.Immune
	result.Defences.Resistance = object.Defences.Resistance
	result.Defences.Vulnerable = object.Defences.Vulnerable
	result.Enemies = object.Enemies
	result.Expertises = []string{}
	for _, expert := range object.Expertises {
		result.Expertises = append(result.Expertises, string(expert))
	}
	result.Eyes = object.Eyes
	result.Faith = object.Faith
	result.FeatsJson, err = json.Marshal(object.Feats)
	if err != nil {
		return result, err
	}
	result.Flaws = object.Flaws
	result.Gender = object.Gender
	result.Hair = object.Hair
	result.Height = object.Height
	result.Hp.Current = object.Hp.Current
	result.Hp.Max = object.Hp.Max
	result.Hp.Temp = object.Hp.Temp
	result.Hp.HitDice.Left = object.Hp.HitDice.Left
	result.Hp.HitDice.Max = object.Hp.HitDice.Max
	result.Hp.HitDice.Value = uint(object.Hp.HitDice.Value)
	result.Ideals = object.Ideals
	result.InitiativeMod = object.InitiativeMod
	result.InventoryJson, err = json.Marshal(object.Inventory)
	if err != nil {
		return result, err
	}
	result.Languages = object.Languages
	result.Level = object.Level
	result.Money = struct {
		Cp uint
		Ep uint
		Gp uint
		Pp uint
		Sp uint
	}(object.Money)
	result.Name = object.Name
	result.Notes = object.Notes
	result.Organizations = object.Organizations
	result.Personality = object.Personality
	result.Picture = object.Picture
	result.Proficiencies = []string{}
	for _, prof := range object.Proficiencies {
		result.Proficiencies = append(result.Proficiencies, string(prof))
	}
	result.ProficiencyMod = object.ProficiencyMod
	result.Race = object.Race
	result.RaceTraitsJson, err = json.Marshal(object.RaceTraits)
	if err != nil {
		return result, err
	}
	result.SavingThrowProf = []string{}
	for _, prof := range object.SavingThrowProf {
		result.SavingThrowProf = append(result.SavingThrowProf, string(prof))
	}
	result.Senses = object.Senses
	result.Shielded = object.Shielded
	result.Size = string(object.Size)
	result.Skin = object.Skin
	result.Speed = object.Speed
	result.Spells.CurrentConcentrationJson, err = json.Marshal(object.Spells.CurrentConcentration)
	if err != nil {
		return result, err
	}
	result.Spells.KnownSpellsJson, err = json.Marshal(object.Spells.KnownSpells)
	if err != nil {
		return result, err
	}
	result.Spells.ReadySpellsJson, err = json.Marshal(object.Spells.ReadySpells)
	if err != nil {
		return result, err
	}
	result.Spells.SlotsJson, err = json.Marshal(object.Spells.Slots)
	if err != nil {
		return result, err
	}
	result.ToolsProficiency = object.ToolsProficiency
	result.WeaponProficiency = object.WeaponProficiency
	result.Weight = object.Weight
	result.Xp = object.Xp
	result.Version = object.Version
	return result, nil
}
