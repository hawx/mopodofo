package mopodofo

import (
	"testing"

	"hawx.me/code/assert"
)

func TestReadAll(t *testing.T) {
	bundles, err := ReadAll("example/lang")
	assert.Nil(t, err)

	assert.Equal(t, NilBundle, bundles.For("xy"))

	en := bundles.For("en")
	assert.Equal(t, "Home", en.Tr("home"))
	assert.Equal(t, "/en", en.Trc("path", "home"))
	assert.Equal(t, "Running on port 8080", en.Trf("runningOn", "Port", "8080"))
	assert.Equal(t, "You have 5 notes, at Apr 2026", en.Trsf("youHaveXNotesAt", 5, "At", "Apr 2026"))

	cy := bundles.For("cy").StealContext(en, "path", "otherpath")
	assert.Equal(t, "/cy", cy.Trc("path", "home"))
	assert.Equal(t, "/en", cy.Trc("otherpath", "home"))
}
