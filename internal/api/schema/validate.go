package schema

import (
	"fmt"

	"eric-cw-hsu.github.io/internal/shared/logger"
	"github.com/xeipuuv/gojsonschema"
	"go.uber.org/zap"
)

func ValidateJsonSchema(payload map[string]interface{}, schema string) error {
	schemaLoader := gojsonschema.NewStringLoader(schema)
	documentLoader := gojsonschema.NewGoLoader(payload)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		logger.Logger.Error("schema.ValidateJsonSchema: validation execution failed", zap.Error(err))
		return fmt.Errorf("schema validation error: %w", err)
	}
	if !result.Valid() {
		var errors []string
		for _, desc := range result.Errors() {
			errors = append(errors, desc.String())
		}
		logger.Logger.Error("schema.ValidateJsonSchema: payload did not conform", zap.Strings("issues", errors))
		return fmt.Errorf("invalid json input: %v", errors)
	}

	return nil
}
