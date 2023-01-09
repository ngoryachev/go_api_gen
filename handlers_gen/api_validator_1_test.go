package main

import (
	"reflect"
	"testing"
)

func GenericTest(t *testing.T, typeName, input, expected string) {
	fd := &FieldDef{TypeName: typeName}
	a := &FieldValidator{}
	fd.ValidatorMeta = a
	err := a.Parse(input)
	ve := fd.ToValidatorExpression()

	if err != nil {
		t.Errorf("%v", err)
	}

	if expected != ve {
		t.Errorf(`EXPECTED: "%v" but GIVEN: "%v"`, expected, ve)
	}
}

func TestCustomValidator1(t *testing.T) {
	GenericTest(t, "int", `apivalidator:"required"`, "required,type(int)")
	GenericTest(t, "string", `apivalidator:"required,min=10"`, "required,type(string),minstringlength(10)")
}

func TestCustomValidator2(t *testing.T) {
	GenericTest(t, "string", `apivalidator:"paramname=full_name"`, "type(string)")
}

func TestCustomValidator3(t *testing.T) {
	GenericTest(t, "string", `apivalidator:"enum=user|moderator|admin,default=user"`, "type(string),in(user|moderator|admin)")
}

func TestCustomValidator4(t *testing.T) {
	GenericTest(t, "string", `apivalidator:"min=0,max=5"`, "type(string),minstringlength(0),maxstringlength(5)")
	GenericTest(t, "string", `apivalidator:"min=0,max=5,required"`, "required,type(string),minstringlength(0),maxstringlength(5)")
	GenericTest(t, "int", `apivalidator:"min=0,max=5,required"`, "required,type(int),range(0|5)")
	GenericTest(t, "int", `apivalidator:"min=1,required"`, "required,type(int),range(1|)")
	GenericTest(t, "string", `apivalidator:"min=1,required"`, "required,type(string),minstringlength(1)")
	GenericTest(t, "int", `apivalidator:"max=5,required"`, "required,type(int),range(|5)")
	GenericTest(t, "string", `apivalidator:"max=5,required"`, "required,type(string),maxstringlength(5)")
}

//type ProfileParams struct {
//	Login string `apivalidator:"required"`
//}
func TestCustomValidator11(t *testing.T) {
	sd := &StructDef{
		Fields: []*FieldDef{{
			Name:     "Login",
			TypeName: "string",
			Tag:      `apivalidator:"required"`,
		}},
	}
	sd.ParseMeta()

	if !reflect.DeepEqual(sd.TemplateMap(), map[string]interface{}{
		"login": "required,type(string)",
	}) {
		t.Error("!reflect.DeepEqual")
	}
}

//type CreateParams struct {
//	Login  string `apivalidator:"required,min=10"`
//	Name   string `apivalidator:"paramname=full_name"`
//	Status string `apivalidator:"enum=user|moderator|admin,default=user"`
//	Age    int    `apivalidator:"min=0,max=128"`
//}
func TestCustomValidator12(t *testing.T) {
	sd := &StructDef{
		Fields: []*FieldDef{
			{
				Name:     "Login",
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
		},
	}
	sd.ParseMeta()

	if !reflect.DeepEqual(sd.TemplateMap(), map[string]interface{}{
		"login":     "required,type(string),minstringlength(10)",
		"full_name": "type(string)",
		"status":    "type(string),in(user|moderator|admin)",
		"age":       "type(int),range(0|128)",
	}) {
		t.Error("!reflect.DeepEqual")
	}
}

//type OtherCreateParams struct {
//	Username string `apivalidator:"required,min=3"`
//	Name     string `apivalidator:"paramname=account_name"`
//	Class    string `apivalidator:"enum=warrior|sorcerer|rouge,default=warrior"`
//	Level    int    `apivalidator:"min=1,max=50"`
//}
func TestCustomValidator13(t *testing.T) {
	sd := &StructDef{
		Fields: []*FieldDef{
			{
				Name:     "Username",
				TypeName: "string",
				Tag:      `apivalidator:"required,min=3"`,
			}, {
				Name:     "Name",
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

	if !reflect.DeepEqual(sd.TemplateMap(), map[string]interface{}{
		"username":     "required,type(string),minstringlength(3)",
		"account_name": "type(string)",
		"class":        "type(string),in(warrior|sorcerer|rouge)",
		"level":        "type(int),range(1|50)",
	}) {
		t.Error("!reflect.DeepEqual")
	}
}

func TestCustomValidator14(t *testing.T) {
	sd := &StructDef{
		Fields: []*FieldDef{{
			Name:     "Login",
			TypeName: "string",
			Tag:      `apivalidator:"required"`,
		}},
	}
	sd.ParseMeta()

	tms := sd.TemplateMapString()

	t.Log(tms)
}
