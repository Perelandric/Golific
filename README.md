

# Golific

**Golific** is a tool for generating Go code using the `go:generate` tool. Currently there are two types of annotations: **&#64;struct** and **&#64;enum**. See descriptions below.

## &#64;struct

**&#64;struct** functionality has been largely discarded and reduced down to adding a custom JSON marshaler that will omit a field that has `omitempty` if the field has an `IsZero()` method that returns `true`. This is useful for the **&#64;enum** type in this package, as well as types like `time.Time`.

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

Enum descriptor syntax in your source to create an enum named `Animal` that has three variants. Note the double underscore prefix on the name. This is a *requirement*.
``` go
/*
@enum json:"string"
*/
type __Animal struct {
	Dog   int `gString:"doggie", gDescription:"Loves to lick your face"`
	Cat   int `gString:"kitty", gDescription:"Loves to scratch your face"`
	Horse int `gString:"horsie", gDescription:"Has a very long face"`
}
```

Run Go's `generate` tool from the project directory:
```
go generate
```

Use the enum in your code:

``` go
type Resident struct {
 Name string
 Pet  AnimalEnum // The generated type for your Animal enum
}

res := Resident{
 Name: "Charlie Brown",
 Pet:  Animal.Dog, // Assign one of the variants
}

// The `json:"string"` option we included causes our custom `gString` value to be used when marshaled as JSON data
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
  - Yes, by using the `bitflags` option.
 - **Can I choose the numeric representation?**
  - Yes, as long as the `bitflags` option is not used, and the number doesn't match another value in the same enum.
 - **Can negative numbers be used for the numeric representation?**
  - No, the numbers must be `0` or greater and it is recommended that `0` be reserved to denote no value having been set, unless a default variant makes sense.
 - **Can I get the name of a variant as a `string`? If so, can I define a string that differs from the variant name?**
  - Yes, the `.Name()` method gives you the name and the `.String()` method gives you an optional custom string defined using the `gString` flag.
 - **Can meta data be associated with each variant?**
  - Yes, each variant can have a description assigned using the `--description` flag, which is accessed using the `.Description()` method.
 - **Can I have JSON marshaled to and unmarshaled from the string value instead of the number?**
  - Yes, using the `json` option.
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

