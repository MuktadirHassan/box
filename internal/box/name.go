package box

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

var (
	adjectives = []string{"silver", "amber", "quiet", "swift"}
	nouns      = []string{"otter", "falcon", "badger", "maple"}
)

func GenerateName() (string, error) {
	adjective, err := randomWord(adjectives)
	if err != nil {
		return "", err
	}
	noun, err := randomWord(nouns)
	if err != nil {
		return "", err
	}

	return adjective + "-" + noun, nil
}

func randomWord(words []string) (string, error) {
	index, err := rand.Int(rand.Reader, big.NewInt(int64(len(words))))
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrNameGeneration, err)
	}

	return words[index.Int64()], nil
}
