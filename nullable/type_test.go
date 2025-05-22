package nullable

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestType_JSONSchema(t *testing.T) {
	schema := Type[int]{}.JSONSchema()
	jsonSchemaBytes, err := json.Marshal(schema)
	require.NoError(t, err)
	require.Equal(t, `{"$schema":"https://json-schema.org/draft/2020-12/schema","oneOf":[{"type":"integer"},{"type":"null"}],"default":null}`, string(jsonSchemaBytes))
}
