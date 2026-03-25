package paths

import (
	"fmt"

	"hawx.me/code/mopodofo"
)

type Path byte

const (
	Home Path = iota
	Create
	Delete
)

func (p Path) String(bundle *mopodofo.Bundle) string {
	switch p {
	case Home:
		return bundle.Trc("path", "home")
	case Create:
		return bundle.Trc("path", "create")
	case Delete:
		return bundle.Trc("path", "delete")
	default:
		panic(fmt.Sprintf("missing String definition for Path(%v)", p))
	}
}

type Paths struct {
	Home   string
	Create string
	Delete string
}

func For(bundle *mopodofo.Bundle) Paths {
	return Paths{
		Home:   Home.String(bundle),
		Create: Create.String(bundle),
		Delete: Delete.String(bundle),
	}
}
