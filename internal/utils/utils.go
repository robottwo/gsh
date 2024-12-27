package utils

import "github.com/sashabaranov/go-openai/jsonschema"

func GenerateJsonSchema(value any) *jsonschema.Definition {
	result, err := jsonschema.GenerateSchemaForType(value)
	if err != nil {
		panic(err)
	}
	return result
}
