package model

import (
	"dnd_back/api"
	"encoding/json"

	_ "github.com/jinzhu/gorm"
)

type InternalSpell struct {
	ID         uint `gorm:"primary_key"`
	Components struct {
		Material struct {
			Components string
			Status     bool
		} `gorm:"embedded;embedded_prefix:components_material_"`
		Somatic bool
		Verbal  bool
	} `gorm:"embedded;embedded_prefix:components"`
	Concentration bool
	Description   string
	Duration      struct {
		Duration     int
		DurationType string
	} `gorm:"embedded;embedded_prefix:duration"`
	Effect struct {
		Ability string
		Dice    struct {
			Amount uint
			Bonus  uint
			Value  uint
		} `gorm:"embedded;embedded_prefix:effect_dice"`
		DmgType string
		Effect  string
	} `gorm:"embedded;embedded_prefix:effect"`
	Level uint
	Name  string
	Range struct {
		Amount    int
		Notes     string
		RangeType string
	} `gorm:"embedded;embedded_prefix:range"`
	Ritual bool
	School string
	Target struct {
		Size       uint
		TargetType string
	} `gorm:"embedded;embedded_prefix:target"`
	TimeToCast struct {
		Amount uint
		Notes  string
		Time   string
	} `gorm:"embedded;embedded_prefix:time_to_cast"`
	Upcast struct {
		Dice struct {
			Amount uint
			Value  uint
		} `gorm:"embedded;embedded_prefix:upcast_dice"`
		UpcastType string
	} `gorm:"embedded;embedded_prefix:upcast"`
}

func MapInternalSpellToObject(internal InternalSpell) api.SpellObject {
	result := api.SpellObject{}
	result.Components = struct {
		Material struct {
			Components string `json:"components"`
			Status     bool   `json:"status"`
		} `json:"material"`
		Somatic bool `json:"somatic"`
		Verbal  bool `json:"verbal"`
	}(internal.Components)
	result.Concentration = internal.Concentration
	result.Description = internal.Description
	result.Duration.Duration = internal.Duration.Duration
	result.Duration.DurationType = api.SpellObjectDurationDurationType(internal.Duration.DurationType)
	result.Effect.Ability = api.SpellObjectEffectAbility(internal.Effect.Ability)
	result.Effect.Dice.Amount = internal.Effect.Dice.Amount
	result.Effect.Dice.Bonus = internal.Effect.Dice.Bonus
	result.Effect.Dice.Value = api.SpellObjectEffectDiceValue(internal.Effect.Dice.Value)
	result.Level = internal.Level
	result.Name = internal.Name
	result.Range.Amount = internal.Range.Amount
	result.Range.Notes = internal.Range.Notes
	result.Range.RangeType = api.SpellObjectRangeRangeType(internal.Range.RangeType)
	result.Ritual = internal.Ritual
	result.School = internal.School
	result.Target = struct {
		Size       uint   `json:"size"`
		TargetType string `json:"target_type"`
	}(internal.Target)
	result.TimeToCast.Amount = internal.TimeToCast.Amount
	result.TimeToCast.Notes = internal.TimeToCast.Notes
	result.TimeToCast.Time = api.SpellObjectTimeToCastTime(internal.TimeToCast.Time)
	result.Upcast.UpcastType = api.SpellObjectUpcastUpcastType(internal.Upcast.UpcastType)
	result.Upcast.Dice.Value = api.SpellObjectUpcastDiceValue(internal.Upcast.Dice.Value)
	result.Upcast.Dice.Amount = internal.Upcast.Dice.Amount
	return result
}

func MapObjectToInternalSpell(object api.SpellObject) InternalSpell {
	result := InternalSpell{}
	result.Components = struct {
		Material struct {
			Components string
			Status     bool
		} `gorm:"embedded;embedded_prefix:components_material_"`
		Somatic bool
		Verbal  bool
	}(object.Components)
	result.Concentration = object.Concentration
	result.Description = object.Description
	result.Duration.Duration = object.Duration.Duration
	result.Duration.DurationType = string(object.Duration.DurationType)
	result.Effect.Ability = string(object.Effect.Ability)
	result.Effect.Dice.Amount = object.Effect.Dice.Amount
	result.Effect.Dice.Bonus = object.Effect.Dice.Bonus
	result.Effect.Dice.Value = uint(object.Effect.Dice.Value)
	result.Level = object.Level
	result.Name = object.Name
	result.Range.Amount = object.Range.Amount
	result.Range.Notes = object.Range.Notes
	result.Range.RangeType = string(object.Range.RangeType)
	result.Ritual = object.Ritual
	result.School = object.School
	result.Target = struct {
		Size       uint
		TargetType string
	}(object.Target)
	result.TimeToCast.Amount = object.TimeToCast.Amount
	result.TimeToCast.Notes = object.TimeToCast.Notes
	result.TimeToCast.Time = string(object.TimeToCast.Time)
	result.Upcast.UpcastType = string(object.Upcast.UpcastType)
	result.Upcast.Dice.Value = uint(object.Upcast.Dice.Value)
	result.Upcast.Dice.Amount = object.Upcast.Dice.Amount
	return result
}

type InternalItem struct {
	ID       uint `gorm:"primary_key"`
	Cost     uint
	DataJson []byte
	ItemType string
	Name     string
	Weight   float32
}

func MapInternalItemToObject(internal InternalItem) (api.ItemObject, error) {
	result := api.ItemObject{}
	result.Name = internal.Name
	result.Cost = internal.Cost
	result.ItemType = api.ItemObjectItemType(internal.ItemType)
	result.Weight = internal.Weight
	err := json.Unmarshal([]byte(internal.DataJson), &result.Data)
	if err != nil {
		return result, err
	}
	return result, nil
}

func MapObjectToInternalItem(object api.ItemObject) (InternalItem, error) {
	result := InternalItem{}
	result.Name = object.Name
	result.Cost = object.Cost
	result.ItemType = string(object.ItemType)
	result.Weight = object.Weight
	var err error
	result.DataJson, err = json.Marshal(object.Data)
	if err != nil {
		return result, err
	}
	return result, nil
}
