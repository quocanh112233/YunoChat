package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// CustomValidator manages struct validation rules
type CustomValidator struct {
	validate *validator.Validate
}

// New returns a new custom validator instance
func New() *CustomValidator {
	v := validator.New()

	// Use JSON tag as the field name in errors instead of Go struct field name
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &CustomValidator{
		validate: v,
	}
}

// ValidateStruct validates a struct and returns a map of validation errors.
// Field name -> User friendly error message.
func (cv *CustomValidator) ValidateStruct(s interface{}) map[string]string {
	err := cv.validate.Struct(s)
	if err == nil {
		return nil
	}

	errors := make(map[string]string)
	
	for _, err := range err.(validator.ValidationErrors) {
		errors[err.Field()] = getFriendlyErrorMessage(err)
	}

	return errors
}

// getFriendlyErrorMessage translates standard tags intro readable strings
func getFriendlyErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "Trường này là bắt buộc"
	case "email":
		return "Định dạng email không hợp lệ"
	case "min":
		if err.Type().String() == "string" {
			return fmt.Sprintf("Độ dài tối thiểu là %s ký tự", err.Param())
		}
		return fmt.Sprintf("Giá trị nhỏ nhất là %s", err.Param())
	case "max":
		if err.Type().String() == "string" {
			return fmt.Sprintf("Độ dài tối đa là %s ký tự", err.Param())
		}
		return fmt.Sprintf("Giá trị lớn nhất là %s", err.Param())
	case "uuid":
		return "UUID không hợp lệ"
	case "url":
		return "URL không hợp lệ"
	case "oneof":
		return fmt.Sprintf("Chỉ chấp nhận các giá trị: %s", err.Param())
	case "eqfield":
		return fmt.Sprintf("Phải trùng khớp với %s", err.Param())
	default:
		return fmt.Sprintf("Không hợp lệ (%s)", err.Tag())
	}
}
