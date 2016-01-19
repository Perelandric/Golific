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

 This short example of how the enum descriptor syntax looks does not utilize all available features. See the documentation for more info.

 Installation:
 ```
 go install github.com/Perelandric/GoEnum
 ```

 Top of file (below imports):
 ```
 //go:generate GoEnum $GOFILE
 ```

 Enum descriptor syntax:
 ```
 /*
 @enum --name=Animal --json=string
 Dog --string=dog --description="Your best friend, and you know it."
 Cat --string=cat --description="Your best friend, but doesn't always show it."
 Horse --string=horse --description="Everyone loves horses."
 */
 ```

 Run the generate tool from the project directory:
 ```
 go generate
 ```

 Use the enum in your code:

 ```
 type Resident struct {
   Name string
   Pet AnimalEnum
 }

 res := Resident{
   Name: "Charlie Brown",
   Pet: Animal.Dog,
 }

 j, err := json.Marshal(&res)
 fmt.Printf("%s\n", j) // {"Name":"Charlie Brown","Pet":"dog"}

 for _, animal := range AnimalValues {
   fmt.Printf("Kind: %s, Description: %q\n", animal, animal.Description())
 }
 ```

# FAQ
*General*
 - **Why was this created?**
  - Primarily in order to achieve greater type safety by restricting values of an enum type to only those variants provided.
 - **Can't this be done with `const` and a type alias?**
  - Yes, however a value of the base type can be substituted accidentally, resulting in bugs. Also, using `consts` pollutes the variable namespace, which can be an issue when overlapping names are needed in different categories.

*Functionality*
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

*Efficiency*
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

*Safety*
 - **Is it still possible to use a value of the base (struct) type in place of one of the variants?**
  - Technically yes, however the variants for each enum use a struct type with a `value` field that has a unique identifier appended to it, i.e. `value_1cn7iw6qxr8ad`, so substituting a base value would be cumbersome and never accidental.
 - **Are the new unique identifiers used in the variants' structs generated every time `generate` is run?**
  - Yes. A pseudo-random number is used with a time-based seed, so it is non-deterministic.
 - **Is it possible to overwrite one variant with another from the same enum?**
  - Unfortunately, Go does not allow struct values to be assigned to a `const`, so yes. However it would require `Animal.Cat = Animal.Horse`, which seems like an unlikely mistake.

# Documentation

*Installation*

To install GoEnum, use the `install` command from the Go toolchain.

```
go install github.com/Perelandric/GoEnum
```

*Using the `generate` command*

After your source code has been properly annotated as described below, then from your source directory run the `generate` command.

```
go generate
```

This creates a new file for every file that had the proper annotations. The new file has the same name as the original, but with an `enum____` prefix, so make sure you don't already have a file with a conflicting name. The generated file also gets the same `package` name as the original.

Do not edit the generated file, as it will be overwritten every time you call `generate`.

*Annotations*

As with all files that rely on Go's `generate` command, your file must have the `go:generate` directive below the file package name and imports. For GoEnum, it should look like this:

```
//go:generate GoEnum $GOFILE
```

Notice that there's no space after the `//` and before `go:`. This is required for the `generate` tool.

*Enum descriptors*

The actual Enum descriptors are defined entirely in code comment blocks at the top-level namespace of your code.

It doesn't matter if you use multiple single-line or multi-line comment, as long as the comment begins with `@enum`. *(Comment lines that are empty or have only whitespace are ignored.)* If using multiple single-line comments, any non-comment line marks the end of the block.

Multiple `@enum` descriptors may be defined in a single comment block.

The beginning of an descriptor can look like either of these *(incomplete)* examples:

```
/*
@enum
*/
```
```
//
// @enum
```

Notice that both have an empty line above the `@enum`. This is fine since empty or whitespace-only comment lines are ignored.

*Descriptor flags*

The next thing to come after the `@enum` are the descriptor flags. These flags begin with `--` and are followed by a word and in some cases a `=` with a value. Flags must be separated by at least 1 white space, and can span multiple lines. They can also begin directly after the `@enum` on the same line.

There's always at least one flag required. That's the `--name` flag. This provides the name that will be used to reference your enum as well as the type of the variants *(to which the word "Enum" will be added)*.

So adding the `name` flag, our *(still incomplete)* examples now look like this:

```
/*
@enum --name=Animal
*/
```
```
//
// @enum
// --name="Animal"
```

In both cases we define an enum named `Animal`. This will create a `var Animal` in the generated file, a `type AnimalEnum struct {...}` and a `var AnimalValues`, so all these names must be available.

Notice two differences in the above examples.
 - The first one uses quotes around `Animal`, and the second does not. As long as a flag value does not contain space characters, the quotation marks are optional.
 - The first one begins its flags on the same line as the `@enum`, and the second starts on the next line. Either way is valid.

Because the `--name` is used as identifiers in code, the name must be a valid identifier. As usual, the capitalization will determine whether or not the items are exported.

The rest of the flags are optional, and are described in the table below.

*Variant definitions*

After all the descriptor flags have been defined, you'll need to define the enum variants. At least one variant is required per enum.

Each variant must be entirely defined on its own line. Lines that are long because of variant flags must not be split across multiple lines.

Adding to the examples above, they may now look like this:

```
/*
@enum --name=Animal
Dog --string=dog --description="Your best friend, and you know it."
Cat --string=cat --description="Your best friend, but doesn't always show it."
Horse --string=horse --description="Everyone loves horses."
*/
```
```
//
// @enum
// --name="Animal"
// Dog --string=dog --description="Your best friend, and you know it."
// Cat --string=cat --description="Your best friend, but doesn't always show it."
// Horse --string=horse --description="Everyone loves horses."
```

So our examples are now fully valid enum descriptors. As long as you have the `go:generate` annotation previously defined, you'll be able to run `go generate` and your new source file will be generated.

The rest of the documentation will use the multi-line version of our descriptor example.
