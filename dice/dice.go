package dice

import (
	"errors"
	"math/rand"
	"strconv"
	"strings"
)

type Dice struct {
	Result    int
	Breakdown string
}

func ParseDiceRequest(request string) (*Dice, error) {
	result := 0
	expandedResult := ""

	reqTokens := TokenizeDiceRequest(request)

	operation := ""
	for _, v := range reqTokens {
		if v == "\\" || v == "*" {
			return nil, errors.New("unimplemented operation")
		}

		//Current token isn't a operator.
		if v != "+" && v != "-" {
			roll, err := Roll(v)
			if err != nil {
				return nil, err
			}

			switch operation {
			case "+":
				result += roll
			case "-":
				result -= roll
			case "":
				result = roll
			}

			expandedResult += operation + " " + v + "(" + strconv.Itoa(roll) + ")"

		} else {
			//Set the operation to apply to the next incoming token
			operation = v
		}
	}
	d := &Dice{Result: result, Breakdown: strings.TrimSpace(expandedResult)}
	return d, nil
}

func TokenizeDiceRequest(request string) []string {
	tokens := make([]string, 0)

	cToken := ""
	for _, v := range request {
		if v == '+' || v == '-' || v == '*' || v == '\\' {
			tokens = append(tokens, cToken)
			tokens = append(tokens, string(v))
			cToken = ""
		} else {
			if v != ' ' {
				cToken += string(v)
			}
		}
	}

	if len(cToken) > 0 {
		tokens = append(tokens, cToken)
	}

	return tokens
}

func Roll(dice string) (int, error) {
	result := 0
	dice = strings.ToLower(dice)
	if strings.Contains(dice, "d") {
		split := strings.Split(dice, "d")
		if len(split) == 2 {
			count, err := strconv.Atoi(split[0])
			if err != nil {
				return 0, errors.New("invalid count")
			}

			sides, err := strconv.Atoi(split[1])
			if err != nil {
				return 0, errors.New("invalid sides")
			}
			for i := 0; i < count; i++ {
				result += rand.Intn(sides) + 1
			}
		} else {
			return 0, errors.New("invalid format")
		}
	} else {
		value, err := strconv.Atoi(dice)
		if err != nil {
			return 0, errors.New("invalid value")
		}
		result = value
	}

	return result, nil
}
