package repositories

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/breakfront-planner/auth-service/internal/autherrors"
)

func TestParseFilter(t *testing.T) {
	testID := uuid.New()
	testLogin := "test_user"

	t.Run("valid filter with single field", func(t *testing.T) {
		type Filter struct {
			ID *uuid.UUID `db:"id"`
		}

		filter := Filter{ID: &testID}
		fields, err := ParseFilter(&filter)

		require.NoError(t, err)
		assert.Len(t, fields, 1)
		assert.Equal(t, "ID", fields[0].FilterName)
		assert.Equal(t, "id", fields[0].DBName)
		assert.Equal(t, testID, fields[0].Value)
	})

	t.Run("valid filter with multiple fields", func(t *testing.T) {
		type Filter struct {
			ID    *uuid.UUID `db:"id"`
			Login *string    `db:"login"`
		}

		filter := Filter{ID: &testID, Login: &testLogin}
		fields, err := ParseFilter(&filter)

		require.NoError(t, err)
		assert.Len(t, fields, 2)
		assert.Equal(t, "ID", fields[0].FilterName)
		assert.Equal(t, "id", fields[0].DBName)
		assert.Equal(t, testID, fields[0].Value)
		assert.Equal(t, "Login", fields[1].FilterName)
		assert.Equal(t, "login", fields[1].DBName)
		assert.Equal(t, testLogin, fields[1].Value)
	})

	t.Run("valid filter with partial fields", func(t *testing.T) {
		type Filter struct {
			ID    *uuid.UUID `db:"id"`
			Login *string    `db:"login"`
		}

		filter := Filter{ID: &testID, Login: nil}
		fields, err := ParseFilter(&filter)

		require.NoError(t, err)
		assert.Len(t, fields, 1)
		assert.Equal(t, "ID", fields[0].FilterName)
		assert.Equal(t, "id", fields[0].DBName)
		assert.Equal(t, testID, fields[0].Value)
	})

	t.Run("error on empty filter - all fields nil", func(t *testing.T) {
		type Filter struct {
			ID    *uuid.UUID `db:"id"`
			Login *string    `db:"login"`
		}

		filter := Filter{ID: nil, Login: nil}
		fields, err := ParseFilter(&filter)

		require.Error(t, err)
		assert.ErrorIs(t, err, autherrors.ErrEmptyFilter)
		assert.Nil(t, fields)
	})

	t.Run("error on empty filter - no fields", func(t *testing.T) {
		type Filter struct{}

		filter := Filter{}
		fields, err := ParseFilter(&filter)

		require.Error(t, err)
		assert.ErrorIs(t, err, autherrors.ErrEmptyFilter)
		assert.Nil(t, fields)
	})

	t.Run("error on non-pointer field", func(t *testing.T) {
		type InvalidFilter struct {
			ID uuid.UUID `db:"id"`
		}

		filter := InvalidFilter{ID: testID}
		fields, err := ParseFilter(&filter)

		require.Error(t, err)
		assert.ErrorIs(t, err, autherrors.ErrNoPtrsFilterFields)
		assert.Nil(t, fields)
	})

	t.Run("error on mixed pointer and non-pointer fields", func(t *testing.T) {
		type InvalidFilter struct {
			ID    uuid.UUID `db:"id"`
			Login *string   `db:"login"`
		}

		filter := InvalidFilter{ID: testID, Login: &testLogin}
		fields, err := ParseFilter(&filter)

		require.Error(t, err)
		assert.ErrorIs(t, err, autherrors.ErrNoPtrsFilterFields)
		assert.Nil(t, fields)
	})

	t.Run("ignore fields without db tag", func(t *testing.T) {
		type Filter struct {
			ID          *uuid.UUID `db:"id"`
			IgnoredData *string
		}

		ignoredValue := "should_be_ignored"
		filter := Filter{ID: &testID, IgnoredData: &ignoredValue}
		fields, err := ParseFilter(&filter)

		require.NoError(t, err)
		assert.Len(t, fields, 1)
		assert.Equal(t, "ID", fields[0].FilterName)
		assert.Equal(t, "id", fields[0].DBName)
		assert.Equal(t, testID, fields[0].Value)
	})

	t.Run("ignore fields with empty db tag", func(t *testing.T) {
		type Filter struct {
			ID            *uuid.UUID `db:"id"`
			EmptyTagField *string    `db:""`
		}

		emptyValue := "should_be_ignored"
		filter := Filter{ID: &testID, EmptyTagField: &emptyValue}
		fields, err := ParseFilter(&filter)

		require.NoError(t, err)
		assert.Len(t, fields, 1)
		assert.Equal(t, "ID", fields[0].FilterName)
		assert.Equal(t, "id", fields[0].DBName)
		assert.Equal(t, testID, fields[0].Value)
	})

	t.Run("works with pointer to struct", func(t *testing.T) {
		type Filter struct {
			ID *uuid.UUID `db:"id"`
		}

		filter := &Filter{ID: &testID}
		fields, err := ParseFilter(filter)

		require.NoError(t, err)
		assert.Len(t, fields, 1)
		assert.Equal(t, "ID", fields[0].FilterName)
		assert.Equal(t, "id", fields[0].DBName)
		assert.Equal(t, testID, fields[0].Value)
	})

	t.Run("works with different data types", func(t *testing.T) {
		type Filter struct {
			IntField    *int       `db:"int_field"`
			StringField *string    `db:"string_field"`
			UUIDField   *uuid.UUID `db:"uuid_field"`
			BoolField   *bool      `db:"bool_field"`
		}

		intVal := 42
		strVal := "test"
		uuidVal := uuid.New()
		boolVal := true

		filter := Filter{
			IntField:    &intVal,
			StringField: &strVal,
			UUIDField:   &uuidVal,
			BoolField:   &boolVal,
		}
		fields, err := ParseFilter(&filter)

		require.NoError(t, err)
		assert.Len(t, fields, 4)
		assert.Equal(t, "IntField", fields[0].FilterName)
		assert.Equal(t, "int_field", fields[0].DBName)
		assert.Equal(t, intVal, fields[0].Value)
		assert.Equal(t, "StringField", fields[1].FilterName)
		assert.Equal(t, "string_field", fields[1].DBName)
		assert.Equal(t, strVal, fields[1].Value)
		assert.Equal(t, "UUIDField", fields[2].FilterName)
		assert.Equal(t, "uuid_field", fields[2].DBName)
		assert.Equal(t, uuidVal, fields[2].Value)
		assert.Equal(t, "BoolField", fields[3].FilterName)
		assert.Equal(t, "bool_field", fields[3].DBName)
		assert.Equal(t, boolVal, fields[3].Value)
	})
}
