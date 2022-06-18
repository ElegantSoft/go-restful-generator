package common

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"reflect"
	"strings"
)

type GetError func(err validator.ValidationErrors) string

func ValidateErrors(requestError error) string {
	if reflect.TypeOf(requestError) == reflect.TypeOf(validator.ValidationErrors{}) {
		return validate(requestError.(validator.ValidationErrors))
	}
	return requestError.Error()
}

func validate(errors validator.ValidationErrors) string {

	resultErrors := ""
	for _, err := range errors {
		switch err.Tag() {
		case "required":
			resultErrors += err.Field() + " الحقل مطلوب\n "
		case "email":
			resultErrors += err.Field() + " البريد الالكتروني غير صحيح\n "
		case "min":
			resultErrors += err.Field() + " يجب ان يكون " + err.Param() + " حرفاً علي الأقل\n"
		case "Enum":
			replacer := *strings.NewReplacer("_", ",")
			resultErrors += err.Field() + " must be one of " + replacer.Replace(err.Param())

		default:
			resultErrors += "البيانات بالحقل غير صحيحة " + err.Tag()
		}
	}
	return resultErrors
}

func ConvertErrorToString(err error, errorString string) string {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errorString
	}
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return "wrong password"
	}
	return err.Error()
}

func Enum(
	fl validator.FieldLevel,
) bool {
	enumString := fl.Param()
	value := fl.Field().String()
	enumSlice := strings.Split(enumString, "_")
	for _, v := range enumSlice {
		if value == v {
			return true
		}
	}
	return false
}
