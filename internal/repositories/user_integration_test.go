package repository_test

import (
	"testing"

	error_codes "golang-order-manager-api/internal/errors"
	repository "golang-order-manager-api/internal/repositories"
	"golang-order-manager-api/internal/testhelpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepo_GetByEmail(t *testing.T) {
	db := testDB
	t.Cleanup(func() {
		ClearDatabase(db)
	})

	user := testhelpers.CreateUserFixture(t, db)
	repo := repository.NewUserRepo(db)

	result, err := repo.GetByEmail(user.Email)

	require.NoError(t, err)
	assert.Equal(t, user.ID, result.ID)
	assert.Equal(t, user.Email, result.Email)
}

func TestUserRepo_GetByEmail_NotFound(t *testing.T) {
	db := testDB
	t.Cleanup(func() {
		ClearDatabase(db)
	})

	repo := repository.NewUserRepo(db)

	_, err := repo.GetByEmail("ghost@example.com")

	assert.ErrorIs(t, err, error_codes.ErrUserNotFound)
}

func TestUserRepo_IsEmailAlreadyInUse(t *testing.T) {
	db := testDB
	t.Cleanup(func() {
		ClearDatabase(db)
	})

	user := testhelpers.CreateUserFixture(t, db)
	repo := repository.NewUserRepo(db)

	exists, err := repo.IsEmailAlreadyInUse(user.Email, nil)
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = repo.IsEmailAlreadyInUse(user.Email, &user.ID)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestUserRepo_IsUsernameAlreadyInUse(t *testing.T) {
	db := testDB
	t.Cleanup(func() {
		ClearDatabase(db)
	})

	user := testhelpers.CreateUserFixture(t, db)
	repo := repository.NewUserRepo(db)

	exists, err := repo.IsUsernameAlreadyInUse(user.Username, nil)
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = repo.IsUsernameAlreadyInUse(user.Username, &user.ID)
	require.NoError(t, err)
	assert.False(t, exists)
}
