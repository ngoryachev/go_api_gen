package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"text/template"
)

// HELPERS

func groupFunctionsByReceiver(fs []*FuncDef) map[string][]*FuncDef {
	ret := map[string][]*FuncDef{}
	for _, f := range fs {
		rn := f.ReceiverName

		ret[rn] = append(ret[rn], f)
	}

	return ret
}

func associateFuncArgumentStruct(functions []*FuncDef, defs []*StructDef) *StructDef {
	for _, f := range functions {
		for _, d := range defs {
			if f.ArgumentTypeName == d.Name {
				f.ArgumentStruct = d
			}
		}
	}

	return nil
}

// EOF HELPERS

func genServeHTTP(receiverName string, funcDefs []*FuncDef) string {
	tmpl := template.Must(template.ParseFiles("handlers_gen/serve_http.template"))
	w := bytes.NewBufferString("")
	tmpl.Execute(w,
		struct {
			ReceiverName string
			FuncDefs     []*FuncDef
		}{
			ReceiverName: receiverName,
			FuncDefs:     funcDefs,
		})

	return w.String()
}

func genHandlers(receiverName string, funcDefs []*FuncDef) string {
	tmpl := template.Must(template.ParseFiles("handlers_gen/handle_method.template"))
	w := bytes.NewBufferString("")
	tmpl.Execute(w,
		struct {
			ReceiverName string
			FuncDefs     []*FuncDef
		}{
			ReceiverName: receiverName,
			FuncDefs:     funcDefs,
		})

	return w.String()
}

func genByTemplate(templatePath string, vars interface{}) string {
	tmpl := template.Must(template.ParseFiles("handlers_gen/" + templatePath))
	w := bytes.NewBufferString("")
	tmpl.Execute(w, vars)

	return w.String()
}

type ApiGenArgs struct {
	Url    string `json:"url"`
	Auth   bool   `json:"auth"`
	Method string `json:"method"`
}

func (args *ApiGenArgs) String() string {
	return fmt.Sprintf("{Url: %v, Auth: %v, Method: %v\n}", args.Url, args.Auth, args.Method)
}

func (args *ApiGenArgs) NoMethod() bool {
	return args.Method == ""
}

func (args *ApiGenArgs) SomeMethod() bool {
	return args.Method != ""
}

func (args *ApiGenArgs) Parse(s string) {
	ss := strings.TrimLeft(s, "apigen:api ")
	data := []byte(ss)
	json.Unmarshal(data, args)

	//if args.Method == "" {
	//	args.Method = "GET"
	//}
}

type FieldValidator struct {
	Parsed bool
	// поле не должно быть пустым (не должно иметь значение по-умолчанию)
	Required bool
	// если указано - то брать из параметра с этим именем, иначе lowercase от имени
	ParamName string
	// "одно из"
	Enum []string
	// если указано и приходит пустое значение (значение по-умолчанию) - устанавливать то что написано указано в default
	Default string
	IsMin   bool
	IsMax   bool
	//для типа int
	//>= X
	//для строк
	//len(str) >=
	Min int
	Max int
}

func (validator *FieldValidator) HasDefault() bool {
	return len(validator.Default) > 0
}

func (validator *FieldValidator) Parse(s string) error {
	if !strings.Contains(s, "apivalidator:") {
		return nil
	}

	body := s[len(`"apivalidator:"`)-1 : len(s)-1]
	xs := strings.Split(body, ",")

	for _, x := range xs {
		bundle := strings.Split(x, "=")

		if len(bundle) < 1 {
			return fmt.Errorf("no key")
		}

		if len(bundle) < 2 && bundle[0] != "required" {
			return fmt.Errorf("no value for " + bundle[0])
		}

		switch bundle[0] {
		// required
		//paramname
		//enum
		//default
		//min
		//max
		case "required":
			validator.Required = true
		case "paramname":
			validator.ParamName = bundle[1]
		case "enum":
			split := strings.Split(bundle[1], "|")

			if len(split) < 1 {
				return fmt.Errorf("no enum values")
			}

			validator.Enum = split
		case "default":
			validator.Default = bundle[1]
		case "min":
			min, e := strconv.Atoi(bundle[1])

			if e != nil {
				return e
			}

			validator.IsMin = true
			validator.Min = min
		case "max":
			max, e := strconv.Atoi(bundle[1])

			if e != nil {
				return e
			}

			validator.IsMax = true
			validator.Max = max
		}
	}

	validator.Parsed = true

	return nil
}

type FuncDef struct {
	CommentText      string
	ApiArgs          *ApiGenArgs
	ReceiverName     string
	MethodName       string
	ArgumentName     string
	ArgumentTypeName string
	ArgumentStruct   *StructDef
	ResulTypeName    string
}

func (p *FuncDef) TemplateMapString() string {
	return p.ArgumentStruct.TemplateMapString()
}

type FieldDef struct {
	Name          string
	TypeName      string
	Tag           reflect.StructTag
	ValidatorMeta *FieldValidator
	Value         interface{}
}

func (def *FieldDef) ParamName() string {
	var key string

	if len(def.ValidatorMeta.ParamName) > 0 {
		key = def.ValidatorMeta.ParamName
	} else {
		key = strings.ToLower(def.Name)
	}

	return key
}

func (def *FieldDef) GenInputValue() string {
	return fmt.Sprintf(`{ ParamName:  "%s", Def: "%s", TypeName: "%s", HasDefault: %v }`,
		ParamName(def.ValidatorMeta.ParamName, def.Name),
		def.ValidatorMeta.Default,
		def.TypeName,
		def.ValidatorMeta.HasDefault(),
	)
}

func (def *FieldDef) GenParamKeyVal() string {
	return fmt.Sprintf(`%s:  inputMap["%s"].(%s)`,
		def.Name,
		ParamName(def.ValidatorMeta.ParamName, def.Name),
		def.TypeName,
	)
}

type StructDef struct {
	Name   string
	Fields []*FieldDef
}

func (def *StructDef) ParseMeta() error {
	for _, f := range def.Fields {
		t := f.Tag
		f.ValidatorMeta = &FieldValidator{}
		e := f.ValidatorMeta.Parse(string(t))

		if e != nil {
			return e
		}
	}

	return nil
}

func (def *StructDef) TemplateMap() map[string]interface{} {
	ret := map[string]interface{}{}
	for _, f := range def.Fields {
		ret[f.ParamName()] = f.ToValidatorExpression()
	}

	return ret
}

// TemplateMapString should be according to TemplateMap
func (def *StructDef) TemplateMapString() string {
	w := bytes.NewBufferString("")
	w.WriteString("map[string]interface{}{")

	for _, f := range def.Fields {
		fmt.Fprintf(w, `"%s": "%s",`, f.ParamName(), f.ToValidatorExpression())
	}

	w.WriteString("}")

	return w.String()
}

func (def *StructDef) InputMap(t *testing.T, values url.Values) map[string]interface{} {
	ret := map[string]interface{}{}
	for _, f := range def.Fields {
		val, _ := f.ToInputValue(t, values)

		if val != nil {
			ret[f.ParamName()] = val
		}
	}

	return ret
}

// GENERATE PART

func ToInputValue(paramName, def, typeName string, hasDefault bool, values url.Values) (interface{}, error) {
	var ret interface{}

	has := values.Has(paramName)
	sv := values.Get(paramName)

	if has && len(sv) < 1 {
		has = false
	}

	if !has {
		if hasDefault {
			sv = def
		} else {
			return nil, nil
		}
	}

	switch typeName {
	case "string":
		ret = sv
	case "int":
		fallthrough
	case "uint64":
		var err error
		ret, err = strconv.Atoi(sv)

		if err != nil {
			return nil, fmt.Errorf("!strconv.Atoi(sv)")
		}
	}

	return ret, nil
}

func ParamName(paramName, name string) string {
	var key string

	if len(paramName) > 0 {
		key = paramName
	} else {
		key = strings.ToLower(name)
	}

	return key
}

type InputValue struct {
	ParamName  string
	Def        string
	TypeName   string
	HasDefault bool
}

func InputMap(fields []InputValue, values url.Values) map[string]interface{} {
	ret := map[string]interface{}{}
	for _, f := range fields {
		val, _ := ToInputValue(f.ParamName, f.Def, f.TypeName, f.HasDefault, values)

		if val != nil {
			ret[f.ParamName] = val
		}
	}

	return ret
}

// EOF GENERATE PART

func (def *FieldDef) String() string {
	return fmt.Sprintf("%s %s %s\n", def.Name, def.TypeName, def.Tag)
}

func (def *FieldDef) ToValidatorExpression() string {
	if !def.ValidatorMeta.Parsed {
		return "-"
	}

	var sb []string

	if def.ValidatorMeta.Required {
		sb = append(sb, "required")
	}

	sb = append(sb, fmt.Sprintf("type(%s)", def.TypeName))

	if len(def.ValidatorMeta.Enum) > 0 {
		sb = append(sb, "in("+strings.Join(def.ValidatorMeta.Enum, "|")+")")
	}

	if def.ValidatorMeta.IsMin {
		switch def.TypeName {
		case "string":
			sb = append(sb, fmt.Sprintf("minstringlength(%d)", def.ValidatorMeta.Min))
		case "int":
			fallthrough
		case "uint64":
			if !def.ValidatorMeta.IsMax {
				sb = append(sb, fmt.Sprintf("range(%d|)", def.ValidatorMeta.Min))
			} else {
				sb = append(sb, fmt.Sprintf("range(%d|%d)", def.ValidatorMeta.Min, def.ValidatorMeta.Max))
			}
		}
	}

	if def.ValidatorMeta.IsMax {
		switch def.TypeName {
		case "string":
			sb = append(sb, fmt.Sprintf("maxstringlength(%d)", def.ValidatorMeta.Max))
		case "int":
			fallthrough
		case "uint64":
			if !def.ValidatorMeta.IsMin {
				sb = append(sb, fmt.Sprintf("range(|%d)", def.ValidatorMeta.Max))
			}
		}
	}

	return strings.Join(sb, ",")
}

func (def *FieldDef) ToInputValue(t *testing.T, values url.Values) (interface{}, error) {
	var ret interface{}

	if !def.ValidatorMeta.Parsed {
		return nil, fmt.Errorf("!def.ValidatorMeta.Parsed")
	}

	paramName := def.ParamName()
	has := values.Has(paramName)
	sv := values.Get(paramName)

	if has && len(sv) < 1 {
		has = false
	}

	t.Logf("paramName: %v", paramName)
	t.Logf("has: %v", has)
	t.Logf("sv: %v", sv)

	if !has {
		if def.ValidatorMeta.HasDefault() {
			sv = def.ValidatorMeta.Default
		} else {
			return nil, nil
		}
	}

	switch def.TypeName {
	case "string":
		ret = sv
	case "int":
		fallthrough
	case "uint64":
		var err error
		ret, err = strconv.Atoi(sv)

		if err != nil {
			return nil, fmt.Errorf("!strconv.Atoi(sv)")
		}
	}

	return ret, nil
}

func inspectFuncSignature(f ast.Node, funcCall *FuncDef) {
	ast.Inspect(f, func(n ast.Node) bool {
		if fd, ok := n.(*ast.FuncDecl); ok {
			funcCall.MethodName = fd.Name.Name
			funcCall.CommentText = fd.Doc.Text()
			funcCall.ApiArgs = &ApiGenArgs{}
			funcCall.ApiArgs.Parse(funcCall.CommentText)

			// param type
			for _, param := range fd.Type.Params.List {
				funcCall.ArgumentName = fmt.Sprintf("%s\n", param.Names[0])
				funcCall.ArgumentTypeName = fmt.Sprintf("%+v", param.Type)
			}

			// result type
			for _, result := range fd.Type.Results.List {
				switch recType := result.Type.(type) {
				case *ast.StarExpr:
					if si, ok := recType.X.(*ast.Ident); ok {
						funcCall.ResulTypeName = si.Name
					}
				case *ast.Ident:
					funcCall.ResulTypeName = recType.Name
				}
				break // only first iteration
			}

			for _, receiver := range fd.Recv.List {
				switch recType := receiver.Type.(type) {
				case *ast.StarExpr:
					if si, ok := recType.X.(*ast.Ident); ok {
						funcCall.ReceiverName = si.Name
					}
				case *ast.Ident:
					funcCall.ReceiverName = recType.Name
				}
			}
		}
		return true
	})
}

//func populateStructs(s []*StructDef, f []*FuncDef) *StructDef {
//	for _, v := range s {
//		if v.Name == name {
//			return v
//		}
//	}
//
//	return nil
//}

func main() {
	inFileName := os.Args[1]
	outFileName := os.Args[2]

	fileSet := token.NewFileSet()

	node, err := parser.ParseFile(fileSet, inFileName, nil, parser.ParseComments)

	if err != nil {
		log.Fatal(err)
	}

	outFile, _ := os.Create(outFileName)

	fmt.Fprintln(outFile, `package `+node.Name.Name)
	fmt.Fprintln(outFile)                                               // empty line
	fmt.Fprintln(outFile, `import "encoding/json"`)                     // empty line
	fmt.Fprintln(outFile, `import "fmt"`)                               // empty line
	fmt.Fprintln(outFile, `import "github.com/asaskevich/govalidator"`) // empty line
	fmt.Fprintln(outFile, `import "net/http"`)                          // empty line
	fmt.Fprintln(outFile, `import "net/url"`)                           // empty line
	fmt.Fprintln(outFile, `import "strconv"`)                           // empty line
	fmt.Fprintln(outFile, `import "strings"`)                           // empty line
	fmt.Fprintln(outFile)

	var funcCalls []*FuncDef
	var structs []*StructDef

	// BadDecl | FuncDecl | GenDecl
	for _, d := range node.Decls {
		switch d.(type) {
		case *ast.BadDecl:
		case *ast.FuncDecl:
			f, _ := d.(*ast.FuncDecl)

			var funcCall = &FuncDef{}

			if f.Recv != nil && strings.Contains(f.Doc.Text(), "apigen:api") {
				fmt.Println("M")

				inspectFuncSignature(f, funcCall)

				funcCalls = append(funcCalls, funcCall)
			}
		case *ast.GenDecl:
			fmt.Println("G")

			g, _ := d.(*ast.GenDecl)

		SPECS_LOOP:
			// *ImportSpec | *ValueSpec | *TypeSpec
			for _, spec := range g.Specs {
				currType, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				currStruct, ok := currType.Type.(*ast.StructType)
				if !ok {
					continue
				}

				struc := &StructDef{Name: currType.Name.Name}
			FIELDS_LOOP:
				for i, field := range currStruct.Fields.List {
					if field.Tag == nil {
						continue SPECS_LOOP
					}

					tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
					apiValidatorTag := tag.Get("apivalidator")
					jsonTag := tag.Get("json")
					empty := ""
					apiValidatorTagNotEmpty := apiValidatorTag != empty
					jsonTagNotEmpty := jsonTag != empty

					if !apiValidatorTagNotEmpty && !jsonTagNotEmpty {
						continue FIELDS_LOOP
					}

					fieldName := field.Names[0].Name
					fieldType := field.Type.(*ast.Ident).Name

					switch fieldType {
					case "int":
					case "uint64":
					case "string":
					default:
						log.Fatalln("unsupported", fieldType)
					}

					field := &FieldDef{
						Name:          fieldName,
						TypeName:      fieldType,
						Tag:           tag,
						ValidatorMeta: &FieldValidator{},
					}

					stag := string(field.Tag)

					if err := field.ValidatorMeta.Parse(stag); err != nil {
						log.Fatalln("ParsedValidator.Parse failed for tag " + stag)
					}

					struc.Fields = append(struc.Fields, field)

					if i == len(currStruct.Fields.List)-1 {
						structs = append(structs, struc)
					}
				}
			}
		}
	}

	associateFuncArgumentStruct(funcCalls, structs)

	//fmt.Println("METHODS")
	//for _, fc := range funcCalls {
	//	fmt.Printf("%+v\n", fc)
	//}
	//
	//fmt.Println()
	//
	//fmt.Println("STRUCTS")
	//for _, x := range structs {
	//	fmt.Printf("%+v\n", x)
	//}

	grouped := groupFunctionsByReceiver(funcCalls)

	for k, v := range grouped {
		fmt.Fprintln(outFile, genServeHTTP(k, v))
	}

	for k, v := range grouped {
		fmt.Fprintln(outFile, genHandlers(k, v))
	}

	fmt.Fprintln(outFile, genByTemplate("middleware.template", nil))
}
