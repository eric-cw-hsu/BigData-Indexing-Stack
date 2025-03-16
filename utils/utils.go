package utils

import (
	"crypto/sha256"
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

func GenerateETag(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%s", hash)
}

func ValidateJSONSchema(payload map[string]interface{}, jsonSchema string) error {
	schemaLoader := gojsonschema.NewStringLoader(jsonSchema)
	documentLoader := gojsonschema.NewGoLoader(payload)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("Schema validation error")
	}

	if !result.Valid() {
		errors := []string{}
		for _, err := range result.Errors() {
			errors = append(errors, err.String())
		}
		return fmt.Errorf("Validation errors: %v", errors)
	}

	return nil
}
