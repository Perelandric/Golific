# GoEnum

**GoEnum** works with the `generate` command to create namespaced enums using structs, providing greater type safety as well as other features, including:

 - A `String()` method, allowing you to print the name of the enum variant or the optional custom value.
 - A `Description()` method, that will print the optional description of each variant.
 - The ability to marshal and/or unmarshal JSON (TODO: and XML) to the string value instead of the number.
 - A `Value()` method for retrieving the numeric representation of the variant.
 - The ability to assign a custom numeric value.
 - The ability to define an enum's variants as bitflags.
 - A generated array of the variants, to be used with a `range` loop.

# Quick start

This is a short example of how the basic syntax looks and how it's used. See the documentation for more info.

Installation:
```
go install github.com/Perelandric/GoEnum
```

Top of your source (below imports):
``` go
//go:generate GoEnum $GOFILE
```

Enum descriptor syntax in your source to create an enum named `Animal` that has 3 variants:
``` go
/*
@enum Animal --json=string
Dog --string=doggie --description="Your best friend, and you know it."
Cat --string=kitty --description="Your best friend, but doesn't always show it."
Horse --string=horsie --description="Everyone loves horses."
*/
```

Run Go's `generate` tool from the project directory:
```
go generate
```

Use the enum in your code:

``` go
type Resident struct {
 Name string
 Pet AnimalEnum // The generated type for your Animal enum
}

res := Resident{
 Name: "Charlie Brown",
 Pet: Animal.Dog, // Assign one of the variants
}

// The `--json=string` flag causes our custom `--string` value to be used in the resulting JSON
j, err := json.Marshal(&res)
fmt.Printf("%s %s\n", j, err) // {"Name":"Charlie Brown","Pet":"doggie"} <nil>

// Enumerate all the variants in a range loop
for _, animal := range AnimalValues {
 fmt.Printf("Kind: %s, Description: %q\n", animal, animal.Description())
}
```

# FAQ
###General
 - **Why was this created?**
  - Primarily in order to achieve greater type safety by restricting values of an enum type to only those variants provided.
 - **Can't this be done with `const` and a type alias?**
  - Yes, however a value of the base type can be substituted accidentally, resulting in bugs. Also, using `consts` pollutes the variable namespace, which can be an issue when overlapping names are needed in different categories.

###Functionality
 - **How are the variants stored and referenced?**
  - For each enum, the variants are stored together in an anonymous struct value assigned to a variable. They are referenced as `Animal.Dog`.
 - **How do I access the numeric representation of a variant?**
  - Use the `.Value()` method.
 - **Will GoEnum generate bitflag numbers for me?**
  - Yes, using the `--bitflags` flag.
 - **Can I choose the numeric representation?**
  - Yes, as long as the `--bitflags` option is not used, and the number doesn't match another value in the same enum.
 - **Can negative numbers be used for the numeric representation?**
  - No, the numbers must be `0` or greater and it is recommended that `0` be reserved to denote no value having been set, unless a default variant makes sense.
 - **Can I get the name of a variant as a `string`? If so, can I define a string that differs from the variant name?**
  - Yes, using the `.String()` method and yes, a custom string can be defined using the `--string` flag.
 - **Can meta data be associated with each variant?**
  - Yes, each variant can have a description assigned using the `--description` flag, which is accessed using the `.Description()` method.
 - **Can I have JSON (TODO: and XML) marshaled to and unmarshaled from the string value instead of the number?**
  - Yes, using the `--json` and `--xml` flags. There are also flags to set specifically JSON or XML marshaling/unmarshaling, though you'll get a warning if the marshaler doesn't match the unmarshaler.
 - **Can I enumerate the variants of an enum using a `range` loop?**
  - Yes, an array holding the variants is generated, which can be used in a `range` loop.

###Efficiency
 - **How are the variants represented in memory?**
  - The individual variants are stored as a value of a struct type that has a single `uint` field, sized to the smallest size needed for each given enum.
 - **Does each variant being a struct value add extra memory overhead?**
  - No, variants that use a `uint8` will still use only 8 bits.
 - **Does the `.Value()` call add overhead when getting the underlying number?**
  - Only if the compiler does not inline the call. However, the method simply returns the value of the field, so it would seem about as likely a candidate for inlining as one can hope to find.
 - **Does GoEnum use reflection?**
  - No, because the code is generated, we can hardcode necessary values into `switch` statements where needed, making the generated code longer, but faster.
 - **Does GoEnum use interfaces or pointers as the type of its variants?**
  - No, the `type` of the variants is a concrete, value type. Assigning or passing makes a copy, which equals the specific size of the `uint` used for that enum.

###Safety
 - **Is it still possible to use a value of the base (struct) type in place of one of the variants?**
  - Technically yes, however the variants for each enum use a struct type with a `value` field that has a unique identifier appended to it, i.e. `value_1cn7iw6qxr8ad`, so substituting a base value would be cumbersome and never accidental.
 - **Are the new unique identifiers used in the variants' structs generated every time `generate` is run?**
  - Yes. A pseudo-random number is used with a time-based seed, so it is non-deterministic.
 - **Is it possible to overwrite one variant with another from the same enum?**
  - Unfortunately, Go does not allow struct values to be assigned to a `const`, so yes. However it would require `Animal.Cat = Animal.Horse`, which seems like an unlikely mistake.

# Documentation

###Installation

To install GoEnum, use the `install` command from the Go toolchain.

```
go install github.com/Perelandric/GoEnum
```

###Using the `generate` command

After your source code has been properly annotated as described below, then from your source directory run the `generate` command.

```
go generate
```

This creates a new file for every file that had the proper annotations. The new file has the same name as the original, but with an `enum____` prefix, so make sure you don't already have a file with a conflicting name. The generated file also gets the same `package` name as the original.

Do not edit the generated file, as it will be overwritten every time you call `generate`.

###Setup for the `generate` command

As with all files that rely on Go's `generate` command, your file must have the `go:generate` directive below the file package name and imports. For GoEnum, it should look like this:

``` go
//go:generate GoEnum $GOFILE
```

Notice that there's no space after the `//` and before `go:`. This is required for the `generate` tool.

###Enum descriptor syntax

The actual Enum descriptors are defined entirely in code comment blocks at the top-level namespace of your code. A comment block is either a `/* multi line comment */` or several adjacent `// single line comments`.

The comment block must begin with `@enum` *(comment lines that are empty or have only whitespace are ignored)* followed by an identifier, which provides the name that will be used to reference your enum as well as the type of the variants *(to which the word "Enum" will be added)*.

Multiple enum descriptors may be defined in a single comment block, where `@enum` at the beginning of a line marks the start of a new enum.

The beginning of a descriptor can look like either of these *(incomplete)* examples:

``` go
/*
@enum Animal
*/
```
``` go
//
// @enum Animal
```

In both cases we define an enum named `Animal`. This will create a `var Animal`, a `type AnimalEnum struct {...}` and a `var AnimalValues` in the generated file, so all these names must be available to avoid conflicts and must be a valid identifier. As usual, the capitalization will determine whether or not the items are exported, so `animal` could be used instead.

Notice that both have an empty line above the `@enum Animal`. This is fine since empty or whitespace-only comment lines are ignored.

#####*Descriptor flags*

The next thing to come are the descriptor flags. These flags begin with `--` and are followed by a word and in some cases a `=` with a value. Flags must be separated by at least 1 white space, and can optionally be defined on separate lines as seen in these *(still incomplete)* examples.

``` go
/*
@enum Animal --json="string"
*/
```
``` go
//
// @enum Animal
// --json=string
```

Notice two differences in the above examples.
 - The first one uses quotes around `string`, and the second does not. As long as a flag value does not contain space characters, the quotation marks are optional.
 - The first one begins its flags on the same line as the `@enum`, and the second starts on the next line. Either way is valid.

All flags are optional, and are described in the table below.

#####*Variant definitions*

After the descriptor flags have been defined, you'll need to define the enum variants. At least one variant is required per enum.

Each variant must be entirely defined on its own line. Lines that are long because of variant flags must not be split across multiple lines.

Adding to the examples above, they may now look like this:

``` go
/*
@enum Animal --json="string"
Dog --string=doggie --description="Your best friend, and you know it."
Cat --string=kitty --description="Your best friend, but doesn't always show it."
Horse --string=horsie --description="Everyone loves horses."
*/
```
``` go
//
// @enum
// "Animal"
// --json=string
// Dog --string=doggie --description="Your best friend, and you know it."
// Cat --string=kitty --description="Your best friend, but doesn't always show it."
// Horse --string=horsie --description="Everyone loves horses."
```

So our examples are now fully valid enum descriptors. As long as you have the `go:generate` annotation previously defined, you'll be able to run `go generate` and your new source file will be generated.

The rest of the documentation will use the multi-line version of our descriptor example.

#Flags and Methods

###Descriptor flags

These are the flags available for use in the main `@enum` descriptor. They are distinct from the variant flags, which are listed in a separate table. All of the flags are optional.

| Flag | Value | Behavior |
| :--: | ----- | -------- |
| `bitflags` | *(no value)* | Causes the numeric values generated to able to be used as bitflags. When used, a maximum of 64 variants is allowed. |
| `bitflag_separator` | Any non-empty string of text. | Only valid when `--bitflags` is used. Defines the separator used when the `.String()` method is called on values that have multiple bits set, as well as when `string` is used for JSON/XML marshaling and/or unmarshaling. Default value is `,`. |
| `iterator_name` | Any valid Go identifier | Alternate identifier name used for the array of variants generated. Used to resolve conflicts. The default name is `Values` |
| `json` | Allowed values: `string` or `value` | Sets the type of marshaler and unmarshaler to use for JSON. The `string` option will use the `.String()` representation of the variant, whereas the `number` will use the numeric value. |
| `xml` | Allowed values: `string` or `value` | Sets the type of marshaler and unmarshaler to use for XML. The `string` option will use the `.String()` representation of the variant, whereas the `number` will use the numeric value. |
| `json_marshal` | Allowed values: `string` or `value` | Same as the `--json` flag but only sets the marshaler. |
| `json_unmarshal` | Allowed values: `string` or `value` | Same as the `--json` flag but only sets the unmarshaler. |
| `xml_marshal` | Allowed values: `string` or `value` | Same as the `--xml` flag but only sets the marshaler. |
| `xml_unmarshal` | Allowed values: `string` or `value` | Same as the `--xml` flag but only sets the unmarshaler. |
| `drop_json` | n/a | Prevent JSON marshaling methods from being generated. |
| `drop_xml` | n/a | Prevent XML marshaling methods from being generated. |


###Variant flags

These are the flags available for use in each variant. All of the flags are optional.

| Flag | Value | Behavior |
| :--: | ----- | -------- |
| `string` | Any string of text. | This sets the the value returned by the `.String()` method. If empty, the variant name is used. |
| `description` | Any string of text. | This sets the value returned by the `.Description()` method. If empty, the `--string` value is used. |
| `value` | Any integer >= 0 | This overrides the default numeric value of the variant. It may *not* be used when `--bitflags` is used. The value must not have been already assigned to another variant. The value `0` should only be used for a variant that makes sense to use as a default value. |

###Variant Methods

`func Name() string` - Returns the name of the variant as a string.

`func String() string` - Returns the given `--string` value of the variant. If none has been set, its return value is as though `Name()` had been called.

`func Description() string` - Returns the given `--description` value of the variant. If None has been set, its return value is as though `String()` had been called.

`func Value() uint?` - Returns the numeric value of the variant as the specific `uint` size that it the variant uses.

`func IntValue() int` - Same as `Value()`, except that the value is cast to an `int`.

###Variant Methods (for bitflag enums)

*The methods in this section are only generated when the enum uses the `--bitflags` option.*

*Coming soon...*
