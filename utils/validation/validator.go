package validator

import (
	"reflect"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

var commonPassword= map[string]bool{

"password":   true,
"123456":     true,
"12345678":   true,
"qwerty":     true,
"admin":      true,
"password1":  true,
"welcome":    true,
"letmein":    true,

}
func init() {
	v := validator.New()

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	Validate = v
}

func IsValidPassword(pass string)bool{

	if len(pass) < 8 || len(pass) > 64{
		return false
	}
	if commonPassword[strings.ToLower(pass)]{
		return false
	}
	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false

    for _,ch:=range pass{
    
	switch{
	  
	case unicode.IsUpper(ch):
	hasUpper=true
	case unicode.IsLower(ch):
	hasLower=true
	case unicode.IsDigit(ch):
	hasNumber=true
	case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
	hasSpecial = true

    }
  }
return hasUpper && hasLower && hasNumber && hasSpecial

}