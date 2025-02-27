package fuzz

import (
	"bufio"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/fuzz/lorem"
	"github.com/lucasjones/reggen"
	"github.com/oklog/ulid/v2"
	log "github.com/sirupsen/logrus"
	"github.com/twinj/uuid"
	regen "github.com/zach-klippenstein/goregen"
	"gopkg.in/yaml.v3"
	"math/rand"
	mrand "math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var entropy = ulid.Monotonic(mrand.New(mrand.NewSource(time.Now().UnixNano())), 0)

// RandIntMinMax returns random number between min and max
func RandIntMinMax(min, max int) int {
	return SeededRandIntMax(0, min, max)
}

// RandIntMax returns random number between 0 and max
func RandIntMax(max int) int {
	return SeededRandIntMax(0, 0, max)
}

// SeededRandIntMax returns random number with seed upto a max
func SeededRandIntMax(seed int64, min, max int) int {
	if seed <= 0 {
		seed = time.Now().UnixNano()
	}
	if max == 0 {
		max = 100000
	}
	if min == max {
		return min
	}
	r := rand.New(rand.NewSource(seed))
	return r.Intn(max-min) + min
}

// RandFloatMinMax returns random number between min and max
func RandFloatMinMax(min, max float64) float64 {
	return SeededRandFloatMax(0, min, max)
}

// RandFloatMax returns random number between 0 and max
func RandFloatMax(max float64) float64 {
	return SeededRandFloatMax(0, 0, max)
}

// SeededRandFloatMax returns random number with seed upto a max
func SeededRandFloatMax(seed int64, min, max float64) float64 {
	if seed <= 0 {
		seed = time.Now().UnixNano()
	}
	r := rand.New(rand.NewSource(seed))
	if max == 0 {
		max = 100000
	}
	if min == max {
		return min
	}
	return min + r.Float64()*(max-min)
}

// ULID generator
func ULID() string {
	return ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()
}

// UUID generator
func UUID() string {
	return uuid.NewV4().String()
}

// SeededUUID generates a deterministic UUID v4 based on a seed value
func SeededUUID(seed int64) string {
	// Create a new random source with the seed
	source := rand.NewSource(seed)
	rng := rand.New(source)

	// Generate 16 random bytes
	uuid := make([]byte, 16)
	rng.Read(uuid)

	// Set version (4) and variant bits according to RFC 4122
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant 1

	return fmt.Sprintf("%x-%x-%x-%x-%x",
		uuid[0:4],
		uuid[4:6],
		uuid[6:8],
		uuid[8:10],
		uuid[10:16])
}

// RandBool generator
func RandBool() bool {
	return SeededBool(0)
}

// SeededBool generator
func SeededBool(seed int64) bool {
	bools := []bool{true, false}
	return bools[SeededRandIntMax(seed, 0, 2)]
}

// RandItin generator for random United States Individual Taxpayer Identification Number ITIN).
func RandItin() string {
	return RandRegex(`\d{3}-\d{2}-\d{4}`)
}

// RandEin Generate a random United States Employer Identification Number EIN).
func RandEin() string {
	return RandRegex(`\d{2}-\d{7}`)
}

// RandSsn Generate a random invalid United States Social Security Identification Number SSN).
func RandSsn() string {
	return RandRegex(`\d{3}-\d{2}-\d{4}`)
}

// RandFirstName first name
func RandFirstName() string {
	return SeededFirstName(0)
}

// SeededFirstName first name
func SeededFirstName(seed int64) string {
	if SeededBool(seed) {
		return SeededFirstMaleName(seed)
	}
	return SeededFirstFemaleName(seed)
}

// RandFirstMaleName first name
func RandFirstMaleName() string {
	return SeededFirstMaleName(0)
}

// SeededFirstMaleName first name
func SeededFirstMaleName(seed int64) string {
	names := []string{
		"Aaron",
		"Adam",
		"Adrian",
		"Alan",
		"Albert",
		"Alec",
		"Alejandro",
		"Alex",
		"Alexander",
		"Alexis",
		"Alfred",
		"Allen",
		"Alvin",
		"Andre",
		"Andres",
		"Andrew",
		"Angel",
		"Anthony",
		"Antonio",
		"Arthur",
		"Austin",
		"Barry",
		"Benjamin",
		"Bernard",
		"Bill",
		"Billy",
		"Blake",
		"Bob",
		"Bobby",
		"Brad",
		"Bradley",
		"Brady",
		"Brandon",
		"Brendan",
		"Brent",
		"Brett",
		"Brian",
		"Bruce",
		"Bryan",
		"Bryce",
		"Caleb",
		"Calvin",
		"Cameron",
		"Carl",
		"Carlos",
		"Casey",
		"Cesar",
		"Chad",
		"Charles",
		"Chase",
		"Chris",
		"Christian",
		"Christopher",
		"Clarence",
		"Clayton",
		"Clifford",
		"Clinton",
		"Cody",
		"Cole",
		"Colin",
		"Collin",
		"Colton",
		"Connor",
		"Corey",
		"Cory",
		"Craig",
		"Cristian",
		"Curtis",
		"Dakota",
		"Dale",
		"Dalton",
		"Damon",
		"Dan",
		"Daniel",
		"Danny",
		"Darin",
		"Darius",
		"Darrell",
		"Darren",
		"Darryl",
		"Daryl",
		"Dave",
		"David",
		"Dean",
		"Dennis",
		"Derek",
		"Derrick",
		"Devin",
		"Devon",
		"Dillon",
		"Dominic",
		"Don",
		"Donald",
		"Douglas",
		"Drew",
		"Duane",
		"Dustin",
		"Dwayne",
		"Dylan",
		"Earl",
		"Eddie",
		"Edgar",
		"Eduardo",
		"Edward",
		"Edwin",
		"Elijah",
		"Eric",
		"Erik",
		"Ernest",
		"Ethan",
		"Eugene",
		"Evan",
		"Fernando",
		"Francis",
		"Francisco",
		"Frank",
		"Franklin",
		"Fred",
		"Frederick",
		"Gabriel",
		"Garrett",
		"Gary",
		"Gavin",
		"Gene",
		"Geoffrey",
		"George",
		"Gerald",
		"Gilbert",
		"Glen",
		"Glenn",
		"Gordon",
		"Grant",
		"Greg",
		"Gregg",
		"Gregory",
		"Guy",
		"Harold",
		"Harry",
		"Hayden",
		"Hector",
		"Henry",
		"Herbert",
		"Howard",
		"Hunter",
		"Ian",
		"Isaac",
		"Isaiah",
		"Ivan",
		"Jack",
		"Jackson",
		"Jacob",
		"Jaime",
		"Jake",
		"James",
		"Jamie",
		"Jared",
		"Jason",
		"Javier",
		"Jay",
		"Jeff",
		"Jeffery",
		"Jeffrey",
		"Jeremiah",
		"Jeremy",
		"Jermaine",
		"Jerome",
		"Jerry",
		"Jesse",
		"Jesus",
		"Jim",
		"Jimmy",
		"Joe",
		"Joel",
		"John",
		"Johnathan",
		"Johnny",
		"Jon",
		"Jonathan",
		"Jonathon",
		"Jordan",
		"Jorge",
		"Jose",
		"Joseph",
		"Joshua",
		"Juan",
		"Julian",
		"Justin",
		"Karl",
		"Keith",
		"Kelly",
		"Kenneth",
		"Kent",
		"Kerry",
		"Kevin",
		"Kirk",
		"Kristopher",
		"Kurt",
		"Kyle",
		"Lance",
		"Larry",
		"Lawrence",
		"Lee",
		"Leon",
		"Leonard",
		"Leroy",
		"Leslie",
		"Levi",
		"Logan",
		"Lonnie",
		"Louis",
		"Lucas",
		"Luis",
		"Luke",
		"Malik",
		"Manuel",
		"Marc",
		"Marco",
		"Marcus",
		"Mario",
		"Mark",
		"Martin",
		"Marvin",
		"Mason",
		"Mathew",
		"Matthew",
		"Maurice",
		"Max",
		"Maxwell",
		"Melvin",
		"Michael",
		"Micheal",
		"Miguel",
		"Mike",
		"Mitchell",
		"Nathan",
		"Nathaniel",
		"Neil",
		"Nicholas",
		"Nicolas",
		"Noah",
		"Norman",
		"Omar",
		"Oscar",
		"Parker",
		"Patrick",
		"Paul",
		"Pedro",
		"Perry",
		"Peter",
		"Philip",
		"Phillip",
		"Preston",
		"Ralph",
		"Randall",
		"Randy",
		"Ray",
		"Raymond",
		"Reginald",
		"Ricardo",
		"Richard",
		"Rick",
		"Rickey",
		"Ricky",
		"Riley",
		"Robert",
		"Roberto",
		"Rodney",
		"Roger",
		"Ronald",
		"Ronnie",
		"Ross",
		"Roy",
		"Ruben",
		"Russell",
		"Ryan",
		"Samuel",
		"Scott",
		"Sean",
		"Sergio",
		"Seth",
		"Shane",
		"Shannon",
		"Shaun",
		"Shawn",
		"Spencer",
		"Stanley",
		"Stephen",
		"Steve",
		"Steven",
		"Stuart",
		"Tanner",
		"Taylor",
		"Terrance",
		"Terrence",
		"Terry",
		"Theodore",
		"Thomas",
		"Tim",
		"Timothy",
		"Todd",
		"Tom",
		"Tommy",
		"Tony",
		"Tracy",
		"Travis",
		"Trevor",
		"Tristan",
		"Troy",
		"Tyler",
		"Tyrone",
		"Vernon",
		"Victor",
		"Vincent",
		"Walter",
		"Warren",
		"Wayne",
		"Wesley",
		"William",
		"Willie",
		"Wyatt",
		"Xavier",
		"Zachary",
	}
	return randomArrayElement(names, seed)
}

// RandLastName last name
func RandLastName() string {
	return SeededLastName(0)
}

// SeededLastName last name
func SeededLastName(seed int64) string {
	names := []string{
		"Smith",
		"Johnson",
		"Williams",
		"Brown",
		"Jones",
		"Miller",
		"Davis",
		"Garcia",
		"Rodriguez",
		"Wilson",
		"Martinez",
		"Anderson",
		"Taylor",
		"Thomas",
		"Hernandez",
		"Moore",
		"Martin",
		"Jackson",
		"Thompson",
		"White",
		"Lopez",
		"Lee",
		"Gonzalez",
		"Harris",
		"Clark",
		"Lewis",
		"Robinson",
		"Walker",
		"Perez",
		"Hall",
		"Young",
		"Allen",
		"Sanchez",
		"Wright",
		"King",
		"Scott",
		"Green",
		"Baker",
		"Adams",
		"Nelson",
		"Hill",
		"Ramirez",
		"Campbell",
		"Mitchell",
		"Roberts",
		"Carter",
		"Phillips",
		"Evans",
		"Turner",
		"Torres",
		"Parker",
		"Collins",
		"Edwards",
		"Stewart",
		"Flores",
		"Morris",
		"Nguyen",
		"Murphy",
		"Rivera",
		"Cook",
		"Rogers",
		"Morgan",
		"Peterson",
		"Cooper",
		"Reed",
		"Bailey",
		"Bell",
		"Gomez",
		"Kelly",
		"Howard",
		"Ward",
		"Cox",
		"Diaz",
		"Richardson",
		"Wood",
		"Watson",
		"Brooks",
		"Bennett",
		"Gray",
		"James",
		"Reyes",
		"Cruz",
		"Hughes",
		"Price",
		"Myers",
		"Long",
		"Foster",
		"Sanders",
		"Ross",
		"Morales",
		"Powell",
		"Sullivan",
		"Russell",
		"Ortiz",
		"Jenkins",
		"Gutierrez",
		"Perry",
		"Butler",
		"Barnes",
		"Fisher",
		"Henderson",
		"Coleman",
		"Simmons",
		"Patterson",
		"Jordan",
		"Reynolds",
		"Hamilton",
		"Graham",
		"Kim",
		"Gonzales",
		"Alexander",
		"Ramos",
		"Wallace",
		"Griffin",
		"West",
		"Cole",
		"Hayes",
		"Chavez",
		"Gibson",
		"Bryant",
		"Ellis",
		"Stevens",
		"Murray",
		"Ford",
		"Marshall",
		"Owens",
		"Mcdonald",
		"Harrison",
		"Ruiz",
		"Kennedy",
		"Wells",
		"Alvarez",
		"Woods",
		"Mendoza",
		"Castillo",
		"Olson",
		"Webb",
		"Washington",
		"Tucker",
		"Freeman",
		"Burns",
		"Henry",
		"Vasquez",
		"Snyder",
		"Simpson",
		"Crawford",
		"Jimenez",
		"Porter",
		"Mason",
		"Shaw",
		"Gordon",
		"Wagner",
		"Hunter",
		"Romero",
		"Hicks",
		"Dixon",
		"Hunt",
		"Palmer",
		"Robertson",
		"Black",
		"Holmes",
		"Stone",
		"Meyer",
		"Boyd",
		"Mills",
		"Warren",
		"Fox",
		"Rose",
		"Rice",
		"Moreno",
		"Schmidt",
		"Patel",
		"Ferguson",
		"Nichols",
		"Herrera",
		"Medina",
		"Ryan",
		"Fernandez",
		"Weaver",
		"Daniels",
		"Stephens",
		"Gardner",
		"Payne",
		"Kelley",
		"Dunn",
		"Pierce",
		"Arnold",
		"Tran",
		"Spencer",
		"Peters",
		"Hawkins",
		"Grant",
		"Hansen",
		"Castro",
		"Hoffman",
		"Hart",
		"Elliott",
		"Cunningham",
		"Knight",
		"Bradley",
		"Carroll",
		"Hudson",
		"Duncan",
		"Armstrong",
		"Berry",
		"Andrews",
		"Johnston",
		"Ray",
		"Lane",
		"Riley",
		"Carpenter",
		"Perkins",
		"Aguilar",
		"Silva",
		"Richards",
		"Willis",
		"Matthews",
		"Chapman",
		"Lawrence",
		"Garza",
		"Vargas",
		"Watkins",
		"Wheeler",
		"Larson",
		"Carlson",
		"Harper",
		"George",
		"Greene",
		"Burke",
		"Guzman",
		"Morrison",
		"Munoz",
		"Jacobs",
		"Obrien",
		"Lawson",
		"Franklin",
		"Lynch",
		"Bishop",
		"Carr",
		"Salazar",
		"Austin",
		"Mendez",
		"Gilbert",
		"Jensen",
		"Williamson",
		"Montgomery",
		"Harvey",
		"Oliver",
		"Howell",
		"Dean",
		"Hanson",
		"Weber",
		"Garrett",
		"Sims",
		"Burton",
		"Fuller",
		"Soto",
		"Mccoy",
		"Welch",
		"Chen",
		"Schultz",
		"Walters",
		"Reid",
		"Fields",
		"Walsh",
		"Little",
		"Fowler",
		"Bowman",
		"Davidson",
		"May",
		"Day",
		"Schneider",
		"Newman",
		"Brewer",
		"Lucas",
		"Holland",
		"Wong",
		"Banks",
		"Santos",
		"Curtis",
		"Pearson",
		"Delgado",
		"Valdez",
		"Pena",
		"Rios",
		"Douglas",
		"Sandoval",
		"Barrett",
		"Hopkins",
		"Keller",
		"Guerrero",
		"Stanley",
		"Bates",
		"Alvarado",
		"Beck",
		"Ortega",
		"Wade",
		"Estrada",
		"Contreras",
		"Barnett",
		"Caldwell",
		"Santiago",
		"Lambert",
		"Powers",
		"Chambers",
		"Nunez",
		"Craig",
		"Leonard",
		"Lowe",
		"Rhodes",
		"Byrd",
		"Gregory",
		"Shelton",
		"Frazier",
		"Becker",
		"Maldonado",
		"Fleming",
		"Vega",
		"Sutton",
		"Cohen",
		"Jennings",
		"Parks",
		"Mcdaniel",
		"Watts",
		"Barker",
		"Norris",
		"Vaughn",
		"Vazquez",
		"Holt",
		"Schwartz",
		"Steele",
		"Benson",
		"Neal",
		"Dominguez",
		"Horton",
		"Terry",
		"Wolfe",
		"Hale",
		"Lyons",
		"Graves",
		"Haynes",
		"Miles",
		"Park",
		"Warner",
		"Padilla",
		"Bush",
		"Thornton",
		"Mccarthy",
		"Mann",
		"Zimmerman",
		"Erickson",
		"Fletcher",
		"Mckinney",
		"Page",
		"Dawson",
		"Joseph",
		"Marquez",
		"Reeves",
		"Klein",
		"Espinoza",
		"Baldwin",
		"Moran",
		"Love",
		"Robbins",
		"Higgins",
		"Ball",
		"Cortez",
		"Le",
		"Griffith",
		"Bowen",
		"Sharp",
		"Cummings",
		"Ramsey",
		"Hardy",
		"Swanson",
		"Barber",
		"Acosta",
		"Luna",
		"Chandler",
		"Daniel",
		"Blair",
		"Cross",
		"Simon",
		"Dennis",
		"Oconnor",
		"Quinn",
		"Gross",
		"Navarro",
		"Moss",
		"Fitzgerald",
		"Doyle",
		"Mclaughlin",
		"Rojas",
		"Rodgers",
		"Stevenson",
		"Singh",
		"Yang",
		"Figueroa",
		"Harmon",
		"Newton",
		"Paul",
		"Manning",
		"Garner",
		"Mcgee",
		"Reese",
		"Francis",
		"Burgess",
		"Adkins",
		"Goodman",
		"Curry",
		"Brady",
		"Christensen",
		"Potter",
		"Walton",
		"Goodwin",
		"Mullins",
		"Molina",
		"Webster",
		"Fischer",
		"Campos",
		"Avila",
		"Sherman",
		"Todd",
		"Chang",
		"Blake",
		"Malone",
		"Wolf",
		"Hodges",
		"Juarez",
		"Gill",
		"Farmer",
		"Hines",
		"Gallagher",
		"Duran",
		"Hubbard",
		"Cannon",
		"Miranda",
		"Wang",
		"Saunders",
		"Tate",
		"Mack",
		"Hammond",
		"Carrillo",
		"Townsend",
		"Wise",
		"Ingram",
		"Barton",
		"Mejia",
		"Ayala",
		"Schroeder",
		"Hampton",
		"Rowe",
		"Parsons",
		"Frank",
		"Waters",
		"Strickland",
		"Osborne",
		"Maxwell",
		"Chan",
		"Deleon",
		"Norman",
		"Harrington",
		"Casey",
		"Patton",
		"Logan",
		"Bowers",
		"Mueller",
		"Glover",
		"Floyd",
		"Hartman",
		"Buchanan",
		"Cobb",
		"French",
		"Kramer",
		"Mccormick",
		"Clarke",
		"Tyler",
		"Gibbs",
		"Moody",
		"Conner",
		"Sparks",
		"Mcguire",
		"Leon",
		"Bauer",
		"Norton",
		"Pope",
		"Flynn",
		"Hogan",
		"Robles",
		"Salinas",
		"Yates",
		"Lindsey",
		"Lloyd",
		"Marsh",
		"Mcbride",
		"Owen",
		"Solis",
		"Pham",
		"Lang",
		"Pratt",
		"Lara",
		"Brock",
		"Ballard",
		"Trujillo",
		"Shaffer",
		"Drake",
		"Roman",
		"Aguirre",
		"Morton",
		"Stokes",
		"Lamb",
		"Pacheco",
		"Patrick",
		"Cochran",
		"Shepherd",
		"Cain",
		"Burnett",
		"Hess",
		"Li",
		"Cervantes",
		"Olsen",
		"Briggs",
		"Ochoa",
		"Cabrera",
		"Velasquez",
		"Montoya",
		"Roth",
		"Meyers",
		"Cardenas",
		"Fuentes",
		"Weiss",
		"Wilkins",
		"Hoover",
		"Nicholson",
		"Underwood",
		"Short",
		"Carson",
		"Morrow",
		"Colon",
		"Holloway",
		"Summers",
		"Bryan",
		"Petersen",
		"Mckenzie",
		"Serrano",
		"Wilcox",
		"Carey",
		"Clayton",
		"Poole",
		"Calderon",
		"Gallegos",
		"Greer",
		"Rivas",
		"Guerra",
		"Decker",
		"Collier",
		"Wall",
		"Whitaker",
		"Bass",
		"Flowers",
		"Davenport",
		"Conley",
		"Houston",
		"Huff",
		"Copeland",
		"Hood",
		"Monroe",
		"Massey",
		"Roberson",
		"Combs",
		"Franco",
		"Larsen",
		"Pittman",
		"Randall",
		"Skinner",
		"Wilkinson",
		"Kirby",
		"Cameron",
		"Bridges",
		"Anthony",
		"Richard",
		"Kirk",
		"Bruce",
		"Singleton",
		"Mathis",
		"Bradford",
		"Boone",
		"Abbott",
		"Charles",
		"Allison",
		"Sweeney",
		"Atkinson",
		"Horn",
		"Jefferson",
		"Rosales",
		"York",
		"Christian",
		"Phelps",
		"Farrell",
		"Castaneda",
		"Nash",
		"Dickerson",
		"Bond",
		"Wyatt",
		"Foley",
		"Chase",
		"Gates",
		"Vincent",
		"Mathews",
		"Hodge",
		"Garrison",
		"Trevino",
		"Villarreal",
		"Heath",
		"Dalton",
		"Valencia",
		"Callahan",
		"Hensley",
		"Atkins",
		"Huffman",
		"Roy",
		"Boyer",
		"Shields",
		"Lin",
		"Hancock",
		"Grimes",
		"Glenn",
		"Cline",
		"Delacruz",
		"Camacho",
		"Dillon",
		"Parrish",
		"Oneill",
		"Melton",
		"Booth",
		"Kane",
		"Berg",
		"Harrell",
		"Pitts",
		"Savage",
		"Wiggins",
		"Brennan",
		"Salas",
		"Marks",
		"Russo",
		"Sawyer",
		"Baxter",
		"Golden",
		"Hutchinson",
		"Liu",
		"Walter",
		"Mcdowell",
		"Wiley",
		"Rich",
		"Humphrey",
		"Johns",
		"Koch",
		"Suarez",
		"Hobbs",
		"Beard",
		"Gilmore",
		"Ibarra",
		"Keith",
		"Macias",
		"Khan",
		"Andrade",
		"Ware",
		"Stephenson",
		"Henson",
		"Wilkerson",
		"Dyer",
		"Mcclure",
		"Blackwell",
		"Mercado",
		"Tanner",
		"Eaton",
		"Clay",
		"Barron",
		"Beasley",
		"Oneal",
		"Small",
		"Preston",
		"Wu",
		"Zamora",
		"Macdonald",
		"Vance",
		"Snow",
		"Mcclain",
		"Stafford",
		"Orozco",
		"Barry",
		"English",
		"Shannon",
		"Kline",
		"Jacobson",
		"Woodard",
		"Huang",
		"Kemp",
		"Mosley",
		"Prince",
		"Merritt",
		"Hurst",
		"Villanueva",
		"Roach",
		"Nolan",
		"Lam",
		"Yoder",
		"Mccullough",
		"Lester",
		"Santana",
		"Valenzuela",
		"Winters",
		"Barrera",
		"Orr",
		"Leach",
		"Berger",
		"Mckee",
		"Strong",
		"Conway",
		"Stein",
		"Whitehead",
		"Bullock",
		"Escobar",
		"Knox",
		"Meadows",
		"Solomon",
		"Velez",
		"Odonnell",
		"Kerr",
		"Stout",
		"Blankenship",
		"Browning",
		"Kent",
		"Lozano",
		"Bartlett",
		"Pruitt",
		"Buck",
		"Barr",
		"Gaines",
		"Durham",
		"Gentry",
		"Mcintyre",
		"Sloan",
		"Rocha",
		"Melendez",
		"Herman",
		"Sexton",
		"Moon",
		"Hendricks",
		"Rangel",
		"Stark",
		"Lowery",
		"Hardin",
		"Hull",
		"Sellers",
		"Ellison",
		"Calhoun",
		"Gillespie",
		"Mora",
		"Knapp",
		"Mccall",
		"Morse",
		"Dorsey",
		"Weeks",
		"Nielsen",
		"Livingston",
		"Leblanc",
		"Mclean",
		"Bradshaw",
		"Glass",
		"Middleton",
		"Buckley",
		"Schaefer",
		"Frost",
		"Howe",
		"House",
		"Mcintosh",
		"Ho",
		"Pennington",
		"Reilly",
		"Hebert",
		"Mcfarland",
		"Hickman",
		"Noble",
		"Spears",
		"Conrad",
		"Arias",
		"Galvan",
		"Velazquez",
		"Huynh",
		"Frederick",
		"Randolph",
		"Cantu",
		"Fitzpatrick",
		"Mahoney",
		"Peck",
		"Villa",
		"Michael",
		"Donovan",
		"Mcconnell",
		"Walls",
		"Boyle",
		"Mayer",
		"Zuniga",
		"Giles",
		"Pineda",
		"Pace",
		"Hurley",
		"Mays",
		"Mcmillan",
		"Crosby",
		"Ayers",
		"Case",
		"Bentley",
		"Shepard",
		"Everett",
		"Pugh",
		"David",
		"Mcmahon",
		"Dunlap",
		"Bender",
		"Hahn",
		"Harding",
		"Acevedo",
		"Raymond",
		"Blackburn",
		"Duffy",
		"Landry",
		"Dougherty",
		"Bautista",
		"Shah",
		"Potts",
		"Arroyo",
		"Valentine",
		"Meza",
		"Gould",
		"Vaughan",
		"Fry",
		"Rush",
		"Avery",
		"Herring",
		"Dodson",
		"Clements",
		"Sampson",
		"Tapia",
		"Bean",
		"Lynn",
		"Crane",
		"Farley",
		"Cisneros",
		"Benton",
		"Ashley",
		"Mckay",
		"Finley",
		"Best",
		"Blevins",
		"Friedman",
		"Moses",
		"Sosa",
		"Blanchard",
		"Huber",
		"Frye",
		"Krueger",
		"Bernard",
		"Rosario",
		"Rubio",
		"Mullen",
		"Benjamin",
		"Haley",
		"Chung",
		"Moyer",
		"Choi",
		"Horne",
		"Yu",
		"Woodward",
		"Ali",
		"Nixon",
		"Hayden",
		"Rivers",
		"Estes",
		"Mccarty",
		"Richmond",
		"Stuart",
		"Maynard",
		"Brandt",
		"Oconnell",
		"Hanna",
		"Sanford",
		"Sheppard",
		"Church",
		"Burch",
		"Levy",
		"Rasmussen",
		"Coffey",
		"Ponce",
		"Faulkner",
		"Donaldson",
		"Schmitt",
		"Novak",
		"Costa",
		"Montes",
		"Booker",
		"Cordova",
		"Waller",
		"Arellano",
		"Maddox",
		"Mata",
		"Bonilla",
		"Stanton",
		"Compton",
		"Kaufman",
		"Dudley",
		"Mcpherson",
		"Beltran",
		"Dickson",
		"Mccann",
		"Villegas",
		"Proctor",
		"Hester",
		"Cantrell",
		"Daugherty",
		"Cherry",
		"Bray",
		"Davila",
		"Rowland",
		"Madden",
		"Levine",
		"Spence",
		"Good",
		"Irwin",
		"Werner",
		"Krause",
		"Petty",
		"Whitney",
		"Baird",
		"Hooper",
		"Pollard",
		"Zavala",
		"Jarvis",
		"Holden",
		"Hendrix",
		"Haas",
		"Mcgrath",
		"Bird",
		"Lucero",
		"Terrell",
		"Riggs",
		"Joyce",
		"Rollins",
		"Mercer",
		"Galloway",
		"Duke",
		"Odom",
		"Andersen",
		"Downs",
		"Hatfield",
		"Benitez",
		"Archer",
		"Huerta",
		"Travis",
		"Mcneil",
		"Hinton",
		"Zhang",
		"Hays",
		"Mayo",
		"Fritz",
		"Branch",
		"Mooney",
		"Ewing",
		"Ritter",
		"Esparza",
		"Frey",
		"Braun",
		"Gay",
		"Riddle",
		"Haney",
		"Kaiser",
		"Holder",
		"Chaney",
		"Mcknight",
		"Gamble",
		"Vang",
		"Cooley",
		"Carney",
		"Cowan",
		"Forbes",
		"Ferrell",
		"Davies",
		"Barajas",
		"Shea",
		"Osborn",
		"Bright",
		"Cuevas",
		"Bolton",
		"Murillo",
		"Lutz",
		"Duarte",
		"Kidd",
		"Key",
		"Cooke",
	}
	return randomArrayElement(names, seed)
}

// RandFirstFemaleName first name
func RandFirstFemaleName() string {
	return SeededFirstFemaleName(0)
}

// SeededFirstFemaleName first name
func SeededFirstFemaleName(seed int64) string {
	names := []string{
		"April",
		"Abigail",
		"Adriana",
		"Adrienne",
		"Aimee",
		"Alejandra",
		"Alexa",
		"Alexandra",
		"Alexandria",
		"Alexis",
		"Alice",
		"Alicia",
		"Alisha",
		"Alison",
		"Allison",
		"Alyssa",
		"Amanda",
		"Amber",
		"Amy",
		"Ana",
		"Andrea",
		"Angel",
		"Angela",
		"Angelica",
		"Angie",
		"Anita",
		"Ann",
		"Anna",
		"Anne",
		"Annette",
		"Ariana",
		"Ariel",
		"Ashlee",
		"Ashley",
		"Audrey",
		"Autumn",
		"Bailey",
		"Barbara",
		"Becky",
		"Belinda",
		"Beth",
		"Bethany",
		"Betty",
		"Beverly",
		"Bianca",
		"Bonnie",
		"Brandi",
		"Brandy",
		"Breanna",
		"Brenda",
		"Briana",
		"Brianna",
		"Bridget",
		"Brittany",
		"Brittney",
		"Brooke",
		"Caitlin",
		"Caitlyn",
		"Candace",
		"Candice",
		"Carla",
		"Carly",
		"Carmen",
		"Carol",
		"Caroline",
		"Carolyn",
		"Carrie",
		"Casey",
		"Cassandra",
		"Cassidy",
		"Cassie",
		"Catherine",
		"Cathy",
		"Charlene",
		"Charlotte",
		"Chelsea",
		"Chelsey",
		"Cheryl",
		"Cheyenne",
		"Chloe",
		"Christie",
		"Christina",
		"Christine",
		"Christy",
		"Cindy",
		"Claire",
		"Claudia",
		"Colleen",
		"Connie",
		"Courtney",
		"Cristina",
		"Crystal",
		"Cynthia",
		"Daisy",
		"Dana",
		"Danielle",
		"Darlene",
		"Dawn",
		"Deanna",
		"Debbie",
		"Deborah",
		"Debra",
		"Denise",
		"Desiree",
		"Destiny",
		"Diamond",
		"Diana",
		"Diane",
		"Dominique",
		"Donna",
		"Doris",
		"Dorothy",
		"Ebony",
		"Eileen",
		"Elaine",
		"Elizabeth",
		"Ellen",
		"Emily",
		"Emma",
		"Erica",
		"Erika",
		"Erin",
		"Evelyn",
		"Faith",
		"Felicia",
		"Frances",
		"Gabriela",
		"Gabriella",
		"Gabrielle",
		"Gail",
		"Gina",
		"Glenda",
		"Gloria",
		"Grace",
		"Gwendolyn",
		"Hailey",
		"Haley",
		"Hannah",
		"Hayley",
		"Heather",
		"Heidi",
		"Helen",
		"Holly",
		"Isabel",
		"Isabella",
		"Jackie",
		"Jaclyn",
		"Jacqueline",
		"Jade",
		"Jaime",
		"Jamie",
		"Jane",
		"Janet",
		"Janice",
		"Jasmin",
		"Jasmine",
		"Jean",
		"Jeanette",
		"Jeanne",
		"Jenna",
		"Jennifer",
		"Jenny",
		"Jessica",
		"Jill",
		"Jillian",
		"Jo",
		"Joan",
		"Joann",
		"Joanna",
		"Joanne",
		"Jocelyn",
		"Jodi",
		"Jody",
		"Jordan",
		"Joy",
		"Joyce",
		"Judith",
		"Judy",
		"Julia",
		"Julie",
		"Kaitlin",
		"Kaitlyn",
		"Kara",
		"Karen",
		"Kari",
		"Karina",
		"Karla",
		"Katelyn",
		"Katherine",
		"Kathleen",
		"Kathryn",
		"Kathy",
		"Katie",
		"Katrina",
		"Kayla",
		"Kaylee",
		"Kelli",
		"Kellie",
		"Kelly",
		"Kelsey",
		"Kendra",
		"Kerri",
		"Kerry",
		"Kiara",
		"Kim",
		"Kimberly",
		"Kirsten",
		"Krista",
		"Kristen",
		"Kristi",
		"Kristie",
		"Kristin",
		"Kristina",
		"Kristine",
		"Kristy",
		"Krystal",
		"Kylie",
		"Lacey",
		"Latasha",
		"Latoya",
		"Laura",
		"Lauren",
		"Laurie",
		"Leah",
		"Leslie",
		"Linda",
		"Lindsay",
		"Lindsey",
		"Lisa",
		"Loretta",
		"Lori",
		"Lorraine",
		"Lydia",
		"Lynn",
		"Mackenzie",
		"Madeline",
		"Madison",
		"Makayla",
		"Mallory",
		"Mandy",
		"Marcia",
		"Margaret",
		"Maria",
		"Mariah",
		"Marie",
		"Marilyn",
		"Marisa",
		"Marissa",
		"Martha",
		"Mary",
		"Maureen",
		"Mckenzie",
		"Meagan",
		"Megan",
		"Meghan",
		"Melanie",
		"Melinda",
		"Melissa",
		"Melody",
		"Mercedes",
		"Meredith",
		"Mia",
		"Michaela",
		"Michele",
		"Michelle",
		"Mikayla",
		"Mindy",
		"Miranda",
		"Misty",
		"Molly",
		"Monica",
		"Monique",
		"Morgan",
		"Nancy",
		"Natalie",
		"Natasha",
		"Nichole",
		"Nicole",
		"Nina",
		"Norma",
		"Olivia",
		"Paige",
		"Pam",
		"Pamela",
		"Patricia",
		"Patty",
		"Paula",
		"Peggy",
		"Penny",
		"Phyllis",
		"Priscilla",
		"Rachael",
		"Rachel",
		"Raven",
		"Rebecca",
		"Rebekah",
		"Regina",
		"Renee",
		"Rhonda",
		"Rita",
		"Roberta",
		"Robin",
		"Robyn",
		"Rose",
		"Ruth",
		"Sabrina",
		"Sally",
		"Samantha",
		"Sandra",
		"Sandy",
		"Sara",
		"Sarah",
		"Savannah",
		"Selena",
		"Shannon",
		"Shari",
		"Sharon",
		"Shawna",
		"Sheena",
		"Sheila",
		"Shelby",
		"Shelia",
		"Shelley",
		"Shelly",
		"Sheri",
		"Sherri",
		"Sherry",
		"Sheryl",
		"Shirley",
		"Sierra",
		"Sonia",
		"Sonya",
		"Sophia",
		"Stacey",
		"Stacie",
		"Stacy",
		"Stefanie",
		"Stephanie",
		"Sue",
		"Summer",
		"Susan",
		"Suzanne",
		"Sydney",
		"Sylvia",
		"Tabitha",
		"Tamara",
		"Tami",
		"Tammie",
		"Tammy",
		"Tanya",
		"Tara",
		"Tasha",
		"Taylor",
		"Teresa",
		"Terri",
		"Terry",
		"Theresa",
		"Tiffany",
		"Tina",
		"Toni",
		"Tonya",
		"Tracey",
		"Traci",
		"Tracie",
		"Tracy",
		"Tricia",
		"Valerie",
		"Vanessa",
		"Veronica",
		"Vicki",
		"Vickie",
		"Victoria",
		"Virginia",
		"Wanda",
		"Wendy",
		"Whitney",
		"Yesenia",
		"Yolanda",
		"Yvette",
		"Yvonne",
		"Zoe",
	}
	return randomArrayElement(names, seed)
}

// RandUSState Generate US State
func RandUSState() string {
	return SeededUSState(0)
}

// SeededUSState Generate US State
func SeededUSState(seed int64) string {
	states := []string{
		"Alabama",
		"Alaska",
		"Arizona",
		"Arkansas",
		"California",
		"Colorado",
		"Connecticut",
		"Delaware",
		"Florida",
		"Georgia",
		"Hawaii",
		"Idaho",
		"Illinois",
		"Indiana",
		"Iowa",
		"Kansas",
		"Kentucky",
		"Louisiana",
		"Maine",
		"Maryland",
		"Massachusetts",
		"Michigan",
		"Minnesota",
		"Mississippi",
		"Missouri",
		"Montana",
		"Nebraska",
		"Nevada",
		"New Hampshire",
		"New Jersey",
		"New Mexico",
		"New York",
		"North Carolina",
		"North Dakota",
		"Ohio",
		"Oklahoma",
		"Oregon",
		"Pennsylvania",
		"Rhode Island",
		"South Carolina",
		"South Dakota",
		"Tennessee",
		"Texas",
		"Utah",
		"Vermont",
		"Virginia",
		"Washington",
		"West Virginia",
		"Wisconsin",
		"Wyoming",
	}
	return randomArrayElement(states, seed)
}

// RandUSStateAbbr Generate US State
func RandUSStateAbbr() string {
	return SeededUSStateAbbr(0)
}

// SeededUSStateAbbr Generate US State
func SeededUSStateAbbr(seed int64) string {
	statesAbbr := []string{
		"AL",
		"AK",
		"AZ",
		"AR",
		"CA",
		"CO",
		"CT",
		"DE",
		"DC",
		"FL",
		"GA",
		"HI",
		"ID",
		"IL",
		"IN",
		"IA",
		"KS",
		"KY",
		"LA",
		"ME",
		"MD",
		"MA",
		"MI",
		"MN",
		"MS",
		"MO",
		"MT",
		"NE",
		"NV",
		"NH",
		"NJ",
		"NM",
		"NY",
		"NC",
		"ND",
		"OH",
		"OK",
		"OR",
		"PA",
		"RI",
		"SC",
		"SD",
		"TN",
		"TX",
		"UT",
		"VT",
		"VA",
		"WA",
		"WV",
		"WI",
		"WY",
	}
	return randomArrayElement(statesAbbr, seed)
}

// SeededUSPostal Generate US postal code
func SeededUSPostal(state string, seed int64) string {
	statesPostcode := map[string][]int{
		"AL": {35004, 36925},
		"AK": {99501, 99950},
		"AZ": {85001, 86556},
		"AR": {71601, 72959},
		"CA": {90001, 96162},
		"CO": {80001, 81658},
		"CT": {6001, 6389},
		"DE": {19701, 19980},
		"DC": {20001, 20039},
		"FL": {32004, 34997},
		"GA": {30001, 31999},
		"HI": {96701, 96898},
		"ID": {83201, 83876},
		"IL": {60001, 62999},
		"IN": {46001, 47997},
		"IA": {50001, 52809},
		"KS": {66002, 67954},
		"KY": {40003, 42788},
		"LA": {70001, 71232},
		"ME": {3901, 4992},
		"MD": {20812, 21930},
		"MA": {1001, 2791},
		"MI": {48001, 49971},
		"MN": {55001, 56763},
		"MS": {38601, 39776},
		"MO": {63001, 65899},
		"MT": {59001, 59937},
		"NE": {68001, 68118},
		"NV": {88901, 89883},
		"NH": {3031, 3897},
		"NJ": {7001, 8989},
		"NM": {87001, 88441},
		"NY": {10001, 14905},
		"NC": {27006, 28909},
		"ND": {58001, 58856},
		"OH": {43001, 45999},
		"OK": {73001, 73199},
		"OR": {97001, 97920},
		"PA": {15001, 19640},
		"RI": {2801, 2940},
		"SC": {29001, 29948},
		"SD": {57001, 57799},
		"TN": {37010, 38589},
		"TX": {75503, 79999},
		"UT": {84001, 84784},
		"VT": {5001, 5495},
		"VA": {22001, 24658},
		"WA": {98001, 99403},
		"WV": {24701, 26886},
		"WI": {53001, 54990},
		"WY": {82001, 83128},
	}
	minMax := statesPostcode[state]
	if minMax == nil {
		return ""
	}
	if SeededBool(seed) {
		return fmt.Sprintf("%d", RandIntMinMax(minMax[0], minMax[1]))
	}
	return fmt.Sprintf("%d-%0d", RandIntMinMax(minMax[0], minMax[1]), RandIntMax(9999))
}

// RandAddress Generate a random address
func RandAddress() string {
	return SeededAddress(0)
}

// SeededAddress Generate a random address
func SeededAddress(seed int64) string {
	cityPrefixes := []string{"North", "East", "West", "South", "New", "Lake", "Port"}
	citySuffixes := []string{
		"town",
		"ton",
		"land",
		"ville",
		"berg",
		"burgh",
		"borough",
		"bury",
		"view",
		"port",
		"mouth",
		"stad",
		"furt",
		"chester",
		"mouth",
		"fort",
		"haven",
		"side",
		"shire",
	}

	streetSuffixes := []string{
		"Alley",
		"Avenue",
		"Branch",
		"Bridge",
		"Brook",
		"Brooks",
		"Burg",
		"Burgs",
		"Bypass",
		"Camp",
		"Canyon",
		"Cape",
		"Causeway",
		"Center",
		"Centers",
		"Circle",
		"Circles",
		"Cliff",
		"Cliffs",
		"Club",
		"Common",
		"Corner",
		"Corners",
		"Course",
		"Court",
		"Courts",
		"Cove",
		"Coves",
		"Creek",
		"Crescent",
		"Crest",
		"Crossing",
		"Crossroad",
		"Curve",
		"Dale",
		"Dam",
		"Divide",
		"Drive",
		"Drive",
		"Drives",
		"Estate",
		"Estates",
		"Expressway",
		"Extension",
		"Extensions",
		"Fall",
		"Falls",
		"Ferry",
		"Field",
		"Fields",
		"Flat",
		"Flats",
		"Ford",
		"Fords",
		"Forest",
		"Forge",
		"Forges",
		"Fork",
		"Forks",
		"Fort",
		"Freeway",
		"Garden",
		"Gardens",
		"Gateway",
		"Glen",
		"Glens",
		"Green",
		"Greens",
		"Grove",
		"Groves",
		"Harbor",
		"Harbors",
		"Haven",
		"Heights",
		"Highway",
		"Hill",
		"Hills",
		"Hollow",
		"Inlet",
		"Inlet",
		"Island",
		"Island",
		"Islands",
		"Islands",
		"Isle",
		"Isle",
		"Junction",
		"Junctions",
		"Key",
		"Keys",
		"Knoll",
		"Knolls",
		"Lake",
		"Lakes",
		"Land",
		"Landing",
		"Lane",
		"Light",
		"Lights",
		"Loaf",
		"Lock",
		"Locks",
		"Locks",
		"Lodge",
		"Lodge",
		"Loop",
		"Mall",
		"Manor",
		"Manors",
		"Meadow",
		"Meadows",
		"Mews",
		"Mill",
		"Mills",
		"Mission",
		"Mission",
		"Motorway",
		"Mount",
		"Mountain",
		"Mountain",
		"Mountains",
		"Neck",
		"Orchard",
		"Oval",
		"Overpass",
		"Park",
		"Parks",
		"Parkway",
		"Parkways",
		"Pass",
		"Passage",
		"Path",
		"Pike",
		"Pine",
		"Pines",
		"Place",
		"Plain",
		"Plains",
		"Plains",
		"Plaza",
		"Plaza",
		"Point",
		"Points",
		"Port",
		"Port",
		"Ports",
		"Ports",
		"Prairie",
		"Prairie",
		"Radial",
		"Ramp",
		"Ranch",
		"Rapid",
		"Rapids",
		"Rest",
		"Ridge",
		"Ridges",
		"River",
		"Road",
		"Road",
		"Roads",
		"Route",
		"Row",
		"Rue",
		"Run",
		"Shoal",
		"Shoals",
		"Shore",
		"Shores",
		"Skyway",
		"Spring",
		"Springs",
		"Springs",
		"Spur",
		"Spurs",
		"Square",
		"Square",
		"Squares",
		"Squares",
		"Station",
		"Station",
		"Stravenue",
		"Stravenue",
		"Stream",
		"Stream",
		"Street",
		"Street",
		"Streets",
		"Summit",
		"Summit",
		"Terrace",
		"Throughway",
		"Trace",
		"Track",
		"Trafficway",
		"Trail",
		"Trail",
		"Tunnel",
		"Turnpike",
		"Turnpike",
		"Underpass",
		"Union",
		"Unions",
		"Valley",
		"Valleys",
		"Via",
		"Viaduct",
		"View",
		"Views",
		"Village",
		"Village",
		"Villages",
		"Ville",
		"Vista",
		"Vista",
		"Walk",
		"Walks",
		"Wall",
		"Way",
		"Ways",
		"Well",
		"Wells",
	}

	cityPrefix := randomArrayElement(cityPrefixes, seed)
	citySuffix := randomArrayElement(citySuffixes, seed)
	streetSuffix := randomArrayElement(streetSuffixes, seed)
	var city string
	if SeededBool(seed) {
		city = fmt.Sprintf("%s %s%s", cityPrefix, SeededFirstName(seed), citySuffix)
	} else {
		city = fmt.Sprintf("%s%s", SeededFirstName(seed), citySuffix)
	}
	var street string
	if SeededBool(seed) {
		street = fmt.Sprintf("%d %s %s", RandIntMinMax(100, 1000), SeededFirstName(seed), streetSuffix)
	} else {
		street = fmt.Sprintf("%d %s %s", RandIntMinMax(100, 1000), SeededFirstName(seed), streetSuffix)
	}

	if SeededBool(seed) {
		secondaryAddress := []string{"Apt. ", "Suite "}
		apt := randomArrayElement(secondaryAddress, seed)
		street += fmt.Sprintf(" %s %d", apt, RandIntMinMax(100, 10000))
	}
	state := SeededUSStateAbbr(seed)
	return fmt.Sprintf("%s\n%s, %s %s", street, city, state, SeededUSPostal(state, seed))
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
	if strings.Contains(re, PrefixTypeExample) {
		return StripTypeTags(re)
	}
	re = StripTypeTags(re)
	if strings.Contains(re, EmailRegex) ||
		strings.Contains(re, EmailRegex2) ||
		strings.Contains(re, EmailRegex3) ||
		strings.Contains(re, EmailRegex4) {
		return RandEmail()
	}
	re = replaceWordTag(re, `(.+)`)
	re = replaceWordTag(re, `\\w`)
	re = replaceWordTag(re, `\w`)
	re = replaceNumTag(re, `\\d`)
	re = replaceNumTag(re, `\d`)
	if !strings.Contains(re, `\`) &&
		!strings.Contains(re, `+`) &&
		!strings.Contains(re, `*`) &&
		!strings.Contains(re, `[`) &&
		!strings.Contains(re, `(`) {
		return re
	}
	var out string
	var err error
	if strings.Contains(re, ".") {
		out, err = reggen.Generate(re, 64)
	} else {
		out, err = regen.Generate(re)
	}
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
	arr := make([]int, RandIntMinMax(min, max))
	for i := 0; i < len(arr); i++ {
		arr[i] = RandIntMinMax(min, max)
	}
	return arr
}

// RandStringArrayMinMax generator
func RandStringArrayMinMax(min int, max int) []string {
	if max == 0 {
		max = min + 10
	}
	arr := make([]string, RandIntMinMax(min, max))
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
	return RandString(RandIntMinMax(min, max))
}

// RandString generator
func RandString(n int) string {
	if n > 0 {
		return RandWord(n, n)
	}
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

// Airline represents an airline with its code and flight number range
type Airline struct {
	Code      string
	MinNumber int
	MaxNumber int
}

// Major US airlines with their typical flight number ranges
var airlines = []Airline{
	{"AA", 1, 9999},  // American Airlines: 1-9999
	{"DL", 1, 9999},  // Delta Air Lines: 1-9999
	{"UA", 1, 9999},  // United Airlines: 1-9999
	{"WN", 1, 9999},  // Southwest Airlines: 1-9999
	{"AS", 1, 3299},  // Alaska Airlines: 1-3299
	{"B6", 1, 2999},  // JetBlue: 1-2999
	{"NK", 100, 999}, // Spirit: 100-999
	{"F9", 100, 999}, // Frontier: 100-999
	{"HA", 1, 499},   // Hawaiian: 1-499
}

// RandFlightNumber generates a random flight number using real airline patterns
func RandFlightNumber() string {
	// Initialize random seed with current time
	rand.Seed(time.Now().UnixNano())

	// Pick a random airline
	airline := airlines[rand.Intn(len(airlines))]

	// Generate random flight number within airline's range
	flightNum := rand.Intn(airline.MaxNumber-airline.MinNumber+1) + airline.MinNumber

	return fmt.Sprintf("%s%d", airline.Code, flightNum)
}

// SeededFlightNumber generates a deterministic flight number based on a seed
func SeededFlightNumber(seed int64) string {
	// Create seeded random source
	source := rand.NewSource(seed)
	rng := rand.New(source)

	// Use the first few bits for airline selection
	airlineIndex := int(seed % int64(len(airlines)))
	airline := airlines[airlineIndex]

	// Use the random number generator for flight number
	flightNum := rng.Intn(airline.MaxNumber-airline.MinNumber+1) + airline.MinNumber

	return fmt.Sprintf("%s%d", airline.Code, flightNum)
}

// Airport represents an airport with its IATA code, city, and country
type Airport struct {
	Code    string
	City    string
	Country string
	Region  string // NA = North America, EU = Europe, AS = Asia, etc.
}

// Top 100 busiest airports by passenger traffic
var airports = []Airport{
	// North America
	{"ATL", "Atlanta", "USA", "NA"},
	{"DFW", "Dallas/Fort Worth", "USA", "NA"},
	{"DEN", "Denver", "USA", "NA"},
	{"ORD", "Chicago", "USA", "NA"},
	{"LAX", "Los Angeles", "USA", "NA"},
	{"CLT", "Charlotte", "USA", "NA"},
	{"MCO", "Orlando", "USA", "NA"},
	{"LAS", "Las Vegas", "USA", "NA"},
	{"PHX", "Phoenix", "USA", "NA"},
	{"MIA", "Miami", "USA", "NA"},
	{"SEA", "Seattle", "USA", "NA"},
	{"IAH", "Houston", "USA", "NA"},
	{"JFK", "New York", "USA", "NA"},
	{"EWR", "Newark", "USA", "NA"},
	{"SFO", "San Francisco", "USA", "NA"},
	{"MSP", "Minneapolis", "USA", "NA"},
	{"DTW", "Detroit", "USA", "NA"},
	{"BOS", "Boston", "USA", "NA"},
	{"PHL", "Philadelphia", "USA", "NA"},
	{"LGA", "New York", "USA", "NA"},
	{"FLL", "Fort Lauderdale", "USA", "NA"},
	{"BWI", "Baltimore", "USA", "NA"},
	{"DCA", "Washington", "USA", "NA"},
	{"SLC", "Salt Lake City", "USA", "NA"},
	{"MDW", "Chicago", "USA", "NA"},
	{"YYZ", "Toronto", "Canada", "NA"},
	{"YVR", "Vancouver", "Canada", "NA"},
	{"MEX", "Mexico City", "Mexico", "NA"},

	// Europe
	{"LHR", "London", "UK", "EU"},
	{"CDG", "Paris", "France", "EU"},
	{"AMS", "Amsterdam", "Netherlands", "EU"},
	{"FRA", "Frankfurt", "Germany", "EU"},
	{"IST", "Istanbul", "Turkey", "EU"},
	{"MAD", "Madrid", "Spain", "EU"},
	{"BCN", "Barcelona", "Spain", "EU"},
	{"FCO", "Rome", "Italy", "EU"},
	{"MUC", "Munich", "Germany", "EU"},
	{"LGW", "London", "UK", "EU"},
	{"DUB", "Dublin", "Ireland", "EU"},
	{"ZRH", "Zurich", "Switzerland", "EU"},
	{"CPH", "Copenhagen", "Denmark", "EU"},
	{"VIE", "Vienna", "Austria", "EU"},
	{"OSL", "Oslo", "Norway", "EU"},
	{"ARN", "Stockholm", "Sweden", "EU"},

	// Asia
	{"PEK", "Beijing", "China", "AS"},
	{"HND", "Tokyo", "Japan", "AS"},
	{"DXB", "Dubai", "UAE", "AS"},
	{"CAN", "Guangzhou", "China", "AS"},
	{"PVG", "Shanghai", "China", "AS"},
	{"ICN", "Seoul", "South Korea", "AS"},
	{"HKG", "Hong Kong", "China", "AS"},
	{"BKK", "Bangkok", "Thailand", "AS"},
	{"SGN", "Ho Chi Minh City", "Vietnam", "AS"},
	{"KUL", "Kuala Lumpur", "Malaysia", "AS"},
	{"DEL", "Delhi", "India", "AS"},
	{"BOM", "Mumbai", "India", "AS"},
	{"SIN", "Singapore", "Singapore", "AS"},
	{"NRT", "Tokyo", "Japan", "AS"},
	{"MNL", "Manila", "Philippines", "AS"},

	// Australia/Pacific
	{"SYD", "Sydney", "Australia", "OC"},
	{"MEL", "Melbourne", "Australia", "OC"},
	{"BNE", "Brisbane", "Australia", "OC"},
	{"AKL", "Auckland", "New Zealand", "OC"},
}

// RandAirport returns a random airport from the list
func RandAirport() string {
	rand.Seed(time.Now().UnixNano())
	return airports[rand.Intn(len(airports))].Code
}

// SeededAirport returns a deterministic airport based on a seed
func SeededAirport(seed int64) string {
	return airports[int(seed)%len(airports)].Code
}

// RandomAirportByRegion returns a random airport from a specific region
func RandomAirportByRegion(region string) string {
	// Filter airports by region
	filtered := make([]Airport, 0)
	for _, airport := range airports {
		if airport.Region == region {
			filtered = append(filtered, airport)
		}
	}

	if len(filtered) == 0 {
		return ""
	}

	rand.Seed(time.Now().UnixNano())
	return filtered[rand.Intn(len(filtered))].Code
}

// SeededAirportByRegion returns a deterministic airport from a specific region
func SeededAirportByRegion(seed int64, region string) Airport {
	// Filter airports by region
	filtered := make([]Airport, 0)
	for _, airport := range airports {
		if airport.Region == region {
			filtered = append(filtered, airport)
		}
	}

	if len(filtered) == 0 {
		return Airport{}
	}

	return filtered[int(seed)%len(filtered)]
}

// PRIVATE FUNCTIONS

func randomArrayElement(arr []string, seed int64) string {
	return strings.TrimSpace(arr[SeededRandIntMax(seed, 0, len(arr))])
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
	start := strings.Index(str, tag+"+")
	for start != -1 {
		str = strings.Replace(str, tag+"+", RandSentence(3, 6), 1)
		start = strings.Index(str, tag+"+")
	}
	start = strings.Index(str, tag+"{")
	for start != -1 {
		min, max, i := parseRegexMinMax(str, tag, start)
		str = str[0:start] + RandSentence(min, max) + str[i:]
		start = strings.Index(str, tag+"{")
	}
	start = strings.Index(str, tag)
	for start != -1 {
		str = strings.Replace(str, tag, RandWord(3, 6), 1)
		start = strings.Index(str, tag)
	}
	return str
}

func replaceNumTag(str, tag string) string {
	start := strings.Index(str, tag+"+")
	for start != -1 {
		str = strings.Replace(str, tag+"+", strconv.Itoa(RandIntMinMax(1, 10000)), 1)
		start = strings.Index(str, tag+"+")
	}
	start = strings.Index(str, tag+"{")
	for start != -1 {
		min, max, i := parseRegexMinMax(str, tag, start)
		var sb strings.Builder
		limit := RandIntMinMax(min, max)
		for i := 0; i < limit; i++ {
			if i == 0 {
				sb.WriteString(strconv.Itoa(RandIntMinMax(1, 9)))
			} else {
				sb.WriteString(strconv.Itoa(RandIntMinMax(0, 9)))
			}
		}
		str = str[0:start] + sb.String() + str[i:]
		start = strings.Index(str, tag+"{")
	}
	start = strings.Index(str, tag)
	for start != -1 {
		str = strings.Replace(str, tag, strconv.Itoa(RandIntMinMax(0, 9)), 1)
		start = strings.Index(str, tag)
	}
	return str
}

// VariableJsonPath Enhance PropertyMatches to support nested objects with JSON path
func VariableJsonPath(varPath string, expected string, data any) bool {
	// Extract the value using JSON path
	value := extractJsonPath(varPath, data)
	if value == nil {
		return false
	}

	// Support regex matching
	if strings.HasPrefix(expected, "/") && strings.HasSuffix(expected, "/") {
		pattern := expected[1 : len(expected)-1]
		match, err := regexp.MatchString(pattern, fmt.Sprintf("%v", value))
		return err == nil && match
	}

	// Basic equality for different types
	return fmt.Sprintf("%v", value) == expected
}
