package fuzz

import (
	"bufio"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/fuzz/lorem"
	"github.com/lucasjones/reggen"
	log "github.com/sirupsen/logrus"
	"github.com/twinj/uuid"
	regen "github.com/zach-klippenstein/goregen"
	"gopkg.in/yaml.v3"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// RandNumMinMax returns random number between min and max
func RandNumMinMax(min, max int) int {
	if max == 0 {
		max = 100000
	}
	if min == max {
		return min
	}
	return rand.Intn(max-min) + min
}

// Random returns random number between 0 and max
func Random(max int) int {
	return RandNumMinMax(0, max)
}

// SeededRandom returns random number with seed upto a max
func SeededRandom(seed int64, max int) int {
	if seed <= 0 {
		seed = time.Now().UnixNano()
	}
	r := rand.New(rand.NewSource(seed))
	return r.Intn(max)
}

// UUID generator
func UUID() string {
	return uuid.NewV4().String()
}

// SeededUUID generator
func SeededUUID(seed int64) string {
	ids := []string{
		"bec879f5-5072-46a8-b47c-d86ace",
		"688b999f-fd64-4333-b704-1148a4",
		"93bc2b5b-fd37-4217-b0f2-6cdf0f",
		"ecb7c1f4-99a0-4520-be4c-43246e",
		"bf430cfe-e0bd-464c-80c5-8480b7",
		"1e840362-b5bd-4069-a8e9-d0e48a",
		"1ba38373-c76a-492e-9444-530431",
		"4b43c32f-91e8-4404-bd63-2251ed",
		"4c4c123c-484e-449f-bc54-f66ed8",
		"daed2f9c-f485-44b1-bb08-38bc8c",
		"ffa73c6c-acc2-43b9-8a65-a18f59",
		"bc81a1bf-1b64-40a0-b63b-080bb5",
		"92e99380-005a-47c8-9c9f-4db827",
		"a93f4dda-b6c4-4aa0-80dd-cce716",
		"634e521f-cde3-44dc-b18f-acb8bf",
		"26fd7bbe-fe97-4827-95b0-d214d2",
		"fb322c3a-79b9-4d1c-8d07-cb07bb",
		"36c6ffc6-c10a-4870-980a-a4f5e2",
		"bcbe5c9e-2760-4f62-8af7-858082",
		"939215b3-615a-4d1a-a004-51671a",
		"21ebff0f-bda0-43bf-aeb6-5f7518",
		"7e5e8ec9-05c8-4322-8831-768909",
		"8b9195eb-0604-4994-9e14-f69fa7",
		"824a6c68-8b11-4058-b74c-908266",
		"be7cf107-f9fc-43d9-a538-2b2d1a",
		"8d94d137-fb53-47b4-b71e-bb8cf5",
		"f85fc002-f356-4938-b92b-26bbe8",
		"cad6da7e-49f5-4893-b9c8-3a7f9e",
		"33a694b5-b726-4d44-b9b3-bada37",
		"75f50c87-4182-4253-bb76-711fdc",
		"9c35b0c9-21e5-4011-8b6f-28b56c",
		"f649ca59-b0ef-40ef-bfa8-997ffd",
		"85afcb6a-decf-4ce9-b39d-62be64",
		"2b3e91e8-5f49-4a0f-9118-33caf8",
		"e217a204-690e-4c08-ad53-cc0f9a",
		"8808ae5e-5130-4558-9231-efba98",
		"28568257-ed47-4098-b487-031618",
		"4d580297-847a-4aef-8329-bdff91",
		"bc1f7e03-d71f-48ed-9306-7f6dd0",
		"50f9fe32-340b-4a2d-a2df-84902e",
		"994b34e2-a84a-4c6e-b074-fb8a24",
		"1b3a67f1-3201-472b-920d-d15b2a",
		"bc2acb92-64ff-4daf-b33e-1710ee",
		"07611a4e-43c3-47c6-917a-b13634",
		"ad6d1e62-29c1-4471-8700-4e8267",
		"12e95054-7d2c-4a7b-bd2e-acb793",
		"fe49b338-4593-43c9-b1e9-67581d",
		"1a1c36aa-9854-4a6f-a0b4-4c60bc",
		"82d73605-6531-43b3-a793-d48f8a",
		"2204f077-d19b-489d-af52-0fe357",
		"524a6056-c057-4786-9f67-691c84",
		"0be337bb-561d-4393-ac15-53ee83",
		"e6ef6a7a-6b4d-4e7f-a82b-5e6e70",
		"bff24c15-012f-433f-9b46-ff1940",
		"4ffbc758-44be-45c4-8a96-e586d4",
		"edaa05d6-71ed-4e62-9410-f44d51",
		"fcb32984-87b9-490c-8a60-1c28ff",
		"71029b5f-a5b5-45c4-b6a8-23b7ea",
		"72008a0e-9e16-4d5b-8ff4-41bf76",
		"0bbacdc0-8cca-4cc1-ada6-d2735a",
		"132b240e-af77-4015-b503-f676f3",
		"2f8dcf9c-6e39-4e86-ac8a-292906",
		"25923d2e-1f8a-4cb5-8e2f-1fa177",
		"e9379c39-03e1-49f4-9c9a-7958d8",
		"5c761a19-8063-49e7-8f26-7610ab",
		"d145dca8-62bf-408e-88f2-45dcd9",
		"3c0e54b6-99a6-49dc-9e51-e432f8",
		"1fadcf35-846c-44e6-a2e5-3bc14b",
		"29cef7ea-0b24-42fd-9751-b0844c",
		"b6e79fe8-47e6-45dd-88cd-93fcfd",
		"dff0f14e-4e96-4a81-b494-099fe4",
		"4cb905fa-3f43-401d-9ab7-fa5a6b",
		"c5de49f5-49fe-49f7-a883-cc0e64",
		"1f2ab10f-559a-43d1-b534-096e3f",
		"49084552-27fd-4827-bfed-3d848d",
		"9b9b20e5-7d45-46c1-b4e9-37188c",
		"5332dd12-ce18-4d0e-b39f-90fee8",
		"958c7c72-7cf5-40cb-8166-5c49c9",
		"a843b9c9-037f-4666-9b88-ee1f79",
		"cddee08c-9838-4d41-897e-48d153",
		"f7c1a2b1-ab37-4ded-a423-567c7f",
		"c0871e48-c1c7-43cc-a5d7-7af769",
		"c9a50c65-e5a4-4677-b767-df755e",
		"c0a13bfc-b790-4c68-bdac-aec050",
		"16575168-ab96-4ed7-9ae5-8e664c",
		"98714bcd-04e4-48d5-8d6f-5a0ccd",
		"0af6d89d-41a3-43bb-ba7c-95b49f",
		"70a06d51-2d6b-4096-8b12-0ac2d7",
		"bf140ab6-acec-4a7f-878f-df7e56",
		"f6973ff7-d42f-4401-9c0b-d0d8b2",
		"08331734-0d69-42d2-a352-879132",
		"1f72a866-4dd2-4f48-be8a-e13170",
		"aa90bc3d-1523-4b3a-9d25-a42f16",
		"3a357489-e2e6-49ea-9e61-d81119",
		"5d383675-e74a-471a-919e-f6a78f",
		"ee1f789c-6264-42d4-998a-2b6d62",
		"97cd51ae-420a-40cb-82dd-010291",
		"775c74e6-a505-41c3-a8c8-e8a8ae",
		"494e47cc-6416-4125-b2f6-a26b70",
		"221105a3-b322-4f0e-9db9-0837ee",
	}
	suffix := fmt.Sprintf("%06d", seed)
	return randomArrayElement(ids, seed) + suffix
}

// RandBool generator
func RandBool() bool {
	return SeededBool(0)
}

// SeededBool generator
func SeededBool(seed int64) bool {
	bools := []bool{true, false}
	return bools[SeededRandom(seed, 2)]
}

// RandCity generator
func RandCity() string {
	return SeededCity(0)
}

// SeededCity generator
func SeededCity(seed int64) string {
	cities := []string{
		"Paris", "London", "Chicago", "Karachi", "Tokyo", "Lagos", "Delhi", "Shanghai",
		"Mexico City", "Cairo", "Beijing", "Dhaka", "Osaka", "Buenos Aires",
		"Chongqing", "Istanbul", "Kolkata", "Manila", "Rio de Janeiro",
		"Tianjin", "Kinshasa", "Guangzhou", "Los Angeles", "Moscow",
		"Shenzhen", "Lahore", "Bangalore", "Paris", "Bogotá", "Jakarta",
		"Chennai", "Lima", "Bangkok", "Seoul", "Nagoya", "Hyderabad",
		"Chengdu", "Nanjing", "Wuhan", "Ho Chi Minh City", "Luanda",
		"Ahmedabad", "Kuala Lumpur", "Xi'an", "Hong Kong", "Dongguan",
		"Hangzhou", "Foshan", "Shenyang", "Riyadh", "Baghdad", "Santiago",
		"Surat", "Madrid", "Suzhou", "Washington, D.C.", "New York City",
		"Pune", "Harbin", "Houston", "Dallas", "Toronto", "Dar es Salaam",
		"Miami", "Belo Horizonte", "Singapore", "Philadelphia", "Atlanta",
		"Fukuoka", "Khartoum", "Barcelona", "Johannesburg", "Tehran",
		"Saint Petersburg", "Qingdao", "Dalian", "Yangon", "Alexandria", "Jinan", "Guadalajara",
	}
	return randomArrayElement(cities, seed)
}

// EnumString selects substring
func EnumString(anyArr ...any) string {
	var str strings.Builder
	for _, val := range anyArr {
		if str.Len() > 0 {
			str.WriteRune(' ')
		}
		str.WriteString(fmt.Sprintf("%v", val))
	}
	parts := strings.Split(str.String(), " ")
	return randomArrayElement(parts, 0)
}

// EnumInt selects numeric
func EnumInt(anyArr ...any) (n int64) {
	str := EnumString(anyArr...)
	n, _ = strconv.ParseInt(str, 10, 64)
	return
}

// RandCountry country generator
func RandCountry() string {
	return SeededCountry(0)
}

// SeededCountry country generator
func SeededCountry(seed int64) string {
	countries := []string{
		"Afghanistan",
		"Aland Islands",
		"Albania",
		"Algeria",
		"American Samoa",
		"Andorra",
		"Angola",
		"Anguilla",
		"Antarctica",
		"Antigua And Barbuda",
		"Argentina",
		"Armenia",
		"Aruba",
		"Australia",
		"Austria",
		"Azerbaijan",
		"Bahamas",
		"Bahrain",
		"Bangladesh",
		"Barbados",
		"Belarus",
		"Belgium",
		"Belize",
		"Benin",
		"Bermuda",
		"Bhutan",
		"Bolivia",
		"Bosnia And Herzegovina",
		"Botswana",
		"Bouvet Island",
		"Brazil",
		"British Indian Ocean Territory",
		"Brunei Darussalam",
		"Bulgaria",
		"Burkina Faso",
		"Burundi",
		"Cambodia",
		"Cameroon",
		"Canada",
		"Cape Verde",
		"Cayman Islands",
		"Central African Republic",
		"Chad",
		"Chile",
		"China",
		"Christmas Island",
		"Cocos (Keeling) Islands",
		"Colombia",
		"Comoros",
		"Congo",
		"Congo, Democratic Republic",
		"Cook Islands",
		"Costa Rica",
		"Cote D\"Ivoire",
		"Croatia",
		"Cuba",
		"Cyprus",
		"Czech Republic",
		"Denmark",
		"Djibouti",
		"Dominica",
		"Dominican Republic",
		"Ecuador",
		"Egypt",
		"El Salvador",
		"Equatorial Guinea",
		"Eritrea",
		"Estonia",
		"Ethiopia",
		"Falkland Islands (Malvinas)",
		"Faroe Islands",
		"Fiji",
		"Finland",
		"France",
		"French Guiana",
		"French Polynesia",
		"French Southern Territories",
		"Gabon",
		"Gambia",
		"Georgia",
		"Germany",
		"Ghana",
		"Gibraltar",
		"Greece",
		"Greenland",
		"Grenada",
		"Guadeloupe",
		"Guam",
		"Guatemala",
		"Guernsey",
		"Guinea",
		"Guinea-Bissau",
		"Guyana",
		"Haiti",
		"Heard Island & Mcdonald Islands",
		"Holy See (Vatican City State)",
		"Honduras",
		"Hong Kong",
		"Hungary",
		"Iceland",
		"India",
		"Indonesia",
		"Iran, Islamic Republic Of",
		"Iraq",
		"Ireland",
		"Isle Of Man",
		"Israel",
		"Italy",
		"Jamaica",
		"Japan",
		"Jersey",
		"Jordan",
		"Kazakhstan",
		"Kenya",
		"Kiribati",
		"Korea",
		"North Korea",
		"Kuwait",
		"Kyrgyzstan",
		"Lao People\"s Democratic Republic",
		"Latvia",
		"Lebanon",
		"Lesotho",
		"Liberia",
		"Libyan Arab Jamahiriya",
		"Liechtenstein",
		"Lithuania",
		"Luxembourg",
		"Macao",
		"Macedonia",
		"Madagascar",
		"Malawi",
		"Malaysia",
		"Maldives",
		"Mali",
		"Malta",
		"Marshall Islands",
		"Martinique",
		"Mauritania",
		"Mauritius",
		"Mayotte",
		"Mexico",
		"Micronesia, Federated States Of",
		"Moldova",
		"Monaco",
		"Mongolia",
		"Montenegro",
		"Montserrat",
		"Morocco",
		"Mozambique",
		"Myanmar",
		"Namibia",
		"Nauru",
		"Nepal",
		"Netherlands",
		"Netherlands Antilles",
		"New Caledonia",
		"New Zealand",
		"Nicaragua",
		"Niger",
		"Nigeria",
		"Niue",
		"Norfolk Island",
		"Northern Mariana Islands",
		"Norway",
		"Oman",
		"Pakistan",
		"Palau",
		"Palestinian Territory, Occupied",
		"Panama",
		"Papua New Guinea",
		"Paraguay",
		"Peru",
		"Philippines",
		"Pitcairn",
		"Poland",
		"Portugal",
		"Puerto Rico",
		"Qatar",
		"Reunion",
		"Romania",
		"Russian Federation",
		"Rwanda",
		"Saint Barthelemy",
		"Saint Helena",
		"Saint Kitts And Nevis",
		"Saint Lucia",
		"Saint Martin",
		"Saint Pierre And Miquelon",
		"Saint Vincent And Grenadines",
		"Samoa",
		"San Marino",
		"Sao Tome And Principe",
		"Saudi Arabia",
		"Senegal",
		"Serbia",
		"Seychelles",
		"Sierra Leone",
		"Singapore",
		"Slovakia",
		"Slovenia",
		"Solomon Islands",
		"Somalia",
		"South Africa",
		"South Georgia And Sandwich Isl.",
		"Spain",
		"Sri Lanka",
		"Sudan",
		"Suriname",
		"Svalbard And Jan Mayen",
		"Swaziland",
		"Sweden",
		"Switzerland",
		"Syrian Arab Republic",
		"Taiwan",
		"Tajikistan",
		"Tanzania",
		"Thailand",
		"Timor-Leste",
		"Togo",
		"Tokelau",
		"Tonga",
		"Trinidad And Tobago",
		"Tunisia",
		"Turkey",
		"Turkmenistan",
		"Turks And Caicos Islands",
		"Tuvalu",
		"Uganda",
		"Ukraine",
		"United Arab Emirates",
		"United Kingdom",
		"United States",
		"United States Outlying Islands",
		"Uruguay",
		"Uzbekistan",
		"Vanuatu",
		"Venezuela",
		"Vietnam",
		"Virgin Islands, British",
		"Virgin Islands, U.S.",
		"Wallis And Futuna",
		"Western Sahara",
		"Yemen",
		"Zambia",
		"Zimbabwe",
	}
	return randomArrayElement(countries, seed)
}

// RandCountryCode country generator
func RandCountryCode() string {
	return SeededCountryCode(0)
}

// SeededCountryCode country generator
func SeededCountryCode(seed int64) string {
	countryCodes := []string{
		"AF",
		"AX",
		"AL",
		"DZ",
		"AS",
		"AD",
		"AO",
		"AI",
		"AQ",
		"AG",
		"AR",
		"AM",
		"AW",
		"AU",
		"AT",
		"AZ",
		"BS",
		"BH",
		"BD",
		"BB",
		"BY",
		"BE",
		"BZ",
		"BJ",
		"BM",
		"BT",
		"BO",
		"BA",
		"BW",
		"BV",
		"BR",
		"IO",
		"BN",
		"BG",
		"BF",
		"BI",
		"KH",
		"CM",
		"CA",
		"CV",
		"KY",
		"CF",
		"TD",
		"CL",
		"CN",
		"CX",
		"CC",
		"CO",
		"KM",
		"CG",
		"CD",
		"CK",
		"CR",
		"CI",
		"HR",
		"CU",
		"CY",
		"CZ",
		"DK",
		"DJ",
		"DM",
		"DO",
		"EC",
		"EG",
		"SV",
		"GQ",
		"ER",
		"EE",
		"ET",
		"FK",
		"FO",
		"FJ",
		"FI",
		"FR",
		"GF",
		"PF",
		"TF",
		"GA",
		"GM",
		"GE",
		"DE",
		"GH",
		"GI",
		"GR",
		"GL",
		"GD",
		"GP",
		"GU",
		"GT",
		"GG",
		"GN",
		"GW",
		"GY",
		"HT",
		"HM",
		"VA",
		"HN",
		"HK",
		"HU",
		"IS",
		"IN",
		"ID",
		"IR",
		"IQ",
		"IE",
		"IM",
		"IL",
		"IT",
		"JM",
		"JP",
		"JE",
		"JO",
		"KZ",
		"KE",
		"KI",
		"KR",
		"KP",
		"KW",
		"KG",
		"LA",
		"LV",
		"LB",
		"LS",
		"LR",
		"LY",
		"LI",
		"LT",
		"LU",
		"MO",
		"MK",
		"MG",
		"MW",
		"MY",
		"MV",
		"ML",
		"MT",
		"MH",
		"MQ",
		"MR",
		"MU",
		"YT",
		"MX",
		"FM",
		"MD",
		"MC",
		"MN",
		"ME",
		"MS",
		"MA",
		"MZ",
		"MM",
		"NA",
		"NR",
		"NP",
		"NL",
		"AN",
		"NC",
		"NZ",
		"NI",
		"NE",
		"NG",
		"NU",
		"NF",
		"MP",
		"NO",
		"OM",
		"PK",
		"PW",
		"PS",
		"PA",
		"PG",
		"PY",
		"PE",
		"PH",
		"PN",
		"PL",
		"PT",
		"PR",
		"QA",
		"RE",
		"RO",
		"RU",
		"RW",
		"BL",
		"SH",
		"KN",
		"LC",
		"MF",
		"PM",
		"VC",
		"WS",
		"SM",
		"ST",
		"SA",
		"SN",
		"RS",
		"SC",
		"SL",
		"SG",
		"SK",
		"SI",
		"SB",
		"SO",
		"ZA",
		"GS",
		"ES",
		"LK",
		"SD",
		"SR",
		"SJ",
		"SZ",
		"SE",
		"CH",
		"SY",
		"TW",
		"TJ",
		"TZ",
		"TH",
		"TL",
		"TG",
		"TK",
		"TO",
		"TT",
		"TN",
		"TR",
		"TM",
		"TC",
		"TV",
		"UG",
		"UA",
		"AE",
		"GB",
		"US",
		"UM",
		"UY",
		"UZ",
		"VU",
		"VE",
		"VN",
		"VG",
		"VI",
		"WF",
		"EH",
		"YE",
		"ZM",
		"ZW",
	}
	return randomArrayElement(countryCodes, seed)
}

// RandTriString phrase generator
func RandTriString(sep string) string {
	return SeededTriString(0, sep)
}

// SeededTriString phrase generator
func SeededTriString(seed int64, sep string) string {
	words := []string{
		"abandon", "ability", "able", "about", "above", "absent", "absorb", "abstract", "absurd", "abuse", "access", "accident", "account", "accuse", "achieve", "acid", "acoustic", "acquire", "across", "act", "action", "actor", "actress", "actual", "adapt", "add", "addict", "address", "adjust", "admit", "adult", "advance", "advice", "aerobic", "affair", "afford", "afraid", "again", "age", "agent", "agree", "ahead", "aim", "air", "airport", "aisle", "alarm", "album", "alcohol", "alert", "alien", "all", "alley", "allow", "almost", "alone", "alpha", "already", "also", "alter", "always", "amateur", "amazing", "among", "amount", "amused", "analyst", "anchor", "ancient", "anger", "angle", "angry", "animal", "ankle", "announce", "annual", "another", "answer", "antenna", "antique", "anxiety", "any", "apart", "apology", "appear", "apple", "approve", "april", "arch", "arctic", "area", "arena", "argue", "arm", "armed", "armor", "army", "around", "arrange", "arrest", "arrive", "arrow", "art", "artefact", "artist", "artwork", "ask", "aspect", "assault", "asset", "assist", "assume", "asthma", "athlete", "atom", "attack", "attend", "attitude", "attract", "auction", "audit", "august", "aunt", "author", "auto", "autumn", "average", "avocado", "avoid", "awake", "aware", "away", "awesome", "awful", "awkward", "axis",
		"baby", "bachelor", "bacon", "badge", "bag", "balance", "balcony", "ball", "bamboo", "banana", "banner", "bar", "barely", "bargain", "barrel", "base", "basic", "basket", "battle", "beach", "bean", "beauty", "because", "become", "beef", "before", "begin", "behave", "behind", "believe", "below", "belt", "bench", "benefit", "best", "betray", "better", "between", "beyond", "bicycle", "bid", "bike", "bind", "biology", "bird", "birth", "bitter", "black", "blade", "blame", "blanket", "blast", "bleak", "bless", "blind", "blood", "blossom", "blouse", "blue", "blur", "blush", "board", "boat", "body", "boil", "bomb", "bone", "bonus", "book", "boost", "border", "boring", "borrow", "boss", "bottom", "bounce", "box", "boy", "bracket", "brain", "brand", "brass", "brave", "bread", "breeze", "brick", "bridge", "brief", "bright", "bring", "brisk", "broccoli", "broken", "bronze", "broom", "brother", "brown", "brush", "bubble", "buddy", "budget", "buffalo", "build", "bulb", "bulk", "bullet", "bundle", "bunker", "burden", "burger", "burst", "bus", "business", "busy", "butter", "buyer", "buzz",
		"cabbage", "cabin", "cable", "cactus", "cage", "cake", "call", "calm", "camera", "camp", "can", "canal", "cancel", "candy", "cannon", "canoe", "canvas", "canyon", "capable", "capital", "captain", "car", "carbon", "card", "cargo", "carpet", "carry", "cart", "case", "cash", "casino", "castle", "casual", "cat", "catalog", "catch", "category", "cattle", "caught", "cause", "caution", "cave", "ceiling", "celery", "cement", "census", "century", "cereal", "certain", "chair", "chalk", "champion", "change", "contract", "chapter", "charge", "chase", "chat", "cheap", "check", "cheese", "chef", "cherry", "chest", "chicken", "chief", "child", "chimney", "choice", "choose", "chronic", "chuckle", "chunk", "churn", "cigar", "cinnamon", "circle", "citizen", "city", "civil", "claim", "clap", "clarify", "claw", "clay", "clean", "clerk", "clever", "click", "client", "cliff", "climb", "clinic", "clip", "clock", "clog", "close", "cloth", "cloud", "clown", "club", "clump", "cluster", "clutch", "coach", "coast", "coconut", "code", "coffee", "coil", "coin", "collect", "color", "column", "combine", "come", "comfort", "comic", "common", "company", "concert", "conduct", "confirm", "congress", "connect", "consider", "control", "convince", "cook", "cool", "copper", "copy", "coral", "core", "corn", "correct", "cost", "cotton", "couch", "country", "couple", "course", "cousin", "cover", "coyote", "crack", "cradle", "craft", "cram", "crane", "crash", "crater", "crawl", "crazy", "cream", "credit", "creek", "crew", "cricket", "crime", "crisp", "critic", "crop", "cross", "crouch", "crowd", "crucial", "cruel", "cruise", "crumble", "crunch", "crush", "cry", "crystal", "cube", "culture", "cup", "cupboard", "curious", "current", "curtain", "curve", "cushion", "custom", "cute", "cycle",
		"dad", "damage", "damp", "dance", "danger", "daring", "dash", "daughter", "dawn", "day", "deal", "debate", "debris", "decade", "december", "decide", "decline", "decorate", "decrease", "deer", "defense", "define", "defy", "degree", "delay", "deliver", "demand", "demise", "denial", "dentist", "deny", "depart", "depend", "deposit", "depth", "deputy", "derive", "describe", "desert", "design", "desk", "despair", "destroy", "detail", "detect", "develop", "device", "devote", "diagram", "dial", "diamond", "diary", "dice", "diesel", "diet", "differ", "digital", "dignity", "dilemma", "dinner", "dinosaur", "direct", "dirt", "disagree", "discover", "disease", "dish", "dismiss", "disorder", "display", "distance", "divert", "divide", "divorce", "dizzy", "doctor", "document", "dog", "doll", "dolphin", "domain", "donate", "donkey", "donor", "door", "dose", "double", "dove", "draft", "dragon", "drama", "drastic", "draw", "dream", "dress", "drift", "drill", "drink", "drip", "drive", "drop", "drum", "dry", "duck", "dumb", "dune", "during", "dust", "dutch", "duty", "dwarf", "dynamic",
		"eager", "eagle", "early", "earn", "earth", "easily", "east", "easy", "echo", "ecology", "economy", "edge", "edit", "educate", "effort", "egg", "eight", "either", "elbow", "elder", "electric", "elegant", "element", "elephant", "elevator", "elite", "else", "embark", "embody", "embrace", "emerge", "emotion", "employ", "empower", "empty", "enable", "enact", "end", "endless", "endorse", "enemy", "energy", "enforce", "engage", "engine", "enhance", "enjoy", "enlist", "enough", "enrich", "enroll", "ensure", "enter", "entire", "entry", "envelope", "episode", "equal", "equip", "era", "erase", "erode", "erosion", "error", "erupt", "escape", "essay", "essence", "estate", "eternal", "ethics", "evidence", "evil", "evoke", "evolve", "exact", "example", "excess", "exchange", "excite", "exclude", "excuse", "execute", "exercise", "exhaust", "exhibit", "exile", "exist", "exit", "exotic", "expand", "expect", "expire", "explain", "expose", "express", "extend", "extra", "eye", "eyebrow",
		"fabric", "face", "faculty", "fade", "faint", "faith", "fall", "false", "fame", "family", "famous", "fan", "fancy", "fantasy", "farm", "fashion", "fat", "fatal", "father", "fatigue", "fault", "favorite", "feature", "february", "federal", "fee", "feed", "feel", "female", "fence", "festival", "fetch", "fever", "few", "fiber", "fiction", "field", "figure", "file", "film", "filter", "final", "find", "fine", "finger", "finish", "fire", "firm", "first", "fiscal", "fish", "fit", "fitness", "fix", "flag", "flame", "flash", "flat", "flavor", "flee", "flight", "flip", "float", "flock", "floor", "flower", "fluid", "flush", "fly", "foam", "focus", "fog", "foil", "fold", "follow", "food", "foot", "force", "forest", "forget", "fork", "fortune", "forum", "forward", "fossil", "foster", "found", "fox", "fragile", "frame", "frequent", "fresh", "friend", "fringe", "frog", "front", "frost", "frown", "frozen", "fruit", "fuel", "fun", "funny", "furnace", "fury", "future",
		"gadget", "gain", "galaxy", "gallery", "game", "gap", "garage", "garbage", "garden", "garlic", "garment", "gas", "gasp", "gate", "gather", "gauge", "gaze", "general", "genius", "genre", "gentle", "genuine", "gesture", "ghost", "giant", "gift", "giggle", "ginger", "giraffe", "girl", "give", "glad", "glance", "glare", "glass", "glide", "glimpse", "globe", "gloom", "glory", "glove", "glow", "glue", "goat", "goddess", "gold", "good", "goose", "gorilla", "gospel", "gossip", "govern", "gown", "grab", "grace", "grain", "grant", "grape", "grass", "gravity", "great", "green", "grid", "grief", "grit", "grocery", "group", "grow", "grunt", "guard", "guess", "guide", "guilt", "guitar", "gun", "gym",
		"habit", "hair", "half", "hammer", "hamster", "hand", "happy", "harbor", "hard", "harsh", "harvest", "hat", "have", "hawk", "hazard", "head", "health", "heart", "heavy", "hedgehog", "height", "hello", "helmet", "help", "hen", "hero", "hidden", "high", "hill", "hint", "hip", "hire", "history", "hobby", "hockey", "hold", "hole", "holiday", "hollow", "home", "honey", "hood", "hope", "horn", "horror", "horse", "hospital", "host", "hotel", "hour", "hover", "hub", "huge", "human", "humble", "humor", "hundred", "hungry", "hunt", "hurdle", "hurry", "hurt", "husband", "hybrid",
		"ice", "icon", "idea", "identify", "idle", "ignore", "ill", "illegal", "illness", "image", "imitate", "immense", "immune", "impact", "impose", "improve", "impulse", "inch", "include", "income", "increase", "index", "indicate", "indoor", "industry", "infant", "inflict", "inform", "inhale", "inherit", "initial", "inject", "injury", "inmate", "inner", "innocent", "input", "inquiry", "insane", "insect", "inside", "inspire", "install", "intact", "interest", "into", "invest", "invite", "involve", "iron", "island", "isolate", "issue", "item", "ivory",
		"jacket", "jaguar", "jar", "jazz", "jealous", "jeans", "jelly", "jewel", "job", "join", "joke", "journey", "joy", "judge", "juice", "jump", "jungle", "junior", "junk", "just",
		"kangaroo", "keen", "keep", "ketchup", "key", "kick", "kid", "kidney", "kind", "kingdom", "kiss", "kit", "kitchen", "kite", "kitten", "kiwi", "knee", "knife", "knock", "know",
		"lab", "label", "labor", "ladder", "lady", "lake", "lamp", "language", "laptop", "large", "later", "latin", "laugh", "laundry", "lava", "law", "lawn", "lawsuit", "layer", "lazy", "leader", "leaf", "learn", "leave", "lecture", "left", "leg", "legal", "legend", "leisure", "lemon", "lend", "length", "lens", "leopard", "lesson", "letter", "level", "liar", "liberty", "library", "license", "life", "lift", "light", "like", "limb", "limit", "link", "lion", "liquid", "list", "little", "live", "lizard", "load", "loan", "lobster", "local", "lock", "logic", "lonely", "long", "loop", "lottery", "loud", "lounge", "love", "loyal", "lucky", "luggage", "lumber", "lunar", "lunch", "luxury", "lyrics",
		"machine", "mad", "magic", "magnet", "maid", "mail", "main", "major", "make", "mammal", "man", "manage", "mandate", "mango", "mansion", "manual", "maple", "marble", "march", "margin", "marine", "market", "marriage", "mask", "mass", "master", "match", "material", "math", "matrix", "matter", "maximum", "maze", "meadow", "mean", "measure", "meat", "mechanic", "medal", "media", "melody", "melt", "member", "memory", "mention", "menu", "mercy", "merge", "merit", "merry", "mesh", "message", "metal", "method", "middle", "midnight", "milk", "million", "mimic", "mind", "minimum", "minor", "minute", "miracle", "mirror", "misery", "miss", "mistake", "mix", "mixed", "mixture", "mobile", "model", "modify", "mom", "moment", "monitor", "monkey", "monster", "month", "moon", "moral", "more", "morning", "mosquito", "mother", "motion", "motor", "mountain", "mouse", "move", "movie", "much", "muffin", "mule", "multiply", "muscle", "museum", "mushroom", "music", "must", "mutual", "myself", "mystery", "myth",
		"naive", "name", "napkin", "narrow", "nasty", "nation", "nature", "near", "neck", "need", "negative", "neglect", "neither", "nephew", "nerve", "nest", "net", "network", "neutral", "never", "news", "next", "nice", "night", "noble", "noise", "nominee", "noodle", "normal", "north", "nose", "notable", "note", "nothing", "notice", "novel", "now", "nuclear", "number", "nurse", "nut",
		"oak", "obey", "object", "oblige", "obscure", "observe", "obtain", "obvious", "occur", "ocean", "october", "odor", "off", "offer", "office", "often", "oil", "okay", "old", "olive", "olympic", "omit", "once", "one", "onion", "online", "only", "open", "opera", "opinion", "oppose", "option", "orange", "orbit", "orchard", "order", "ordinary", "organ", "orient", "original", "orphan", "ostrich", "other", "outdoor", "outer", "output", "outside", "oval", "oven", "over", "own", "owner", "oxygen", "oyster", "ozone",
		"pact", "paddle", "page", "pair", "palace", "palm", "panda", "panel", "panic", "panther", "paper", "parade", "parent", "park", "parrot", "party", "pass", "patch", "path", "patient", "patrol", "pattern", "pause", "pave", "payment", "peace", "peanut", "pear", "peasant", "pelican", "pen", "penalty", "pencil", "people", "pepper", "perfect", "permit", "person", "pet", "phone", "photo", "phrase", "physical", "piano", "picnic", "picture", "piece", "pig", "pigeon", "pill", "pilot", "pink", "pioneer", "pipe", "pistol", "pitch", "pizza", "place", "planet", "plastic", "plate", "play", "please", "pledge", "pluck", "plug", "plunge", "poem", "poet", "point", "polar", "pole", "police", "pond", "pony", "pool", "popular", "portion", "position", "possible", "post", "potato", "pottery", "poverty", "powder", "power", "practice", "praise", "predict", "prefer", "prepare", "present", "pretty", "prevent", "price", "pride", "primary", "print", "priority", "prison", "private", "prize", "problem", "process", "produce", "profit", "program", "project", "promote", "proof", "property", "prosper", "protect", "proud", "provide", "public", "pudding", "pull", "pulp", "pulse", "pumpkin", "punch", "pupil", "puppy", "purchase", "purity", "purpose", "purse", "push", "put", "puzzle", "pyramid",
		"quality", "quantum", "quarter", "question", "quick", "quit", "quiz", "quote",
		"rabbit", "raccoon", "race", "rack", "radar", "radio", "rail", "rain", "raise", "rally", "ramp", "ranch", "random", "range", "rapid", "rare", "rate", "rather", "raven", "raw", "razor", "ready", "real", "reason", "rebel", "rebuild", "recall", "receive", "recipe", "record", "recycle", "reduce", "reflect", "reform", "refuse", "region", "regret", "regular", "reject", "relax", "release", "relief", "rely", "remain", "remember", "remind", "remove", "render", "renew", "rent", "reopen", "repair", "repeat", "replace", "report", "require", "rescue", "resemble", "resist", "resource", "response", "result", "retire", "retreat", "return", "reunion", "reveal", "review", "reward", "rhythm", "rib", "ribbon", "rice", "rich", "ride", "ridge", "rifle", "right", "rigid", "ring", "riot", "ripple", "risk", "ritual", "rival", "river", "road", "roast", "robot", "robust", "rocket", "romance", "roof", "rookie", "room", "rose", "rotate", "rough", "round", "route", "royal", "rubber", "rude", "rug", "rule", "run", "runway", "rural",
		"sad", "saddle", "sadness", "safe", "sail", "salad", "salmon", "salon", "salt", "salute", "same", "sample", "sand", "satisfy", "satoshi", "sauce", "sausage", "save", "say", "scale", "scan", "scare", "scatter", "scene", "scheme", "school", "science", "scissors", "scorpion", "scout", "scrap", "screen", "script", "scrub", "sea", "search", "season", "seat", "second", "secret", "section", "security", "seed", "seek", "segment", "select", "sell", "seminar", "senior", "sense", "sentence", "series", "service", "session", "settle", "setup", "seven", "shadow", "shaft", "shallow", "share", "shed", "shell", "sheriff", "shield", "shift", "shine", "ship", "shiver", "shock", "shoe", "shoot", "shop", "short", "shoulder", "shove", "shrimp", "shrug", "shuffle", "shy", "sibling", "sick", "side", "siege", "sight", "sign", "silent", "silk", "silly", "silver", "similar", "simple", "since", "sing", "siren", "sister", "situate", "six", "size", "skate", "sketch", "ski", "skill", "skin", "skirt", "skull", "slab", "slam", "sleep", "slender", "slice", "slide", "slight", "slim", "slogan", "slot", "slow", "slush", "small", "smart", "smile", "smoke", "smooth", "snack", "snake", "snap", "sniff", "snow", "soap", "soccer", "social", "sock", "soda", "soft", "solar", "soldier", "solid", "solution", "solve", "someone", "song", "soon", "sorry", "sort", "soul", "sound", "soup", "source", "south", "space", "spare", "spatial", "spawn", "speak", "special", "speed", "spell", "spend", "sphere", "spice", "spider", "spike", "spin", "spirit", "split", "spoil", "sponsor", "spoon", "sport", "spot", "spray", "spread", "spring", "spy", "square", "squeeze", "squirrel", "stable", "stadium", "staff", "stage", "stairs", "stamp", "stand", "start", "state", "stay", "steak", "steel", "stem", "step", "stereo", "stick", "still", "sting", "stock", "stomach", "stone", "stool", "story", "stove", "strategy", "street", "strike", "strong", "struggle", "student", "stuff", "stumble", "style", "subject", "submit", "subway", "success", "such", "sudden", "suffer", "sugar", "suggest", "suit", "summer", "sun", "sunny", "sunset", "super", "supply", "supreme", "sure", "surface", "surge", "surprise", "surround", "survey", "suspect", "sustain", "swallow", "swamp", "swap", "swarm", "swear", "sweet", "swift", "swim", "swing", "switch", "sword", "symbol", "symptom", "syrup", "system",
		"table", "tackle", "tag", "tail", "talent", "talk", "tank", "tape", "target", "task", "taste", "tattoo", "taxi", "teach", "team", "tell", "ten", "tenant", "tennis", "tent", "term", "test", "text", "thank", "that", "theme", "then", "theory", "there", "they", "thing", "this", "thought", "three", "thrive", "throw", "thumb", "thunder", "ticket", "tide", "tiger", "tilt", "timber", "time", "tiny", "tip", "tired", "tissue", "title", "toast", "tobacco", "today", "toddler", "toe", "together", "toilet", "token", "tomato", "tomorrow", "tone", "tongue", "tonight", "tool", "tooth", "top", "topic", "topple", "torch", "tornado", "tortoise", "toss", "total", "tourist", "toward", "tower", "town", "toy", "track", "trade", "traffic", "tragic", "train", "transfer", "trap", "trash", "travel", "tray", "treat", "tree", "trend", "trial", "tribe", "trick", "trigger", "trim", "trip", "trophy", "trouble", "truck", "true", "truly", "trumpet", "trust", "truth", "try", "tube", "tuition", "tumble", "tuna", "tunnel", "turkey", "turn", "turtle", "twelve", "twenty", "twice", "twin", "twist", "two", "type", "typical",
		"ugly", "umbrella", "unable", "unaware", "uncle", "uncover", "under", "undo", "unfair", "unfold", "unhappy", "uniform", "unique", "unit", "universe", "unknown", "unlock", "until", "unusual", "unveil", "update", "upgrade", "uphold", "upon", "upper", "upset", "urban", "urge", "usage", "use", "used", "useful", "useless", "usual", "utility",
		"vacant", "vacuum", "vague", "valid", "valley", "valve", "van", "vanish", "vapor", "various", "vast", "vault", "vehicle", "velvet", "vendor", "venture", "venue", "verb", "verify", "version", "very", "vessel", "veteran", "viable", "vibrant", "vicious", "victory", "video", "view", "village", "vintage", "violin", "virtual", "virus", "visa", "visit", "visual", "vital", "vivid", "vocal", "voice", "void", "volcano", "volume", "vote", "voyage",
		"wage", "wagon", "wait", "walk", "wall", "walnut", "want", "warfare", "warm", "warrior", "wash", "wasp", "waste", "water", "wave", "way", "wealth", "weapon", "wear", "weasel", "weather", "web", "wedding", "weekend", "weird", "welcome", "west", "wet", "whale", "what", "wheat", "wheel", "when", "where", "whip", "whisper", "wide", "width", "wife", "wild", "will", "win", "window", "wine", "wing", "wink", "winner", "winter", "wire", "wisdom", "wise", "wish", "witness", "wolf", "woman", "wonder", "wood", "wool", "word", "work", "world", "worry", "worth", "wrap", "wreck", "wrestle", "wrist", "write", "wrong",
		"yard", "year", "yellow", "you", "young", "youth",
		"zebra", "zero", "zone", "zoo",
	}
	return randomArrayElement(words, seed) + sep + randomArrayElement(words, seed) + sep + randomArrayElement(words, seed)
}

// RandName name generator
func RandName() string {
	return SeededName(0)
}

// RandWord generate a word with at least min letters and at most max letters.
func RandWord(min, max int) string {
	return lorem.Word(min, max)
}

// RandSentence generate a sentence with at least min words and at most max words.
func RandSentence(min, max int) string {
	return lorem.Sentence(min, max)
}

// RandParagraph generate a paragraph with at least min sentences and at most max sentences.
func RandParagraph(min, max int) string {
	return lorem.Paragraph(min, max)
}

// SeededName name generator
func SeededName(seed int64) string {
	names := []string{
		"Aaron", "Abigail", "Adam", "Alan", "Albert", "Alexander", "Alexis", "Alice",
		"Amanda", "Amber", "Amy", "Andrea", "Andrew", "Angela", "Ann", "Anna",
		"Anthony", "Arthur", "Ashley", "Austin", "Barbara", "Benjamin",
		"Betty", "Beverly", "Billy", "Bobby", "Bradley", "Brandon",
		"Brenda", "Brian", "Brittany", "Bruce", "Bryan", "Carl", "Carol",
		"Carolyn", "Catherine", "Charles", "Cheryl", "Christian",
		"Christina", "Christine", "Christopher", "Cynthia", "Daniel",
		"Danielle", "David", "Deborah", "Debra", "Denise", "Dennis",
		"Diana", "Diane", "Donald", "Donna", "Doris", "Dorothy", "Douglas",
		"Dylan", "Edward", "Elizabeth", "Emily", "Emma", "Eric", "Ethan",
		"Eugene", "Evelyn", "Frances", "Frank", "Gabriel", "Gary", "George",
		"Gerald", "Gloria", "Grace", "Gregory", "Hannah", "Harold",
		"Heather", "Helen", "Henry", "Jack", "Jacob", "Jacqueline", "James",
		"Jane", "Janet", "Janice", "Jason", "Jean", "Jeffrey", "Jennifer",
		"Jeremy", "Jerry", "Jesse", "Jessica", "Joan", "Joe", "John",
		"Johnny", "Jonathan", "Jordan", "Jose", "Joseph", "Joshua", "Joyce",
		"Juan", "Judith", "Judy", "Julia", "Julie", "Justin", "Karen",
		"Katherine", "Kathleen", "Kathryn", "Kayla", "Keith", "Kelly",
		"Kenneth", "Kevin", "Kimberly", "Kyle", "Larry", "Laura", "Lauren",
		"Lawrence", "Linda", "Lisa", "Logan", "Lori", "Louis", "Madison",
		"Margaret", "Maria", "Marie", "Marilyn", "Mark", "Martha", "Mary",
		"Matthew", "Megan", "Melissa", "Michael", "Michelle", "Nancy",
		"Natalie", "Nathan", "Nicholas", "Nicole", "Noah", "Olivia",
		"Pamela", "Patricia", "Patrick", "Paul", "Peter", "Philip",
		"Rachel", "Ralph", "Randy", "Raymond", "Rebecca", "Richard",
		"Robert", "Roger", "Ronald", "Rose", "Roy", "Russell", "Ruth",
		"Ryan", "Samantha", "Samuel", "Sandra", "Sara", "Sarah", "Scott",
		"Sean", "Sharon", "Shirley", "Sophia", "Stephanie", "Stephen",
		"Steven", "Susan", "Teresa", "Terry", "Theresa", "Thomas",
		"Timothy", "Tyler", "Victoria", "Vincent", "Virginia", "Walter",
		"Wayne", "William", "Willie", "Zachary",
	}
	return randomArrayElement(names, seed)
}

// RandRegex generator
func RandRegex(re string) string {
	re = StripTypeTags(re)
	if strings.Contains(re, EmailRegex) ||
		strings.Contains(re, EmailRegex2) ||
		strings.Contains(re, EmailRegex3) ||
		strings.Contains(re, EmailRegex4) {
		return RandEmail()
	}
	if !strings.Contains(re, `\`) &&
		strings.Contains(re, `+`) &&
		strings.Contains(re, `*`) &&
		strings.Contains(re, `(`) {
		return re
	}
	re = replaceWordTag(re, `(.+)`)
	re = replaceWordTag(re, `\\w`)
	re = replaceWordTag(re, `\w`)
	re = replaceNumTag(re, `\\d`)
	re = replaceNumTag(re, `\d`)
	out, err := regen.Generate(re)
	if err != nil {
		out, err = reggen.Generate(re, 64)
		if err != nil {
			log.WithFields(log.Fields{
				"Error":   err,
				"Pattern": re,
			}).Warnf("failed to parse regex")
			return RandSentence(1, 5)
		}
	}
	return out
}

// RandEmail generator
func RandEmail() string {
	return strings.ToLower(RandName() + "." + RandWord(5, 10) + `@` + RandHost())
}

// RandURL generator
func RandURL() string {
	protocol, err := regen.Generate(`(ftp|http|https|mailto)`)
	if err != nil {
		return err.Error()
	}
	return strings.ToLower(protocol + `://` + RandHost())
}

// RandHost generator
func RandHost() string {
	domain, err := regen.Generate(`(com|org|net|io|gov|edu)`)
	if err != nil {
		return err.Error()
	}
	return strings.ToLower(RandWord(5, 10) + `.` + domain)
}

// RandPhone generator
func RandPhone() string {
	return RandRegex(`1-\d{3}-\d{3}-\d{4}`)
}

// RandIntArrayMinMax generator
func RandIntArrayMinMax(min int, max int) []int {
	if max == 0 {
		max = min + 10
	}
	arr := make([]int, RandNumMinMax(min, max))
	for i := 0; i < len(arr); i++ {
		arr[i] = RandNumMinMax(min, max)
	}
	return arr
}

// RandStringArrayMinMax generator
func RandStringArrayMinMax(min int, max int) []string {
	if max == 0 {
		max = min + 10
	}
	arr := make([]string, RandNumMinMax(min, max))
	for i := 0; i < len(arr); i++ {
		arr[i] = RandTriString("_")
	}
	return arr
}

// RandStringMinMax generator
func RandStringMinMax(min int, max int) string {
	if max == 0 {
		return RandTriString("_")
	}
	if max > 200 {
		max = 200
	}
	return RandString(RandNumMinMax(min, max))
}

// RandString generator
func RandString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

// FileProperty generator
func FileProperty(fileName string, name string) any {
	file, err := os.ReadFile(fileName)
	if err != nil {
		return err.Error()
	}

	data := make(map[string]any)

	err = yaml.Unmarshal(file, &data)
	if err != nil {
		return err.Error()
	}
	return data[name]
}

// RandFileLine generator
func RandFileLine(fileName string) string {
	return SeededFileLine(fileName, 0)
}

// SeededFileLine generator
func SeededFileLine(fileName string, seed int64) string {
	lines, err := fileLines(fileName)
	if err != nil {
		return err.Error()
	}
	return strings.TrimSpace(randomArrayElement(lines, seed))
}

// PRIVATE FUNCTIONS

func randomArrayElement(arr []string, seed int64) string {
	return strings.TrimSpace(arr[SeededRandom(seed, len(arr))])
}

func fileLines(fileName string) ([]string, error) {
	readFile, err := os.Open(fileName)

	if err != nil {
		return nil, err
	}
	var lines []string

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		lines = append(lines, fileScanner.Text())
	}
	_ = readFile.Close()
	return lines, nil
}

func parseRegexMinMax(str string, tag string, start int) (int, int, int) {
	min := -1
	max := -1
	i := start
	var sb strings.Builder
	runes := []rune(str)
	for i = start + len(tag) + 1; i < len(runes); i++ {
		if runes[i] == ',' {
			min, _ = strconv.Atoi(sb.String())
			sb.Reset()
		} else if runes[i] == '}' {
			i++
			break
		} else if unicode.IsDigit(runes[i]) {
			sb.WriteRune(runes[i])
		}
	}
	max, _ = strconv.Atoi(sb.String())
	if min == -1 {
		min = max
	}
	return min, max, i
}

func replaceWordTag(str, tag string) string {
	for {
		start := strings.Index(str, tag+"+")
		if start != -1 {
			str = strings.Replace(str, tag+"+", RandSentence(5, 10), 1)
		}
		start = strings.Index(str, tag+"{")
		if start != -1 {
			min, max, i := parseRegexMinMax(str, tag, start)
			str = str[0:start] + RandSentence(min, max) + str[i:]
		}
		start = strings.Index(str, tag)
		if start == -1 {
			return str
		}
		str = strings.Replace(str, tag, RandWord(5, 10), 1)
	}
}

func replaceNumTag(str, tag string) string {
	for {
		start := strings.Index(str, tag+"+")
		if start != -1 {
			str = strings.Replace(str, tag+"+", strconv.Itoa(RandNumMinMax(1, 10000)), 1)
		}
		start = strings.Index(str, tag+"{")
		if start != -1 {
			min, max, i := parseRegexMinMax(str, tag, start)
			var sb strings.Builder
			limit := RandNumMinMax(min, max)
			for i := 0; i < limit; i++ {
				if i == 0 {
					sb.WriteString(strconv.Itoa(RandNumMinMax(1, 9)))
				} else {
					sb.WriteString(strconv.Itoa(RandNumMinMax(0, 9)))
				}
			}
			str = str[0:start] + sb.String() + str[i:]
		}
		start = strings.Index(str, tag)
		if start == -1 {
			return str
		}
		str = strings.Replace(str, tag, strconv.Itoa(RandNumMinMax(0, 9)), 1)
	}
}
