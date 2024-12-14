package util

import (
	"math/rand"
	"strings"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

func init() {
	source := rand.NewSource(time.Now().UnixNano())
	rand.New(source)
}

// RandomInt generates a ramdom integer between  min and max
func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

// RandomString generates a random string of length n
func RandomString(n int) string {
	// Create a string builder to be more efficient
	var sb strings.Builder
	// Get the length of the alphabet
	k := len(alphabet)
	// Loop n times and add a random character
	for i := 0; i < n; i++ {
		// Get a random index of the alphabet and add it to the string builder
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}
	// Return the string
	return sb.String()
}

// RandomOwner generates a random owner
func RandomOwner() string {
	return RandomString(6)
}

// RandomMoney generates a random amount of money
func RandomMoney() int64 {
	return RandomInt(0, 1000)
}

// RandomCurrency generates a random currency code
func RandomCurrency() string {
	currencies := []string{USD, EUR, VND}
	n := len(currencies)
	return currencies[rand.Intn(n)]
}

func RandomEmail() string { return RandomString(6) + "@gmail.com" } // RandomEmail
