package test

import (
	"encoding/json"
	"fmt"
)

//go:generate GoEnum $GOFILE

/*
@enum --name=Foo --bitflags --bitflag_separator="," --iterator_name="foobar" --marshaler=string --unmarshaler=string
Bar --string=bar
Baz --string=baz --description="This is the description"
Buz

@enum --name=Oof
Bar --string="bar"
Baz --value = 123
Buz --description="Some description"
*/

type tester struct {
	F FooEnum
	O OofEnum
}

func init() {
	var t = tester{
		F: Foo.Baz,
		O: Oof.Buz,
	}

	var j, err = json.Marshal(&t)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(j))
}
