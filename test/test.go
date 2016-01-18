package test

//go:generate generate_enum $GOFILE

// @enum
// --name=Foo --bitflags --bitflag_separator="," --iterator_name="foobar" --marshaler=string --unmarshaler=string
// Bar --string=bar
// Baz --string=baz --description="This is the description"
// Buz

// @enum
// --name=Oof --unmarshaler=string
// Bar --string="bar"
// Baz --value = 123
// Buz --description="Some description"
