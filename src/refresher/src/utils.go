package main

import "unicode"

func ContainsNum(StringForCheck string) bool {
	for _, rune := range StringForCheck {
		if unicode.IsNumber(rune) {
			return true
		} else {
			continue
		}
	}

	return false
}
