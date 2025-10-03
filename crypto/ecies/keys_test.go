package ecies_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sonr-io/sonr/crypto/ecies"
)

func TestGenerateKey(t *testing.T) {
	_, err := ecies.GenerateKey()
	assert.NoError(t, err)
}

func TestGenerateFromSeed(t *testing.T) {
	seed := ecies.HashSeed(
		[]byte("testasdfasdfasdfasdfasdfw234453412341testasdfasdfasdfasdfasdfw234453412341"),
	)
	_, err := ecies.GenerateKeyFromSeed(seed)
	assert.NoError(t, err)
	_, err = ecies.GenerateKeyFromSeed(seed)
	assert.NoError(t, err)
}
