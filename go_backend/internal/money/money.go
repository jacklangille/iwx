package money

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func ParseCents(value string) (int64, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, fmt.Errorf("price is required")
	}

	negative := false
	if strings.HasPrefix(trimmed, "-") {
		negative = true
		trimmed = strings.TrimPrefix(trimmed, "-")
	}

	parts := strings.Split(trimmed, ".")
	if len(parts) > 2 {
		return 0, fmt.Errorf("invalid decimal %q", value)
	}

	wholePart := parts[0]
	if wholePart == "" {
		wholePart = "0"
	}

	whole, err := strconv.ParseInt(wholePart, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid decimal %q", value)
	}

	fraction := int64(0)
	if len(parts) == 2 {
		fractionPart := parts[1]
		if len(fractionPart) > 2 {
			return 0, fmt.Errorf("price supports at most 2 decimal places")
		}

		fractionPart = fractionPart + strings.Repeat("0", 2-len(fractionPart))
		fraction, err = strconv.ParseInt(fractionPart, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid decimal %q", value)
		}
	}

	total := whole*100 + fraction
	if negative {
		total *= -1
	}

	return total, nil
}

func FormatCents(cents int64) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = int64(math.Abs(float64(cents)))
	}

	return fmt.Sprintf("%s%d.%02d", sign, cents/100, cents%100)
}

func Average(a, b int64) int64 {
	return (a + b) / 2
}

func Abs(value int64) int64 {
	if value < 0 {
		return -value
	}

	return value
}
