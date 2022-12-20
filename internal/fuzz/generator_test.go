package fuzz

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func Test_ShouldPopulateRandomDataNil(t *testing.T) {
	val := ExtractTypes(nil, NewDataTemplateRequest(false, 1, 1))
	val = GenerateFuzzData(nil)
	require.Nil(t, val)
}

func Test_ShouldPopulateRandomDataStringEmpty(t *testing.T) {
	val := ExtractTypes("", NewDataTemplateRequest(false, 1, 1))
	val = GenerateFuzzData(val)
	require.Equal(t, "string", reflect.TypeOf(val).String())
}

func Test_ShouldPopulateRandomDataString(t *testing.T) {
	val := ExtractTypes("", NewDataTemplateRequest(false, 1, 2))
	val = GenerateFuzzData(val)
	require.Equal(t, "string", reflect.TypeOf(val).String())
	require.True(t, len(val.(string)) > 0)
}

func Test_ShouldPopulateRandomDataStringValue(t *testing.T) {
	val := ExtractTypes("abc", NewDataTemplateRequest(false, 1, 1))
	val = GenerateFuzzData(val)
	require.Equal(t, "string", reflect.TypeOf(val).String())
	require.True(t, len(val.(string)) > 0)
}

func Test_ShouldPopulateRandomDataBool(t *testing.T) {
	val := ExtractTypes(true, NewDataTemplateRequest(false, 1, 1))
	val = GenerateFuzzData(val)
	require.Equal(t, "bool", reflect.TypeOf(val).String())

	val = GenerateFuzzData(true)
	require.Equal(t, "bool", reflect.TypeOf(val).String())
}

func Test_ShouldPopulateRandomDataInt(t *testing.T) {
	val := ExtractTypes(3, NewDataTemplateRequest(false, 1, 1))
	val = GenerateFuzzData(val)
	require.Equal(t, "int", reflect.TypeOf(val).String())

	val = GenerateFuzzData(3)
	require.Equal(t, "int", reflect.TypeOf(val).String())
}

func Test_ShouldPopulateRandomDataInt8(t *testing.T) {
	val := ExtractTypes(int8(3), NewDataTemplateRequest(false, 1, 1))
	val = GenerateFuzzData(val)
	require.Equal(t, "int", reflect.TypeOf(val).String())

	val = GenerateFuzzData(int8(5))
	require.Equal(t, "int8", reflect.TypeOf(val).String())
}

func Test_ShouldPopulateRandomDataInt16(t *testing.T) {
	val := ExtractTypes(int16(3), NewDataTemplateRequest(false, 1, 1))
	val = GenerateFuzzData(val)
	require.Equal(t, "int", reflect.TypeOf(val).String())

	val = GenerateFuzzData(int16(1))
	require.Equal(t, "int16", reflect.TypeOf(val).String())
}

func Test_ShouldPopulateRandomDataInt32(t *testing.T) {
	val := ExtractTypes(int32(3), NewDataTemplateRequest(false, 1, 1))
	val = GenerateFuzzData(val)
	require.Equal(t, "int", reflect.TypeOf(val).String())

	val = GenerateFuzzData(int32(5))
	require.Equal(t, "int32", reflect.TypeOf(val).String())
}

func Test_ShouldPopulateRandomDataInt64(t *testing.T) {
	val := ExtractTypes(int64(3), NewDataTemplateRequest(false, 1, 1))
	val = GenerateFuzzData(val)
	require.Equal(t, "int", reflect.TypeOf(val).String())

	val = GenerateFuzzData(int64(1))
	require.Equal(t, "int64", reflect.TypeOf(val).String())
}

func Test_ShouldPopulateRandomDataUInt(t *testing.T) {
	val := ExtractTypes(uint(3), NewDataTemplateRequest(false, 1, 1))
	val = GenerateFuzzData(val)
	require.Equal(t, "int", reflect.TypeOf(val).String())

	val = GenerateFuzzData(uint(1))
	require.Equal(t, "uint", reflect.TypeOf(val).String())
}

func Test_ShouldPopulateRandomDataUInt8(t *testing.T) {
	val := ExtractTypes(uint8(3), NewDataTemplateRequest(false, 1, 1))
	val = GenerateFuzzData(val)
	require.Equal(t, "int", reflect.TypeOf(val).String())

	val = GenerateFuzzData(uint8(1))
	require.Equal(t, "uint8", reflect.TypeOf(val).String())
}

func Test_ShouldPopulateRandomDataUInt16(t *testing.T) {
	val := ExtractTypes(uint16(3), NewDataTemplateRequest(false, 1, 1))
	val = GenerateFuzzData(val)
	require.Equal(t, "int", reflect.TypeOf(val).String())

	val = GenerateFuzzData(uint16(1))
	require.Equal(t, "uint16", reflect.TypeOf(val).String())
}

func Test_ShouldPopulateRandomDataUInt32(t *testing.T) {
	val := ExtractTypes(uint32(3), NewDataTemplateRequest(false, 1, 1))
	val = GenerateFuzzData(val)
	require.Equal(t, "int", reflect.TypeOf(val).String())

	val = GenerateFuzzData(uint32(1))
	require.Equal(t, "uint32", reflect.TypeOf(val).String())
}

func Test_ShouldPopulateRandomDataUInt64(t *testing.T) {
	val := ExtractTypes(uint64(3), NewDataTemplateRequest(false, 1, 1))
	val = GenerateFuzzData(val)
	require.Equal(t, "int", reflect.TypeOf(val).String())

	val = GenerateFuzzData(uint64(1))
	require.Equal(t, "uint64", reflect.TypeOf(val).String())
}

func Test_ShouldPopulateRandomDataFloat32(t *testing.T) {
	val := ExtractTypes(float32(-13.5), NewDataTemplateRequest(false, 1, 1))
	require.Equal(t, "string", reflect.TypeOf(val).String())
	val = GenerateFuzzData(val)
	require.Equal(t, "float64", reflect.TypeOf(val).String())

	val = GenerateFuzzData(float32(1.1))
	require.Equal(t, "float32", reflect.TypeOf(val).String())
}

func Test_ShouldPopulateRandomDataFloat64(t *testing.T) {
	val := ExtractTypes(-13.5, NewDataTemplateRequest(false, 1, 1))
	val = GenerateFuzzData(val)
	require.Equal(t, "float64", reflect.TypeOf(val).String())

	val = GenerateFuzzData(1.2)
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
	val = GenerateFuzzData(val)
	require.Equal(t, "string", reflect.TypeOf(val).String())
	require.True(t, len(val.(string)) > 5)
}

func Test_ShouldPopulateRandomDataNegativeFloat(t *testing.T) {
	val := ExtractTypes("-1234.12", NewDataTemplateRequest(false, 1, 1))
	val = GenerateFuzzData(val)
	require.Equal(t, "string", reflect.TypeOf(val).String())
	require.True(t, len(val.(string)) > 0)
}

func Test_ShouldPopulateRandomDataNegative(t *testing.T) {
	val := ExtractTypes("-1234", NewDataTemplateRequest(false, 1, 1))
	val = GenerateFuzzData(val)
	require.Equal(t, "string", reflect.TypeOf(val).String())
	require.True(t, len(val.(string)) > 0)
}

func Test_ShouldPopulateRandomDataObject(t *testing.T) {
	j := `{"userId": 1, "id": 1, "title": "sunt aut", "body": "quia et rem eveniet architecto"}`
	res, err := UnmarshalArrayOrObject([]byte(j))
	require.NoError(t, err)
	val := ExtractTypes(res, NewDataTemplateRequest(false, 1, 1))
	actual := GenerateFuzzData(val).(map[string]any)
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
	actual := GenerateFuzzData(val).([]any)
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
		actual := GenerateFuzzData(val).(map[string]any)
		require.Equal(t, expected[i][1], reflect.TypeOf(actual["id"]).String(), fmt.Sprintf("test %d", i))
	}
}

func Test_ShouldPopulateRandomDataStringMapDta(t *testing.T) {
	val := GenerateFuzzData(map[string]string{"k": "1"})
	require.Equal(t, "map[string]interface {}", reflect.TypeOf(val).String())
}

func Test_ShouldPopulateRandomDataStringUnknown(t *testing.T) {
	require.NotNil(t, GenerateFuzzData(complex(1, 1)))
}

func Test_ShouldPopulateRandomDataStringMapRegex(t *testing.T) {
	strJSON := `{"completed":"(__boolean__(false|true))","id":"(__number__[+-]?[0-9]{1,10})","title":"(__string__\\w+)","userId":"(__number__[+-]?[0-9]{1,10})"}`
	val, err := UnmarshalArrayOrObject([]byte(strJSON))
	require.NoError(t, err)
	val = GenerateFuzzData(val)
	require.Equal(t, "map[string]interface {}", reflect.TypeOf(val).String())
}

func Test_ShouldNotPopulateRandomDataStringWithoutRegex(t *testing.T) {
	require.Equal(t, "__1", GenerateFuzzData("__1"))
	require.Equal(t, "1", GenerateFuzzData("(1)"))
	require.Equal(t, "1", GenerateFuzzData("1"))
}
