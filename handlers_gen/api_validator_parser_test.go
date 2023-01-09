package main

import (
	"reflect"
	"testing"
)

func all(cond []bool) bool {
	var ret bool

	if len(cond) > 0 {
		ret = cond[0]
	} else {
		ret = false
	}

	for _, v := range cond {
		ret = ret && v
	}

	return ret
}

func ApiValidatorGeneric(t *testing.T, s string, provider func(fv *FieldValidator) []bool) {
	a := &FieldValidator{}
	err := a.Parse(s)

	cond := provider(a)

	if err != nil || !all(cond) {
		t.Errorf("%v", cond)
	}
}

//Login string `apivalidator:"required"`
//Login  string `apivalidator:"required,min=10"`
func TestApiValidatorRequired(t *testing.T) {
	ApiValidatorGeneric(t, `apivalidator:"required"`, func(fv *FieldValidator) []bool {
		return []bool{fv.Required == true}
	})
	ApiValidatorGeneric(t, `apivalidator:"required,min=10"`, func(fv *FieldValidator) []bool {
		return []bool{
			fv.Required == true,
			fv.Min == 10,
		}
	})
}

//Name   string `apivalidator:"paramname=full_name"`
//Name     string `apivalidator:"paramname=account_name"`
func TestApiValidatorParamName(t *testing.T) {
	ApiValidatorGeneric(t, `apivalidator:"paramname=full_name"`, func(fv *FieldValidator) []bool {
		return []bool{
			fv.ParamName == "full_name",
		}
	})
	ApiValidatorGeneric(t, `apivalidator:"paramname=account_name"`, func(fv *FieldValidator) []bool {
		return []bool{
			fv.ParamName == "account_name",
		}
	})
}

//Status string `apivalidator:"enum=user|moderator|admin,default=user"`
//Class    string `apivalidator:"enum=warrior|sorcerer|rouge,default=warrior"`
func TestApiValidatorEnum(t *testing.T) {
	ApiValidatorGeneric(t, `apivalidator:"enum=user|moderator|admin,default=user"`, func(fv *FieldValidator) []bool {
		return []bool{
			fv.Default == "user",
			reflect.DeepEqual([]string{
				"user",
				"moderator",
				"admin",
			}, fv.Enum),
		}
	})
}

//Age    int    `apivalidator:"min=0,max=128"`
//Level    int    `apivalidator:"min=1,max=50"`
//Level    int    `apivalidator:"min=1"`
func TestApiValidatorMinMax(t *testing.T) {
	ApiValidatorGeneric(t, `apivalidator:"min=0,max=128"`, func(fv *FieldValidator) []bool {
		return []bool{
			fv.IsMin == true,
			fv.IsMax == true,
			fv.Min == 0,
			fv.Max == 128,
		}
	})
	ApiValidatorGeneric(t, `apivalidator:"min=1,max=50"`, func(fv *FieldValidator) []bool {
		return []bool{
			fv.IsMin == true,
			fv.IsMax == true,
			fv.Min == 1,
			fv.Max == 50,
		}
	})
	ApiValidatorGeneric(t, `apivalidator:"min=1"`, func(fv *FieldValidator) []bool {
		return []bool{
			fv.IsMin == true,
			fv.IsMax == false,
			fv.Min == 1,
			fv.Max == 0,
		}
	})
}

func TestApiValidatorDefault(t *testing.T) {
	ApiValidatorGeneric(t, `apivalidator:"min=1,max=2,default=3,enum=ok|nok"`, func(fv *FieldValidator) []bool {
		return []bool{
			fv.Default == "3",
		}
	})
}
