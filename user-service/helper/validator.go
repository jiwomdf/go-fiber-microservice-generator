package helper

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidationError represents a single validation error message.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

var validate = validator.New()

// ValidateStruct validates a struct based on the 'validate' tags.
// It returns a slice of validation errors if any are found.
func ValidateStruct(s interface{}) []*ValidationError {
	var errors []*ValidationError
	err := validate.Struct(s)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			// Get the JSON tag name if available
			fieldName := getJSONFieldName(s, err.Field())
			errors = append(errors, &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("Field validation for '%s' failed on the '%s' tag", err.Field(), err.Tag()),
			})
		}
	}
	return errors
}

// getJSONFieldName returns the JSON tag name for a struct field, or the lowercase field name if no JSON tag exists.
func getJSONFieldName(s interface{}, fieldName string) string {
	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return strings.ToLower(fieldName)
	}

	field, exists := val.Type().FieldByName(fieldName)
	if !exists {
		return strings.ToLower(fieldName)
	}

	jsonTag := field.Tag.Get("json")
	if jsonTag != "" && jsonTag != "-" {
		// Remove any options after comma (e.g., "property_code,omitempty" -> "property_code")
		jsonTag = strings.Split(jsonTag, ",")[0]
		return jsonTag
	}

	return strings.ToLower(fieldName)
}
