package main

import "encoding/json"
import "fmt"
import "github.com/asaskevich/govalidator"
import "net/http"
import "net/url"
import "strconv"
import "strings"

func (srv *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if "/user/profile" == r.URL.Path {
		if true {

			errorMiddleware(http.HandlerFunc(srv.handleProfile)).ServeHTTP(w, r)

			return
		} else {
			handleServerError(w, http.StatusNotAcceptable, fmt.Errorf("bad method"))

			return
		}
	}

	if "/user/create" == r.URL.Path {
		if "POST" == r.Method {

			errorMiddleware(authMiddleware(http.HandlerFunc(srv.handleCreate))).ServeHTTP(w, r)

			return
		} else {
			handleServerError(w, http.StatusNotAcceptable, fmt.Errorf("bad method"))

			return
		}
	}

	handleServerError(w, http.StatusNotFound, fmt.Errorf("unknown method"))
}
func (srv *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if "/user/create" == r.URL.Path {
		if "POST" == r.Method {

			errorMiddleware(authMiddleware(http.HandlerFunc(srv.handleCreate))).ServeHTTP(w, r)

			return
		} else {
			handleServerError(w, http.StatusNotAcceptable, fmt.Errorf("bad method"))

			return
		}
	}

	handleServerError(w, http.StatusNotFound, fmt.Errorf("unknown method"))
}

func (srv *MyApi) handleProfile(w http.ResponseWriter, r *http.Request) {
	templateMap := map[string]interface{}{"login": "required,type(string)"}
	inputValues := []InputValue{

		{ParamName: "login", Def: "", TypeName: "string", HasDefault: false},
	}
	r.ParseForm()
	inputMap, e := InputMap(inputValues, r.Form)
	if e != nil {
		handleServerError(w, http.StatusBadRequest, e)

		return
	}

	valid, err := govalidator.ValidateMap(inputMap, templateMap)

	if !valid {
		handleServerError(w, http.StatusBadRequest, err)

		return
	}

	v, err := srv.Profile(r.Context(), ProfileParams{

		Login: inputMap["login"].(string),
	})

	if err != nil {
		handleServerError(w, err.(ApiError).HTTPStatus, err)

		return
	}

	handleServerResponse(w, v)
}

func (srv *MyApi) handleCreate(w http.ResponseWriter, r *http.Request) {
	templateMap := map[string]interface{}{"login": "required,type(string),minstringlength(10)", "full_name": "type(string)", "status": "type(string),in(user|moderator|admin)", "age": "type(int),range(0|128)"}
	inputValues := []InputValue{

		{ParamName: "login", Def: "", TypeName: "string", HasDefault: false},
		{ParamName: "full_name", Def: "", TypeName: "string", HasDefault: false},
		{ParamName: "status", Def: "user", TypeName: "string", HasDefault: true},
		{ParamName: "age", Def: "", TypeName: "int", HasDefault: false},
	}
	r.ParseForm()
	inputMap, e := InputMap(inputValues, r.Form)
	if e != nil {
		handleServerError(w, http.StatusBadRequest, e)

		return
	}

	valid, err := govalidator.ValidateMap(inputMap, templateMap)

	if !valid {
		handleServerError(w, http.StatusBadRequest, err)

		return
	}

	v, err := srv.Create(r.Context(), CreateParams{

		Login:  inputMap["login"].(string),
		Name:   inputMap["full_name"].(string),
		Status: inputMap["status"].(string),
		Age:    inputMap["age"].(int),
	})

	if err != nil {
		handleServerError(w, err.(ApiError).HTTPStatus, err)

		return
	}

	handleServerResponse(w, v)
}

func (srv *OtherApi) handleCreate(w http.ResponseWriter, r *http.Request) {
	templateMap := map[string]interface{}{"username": "required,type(string),minstringlength(3)", "account_name": "type(string)", "class": "type(string),in(warrior|sorcerer|rouge)", "level": "type(int),range(1|50)"}
	inputValues := []InputValue{

		{ParamName: "username", Def: "", TypeName: "string", HasDefault: false},
		{ParamName: "account_name", Def: "", TypeName: "string", HasDefault: false},
		{ParamName: "class", Def: "warrior", TypeName: "string", HasDefault: true},
		{ParamName: "level", Def: "", TypeName: "int", HasDefault: false},
	}
	r.ParseForm()
	inputMap, e := InputMap(inputValues, r.Form)
	if e != nil {
		handleServerError(w, http.StatusBadRequest, e)

		return
	}

	valid, err := govalidator.ValidateMap(inputMap, templateMap)

	if !valid {
		handleServerError(w, http.StatusBadRequest, err)

		return
	}

	v, err := srv.Create(r.Context(), OtherCreateParams{

		Username: inputMap["username"].(string),
		Name:     inputMap["account_name"].(string),
		Class:    inputMap["class"].(string),
		Level:    inputMap["level"].(int),
	})

	if err != nil {
		handleServerError(w, err.(ApiError).HTTPStatus, err)

		return
	}

	handleServerResponse(w, v)
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("authMiddleware", r.URL.Path)
		xAuthHeader := r.Header.Get("X-Auth")

		if xAuthHeader != "100500" {
			fmt.Println("no auth at", r.URL.Path)

			handleServerError(w, http.StatusForbidden, fmt.Errorf("unauthorized"))

			return
		}
		next.ServeHTTP(w, r)
	})
}

func errorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("errorMiddleware", r.URL.Path)
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("recovered", err)

				e := fmt.Errorf("%s", err)
				handleServerError(w, http.StatusInternalServerError, e)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type ServerResponse struct {
	Error    string      `json:"error"`
	Response interface{} `json:"response,omitempty"`
}

func (sr ServerResponse) Marshal() []byte {
	b, _ := json.Marshal(sr)

	return b
}

func handleServerError(w http.ResponseWriter, httpStatus int, err error) {
	w.WriteHeader(httpStatus)
	w.Write(ServerResponse{
		Error: mapError(ApiError{
			httpStatus,
			err,
		}.Error()),
	}.Marshal())
}

func handleServerResponse(w http.ResponseWriter, response interface{}) {
	//w.WriteHeader(http.StatusOK)
	w.Write(ServerResponse{
		Error:    "",
		Response: response,
	}.Marshal())
}

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

func InputMap(fields []InputValue, values url.Values) (map[string]interface{}, error) {
	ret := map[string]interface{}{}
	for _, f := range fields {
		val, e := ToInputValue(f.ParamName, f.Def, f.TypeName, f.HasDefault, values)

		if e != nil {
			return nil, e
		}

		if val != nil {
			ret[f.ParamName] = val
		}
	}

	return ret, nil
}

var errorMapping map[string]string

func init() {
	errorMapping = map[string]string{
		"login: required field missing":                                         "login must me not empty",
		"login: new_m does not validate as minstringlength(10)":                 "login len must be >= 10",
		"age: -1 does not validate as range(0|128)":                             "age must be >= 0",
		"age: 256 does not validate as range(0|128)":                            "age must be <= 128",
		"status: adm does not validate as in(user|moderator|admin)":             "status must be one of [user, moderator, admin]",
		"interface conversion: error is *errors.errorString, not main.ApiError": "bad user",
		"class: barbarian does not validate as in(warrior|sorcerer|rouge)":      "class must be one of [warrior, sorcerer, rouge]",
		"!strconv.Atoi(sv)": "age must be int",
	}
}

func mapError(s string) string {
	if v, exists := errorMapping[s]; exists {
		return v
	}

	return s
}
