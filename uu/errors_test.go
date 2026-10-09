package uu

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrInvalidVersion_Error(t *testing.T) {
	// The message must contain the offending version number
	// so that a failed validation is diagnosable from logs.
	assert.Equal(t, "invalid UUID version: 0", ErrInvalidVersion(0).Error())
	assert.Equal(t, "invalid UUID version: 15", ErrInvalidVersion(15).Error())
}

func TestID_Validate_Errors(t *testing.T) {
	// Callers distinguish the failure reasons with errors.Is / errors.As,
	// so Validate must return the exported error values and types.
	valid := IDMust("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

	t.Run("Nil UUID", func(t *testing.T) {
		err := IDNil.Validate()
		assert.ErrorIs(t, err, ErrNilID)
		assert.Equal(t, "Nil UUID", err.Error())
	})

	t.Run("invalid versions", func(t *testing.T) {
		for _, version := range []int{0, 9, 15} {
			id := valid
			id.SetVersion(version)
			err := id.Validate()
			var versionErr ErrInvalidVersion
			require.True(t, errors.As(err, &versionErr), "version %d: %v", version, err)
			assert.Equal(t, ErrInvalidVersion(version), versionErr)
			assert.False(t, id.Valid())
		}
	})

	t.Run("valid versions", func(t *testing.T) {
		for version := 1; version <= 8; version++ {
			id := valid
			id.SetVersion(version)
			assert.NoError(t, id.Validate(), "version %d", version)
		}
	})

	t.Run("invalid variant", func(t *testing.T) {
		id := valid
		id[8] = 0xe0 // 111x xxxx is a reserved, unsupported variant
		assert.Equal(t, IDVariantInvalid, id.Variant())
		assert.ErrorIs(t, id.Validate(), ErrInvalidVariant)
		assert.False(t, id.Valid())
	})
}
