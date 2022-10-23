package utils

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/twinj/uuid"
)

// RandomMinMax returns random number between min and max
func RandomMinMax(min, max int) int {
	return rand.Intn(max-min) + min
}

// Random returns random number between 0 and max
func Random(max int) int {
	return RandomMinMax(0, max)
}

// SeededRandom returns random number with seed upto a max
func SeededRandom(seed int64, max int) int {
	if seed <= 0 {
		seed = time.Now().UnixNano()
	}
	r := rand.New(rand.NewSource(seed))
	return r.Intn(max)
}

// Udid generator
func Udid() string {
	return uuid.NewV4().String()
}

// SeededUdid generator
func SeededUdid(seed int64) string {
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

// AnySubstring selects substring
func AnySubstring(str string) string {
	parts := strings.Split(str, " ")
	return randomArrayElement(parts, 0)
}

// AnyInt selects numeric
func AnyInt(str string) (n int64) {
	parts := strings.Split(str, " ")
	n, _ = strconv.ParseInt(randomArrayElement(parts, 0), 10, 64)
	return
}

// RandName name generator
func RandName() string {
	return SeededName(0)
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

// RandomString generator
func RandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
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
	return arr[SeededRandom(seed, len(arr))]
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
	readFile.Close()
	return lines, nil
}
