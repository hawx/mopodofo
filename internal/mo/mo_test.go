package mo

import (
	"bytes"
	"os"
	"testing"

	"hawx.me/code/assert"
)

func TestDecode(t *testing.T) {
	file, err := os.ReadFile("testdata/messages.mo")
	assert.Nil(t, err)

	data, err := Decode(bytes.NewReader(file))
	assert.Nil(t, err)

	assert.Equal(t, "en", data.Metadata.Language)
	assert.Equal(t, 2, data.Metadata.NPlurals)
	for i := range 1000 {
		if i == 1 {
			assert.Equal(t, 0, data.Metadata.Plurals(i))
		} else {
			assert.Equal(t, 1, data.Metadata.Plurals(i))
		}
	}

	assert.Equal(t, map[string]string{"goodbyeMessage": "Goodbye", "helloMessage": "Hello", "personSaidSomething": "{{.Person}} said {{.Content}}"}, data.Singles)
	assert.Equal(t, map[string][]string{"youHaveXMessages": {"You have 1 message", "You have {{X}} messages"}}, data.Plurals)
	assert.Equal(t, map[string]map[string]string{"heading": {"welcomeMessage": "Welcome!"}, "title": {"welcomeMessage": "Welcome"}}, data.ContextSingles)
	assert.Equal(t, map[string]map[string][]string(nil), data.ContextPlurals)
}
