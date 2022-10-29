package utils

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ShouldGetRandomMinMax(t *testing.T) {
	require.True(t, RandNumMinMax(1, 10) >= 1)
}

func Test_ShouldGetRandom(t *testing.T) {
	require.True(t, Random(10) <= 10)
}

func Test_ShouldGetSeededRandom(t *testing.T) {
	require.True(t, SeededRandom(1, 10) <= 10)
}

func Test_ShouldGetUdid(t *testing.T) {
	require.Equal(t, 36, len(Udid()))
}

func Test_ShouldGetSeededUdid(t *testing.T) {
	require.Equal(t, 36, len(SeededUdid(1)))
}

func Test_ShouldGetRandBool(t *testing.T) {
	require.True(t, reflect.TypeOf(RandBool()) == reflect.TypeOf(true))
}

func Test_ShouldGetRandCity(t *testing.T) {
	require.True(t, RandCity() != "")
}

func Test_ShouldGetEnumString(t *testing.T) {
	require.Equal(t, "hello", EnumString("hello"))
}

func Test_ShouldGetEnumInt(t *testing.T) {
	require.Equal(t, int64(35), EnumInt("35"))
}

func Test_ShouldGetRandCountry(t *testing.T) {
	require.True(t, RandCountry() != "")
}

func Test_ShouldGetRandCountryCode(t *testing.T) {
	require.True(t, RandCountryCode() != "")
}

func Test_ShouldGetRandName(t *testing.T) {
	require.True(t, RandName() != "")
}

func Test_ShouldGetRandString(t *testing.T) {
	require.True(t, RandString(5) != "")
}

func Test_ShouldGetRandStringMinMax(t *testing.T) {
	require.True(t, RandStringMinMax(5, 10) != "")
}

func Test_ShouldGetRandStringArrayMinMax(t *testing.T) {
	require.True(t, len(RandStringArrayMinMax(5, 10)) >= 5)
}

func Test_ShouldGetRandIntArrayMinMax(t *testing.T) {
	require.True(t, len(RandIntArrayMinMax(5, 10)) >= 5)
}

func Test_ShouldGetRandRegex(t *testing.T) {
	require.Equal(t, "abc", RandRegex("abc"))
}

func Test_ShouldGetRandFileLine(t *testing.T) {
	require.True(t, RandFileLine("../../fixtures/lines.txt") != "")
}
