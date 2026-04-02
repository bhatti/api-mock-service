package fuzz

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_ShouldGetRandomMinMaxInt(t *testing.T) {
	require.True(t, RandIntMinMax(1, 0) >= 1)
	require.True(t, RandIntMinMax(1, 10) >= 1)
}

func Test_ShouldGetRandomInt(t *testing.T) {
	require.True(t, RandIntMax(10) <= 10)
}

func Test_ShouldGetSeededRandomInt(t *testing.T) {
	require.True(t, SeededRandIntMax(1, 0, 10) <= 10)
}

func Test_ShouldGetRandomFloatMinMax(t *testing.T) {
	require.True(t, RandFloatMinMax(1, 0) >= 1)
	require.True(t, RandFloatMinMax(1, 10) >= 1)
}

func Test_ShouldGetFloatRandom(t *testing.T) {
	require.True(t, RandFloatMax(10) <= 10)
}

func Test_ShouldGetSeededFloatRandom(t *testing.T) {
	require.True(t, SeededRandFloatMax(1, 0, 10) <= 10)
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

func Test_ShouldGetRandItin(t *testing.T) {
	require.Equal(t, 11, len(RandItin()))
}

func Test_ShouldGetRandSsn(t *testing.T) {
	require.Equal(t, 11, len(RandSsn()))
}

func Test_ShouldGetRandEin(t *testing.T) {
	require.Equal(t, 10, len(RandEin()))
}

func Test_ShouldGetRandFirstName(t *testing.T) {
	require.True(t, RandFirstName() != "")
}

func Test_ShouldGetRandFirstMaleName(t *testing.T) {
	require.True(t, RandFirstMaleName() != "")
}

func Test_ShouldGetRandFirstFemaleName(t *testing.T) {
	require.True(t, RandFirstFemaleName() != "")
}

func Test_ShouldGetRandLastName(t *testing.T) {
	require.True(t, RandLastName() != "")
}

func Test_ShouldGetRandUSState(t *testing.T) {
	require.True(t, RandUSState() != "")
}

func Test_ShouldGetRandUSStateAbbr(t *testing.T) {
	require.True(t, RandUSStateAbbr() != "")
}

func Test_ShouldGetRandAddress(t *testing.T) {
	addr := RandAddress()
	require.True(t, addr != "")
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
	require.True(t, RandIntMinMax(-10, -1) < 0)
}

func Test_ShouldGetAsciiAny(t *testing.T) {
	require.True(t, RandRegex(`[\x20-\x7F]{1,40}`) != "")
}

func Test_ShouldGetRawRegex(t *testing.T) {
	require.Equal(t, `[a-z]{5}`, RandRegex(PrefixTypeExample+`[a-z]{5}`))
}

func Test_ShouldParseRegexAlphaNumeric(t *testing.T) {
	out := RandRegex("[A-Za-z0-9-_=.]+")
	require.True(t, len(out) > 0)
}

func Test_ShouldParseRegexEscapeS(t *testing.T) {
	out := RandRegex("[\\S]+")
	require.True(t, len(out) > 0)
}

func Test_ShouldParseRegexEscapeP(t *testing.T) {
	out := RandRegex("[\\p{L}\\p{M}\\p{S}\\p{N}\\p{P}]+")
	require.True(t, len(out) > 0)
}

func Test_ShouldBuildRandRegexWords(t *testing.T) {
	str := RandRegex(`\w`)
	require.Equal(t, 0, countSpaces(str))
	str = RandRegex(`\w+`)
	require.True(t, countSpaces(str) >= 2)
	str = RandRegex(`\\w`)
	require.Equal(t, 0, countSpaces(str))
	str = RandRegex(`\\w+`)
	require.True(t, countSpaces(str) >= 2)

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

func Test_ShouldGetRandIPv6(t *testing.T) {
	ip := RandIPv6()
	require.True(t, ip != "")
	// IPv6 has 7 colons separating 8 groups
	require.Equal(t, 7, strings.Count(ip, ":"))
}

func Test_ShouldGetRandMACAddress(t *testing.T) {
	mac := RandMACAddress()
	require.True(t, mac != "")
	require.Equal(t, 5, strings.Count(mac, ":"))
}

func Test_ShouldGetRandHexColor(t *testing.T) {
	color := RandHexColor()
	require.True(t, strings.HasPrefix(color, "#"))
	require.Equal(t, 7, len(color))
}

func Test_ShouldGetRandRGBColor(t *testing.T) {
	color := RandRGBColor()
	require.True(t, strings.HasPrefix(color, "rgb("))
	require.True(t, strings.HasSuffix(color, ")"))
}

func Test_ShouldGetRandCurrencyCode(t *testing.T) {
	code := RandCurrencyCode()
	require.True(t, len(code) == 3)
}

func Test_ShouldGetSeededCurrencyCode(t *testing.T) {
	a := SeededCurrencyCode(42)
	b := SeededCurrencyCode(42)
	require.Equal(t, a, b)
}

func Test_ShouldGetRandSemver(t *testing.T) {
	v := RandSemver()
	require.Equal(t, 2, strings.Count(v, "."))
}

func Test_ShouldGetRandBase64(t *testing.T) {
	s := RandBase64()
	require.True(t, len(s) > 0)
	// base64 of 16 bytes is 24 chars
	require.Equal(t, 24, len(s))
}

func Test_ShouldGetRandSHA256(t *testing.T) {
	h := RandSHA256()
	require.Equal(t, 64, len(h))
}

func Test_ShouldGetRandMD5(t *testing.T) {
	h := RandMD5()
	require.Equal(t, 32, len(h))
}

func Test_ShouldGetRandLatitude(t *testing.T) {
	lat := RandLatitude()
	require.True(t, lat >= -90.0 && lat <= 90.0)
}

func Test_ShouldGetRandLongitude(t *testing.T) {
	lon := RandLongitude()
	require.True(t, lon >= -180.0 && lon <= 180.0)
}

func Test_ShouldGetRandTimezone(t *testing.T) {
	tz := RandTimezone()
	require.True(t, tz != "")
	require.True(t, strings.Contains(tz, "/") || tz == "UTC")
}

func Test_ShouldGetSeededTimezone(t *testing.T) {
	a := SeededTimezone(7)
	b := SeededTimezone(7)
	require.Equal(t, a, b)
}

func Test_ShouldGetRandMimeType(t *testing.T) {
	mt := RandMimeType()
	require.True(t, strings.Contains(mt, "/"))
}

func Test_ShouldGetRandPort(t *testing.T) {
	p := RandPort()
	require.True(t, p >= 1024 && p <= 65535)
}

func Test_ShouldGetRandUnixTimestamp(t *testing.T) {
	ts := RandUnixTimestamp()
	require.True(t, ts > 0)
}

func Test_ShouldGetRandFutureDate(t *testing.T) {
	d := RandFutureDate()
	require.Equal(t, 10, len(d)) // YYYY-MM-DD
	t1, err := time.Parse("2006-01-02", d)
	require.NoError(t, err)
	require.True(t, t1.After(time.Now()))
}

func Test_ShouldGetRandPastDate(t *testing.T) {
	d := RandPastDate()
	require.Equal(t, 10, len(d))
	t1, err := time.Parse("2006-01-02", d)
	require.NoError(t, err)
	require.True(t, t1.Before(time.Now()))
}

func Test_ShouldGetRandUsername(t *testing.T) {
	u := RandUsername()
	require.True(t, u != "")
	// should be lowercase only
	require.Equal(t, u, strings.ToLower(u))
}

func Test_ShouldGetRandPassword(t *testing.T) {
	p := RandPassword()
	require.True(t, len(p) >= 12 && len(p) <= 16)
}

func Test_ShouldGetRandSlug(t *testing.T) {
	s := RandSlug()
	require.True(t, strings.Contains(s, "-"))
	require.Equal(t, s, strings.ToLower(s))
}

func Test_ShouldGetRandHTTPStatus(t *testing.T) {
	code := RandHTTPStatus()
	valid := map[int]bool{200: true, 201: true, 204: true, 301: true, 302: true,
		400: true, 401: true, 403: true, 404: true, 409: true, 422: true, 429: true, 500: true, 503: true}
	require.True(t, valid[code])
}

func Test_ShouldGetRandFileExtension(t *testing.T) {
	ext := RandFileExtension()
	require.True(t, ext != "")
	require.False(t, strings.HasPrefix(ext, "."))
}

func Test_ShouldGetRandFilename(t *testing.T) {
	name := RandFilename()
	require.True(t, strings.Contains(name, "."))
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
