package utils_test

import (
	"testing"

	"github.com/aryuuu/mines-party-server/utils"
)

func TestGenRandomString(t *testing.T) {
	testLength := 5
	randomString := utils.GenRandomString(testLength)

	if len(randomString) != testLength {
		t.Errorf("length of generated random string should be %d instead of %d", testLength, len(randomString))
	}

}

