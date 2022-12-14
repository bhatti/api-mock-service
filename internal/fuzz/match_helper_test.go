package fuzz

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"os"
	"reflect"
	"testing"
)

func Test_ShouldFlatRegexMapWithArray(t *testing.T) {
	// GIVEN a mock scenario loaded from YAML
	b, err := os.ReadFile("../../fixtures/users.yaml")
	require.NoError(t, err)

	scenario := make(map[string]any)
	// AND it should return valid mock scenario
	err = yaml.Unmarshal(b, &scenario)
	require.NoError(t, err)
	response := scenario["response"].(map[string]any)
	contents := response["contents"].(string)
	res, err := UnmarshalArrayOrObject([]byte(contents))
	require.NoError(t, err)
	regexMap := FlatRegexMap(ExtractTypes(res, NewDataTemplateRequest(false, 1, 1)))
	err = ValidateRegexMap(res, regexMap)
	require.NoError(t, err)
}

func Test_ShouldFlatRegexMap(t *testing.T) {
	// GIVEN a mock scenario loaded from YAML
	b, err := os.ReadFile("../../fixtures/users.yaml")
	require.NoError(t, err)

	scenario := make(map[string]any)
	// AND it should return valid mock scenario
	err = yaml.Unmarshal(b, &scenario)
	require.NoError(t, err)
	response := scenario["response"].(map[string]any)
	contents := response["contents"].(string)
	res, err := UnmarshalArrayOrObject([]byte(contents))
	require.NoError(t, err)
	regexMap := FlatRegexMap(ExtractTypes(res, NewDataTemplateRequest(false, 1, 1)))
	arr := res.([]any)
	err = ValidateRegexMap(arr[0], regexMap)
	require.NoError(t, err)
}

func Test_ShouldFlatRegexMapNil(t *testing.T) {
	regex := make(map[string]string)
	flatRegexMap(nil, regex, "")
	require.Equal(t, 0, len(regex))
}

func Test_ShouldFlatRegexMapStringMap(t *testing.T) {
	regex := make(map[string]string)
	flatRegexMap(map[string]string{"k": "v"}, regex, "")
	require.Equal(t, "(v)", regex["k"])
}

func Test_ShouldFlatRegexMapInt(t *testing.T) {
	regex := make(map[string]string)
	flatRegexMap(1, regex, "")
	require.Equal(t, 0, len(regex))
}

func Test_ShouldValidateRegexMapNil(t *testing.T) {
	regex := make(map[string]string)
	require.NoError(t, validateRegexMap(nil, regex, ""))
}

func Test_ShouldValidateRegexMapStringMap(t *testing.T) {
	regex := map[string]string{"k": `\w`}
	require.NoError(t, validateRegexMap(map[string]string{"k": "v"}, regex, ""))
	regex = map[string]string{"k": `\d`}
	require.Error(t, validateRegexMap(map[string]string{"k": "v"}, regex, ""))
}

func Test_ShouldValidateRegexMapInt(t *testing.T) {
	regex := make(map[string]string)
	require.NoError(t, validateRegexMap(1, regex, ""))
}

func Test_ShouldExtractTypesNil(t *testing.T) {
	val := ExtractTypes(nil, NewDataTemplateRequest(false, 1, 1))
	require.Nil(t, val)
}

func Test_ShouldExtractTypesString(t *testing.T) {
	val := ExtractTypes("", NewDataTemplateRequest(false, 1, 1))
	require.Equal(t, PrefixTypeString+`[a-z]{1,10}`, val)
}

func Test_ShouldExtractTypesStringValue(t *testing.T) {
	val := ExtractTypes("abc", NewDataTemplateRequest(false, 1, 1))
	require.Equal(t, PrefixTypeString+`\w+`, val)
}

func Test_ShouldExtractTypesBool(t *testing.T) {
	val := ExtractTypes(true, NewDataTemplateRequest(false, 1, 1))
	require.Equal(t, PrefixTypeBoolean+"(false|true)", val)
}

func Test_ShouldExtractTypesInt(t *testing.T) {
	val := ExtractTypes(3, NewDataTemplateRequest(false, 1, 1))
	require.Equal(t, IntPrefixRegex, val)
}

func Test_ShouldExtractTypesInt8(t *testing.T) {
	val := ExtractTypes(int8(3), NewDataTemplateRequest(false, 1, 1))
	require.Equal(t, IntPrefixRegex, val)
}

func Test_ShouldExtractTypesInt16(t *testing.T) {
	val := ExtractTypes(int16(3), NewDataTemplateRequest(false, 1, 1))
	require.Equal(t, IntPrefixRegex, val)
}

func Test_ShouldExtractTypesInt32(t *testing.T) {
	val := ExtractTypes(int32(3), NewDataTemplateRequest(false, 1, 1))
	require.Equal(t, IntPrefixRegex, val)
}

func Test_ShouldExtractTypesInt64(t *testing.T) {
	val := ExtractTypes(int64(3), NewDataTemplateRequest(false, 1, 1))
	require.Equal(t, IntPrefixRegex, val)
}

func Test_ShouldExtractTypesUInt(t *testing.T) {
	val := ExtractTypes(uint(3), NewDataTemplateRequest(false, 1, 1))
	require.Equal(t, UintPrefixRegex, val)
}

func Test_ShouldExtractTypesUInt8(t *testing.T) {
	val := ExtractTypes(uint8(3), NewDataTemplateRequest(false, 1, 1))
	require.Equal(t, UintPrefixRegex, val)
}

func Test_ShouldExtractTypesUInt16(t *testing.T) {
	val := ExtractTypes(uint16(3), NewDataTemplateRequest(false, 1, 1))
	require.Equal(t, UintPrefixRegex, val)
}

func Test_ShouldExtractTypesUInt32(t *testing.T) {
	val := ExtractTypes(uint32(3), NewDataTemplateRequest(false, 1, 1))
	require.Equal(t, UintPrefixRegex, val)
}

func Test_ShouldExtractTypesUInt64(t *testing.T) {
	val := ExtractTypes(uint64(3), NewDataTemplateRequest(false, 1, 1))
	require.Equal(t, UintPrefixRegex, val)
}

func Test_ShouldExtractTypesFloat32(t *testing.T) {
	val := ExtractTypes(float32(-13.5), NewDataTemplateRequest(false, 1, 1))
	require.Equal(t, NumberPrefixRegex, val)
}

func Test_ShouldExtractTypesFloat64(t *testing.T) {
	val := ExtractTypes(float32(-13.5), NewDataTemplateRequest(false, 1, 1))
	require.Equal(t, NumberPrefixRegex, val)
}

func Test_ShouldExtractTypesObject(t *testing.T) {
	j := `{"userId": 1, "id": 1, "title": "sunt aut", "body": "quia et rem eveniet architecto"}`
	res, err := UnmarshalArrayOrObject([]byte(j))
	require.NoError(t, err)
	actual := ExtractTypes(res, NewDataTemplateRequest(false, 1, 1)).(map[string]any)

	require.Equal(t, 4, len(actual))
}

func Test_ShouldExtractTypesArray(t *testing.T) {
	j := `
[
           {
             "id": 1,
             "name": "Leanne Graham",
             "username": "Bret",
             "email": "Sincere@april.biz",
             "address": {
               "street": "Kulas Light",
               "suite": "Apt. 556",
               "city": "Gwenborough",
               "zipcode": "92998-3874",
               "geo": {
                 "lat": "-37.3159",
                 "lng": "81.1496"
               }
             },
             "phone": "1-770-736-8031 x56442",
             "website": "hildegard.org",
             "company": {
               "name": "Romaguera-Crona",
               "catchPhrase": "Multi-layered client-server neural-net",
               "bs": "harness real-time e-markets"
             }
           },
           {
             "id": 2,
             "name": "Ervin Howell",
             "username": "Antonette",
             "email": "Shanna@melissa.tv",
             "address": {
               "street": "Victor Plains",
               "suite": "Suite 879",
               "city": "Wisokyburgh",
               "zipcode": "90566-7771",
               "geo": {
                 "lat": "-43.9509",
                 "lng": "-34.4618"
               }
             },
             "phone": "010-692-6593 x09125",
             "website": "anastasia.net",
             "company": {
               "name": "Deckow-Crist",
               "catchPhrase": "Proactive didactic contingency",
               "bs": "synergize scalable supply-chains"
             }
           }
]
`
	res, err := UnmarshalArrayOrObject([]byte(j))
	require.NoError(t, err)
	array := res.([]any)
	actual := ExtractTypes(res, NewDataTemplateRequest(false, 1, 1)).([]any)
	require.Equal(t, len(array), len(actual))
}

func Test_ShouldPopulateRandomDataNil(t *testing.T) {
	val := ExtractTypes(nil, NewDataTemplateRequest(false, 1, 1))
	val = PopulateRandomData(nil)
	require.Nil(t, val)
}

func Test_ShouldPopulateRandomDataStringEmpty(t *testing.T) {
	val := ExtractTypes("", NewDataTemplateRequest(false, 1, 1))
	val = PopulateRandomData(val)
	require.Equal(t, "string", reflect.TypeOf(val).String())
}

func Test_ShouldPopulateRandomDataString(t *testing.T) {
	val := ExtractTypes("", NewDataTemplateRequest(false, 1, 2))
	val = PopulateRandomData(val)
	require.Equal(t, "string", reflect.TypeOf(val).String())
	require.True(t, len(val.(string)) > 0)
}

func Test_ShouldPopulateRandomDataStringValue(t *testing.T) {
	val := ExtractTypes("abc", NewDataTemplateRequest(false, 1, 1))
	val = PopulateRandomData(val)
	require.Equal(t, "string", reflect.TypeOf(val).String())
	require.True(t, len(val.(string)) > 0)
}

func Test_ShouldPopulateRandomDataBool(t *testing.T) {
	val := ExtractTypes(true, NewDataTemplateRequest(false, 1, 1))
	val = PopulateRandomData(val)
	require.Equal(t, "bool", reflect.TypeOf(val).String())

	val = PopulateRandomData(true)
	require.Equal(t, "bool", reflect.TypeOf(val).String())
}

func Test_ShouldPopulateRandomDataInt(t *testing.T) {
	val := ExtractTypes(3, NewDataTemplateRequest(false, 1, 1))
	val = PopulateRandomData(val)
	require.Equal(t, "int", reflect.TypeOf(val).String())

	val = PopulateRandomData(3)
	require.Equal(t, "int", reflect.TypeOf(val).String())
}

func Test_ShouldPopulateRandomDataInt8(t *testing.T) {
	val := ExtractTypes(int8(3), NewDataTemplateRequest(false, 1, 1))
	val = PopulateRandomData(val)
	require.Equal(t, "int", reflect.TypeOf(val).String())

	val = PopulateRandomData(int8(5))
	require.Equal(t, "int8", reflect.TypeOf(val).String())
}

func Test_ShouldPopulateRandomDataInt16(t *testing.T) {
	val := ExtractTypes(int16(3), NewDataTemplateRequest(false, 1, 1))
	val = PopulateRandomData(val)
	require.Equal(t, "int", reflect.TypeOf(val).String())

	val = PopulateRandomData(int16(1))
	require.Equal(t, "int16", reflect.TypeOf(val).String())
}

func Test_ShouldPopulateRandomDataInt32(t *testing.T) {
	val := ExtractTypes(int32(3), NewDataTemplateRequest(false, 1, 1))
	val = PopulateRandomData(val)
	require.Equal(t, "int", reflect.TypeOf(val).String())

	val = PopulateRandomData(int32(5))
	require.Equal(t, "int32", reflect.TypeOf(val).String())
}

func Test_ShouldPopulateRandomDataInt64(t *testing.T) {
	val := ExtractTypes(int64(3), NewDataTemplateRequest(false, 1, 1))
	val = PopulateRandomData(val)
	require.Equal(t, "int", reflect.TypeOf(val).String())

	val = PopulateRandomData(int64(1))
	require.Equal(t, "int64", reflect.TypeOf(val).String())
}

func Test_ShouldPopulateRandomDataUInt(t *testing.T) {
	val := ExtractTypes(uint(3), NewDataTemplateRequest(false, 1, 1))
	val = PopulateRandomData(val)
	require.Equal(t, "int", reflect.TypeOf(val).String())

	val = PopulateRandomData(uint(1))
	require.Equal(t, "uint", reflect.TypeOf(val).String())
}

func Test_ShouldPopulateRandomDataUInt8(t *testing.T) {
	val := ExtractTypes(uint8(3), NewDataTemplateRequest(false, 1, 1))
	val = PopulateRandomData(val)
	require.Equal(t, "int", reflect.TypeOf(val).String())

	val = PopulateRandomData(uint8(1))
	require.Equal(t, "uint8", reflect.TypeOf(val).String())
}

func Test_ShouldPopulateRandomDataUInt16(t *testing.T) {
	val := ExtractTypes(uint16(3), NewDataTemplateRequest(false, 1, 1))
	val = PopulateRandomData(val)
	require.Equal(t, "int", reflect.TypeOf(val).String())

	val = PopulateRandomData(uint16(1))
	require.Equal(t, "uint16", reflect.TypeOf(val).String())
}

func Test_ShouldPopulateRandomDataUInt32(t *testing.T) {
	val := ExtractTypes(uint32(3), NewDataTemplateRequest(false, 1, 1))
	val = PopulateRandomData(val)
	require.Equal(t, "int", reflect.TypeOf(val).String())

	val = PopulateRandomData(uint32(1))
	require.Equal(t, "uint32", reflect.TypeOf(val).String())
}

func Test_ShouldPopulateRandomDataUInt64(t *testing.T) {
	val := ExtractTypes(uint64(3), NewDataTemplateRequest(false, 1, 1))
	val = PopulateRandomData(val)
	require.Equal(t, "int", reflect.TypeOf(val).String())

	val = PopulateRandomData(uint64(1))
	require.Equal(t, "uint64", reflect.TypeOf(val).String())
}

func Test_ShouldPopulateRandomDataFloat32(t *testing.T) {
	val := ExtractTypes(float32(-13.5), NewDataTemplateRequest(false, 1, 1))
	require.Equal(t, "string", reflect.TypeOf(val).String())
	val = PopulateRandomData(val)
	require.Equal(t, "float64", reflect.TypeOf(val).String())

	val = PopulateRandomData(float32(1.1))
	require.Equal(t, "float32", reflect.TypeOf(val).String())
}

func Test_ShouldPopulateRandomDataFloat64(t *testing.T) {
	val := ExtractTypes(-13.5, NewDataTemplateRequest(false, 1, 1))
	val = PopulateRandomData(val)
	require.Equal(t, "float64", reflect.TypeOf(val).String())

	val = PopulateRandomData(1.2)
	require.Equal(t, "float64", reflect.TypeOf(val).String())
}

func Test_ShouldExtractTypesStringMap(t *testing.T) {
	val := ExtractTypes(map[string]string{"key": "val"}, NewDataTemplateRequest(false, 1, 1))
	require.Equal(t, 1, len(val.(map[string]string)))
}

func Test_ShouldExtractTypesUnknown(t *testing.T) {
	val := ExtractTypes(complex(1, 1), NewDataTemplateRequest(false, 1, 1))
	require.Nil(t, val)
}

func Test_ShouldPopulateRandomDataZip(t *testing.T) {
	val := ExtractTypes("12345-1234", NewDataTemplateRequest(false, 1, 1))
	val = PopulateRandomData(val)
	require.Equal(t, "string", reflect.TypeOf(val).String())
	require.True(t, len(val.(string)) > 5)
}

func Test_ShouldPopulateRandomDataNegativeFloat(t *testing.T) {
	val := ExtractTypes("-1234.12", NewDataTemplateRequest(false, 1, 1))
	val = PopulateRandomData(val)
	require.Equal(t, "string", reflect.TypeOf(val).String())
	require.True(t, len(val.(string)) > 0)
}

func Test_ShouldPopulateRandomDataNegative(t *testing.T) {
	val := ExtractTypes("-1234", NewDataTemplateRequest(false, 1, 1))
	val = PopulateRandomData(val)
	require.Equal(t, "string", reflect.TypeOf(val).String())
	require.True(t, len(val.(string)) > 0)
}

func Test_ShouldPopulateRandomDataObject(t *testing.T) {
	j := `{"userId": 1, "id": 1, "title": "sunt aut", "body": "quia et rem eveniet architecto"}`
	res, err := UnmarshalArrayOrObject([]byte(j))
	require.NoError(t, err)
	val := ExtractTypes(res, NewDataTemplateRequest(false, 1, 1))
	actual := PopulateRandomData(val).(map[string]any)
	require.Equal(t, 4, len(actual))
}

func Test_ShouldPopulateRandomDataArray(t *testing.T) {
	j := `
[
           {
             "id": 1,
             "name": "Leanne Graham",
             "username": "Bret",
             "email": "Sincere@april.biz",
             "address": {
               "street": "Kulas Light",
               "suite": "Apt. 556",
               "city": "Gwenborough",
               "zipcode": "92998-3874",
               "geo": {
                 "lat": "-37.3159",
                 "lng": "81.1496"
               }
             },
             "phone": "1-770-736-8031 x56442",
             "website": "hildegard.org",
             "company": {
               "name": "Romaguera-Crona",
               "catchPhrase": "Multi-layered client-server neural-net",
               "bs": "harness real-time e-markets"
             }
           },
           {
             "id": 2,
             "name": "Ervin Howell",
             "username": "Antonette",
             "email": "Shanna@melissa.tv",
             "address": {
               "street": "Victor Plains",
               "suite": "Suite 879",
               "city": "Wisokyburgh",
               "zipcode": "90566-7771",
               "geo": {
                 "lat": "-43.9509",
                 "lng": "-34.4618"
               }
             },
             "phone": "010-692-6593 x09125",
             "website": "anastasia.net",
             "company": {
               "name": "Deckow-Crist",
               "catchPhrase": "Proactive didactic contingency",
               "bs": "synergize scalable supply-chains"
             }
           }
]
`
	res, err := UnmarshalArrayOrObject([]byte(j))
	require.NoError(t, err)
	val := ExtractTypes(res, NewDataTemplateRequest(false, 1, 1))
	actual := PopulateRandomData(val).([]any)
	require.Equal(t, 2, len(actual))
}

func Test_ShouldPopulateRandomDataItems(t *testing.T) {
	input := []string{
		`{"id": 1}`,
		`{"id": "Leanne Graham"}`,
		`{"id": "Sincere@april.biz"}`,
		`{"id": "Apt. 556"}`,
		`{"id": "92998-3874"}`,
		`{"id": "-37.3159"}`,
		`{"id": "81.1496"}`,
		`{"id": "1-770-736-8031 x56442"}`,
		`{"id": "hildegard.org"}`,
		`{"id": "Multi-layered client-server neural-net"}`,
	}
	expected := [][]string{
		{"__number__[+-]?[0-9]{1,10}", "int"},
		{`__string__\w+`, "string"},
		{`__string__` + EmailRegex, "string"},
		{`__string__\w+.?\w+[0-9]{3,3}`, "string"},
		{`__string__[0-9]{5,5}[-][0-9]{4,4}`, "string"},
		{`__string__[+-]?[0-9]{2,2}\.[0-9]{4,4}`, "string"},
		{`__string__[0-9]{2,2}\.[0-9]{4,4}`, "string"},
		{`__string__[0-9]{1,1}[-][0-9]{3,3}[-][0-9]{3,3}[-][0-9]{4,4}\w+[0-9]{5,5}`, "string"},
		{`__string__\w+.?\w+`, "string"},
		{`__string__\w+[-]\w+[-]\w+[-]\w+`, "string"},
	}
	for i, j := range input {
		res, err := UnmarshalArrayOrObject([]byte(j))
		require.NoError(t, err)
		val := ExtractTypes(res, NewDataTemplateRequest(false, 1, 1))
		require.Equal(t, expected[i][0], val.(map[string]any)["id"], fmt.Sprintf("test %d", i))
		actual := PopulateRandomData(val).(map[string]any)
		require.Equal(t, expected[i][1], reflect.TypeOf(actual["id"]).String(), fmt.Sprintf("test %d", i))
	}
}

func Test_ShouldUnmarshalArrayOrObjectAndExtractTypesAndMarshal(t *testing.T) {
	// GIVEN a mock scenario loaded from YAML
	b, err := os.ReadFile("../../fixtures/users.yaml")
	require.NoError(t, err)
	scenario := make(map[string]any)
	// AND it should return valid mock scenario
	err = yaml.Unmarshal(b, &scenario)
	require.NoError(t, err)
	response := scenario["response"].(map[string]any)
	contents := response["contents"].(string)
	str, err := UnmarshalArrayOrObjectAndExtractTypesAndMarshal(contents, NewDataTemplateRequest(false, 1, 1))
	require.NoError(t, err)
	require.Contains(t, str, "address")
}

func Test_ShouldPopulateRandomDataStringMapDta(t *testing.T) {
	val := PopulateRandomData(map[string]string{"k": "1"})
	require.Equal(t, "map[string]interface {}", reflect.TypeOf(val).String())
}

func Test_ShouldPopulateRandomDataStringUnknown(t *testing.T) {
	require.NotNil(t, PopulateRandomData(complex(1, 1)))
}

func Test_ShouldPopulateRandomDataStringMapRegex(t *testing.T) {
	strJSON := `{"completed":"(__boolean__(false|true))","id":"(__number__[+-]?[0-9]{1,10})","title":"(__string__\\w+)","userId":"(__number__[+-]?[0-9]{1,10})"}`
	val, err := UnmarshalArrayOrObject([]byte(strJSON))
	require.NoError(t, err)
	val = PopulateRandomData(val)
	require.Equal(t, "map[string]interface {}", reflect.TypeOf(val).String())
}

func Test_ShouldNotPopulateRandomDataStringWithoutRegex(t *testing.T) {
	require.Equal(t, "__1", PopulateRandomData("__1"))
	require.Equal(t, "1", PopulateRandomData("(1)"))
	require.Equal(t, "1", PopulateRandomData("1"))
}
