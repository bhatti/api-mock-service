package fuzz

import (
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"os"
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

func Test_ShouldExtractTypesSubArray(t *testing.T) {
	j := `{"userId": "sample123", "id": "us-west-1-1234", "regions": ["us-west-1", "us-east-1"], "account": "2a3BC", "description": "quia et rem eveniet architecto"}`
	res, err := UnmarshalArrayOrObject([]byte(j))
	require.NoError(t, err)
	actual := ExtractTypes(res, NewDataTemplateRequest(false, 1, 1)).(map[string]any)
	require.Equal(t, 5, len(actual))
	res = GenerateFuzzData(actual)
	require.NotNil(t, res)
}

func Test_ShouldUnmarshalObjectArrays(t *testing.T) {
	j := "{\"creditCard\":{\"balance\":{\"amount\":{{RandNumMinMax 0 0}},\"currency\":\"{{RandRegex `(USD|CAD|EUR|AUD)`}}\"}}}"
	_, err := UnmarshalArrayOrObject([]byte(j))
	require.Error(t, err) // no quotes for {RandNumMinMax 0 0}
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
