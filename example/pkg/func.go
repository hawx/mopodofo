package pkg

import (
	"fmt"
	"time"

	mopo "hawx.me/code/mopodofo"
)

func MyFunc(bundle *mopo.Bundle) {
	fmt.Println(bundle.Trc("title", "welcomeMessage"))
	fmt.Println(bundle.Trc("heading", "welcomeMessage"))
	fmt.Println(bundle.Trs("youHaveXMessages", 5))
	fmt.Println(bundle.Trf("personSaidSomething", "Person", "John", "Content", "Hi!"))

	// This is a comment to be extracted
	s := bundle.Tr("helloMessage")

	if time.Now().Hour() > 18 {
		// This should also be extracted, but maybe not yet
		s = bundle.Tr("goodbyeMessage")
	}

	fmt.Println(s)
}
