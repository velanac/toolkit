package toolkit

import "crypto/rand"

const radndomStringSource = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_+"

// Tools is the type use to instantiate the modules.
// Any variable of this type can be used to access the modules.
// to all the methods with the reciver *Tools.
type Tools struct{}

// RandomString generates a random string of the specified length using a cryptographically secure random number generator.
func (t *Tools) RandomString(length int) string {
	s, r := make([]rune, length), []rune(radndomStringSource)

	for i := range s {
		p, _ := rand.Prime(rand.Reader, len(r))
		x, y := p.Uint64(), uint64(len(r))
		s[i] = r[x%y]
	}

	return string(s)
}
