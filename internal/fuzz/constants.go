package fuzz

// PrefixTypeNumber type
const PrefixTypeNumber = "__number__"

// PrefixTypeBoolean type
const PrefixTypeBoolean = "__boolean__"

// PrefixTypeString type
const PrefixTypeString = "__string__"

// PrefixTypeExample type
const PrefixTypeExample = "__example__"

// PrefixTypeObject type
const PrefixTypeObject = "__object__"

// PrefixTypeArray type
const PrefixTypeArray = "__array__"

// UintPrefixRegex constant
const UintPrefixRegex = PrefixTypeNumber + `\d{1,10}`

// IntPrefixRegex constant
const IntPrefixRegex = PrefixTypeNumber + `[+-]?\d{1,10}`

// NumberPrefixRegex constant
const NumberPrefixRegex = PrefixTypeNumber + `[+-]?((\d{1,10}(\.\d{1,5})?)|(\.\d{1,10}))`

// BooleanPrefixRegex constant
const BooleanPrefixRegex = PrefixTypeBoolean + `(false|true)`

// EmailRegex constant
const EmailRegex = `\w+@\w+\\.\w+`

// EmailRegex2 constant
const EmailRegex2 = `\w+@\w+.?\w+`

// EmailRegex3 constant
const EmailRegex3 = `.+@.+\..+`

// EmailRegex4 constant
const EmailRegex4 = `.+@.+\\..+`

// AnyWordRegex  constant
const AnyWordRegex = `\w+`

// WildRegex  constant
const WildRegex = `.+`

// UnescapeHTML flag
const UnescapeHTML = "UnescapeHTML"

// RequestCount name
const RequestCount = "_RequestCount"

// FixtureDataExt extension
const FixtureDataExt = ".dat"
