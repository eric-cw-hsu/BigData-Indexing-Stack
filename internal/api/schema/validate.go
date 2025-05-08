package schema

import (
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

func ValidateJsonSchema(payload map[string]interface{}, schema string) error {
	schemaLoader := gojsonschema.NewStringLoader(schema)
	documentLoader := gojsonschema.NewGoLoader(payload)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil || !result.Valid() {
		var errors []string
		for _, desc := range result.Errors() {
			errors = append(errors, desc.String())
		}

		return fmt.Errorf("invalid json input: %v", errors)
	}

	return nil
}
