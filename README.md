# Golific

**Golific** is a tool for generating Go code using the `go:generate` tool. Currently there are two types of annotations: **&#64;struct** and **&#64;enum**. See descriptions below.

## &#64;struct

**&#64;struct** is used as an alternate syntax for generating `struct` types, providing the ability to generate getter and setters, a default constructor with default values and the ability to keep private fields private yet still be able to marshal and unmarshal JSON.

*(docs are forthcoming)*

## &#64;enum

**&#64;enum** is used to create namespaced enums using structs, providing greater type safety and offering several other features.

# Quick start

This is a short example of how the basic syntax looks and how it's used. See the documentation for more info.

Installation:
```
go install github.com/Perelandric/Golific
```

Top of your source (below imports):
``` go
//go:generate Golific $GOFILE
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

// The `--json=string` flag we included causes our custom `--string` value to be used in the JSON
j, err := json.Marshal(&res)

fmt.Printf("%s %s\n", j, err) // {"Name":"Charlie Brown","Pet":"doggie"} <nil>

// Enumerate all the variants in a range loop
for _, animal := range Animal.Values {
 fmt.Printf("Kind: %s, Description: %q\n", animal, animal.Description())
}
```

**Please note:** This will create a new file with the same name as the original, except that it will have the prefix `golific____` added, so if your file is `animal.go`, the file `golific____animal.go` will be created, ***overwriting*** any existing file.

# FAQ
###General
 - **Why was this created?**
  - Primarily in order to achieve greater type safety by restricting values of an enum type to only those variants provided.
 - **Can't this be done with `const` and a type alias?**
  - Yes, however a value of the base type can be substituted accidentally, resulting in bugs. Also, using `const` pollutes the variable namespace, which can be an issue when overlapping names are needed in different categories.

###Functionality
 - **How are the variants stored and referenced?**
  - For each enum, the variants are stored together in an anonymous struct value assigned to a variable. They are referenced as `Animal.Dog`.
 - **Can I get the numeric representation of a variant?**
  - Yes, by using the `.Value()` or `.IntValue()` method.
 - **Will Golific generate bitflag numbers for me?**
  - Yes, by using the `--bitflags` flag.
 - **Can I choose the numeric representation?**
  - Yes, as long as the `--bitflags` option is not used, and the number doesn't match another value in the same enum.
 - **Can negative numbers be used for the numeric representation?**
  - No, the numbers must be `0` or greater and it is recommended that `0` be reserved to denote no value having been set, unless a default variant makes sense.
 - **Can I get the name of a variant as a `string`? If so, can I define a string that differs from the variant name?**
  - Yes, the `.Name()` method gives you the name and the `.String()` method gives you an optional custom string defined using the `--string` flag.
 - **Can meta data be associated with each variant?**
  - Yes, each variant can have a description assigned using the `--description` flag, which is accessed using the `.Description()` method.
 - **Can I have JSON marshaled to and unmarshaled from the string value instead of the number?**
  - Yes, using the `--json` flag. There are also flags to set specifically JSON marshaling/unmarshaling, though you'll get a warning if the marshaler doesn't match the unmarshaler.
 - **Can I enumerate the variants of an enum using a `range` loop?**
  - Yes, an array holding the variants is generated, which can be used in a `range` loop.

###Efficiency
 - **How are the variants represented in memory?**
  - The individual variants are stored as a value of a struct type that has a single `uint` field, sized to the smallest size needed for each given enum.
 - **Does each variant being a struct value add extra memory overhead?**
  - No, variants that, for example, use a `uint8`, will still use only 8 bits.
 - **Does the `.Value()` call add overhead when getting the underlying number?**
  - Only if the compiler does not inline the call. However, the method simply returns the value of the field, so it would seem about as likely a candidate for inlining as one can hope to find.
 - **Does Golific use reflection?**
  - No, because the code is generated, we can hardcode necessary values into `switch` statements where needed, making the generated code longer, but faster.
 - **Does Golific use interfaces or pointers as the type of its variants?**
  - No, the `type` of the variants is a concrete, value type. Assigning or passing makes a copy, which equals the specific size of the `uint` used for that enum.

###Safety
 - **Is it still possible to use a value of the base (struct) type in place of one of the variants?**
  - Technically yes, however the variants for each enum use a struct type with a `value` field that has a unique identifier appended to it, e.g. `value_1cn7iw6qxr8ad`, so substituting a base value would be cumbersome and never accidental.
 - **Are the new unique identifiers used in the variants' structs generated every time `generate` is run?**
  - Yes. A pseudo-random number is used with a time-based seed, so it is non-deterministic.
 - **Is it possible to overwrite one variant with another from the same enum?**
  - Unfortunately, Go does not allow struct values to be assigned to a `const`, so yes. However it would require `Animal.Cat = Animal.Horse`, which seems like an unlikely mistake.

# Documentation

###Installation

To install Golific, use the `install` command from the Go toolchain.

```
go install github.com/Perelandric/Golific
```

###Setup for the `generate` command

As with all files that rely on Go's `generate` command, your file must have the `go:generate` directive below the file package name and imports. For Golific, it should look like this:

``` go
//go:generate Golific $GOFILE
```

Notice that there's no space between `//` and `go:generate`. This is required for the `generate` tool.

###Enum descriptor syntax

The actual Enum descriptors are defined entirely in `/* multi-line code comments */` at the top-level scope of your code. Comment lines that are empty or have only whitespace are ignored.

The comment block must begin with `@enum` followed by an identifier, which provides the name that will be used to reference your enum as well as the type of the variants *(to which the word "Enum" will be added)*.

Multiple enum descriptors may be defined within a single comment, where `@enum` at the beginning of a line marks the start of a new enum.

The beginning of a descriptor can look like this *(incomplete)* example:

``` go
/*
@enum Animal
*/
```

Here we started to define an enum named `Animal`. This will create a `var Animal` and a `type AnimalEnum struct {...}` in the generated file, so these names must be available to avoid conflicts, and the name you provide must be a valid identifier. As usual, the capitalization determines whether or not the items are exported, so `animal` could be used instead of `Animal`.

#####*Descriptor flags*

Next is the set of descriptor flags. These flags begin with `--` and are followed by a word and in some cases an `=` with a value. Flags must be separated by at least 1 white space, and can optionally be defined on separate lines as seen in this *(still incomplete)* example.

``` go
/*
@enum Animal --json="string"
*/
```

Notice it uses quotes around `string`. These are optional as long as the value does not contain space characters. Single quotes (`'`), double quotes (`"`) or backticks (`&#96`) may be used. Flags may be on the same line and/or on subsequent lines.

All flags are optional, and are described in the tables below.

#####*Variant definitions*

After the descriptor flags have been defined, you'll need to define the enum variants. At least one variant is required per enum. Again, the flags may be defined on the same line or on different lines.

Adding to the example above, it may now look like this:

``` go
/*
@enum Animal --json="string"
Dog --string=doggie --description="Your best friend, and you know it."
Cat --string=kitty --description="Your best friend, but doesn't always show it."
Horse --string=horsie --description="Everyone loves horses."
*/
```

So our example is now a fully valid enum descriptor.

###Using the `generate` command

After your source code has been properly annotated as described above, from your source directory run the `generate` command.

```
go generate
```

This creates a new file for every file that had the proper annotations. The new file has the same name as the original, but with a `golific____` prefix, so make sure you don't already have a file with a conflicting name. The generated file also gets the same `package` name as the original.

Do not edit the generated file, as it will be overwritten every time you call `generate`.

#Flags, Methods and everything else

###Descriptor flags

These are the flags available for use in the main `@enum` descriptor. They are distinct from the variant flags, which are listed in a separate table. All of the flags are optional.

The word *boolean* in the **Value** column means the allowed values are `true` or `false`, or if no equal sign and value is provide, it will be considered to be `true`.

| Flag | Value | Behavior |
| :--: | ----- | -------- |
| `bitflags` | *boolean* | Causes the numeric values generated to able to be used as bitflags. When used, a maximum of 64 variants is allowed. |
| `bitflag_separator` | Any non-empty string of text. | Only valid when `--bitflags` is used. Defines the separator used when the `.String()` method is called on values that have multiple bits set, as well as when `string` is used for JSON marshaling and/or unmarshaling. Default value is `,`. |
| `iterator_name` | Any valid Go identifier | Alternate identifier name used for the array of variants generated. Used to resolve conflicts. The default name is `Values` |
| `summary` | *boolean* | ***(This may be relocated, so is considered unstable at the moment.)*** Include a summary of this enum at the top of the generated file. |
| `json` | Allowed values: `string` or `value` | Sets the type of marshaler and unmarshaler to use for JSON. The `string` option will use the `.String()` representation of the variant, whereas the `number` will use the numeric value. |
| `json_marshal` | Allowed values: `string` or `value` | Same as the `--json` flag but only sets the marshaler. |
| `json_unmarshal` | Allowed values: `string` or `value` | Same as the `--json` flag but only sets the unmarshaler. |
| `drop_json` | *boolean* | Prevent JSON marshaling methods from being generated. |


###Variant flags

These are the flags available for use in each variant. All of the flags are optional.

| Flag | Value | Behavior |
| :--: | ----- | -------- |
| `string` | Any text. | This sets the the value returned by the `.String()` method. If empty, the variant name is used. |
| `description` | Any text. | This sets the value returned by the `.Description()` method. If empty, the `--string` value is used. |
| `value` | Any integer >= 0 | This overrides the default numeric value of the variant. It may *not* be used when `--bitflags` is used. The value must not have been already assigned to another variant. The value `0` should only be used for a variant that makes sense to use as a default value. |

###Variant Methods

The methods described in this section use `MyEnum` as a placeholder name for the actual enum type generated by Golific

`func (me MyEnum) Name() string` - Name returns the name of the variant as a string.

`func (me MyEnum) String() string` - String returns the given `--string` value of the variant. If none has been set, its return value is as though `Name()` had been called. If the `--bitflags` option is enabled, multiple values will be joined using the `--bitflag_separator` value.

`func (me MyEnum) Description() string` - Description returns the given `--description` value of the variant. If None has been set, its return value is as though `String()` had been called.

`func (me MyEnum) Value() uint?` - Value returns the numeric value of the variant as the specific `uint` size that it the variant uses.

`func (me MyEnum) IntValue() int` - IntValue is the same as `Value()`, except that the value is cast to an `int`.

`func (me MyEnum) Type() string` - Type returns the variant's type name as a string.

`func (me MyEnum) Namespace() string` - Namespace returns the variant's namespace name as a string.

#####*Methods for bitflag enums*

The methods below are only generated when the enum uses the `--bitflags` option.

`func (me MyEnum) Add(v MyEnum) MyEnum` - Add returns a copy of the variant with the value of `v` added to it.

`func (me MyEnum) AddAll(v ...MyEnum) MyEnum` - AddAll returns a copy of the variant with all the values of `v` added to it.

`func (me MyEnum) Remove(v MyEnum) MyEnum` - Remove returns a copy of the variant with the value of `v` removed from it.

`func (me MyEnum) RemoveAll(v ...MyEnum) MyEnum` - RemoveAll returns a copy of the variant with all the values of `v` removed from it.

`func (me MyEnum) Has(v MyEnum) MyEnum` - Has returns `true` if the receiver contains the value of `v`, otherwise `false`.

`func (me MyEnum) HasAny(v MyEnum) MyEnum` - HasAny returns `true` if the receiver contains any of the values of `v`, otherwise `false`.

`func (me MyEnum) HasAll(v MyEnum) MyEnum` - HasAll returns `true` if the receiver contains all the values of `v`, otherwise `false`.

#####*Other items*

`MyEnum.Values` - By default, the `Values` field of your enum will reference an array of all the enum variants in their defined order. This can be used to iterate the variants in a `range` loop.

If the name `Values` conflicts with a variant name you've defined, you can define an alternate name for the array by using the `--iterator_name` option.
