package fuzz

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// CardType represents different credit card types
type CardType string

const (
	Visa            CardType = "Visa"
	Mastercard      CardType = "Mastercard"
	AmericanExpress CardType = "American Express"
	Discover        CardType = "Discover"
	JCB             CardType = "JCB"
	DinersClub      CardType = "Diners Club"
	Unknown         CardType = "Unknown"
)

// CardInfo stores credit card type prefix information
type CardInfo struct {
	Type   CardType
	Prefix []string
	Length []int
}

var cardTypes = []CardInfo{
	{Visa, []string{"4"}, []int{16, 13, 19}},
	{Mastercard, []string{"51", "52", "53", "54", "55", "2221", "2222", "2223", "2224", "2225", "2226", "2227", "2228", "2229", "223", "224", "225", "226", "227", "228", "229", "23", "24", "25", "26", "270", "271", "2720"}, []int{16}},
	{AmericanExpress, []string{"34", "37"}, []int{15}},
	{Discover, []string{"6011", "644", "645", "646", "647", "648", "649", "65"}, []int{16, 19}},
	{JCB, []string{"3528", "3529", "353", "354", "355", "356", "357", "358"}, []int{16, 19}},
	{DinersClub, []string{"300", "301", "302", "303", "304", "305", "36", "38", "39"}, []int{14, 16, 19}},
}

func RandCreditCard() string {
	return SeededCreditCard(time.Now().UnixNano())
}

func SeededCreditCard(seed int64) string {
	return GenerateCreditCard(cardTypes[RandIntMax(len(cardTypes))].Type, seed)
}

// NewRandomSource creates a new random source with a seed
func NewRandomSource(seed int64) *rand.Rand {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	return rand.New(rand.NewSource(seed))
}

// GenerateCreditCard generates a valid credit card number for the specified type
func GenerateCreditCard(cardType CardType, seed int64) string {
	rng := NewRandomSource(seed)

	// Find the requested card type
	var selectedCard CardInfo
	found := false

	for _, card := range cardTypes {
		if card.Type == cardType {
			selectedCard = card
			found = true
			break
		}
	}

	// If not found, default to Visa
	if !found {
		for _, card := range cardTypes {
			if card.Type == Visa {
				selectedCard = card
				break
			}
		}
	}

	// Select a random prefix
	prefix := selectedCard.Prefix[rng.Intn(len(selectedCard.Prefix))]

	// Select a random length
	length := selectedCard.Length[rng.Intn(len(selectedCard.Length))]

	// Generate the number
	return generateCreditCardWithPrefix(prefix, length, rng)
}

// GenerateRandomCreditCard generates a random valid credit card number of any type
func GenerateRandomCreditCard(seed int64) (string, CardType) {
	rng := NewRandomSource(seed)

	// Choose a random card type
	selectedCard := cardTypes[rng.Intn(len(cardTypes))]

	// Select a random prefix
	prefix := selectedCard.Prefix[rng.Intn(len(selectedCard.Prefix))]

	// Select a random length
	length := selectedCard.Length[rng.Intn(len(selectedCard.Length))]

	// Generate the number
	ccNum := generateCreditCardWithPrefix(prefix, length, rng)

	return ccNum, selectedCard.Type
}

// Helper function to generate a card with a specific prefix and length
func generateCreditCardWithPrefix(prefix string, length int, rng *rand.Rand) string {
	// Initialize with the prefix
	ccNum := prefix

	// Generate random digits until we're 1 short of the full length
	for len(ccNum) < length-1 {
		ccNum += fmt.Sprintf("%d", rng.Intn(10))
	}

	// Calculate the check digit (Luhn algorithm)
	checkDigit := calculateLuhnCheckDigit(ccNum)

	// Append the check digit
	return ccNum + fmt.Sprintf("%d", checkDigit)
}

// IdentifyCardType attempts to identify the card type from a card number
func IdentifyCardType(cardNumber string) CardType {
	// Remove spaces and dashes
	cardNumber = strings.ReplaceAll(cardNumber, " ", "")
	cardNumber = strings.ReplaceAll(cardNumber, "-", "")

	for _, card := range cardTypes {
		for _, prefix := range card.Prefix {
			if strings.HasPrefix(cardNumber, prefix) {
				for _, length := range card.Length {
					if len(cardNumber) == length {
						return card.Type
					}
				}
			}
		}
	}

	return Unknown
}

// IsValidCreditCard checks if a credit card number is valid using the Luhn algorithm
func IsValidCreditCard(cardNumber string) bool {
	// Remove spaces and dashes
	cardNumber = strings.ReplaceAll(cardNumber, " ", "")
	cardNumber = strings.ReplaceAll(cardNumber, "-", "")

	// Check if it contains only digits
	for _, char := range cardNumber {
		if char < '0' || char > '9' {
			return false
		}
	}

	// Check length
	if len(cardNumber) < 13 || len(cardNumber) > 19 {
		return false
	}

	// Luhn algorithm check
	sum := 0
	alternate := false

	for i := len(cardNumber) - 1; i >= 0; i-- {
		digit := int(cardNumber[i] - '0')

		if alternate {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}

// calculateLuhnCheckDigit calculates the check digit using the Luhn algorithm
func calculateLuhnCheckDigit(partialNumber string) int {
	// Double every second digit, from right to left
	sum := 0
	alternate := true

	for i := len(partialNumber) - 1; i >= 0; i-- {
		digit := int(partialNumber[i] - '0')

		if alternate {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		alternate = !alternate
	}

	// The check digit is what we need to add to make the sum divisible by 10
	return (10 - (sum % 10)) % 10
}

// FormatCreditCard formats a credit card number with spaces based on its type
func FormatCreditCard(cardNumber string) string {
	// Remove existing spaces and dashes
	cardNumber = strings.ReplaceAll(cardNumber, " ", "")
	cardNumber = strings.ReplaceAll(cardNumber, "-", "")

	cardType := IdentifyCardType(cardNumber)

	switch cardType {
	case AmericanExpress:
		if len(cardNumber) == 15 {
			return fmt.Sprintf("%s %s %s",
				cardNumber[:4],
				cardNumber[4:10],
				cardNumber[10:])
		}
	case DinersClub:
		if len(cardNumber) == 14 {
			return fmt.Sprintf("%s %s %s",
				cardNumber[:4],
				cardNumber[4:10],
				cardNumber[10:])
		}
	default:
		// Most cards use 4-digit grouping
		var formatted strings.Builder
		for i, char := range cardNumber {
			if i > 0 && i%4 == 0 {
				formatted.WriteRune(' ')
			}
			formatted.WriteRune(char)
		}
		return formatted.String()
	}

	// Default case for any unusual length
	return cardNumber
}
