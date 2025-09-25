package http

import (
	"fmt"
	"reflect"
	"strings"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type CustomValidator struct {
	validator *validator.Validate
}

func NewValidator() *CustomValidator {
	return &CustomValidator{
		validator: validator.New(),
	}
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		var validationErrors []string
		
		for _, err := range err.(validator.ValidationErrors) {
			field := err.Field()
			tag := err.Tag()
			
			switch tag {
			case "required":
				validationErrors = append(validationErrors, fmt.Sprintf("%s is required", field))
			case "email":
				validationErrors = append(validationErrors, fmt.Sprintf("%s must be a valid email", field))
			case "min":
				validationErrors = append(validationErrors, fmt.Sprintf("%s must be at least %s characters", field, err.Param()))
			case "max":
				validationErrors = append(validationErrors, fmt.Sprintf("%s must be no more than %s characters", field, err.Param()))
			case "url":
				validationErrors = append(validationErrors, fmt.Sprintf("%s must be a valid URL", field))
			case "oneof":
				validationErrors = append(validationErrors, fmt.Sprintf("%s must be one of: %s", field, err.Param()))
			default:
				validationErrors = append(validationErrors, fmt.Sprintf("%s is invalid", field))
			}
		}
		
		return fmt.Errorf("%s", strings.Join(validationErrors, ", "))
	}
	return nil
}

func BindAndValidate(c echo.Context, req interface{}) error {
	if err := c.Bind(req); err != nil {
		return BadRequestResponse(c, "Invalid request format", err.Error())
	}
	
	if err := c.Validate(req); err != nil {
		return BadRequestResponse(c, "Validation failed", err.Error())
	}
	
	return nil
}

func GetStructName(v interface{}) string {
	if t := reflect.TypeOf(v); t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	} else {
		return t.Name()
	}
}