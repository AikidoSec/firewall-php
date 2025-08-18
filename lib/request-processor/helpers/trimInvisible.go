package helpers

import (
	"strings"
	"unicode"
)

// TrimInvisible trims leading/trailing whitespace, control, and specific
// invisible/format runes similar to your PHP StringHelper::trim.
func TrimInvisible(s string) string {
	return strings.TrimFunc(s, isTrimRune)
}

func isTrimRune(r rune) bool {
	// 1) All Unicxde whitespace
	if unicode.IsSpace(r) {
		return true
	}
	// 2) Control characters (Cc)
	if unicode.IsControl(r) {
		return true
	}
	// 3) Specific invisible/format chars from your PHP constant
	switch r {
	// Soft hyphen, combining grapheme joiner, Arabic mark, Hangul fillers, Khmer indep. vowels, Mongolian, etc.
	case 0x00AD, 0x034F, 0x061C, 0x115F, 0x1160, 0x17B4, 0x17B5, 0x180E,
		// General Punctuation & Format (ZWSP, ZWNJ, ZWJ, LRM/RLM, NNBSP, Word Joiner, etc.)
		0x2000, 0x2001, 0x2002, 0x2003, 0x2004, 0x2005, 0x2006, 0x2007, 0x2008, 0x2009, 0x200A,
		0x200B, 0x200C, 0x200D, 0x200E, 0x200F, 0x202F, 0x205F, 0x2060, 0x2061, 0x2062, 0x2063, 0x2064,
		0x2065, 0x206A, 0x206B, 0x206C, 0x206D, 0x206E, 0x206F,
		// Braille blank, Ideographic space, Hangul Filler, BOM, NBSP variants
		0x2800, 0x3000, 0x3164, 0xFEFF, 0xFFA0,
		// Musical symbol format-ish
		0x1D159, 0x1D173, 0x1D174, 0x1D175, 0x1D176, 0x1D177, 0x1D178, 0x1D179, 0x1D17A,
		// Tag Space (Plane 14)
		0xE0020:
		return true
	}
	return false
}
