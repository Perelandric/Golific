# GoEnum

**GoEnum** works with the `generate` command to create namespaced enums using structs, providing greater type safety as well as other features, including:

 - A `String()` method, allowing you to print the name of the enum variant.
 - The ability to optionally assign a custom value for `String()`.
 - A `Description()` method, that will print an optional description of each variant.
 - The ability to marshal and/or unmarshal JSON (TODO: and XML) to the variant name or custom string value instead of the number.
 - A `Value()` method for retreiving the numeric representation of the variant.
 - The ability to assign a custom numeric value.
 - The ability to define an enum's variants as bitflags.
 - A generated array of the variants, to be used with a `range` loop.
 

# FAQ
*General*
 - **Why was this created?**
  - Primarily in order to achieve greater type safety by restricting values of an enum type to only those variants provided.
 - **Can't this be done with `const` and a type alias?**
  - Yes, however a value of the base type can be substituted accidentally, resulting in bugs. Also, using `consts` pollutes the variable namespace, which can be an issue when overlapping names are needed in different categories.

*Functionality*
 - **How do I access the numeric representation of a variant?**
  - Use the `.Value()` method.
 - **Can I choose the numeric representation?**
  - Yes, as long as it doesn't overlap another within the same enum.
 - **Will GoEnum generate bitflag numbers for me?**
  - Yes, using the `--bitflags` flag.
 - **Can I get the name of a variant as a `string`? If so, can I define a string that differs from the variant name?**
  - Yes, using the `.String()` method and yes, a custom string can be defined using the `--string` flag.
 - **Can any meta data be associated with each variant?**
  - Yes, each variant can have a description assigned using the `--description` flag, which is accessed using the `.Description()` method.
 - **Can I have JSON (TODO: and XML) marshaled to and unmarshaled from the string value instead of the number?**
  - Yes, using the `--marshal` and `--unmarshal` flags. There are also flags to set specifically JSON or XML.
 - **Can I enumerate the variants of an enum using a `range` loop?**
  - Yes, an array holding the variants is generated, which can be used in a typical `range` loop.

*Efficiency*
 - **How are the variants stored and referenced?**
  - For each enum, the variants are stored together in an anonymous struct value assigned to a variable.
 - **How are the variants represented in memory?**
  - The individual variants are stored as a value of a struct type that has a single `uint` field, sized to the smallest size needed for each given enum.
 - **Does each variant being a struct value add extra memory overhead?**
  - No, variants that use a `uint8` will still use only 8 bits.
 - **Doesn't the `.Value()` call add overhead when getting the underlying number?**
  - Only if the compiler does not inline the call. However, the method simply returns the value of the field, so it would seem about as likely a candidate for inlining as one can hope to find.
 - **Does GoEnum use reflection?**
  - No, because the code is generated, we can hardcode necessary values into `switch` statements where needed, making the generated code longer, but faster.

*Safety*
 - **Isn't it still possible to use a value of the base type in place of one of the variants?**
  - Technically yes, however the variants for each enum are generated with a `value` field that has a unique identifier appended to it, like `value_1cn7iw6qxr8ad`, so using a base value would be cumbersome and never accidental.
 - **Are new unique identifiers used in the variants' structs generated every time `generate` is run?**
  - Yes. A pseudo-random number is used, so it is non-deterministic.
 - **Isn't it possible to overwrite one variant with another from the same enum?**
  - Unfortunately, Go does not allow struct values to be assigned to a `const`, so yes. However it would require `MyEnum.Foo = MyEnum.Bar`, which seems like an unlikely mistake.


# Documentation

*(coming soon)*
