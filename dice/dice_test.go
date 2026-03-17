package dice

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenizeDiceRequest(t *testing.T) {
	input := "1d6 + 4d8"
	output := TokenizeDiceRequest(input)

	assert.Equal(t, 3, len(output), "should have 3 tokens")

	assert.Equal(t, "1d6", output[0], "should be 1d6")

	assert.Equal(t, "+", output[1], "should be +")

	assert.Equal(t, "4d8", output[2], "should be 4d8")

}

func TestParseTokenRequestBasicMath(t *testing.T) {
	input := "1 + 2"
	output, err := ParseDiceRequest(input)
	assert.Nil(t, err, "Shoudn't of errored")

	assert.Equal(t, 3, output, "output should be 3")

	input = "1 - 2"
	output, err = ParseDiceRequest(input)
	assert.Nil(t, err, "Shoudn't of errored")

	assert.Equal(t, -1, output, "output should be 1")
}

func TestParseTokenRequestRandomRolls(t *testing.T) {
	input := "1d6"
	output, err := ParseDiceRequest(input)
	assert.Nil(t, err, "Shoudn't of errored")

	assert.NotZero(t, output, "output shouldn't be 0")

	input = "1d6 - 2"
	_, err = ParseDiceRequest(input)
	assert.Nil(t, err, "Shoudn't of errored")

}

func TestParseTokenRequestErrors(t *testing.T) {
	input := "d6"
	_, err := ParseDiceRequest(input)
	assert.EqualError(t, err, "invalid count", "should of given invalid count error")

	input = "1d"
	_, err = ParseDiceRequest(input)
	assert.EqualError(t, err, "invalid sides", "should of given invalid sides error")

	input = "1a2"
	_, err = ParseDiceRequest(input)
	assert.EqualError(t, err, "invalid value", "should of given invalid value error")

	input = "1d2d2"
	_, err = ParseDiceRequest(input)
	assert.EqualError(t, err, "invalid format", "should of given invalid format error")

	input = "1d20 * 4"
	_, err = ParseDiceRequest(input)
	assert.EqualError(t, err, "unimplemented operation", "should of given unimplemented operation error")
}
