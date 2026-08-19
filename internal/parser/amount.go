package parser

import (
	"regexp"
	"strconv"
	"strings"
)

var amountPatternIDR = regexp.MustCompile(`(?i)\b(\d+(?:[.,]\d+)?)(rb|ribu|k|jt|juta)?\b`)

func ParseAmountAndDescription(text string, currency string) (int64, string, error) {
	var matches [][]int

	if currency == "IDR" {
		matches = amountPatternIDR.FindAllStringSubmatchIndex(text, -1)
	}
	
	if matches == nil {
		return 0, "", ErrAmountNotFound
	}

	match := matches[len(matches)-1]
	// fmt.Println(match)
	rawAmount := text[match[2]:match[3]]

	suffix := ""
	if match[4] > 0 && match[5] > 0 {
		suffix = text[match[4]:match[5]]
	}

	amount, err := normalizeAmount(rawAmount, suffix)
	if err != nil {
		return 0, "", err
	}

	description := text[:match[0]] + text[match[1]:]
	description = cleanupDescription(description)
	if description == "" {
		description = "transaction"
	}
	
	return amount, description, nil
}

func normalizeAmount(raw string, suffix string) (int64, error) {
	value := strings.ReplaceAll(raw, ".", "")
	value = strings.ReplaceAll(value, ",", "")

	amount, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, err
	}

	switch strings.ToLower(suffix) {
	case "rb", "k", "ribu":
		amount *= 1000
	case "jt", "juta":
		amount *= 1000000
	}

	if amount <= 0 {
		return 0, ErrInvalidAmount
	}

	return amount, nil
}

func cleanupDescription(text string) string {
	description := strings.TrimSpace(text)
	description = strings.Join(strings.Fields(description), " ")
	return description
}