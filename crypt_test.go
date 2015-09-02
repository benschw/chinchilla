package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryption(t *testing.T) {
	// given
	keyringPath := "./test-keys/.pubring.gpg"
	secretKeyringPath := "./test-keys/.secring.gpg"

	in := "Secret Data"

	// when
	encrypted, err1 := encrypt(keyringPath, in)
	decrypted, err2 := decrypt(secretKeyringPath, encrypted)

	// then
	assert.Nil(t, err1, "err should be nil")
	assert.Nil(t, err2, "err should be nil")

	assert.Equal(t, in, decrypted, "values should be equal")
	assert.NotEqual(t, in, encrypted, "values shouldn't be equal")

}
