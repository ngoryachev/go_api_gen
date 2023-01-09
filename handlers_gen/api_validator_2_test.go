package main

import (
	"github.com/asaskevich/govalidator"
	"reflect"
	"strings"
	"testing"
)

func init() {
	//govalidator.SetFieldsRequiredByDefault(true)
}

func TestGoValidator1(t *testing.T) {
	var mapTemplate = map[string]interface{}{
		"name":       "required,alpha",
		"family":     "required,alpha",
		"email":      "required,email",
		"cell-phone": "numeric",
		"age":        "required,type(int)",
	}

	var inputMap = map[string]interface{}{
		"name":   "Bob",
		"family": "Smith",
		"email":  "foo@bar.baz",
		"age":    1,
	}

	result, err := govalidator.ValidateMap(inputMap, mapTemplate)
	if err != nil {
		t.Errorf("%v", err)
	}

	if !result {
		t.Error("!result")
	}
}

func TestGoValidator2(t *testing.T) {
	sd := &StructDef{
		Fields: []*FieldDef{
			{
				Name:     "Login",
				TypeName: "string",
				Tag:      `apivalidator:"required"`,
			},
			{
				Name:     "LoginMin",
				TypeName: "string",
				Tag:      `apivalidator:"required,min=10"`,
			}, {
				Name:     "Name",
				TypeName: "string",
				Tag:      `apivalidator:"paramname=full_name"`,
			},
			{
				Name:     "Status",
				TypeName: "string",
				Tag:      `apivalidator:"enum=user|moderator|admin,default=user"`,
			},
			{
				Name:     "Age",
				TypeName: "int",
				Tag:      `apivalidator:"min=0,max=128"`,
			},
			{
				Name:     "Username",
				TypeName: "string",
				Tag:      `apivalidator:"required,min=3"`,
			}, {
				Name:     "AccountName",
				TypeName: "string",
				Tag:      `apivalidator:"paramname=account_name"`,
			},
			{
				Name:     "Class",
				TypeName: "string",
				Tag:      `apivalidator:"enum=warrior|sorcerer|rouge,default=warrior"`,
			},
			{
				Name:     "Level",
				TypeName: "int",
				Tag:      `apivalidator:"min=1,max=50"`,
			},
		},
	}
	sd.ParseMeta()

	var mapTemplate = sd.TemplateMap()

	inputMap := sd.InputMap(t, map[string][]string{
		strings.ToLower("Login"):       {"l"},          // apivalidator:"required"
		strings.ToLower("LoginMin"):    {"loginLogin"}, // apivalidator:"required,min=10"
		strings.ToLower("FullName"):    {""},           // apivalidator:"paramname=full_name" //
		strings.ToLower("Status"):      {""},           // apivalidator:"enum=user|moderator|admin,default=user" //
		strings.ToLower("Age"):         {"42"},         // apivalidator:"min=0,max=128"
		strings.ToLower("Username"):    {"nick"},       // apivalidator:"required,min=3"
		strings.ToLower("AccountName"): {""},           // apivalidator:"paramname=account_name" //
		strings.ToLower("Class"):       {""},           // apivalidator:"enum=warrior|sorcerer|rouge,default=warrior" //
		strings.ToLower("Level"):       {""},           // apivalidator:"min=1,max=50"
	})

	t.Logf("inputMap: %v", inputMap)

	valid, err := govalidator.ValidateMap(inputMap, mapTemplate)
	if err != nil {
		t.Errorf("%v", err)
	}
	if !valid {
		t.Error("!valid")
	}
}

func TestGoValidatorPositive(t *testing.T) {
	valid, err := govalidator.ValidateMap(
		map[string]interface{}{
			"name":   "Bob",
			"family": "Smith",
			"email":  "foo@bar.baz",
			"age":    1,
		},
		map[string]interface{}{
			"name":       "required,alpha",
			"family":     "required,alpha",
			"email":      "required,email",
			"cell-phone": "numeric",
			"age":        "required,type(int)",
		},
	)

	if !valid {
		errs := err.(govalidator.Errors).Errors()
		for _, e := range errs {
			t.Error(e.Error())
		}
	}
}

func TestGoValidatorNegative(t *testing.T) {
	valid, err := govalidator.ValidateMap(
		map[string]interface{}{
			"name1": "", // required
			"name3": 1,  // wrong type
			"name4": "", // not error, cause not required
		},
		map[string]interface{}{
			"name1": "required,type(string)",
			"name2": "required,type(string)", // missing in input = error
			"name3": "required,type(string)",
			"name4": "type(string)",
		},
	)

	if !valid {
		errs := err.(govalidator.Errors).Errors()
		for _, e := range errs {
			t.Log(e.Error())
		}
	}
}

func Map(vs []*FieldDef, f func(*FieldDef) InputValue) []InputValue {
	vsm := make([]InputValue, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}

func TestGoValidatorString(t *testing.T) {
	sds := &StructDef{
		Fields: []*FieldDef{
			{
				Name:     "Login",
				TypeName: "string",
				Tag:      `apivalidator:"required"`,
			},
			{
				Name:     "LoginMin",
				TypeName: "string",
				Tag:      `apivalidator:"required,min=10"`,
			}, {
				Name:     "Name",
				TypeName: "string",
				Tag:      `apivalidator:"paramname=full_name"`,
			},
			{
				Name:     "Status",
				TypeName: "string",
				Tag:      `apivalidator:"enum=user|moderator|admin,default=user"`,
			},
			{
				Name:     "Age",
				TypeName: "int",
				Tag:      `apivalidator:"min=0,max=128"`,
			},
			{
				Name:     "Username",
				TypeName: "string",
				Tag:      `apivalidator:"required,min=3"`,
			}, {
				Name:     "AccountName",
				TypeName: "string",
				Tag:      `apivalidator:"paramname=account_name"`,
			},
			{
				Name:     "Class",
				TypeName: "string",
				Tag:      `apivalidator:"enum=warrior|sorcerer|rouge,default=warrior"`,
			},
			{
				Name:     "Level",
				TypeName: "int",
				Tag:      `apivalidator:"min=1,max=50"`,
			},
		},
	}

	sds.ParseMeta()

	rVals := map[string][]string{
		strings.ToLower("Login"):       {"l"},          // apivalidator:"required"
		strings.ToLower("LoginMin"):    {"loginLogin"}, // apivalidator:"required,min=10"
		strings.ToLower("FullName"):    {""},           // apivalidator:"paramname=full_name" //
		strings.ToLower("Status"):      {""},           // apivalidator:"enum=user|moderator|admin,default=user" //
		strings.ToLower("Age"):         {"42"},         // apivalidator:"min=0,max=128"
		strings.ToLower("Username"):    {"nick"},       // apivalidator:"required,min=3"
		strings.ToLower("AccountName"): {""},           // apivalidator:"paramname=account_name" //
		strings.ToLower("Class"):       {""},           // apivalidator:"enum=warrior|sorcerer|rouge,default=warrior" //
		strings.ToLower("Level"):       {""},           // apivalidator:"min=1,max=50"
	}

	inputMap := InputMap(Map(sds.Fields, func(def *FieldDef) InputValue {
		return InputValue{
			ParamName:  ParamName(def.ValidatorMeta.ParamName, def.Name),
			Def:        def.ValidatorMeta.Default,
			TypeName:   def.TypeName,
			HasDefault: def.ValidatorMeta.HasDefault(),
		}
	}),
		rVals,
	)

	inputMap2 := sds.InputMap(t, rVals)

	t.Logf("inputMap: %v", inputMap)

	if !reflect.DeepEqual(inputMap, inputMap2) {
		t.Error("!reflect.DeepEqual(inputMap, inputMap2)")
	}
}
