package fuzz

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ShouldGetRandomMinMax(t *testing.T) {
	require.True(t, RandNumMinMax(1, 0) >= 1)
	require.True(t, RandNumMinMax(1, 10) >= 1)
}

func Test_ShouldGetRandom(t *testing.T) {
	require.True(t, Random(10) <= 10)
}

func Test_ShouldGetSeededRandom(t *testing.T) {
	require.True(t, SeededRandom(1, 10) <= 10)
}

func Test_ShouldGetUUID(t *testing.T) {
	require.Equal(t, 36, len(UUID()))
}

func Test_ShouldGetSeededUUID(t *testing.T) {
	require.Equal(t, 36, len(SeededUUID(1)))
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
	require.Equal(t, "", RandString(0))
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

func Test_ShouldGetRandIntMinMax(t *testing.T) {
	require.True(t, RandNumMinMax(-10, -1) < 0)
}

func Test_ShouldGetAsciiAny(t *testing.T) {
	require.True(t, RandRegex(`[\x20-\x7F]{1,40}`) != "")
}

func Test_ShouldBuildRandRegexWords(t *testing.T) {
	str := RandRegex(`\w`)
	require.Equal(t, 0, countSpaces(str))
	str = RandRegex(`\w+`)
	require.True(t, countSpaces(str) >= 4)
	str = RandRegex(`\\w`)
	require.Equal(t, 0, countSpaces(str))
	str = RandRegex(`\\w+`)
	require.True(t, countSpaces(str) >= 4)

	str = RandRegex(`\w{3,4}`)
	require.True(t, countSpaces(str) >= 2, str)
	str = RandRegex(`\w{3}`)
	require.Equal(t, 2, countSpaces(str), str)
	str = RandRegex(`\w{2,2`)
	require.Equal(t, 1, countSpaces(str), str)
}

func Test_ShouldBuildRandRegexDigits(t *testing.T) {
	num, _ := strconv.Atoi(RandRegex(`\d`))
	require.True(t, num >= 0 && num <= 9)
	num, _ = strconv.Atoi(RandRegex(`\d+`))
	require.True(t, num >= 0)
	num, _ = strconv.Atoi(RandRegex(`\\d`))
	require.True(t, num >= 0 && num <= 9)
	num, _ = strconv.Atoi(RandRegex(`\\d+`))
	require.True(t, num >= 0)

	num, _ = strconv.Atoi(RandRegex(`\d{3,4}`))
	require.True(t, num >= 1 && num <= 9999, fmt.Sprintf("%d", num))
	num, _ = strconv.Atoi(RandRegex(`\d{3}`))
	require.True(t, num >= 1 && num <= 999, fmt.Sprintf("%d", num))
	num, _ = strconv.Atoi(RandRegex(`\d{3`))
	require.True(t, num >= 1 && num <= 999, fmt.Sprintf("%d", num))
}

func Test_ShouldGetRandRegex(t *testing.T) {
	require.Equal(t, "abc", RandRegex("abc"))
}

func Test_ShouldGetRandRegexWord(t *testing.T) {
	require.True(t, RandRegex(AnyWordRegex) != "")
	require.True(t, RandRegex(`\w`) != "")
	require.True(t, RandRegex(`\w+@\w+\.\w+`) != "")
}

func Test_ShouldGetRandPhone(t *testing.T) {
	require.Contains(t, RandPhone(), "1-")
}

func Test_ShouldGetRandEmail(t *testing.T) {
	require.Contains(t, RandEmail(), "@")
}

func Test_ShouldGetRandHost(t *testing.T) {
	require.Contains(t, RandHost(), ".")
}

func Test_ShouldGetRandURL(t *testing.T) {
	require.Contains(t, RandURL(), "://")
}

func Test_ShouldGetRandFileLine(t *testing.T) {
	require.True(t, RandFileLine("../../fixtures/lines.txt") != "")
}

func Test_ShouldGetRandPropertyLine(t *testing.T) {
	require.Equal(t, "sample token", FileProperty("../../fixtures/props.yaml", "token"))
}

func Test_ShouldGenerateWord(t *testing.T) {
	require.True(t, RandWord(1, 10) != "")
}

func Test_ShouldGenerateWordEmpty(t *testing.T) {
	require.Equal(t, 1, len(RandWord(0, 0)))
}

func Test_ShouldGenerateSentence(t *testing.T) {
	require.True(t, RandSentence(1, 10) != "")
}

func Test_ShouldGenerateParagraph(t *testing.T) {
	require.True(t, RandParagraph(1, 10) != "")
}

func countSpaces(str string) int {
	count := 0
	for _, ch := range str {
		if ch == ' ' {
			count++
		}

	}
	return count
}
