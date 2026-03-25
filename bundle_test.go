package mopodofo

import (
	"testing"

	"hawx.me/code/assert"
)

func TestReadAll(t *testing.T) {
	bundles, err := ReadAll("lang")
	assert.Nil(t, err)

	en, ok := bundles.For("en")
	if assert.True(t, ok) {
		assert.Equal(t, "Hello", en.Tr("helloMessage"))
	}
}
