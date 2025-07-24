package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPassword(t *testing.T) {
	password := RandomString(10)
	hashedPassword1, err := HashedPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword1)

	err = CheckPasswordHash(password, hashedPassword1)
	require.NoError(t, err)

	wrongPassword := RandomString(10)
	err = CheckPasswordHash(wrongPassword, hashedPassword1)
	require.EqualError(t, err, "password does not match: crypto/bcrypt: hashedPassword is not the hash of the given password")

	hashedPassword2, err := HashedPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword2)
	require.NotEqual(t, hashedPassword1, hashedPassword2, "hashed passwords should not be the same even for the same input")
}
