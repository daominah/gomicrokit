package textproc

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"sort"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// Lists of characters (rune) group by types
var (
	NumericChars       = make(map[rune]bool)
	LowerAlphaChars    = make(map[rune]bool)
	UpperAlphaChars    = make(map[rune]bool)
	AlphaNumericChars  = make(map[rune]bool)
	AlphaNumericChars2 = make([]string, 0)
)

func runeToString(r rune) string {
	return fmt.Sprintf("%c", r)
}

func init() {
	lowerAlphas := "aàáãạảăắằẳẵặâấầẩẫậbcdđeèéẹẻẽêếềểễệfghiìíĩỉịjklmn" +
		"oòóõọỏôốồổỗộơớờởỡợpqrstuùúũụủưứừửữựvwxyýỳỵỷỹz"
	for _, char := range []rune(lowerAlphas) {
		upper := unicode.ToUpper(char)
		LowerAlphaChars[char], UpperAlphaChars[upper] = true, true
		AlphaNumericChars[char], AlphaNumericChars[upper] = true, true
		AlphaNumericChars2 = append(AlphaNumericChars2,
			runeToString(char), runeToString(upper))
	}
	numerics := "0123456789"
	for _, char := range []rune(numerics) {
		NumericChars[char] = true
		AlphaNumericChars[char] = true
		AlphaNumericChars2 = append(AlphaNumericChars2, runeToString(char))
	}
	sort.Strings(AlphaNumericChars2)
}

// TODO: RemoveRedundantSpace performance
func RemoveRedundantSpace(text string) string {
	// newline is special case, must the last filter
	spaces := []rune{'\t', '\v', '\f', '\r', ' ', 0x85, 0xA0, '\n'}
	for _, space := range spaces {
		spaceStr := fmt.Sprintf("%c", space)
		tokens := strings.Split(text, spaceStr)
		realTokens := make([]string, 0)
		for _, token := range tokens {
			isSpaceLine := true
			if token == "\n" {
				isSpaceLine = false
			} else {
				for _, char := range []rune(token) {
					if !unicode.IsSpace(char) {
						isSpaceLine = false
						break
					}
				}
			}

			if !isSpaceLine {
				realTokens = append(realTokens, token)
			}
		}
		text = strings.Join(realTokens, spaceStr)
	}
	return text
}

// HashTextToInt is a unique and fast hash func
func HashTextToInt(word string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(word))
	return int64(h.Sum64())
}

//
func GenRandomWord(minLen int, maxLen int) string {
	if minLen <= 0 {
		minLen = 0
	}
	if maxLen < minLen {
		maxLen = minLen
	}
	wordLen := minLen + rand.Intn(maxLen+1-minLen)
	chars := make([]string, wordLen)
	for i, _ := range chars {
		chars[i] = AlphaNumericChars2[rand.Intn(len(AlphaNumericChars2))]
	}
	word := strings.Join(chars, "")
	return word
}

func checkIsSeparator(r rune) bool {
	// u00a0 is non-breaking space character that is usually seen in HTML
	return r == ' ' || r == '\n' || r == '\t' || r == '\u00a0'
}

// TextToWords splits a text to list of words (punctuations removed)
func TextToWords(text string) []string {
	ret := make([]string, 0)
	wordsWithPun := strings.FieldsFunc(text, checkIsSeparator)
	for _, wordWP := range wordsWithPun {
		runes := []rune(wordWP)
		firstAlphaNumeric := -1
		lastAlphaNumeric := len(runes)
		for i, r := range runes {
			if AlphaNumericChars[r] {
				firstAlphaNumeric = i
				break
			}
		}
		for i := len(runes) - 1; i >= 0; i-- {
			if AlphaNumericChars[runes[i]] {
				lastAlphaNumeric = i
				break
			}
		}
		word := ""
		if firstAlphaNumeric != -1 && lastAlphaNumeric != len(runes) {
			word = string(runes[firstAlphaNumeric : lastAlphaNumeric+1])
		}
		if word != "" {
			ret = append(ret, word)
		}
	}
	return ret
}

// WordsToNGrams creates a set of n-gram from input words,
// (A n-gram is a contiguous sequence of n words)
func WordsToNGrams(words []string, n int) map[string]int {
	result := make(map[string]int)
	for i := 0; i < len(words)-n+1; i++ {
		nGram := strings.Join(words[i:i+n], " ")
		result[nGram] += 1
	}
	return result
}

// TextToNGrams creates a set of n-gram from lowered input text
func TextToNGrams(text string, n int) map[string]int {
	text = strings.ToLower(text)
	words := TextToWords(text)
	return WordsToNGrams(words, n)
}

// There are often several ways to represent the same string. For example,
// an "é" (e-acute) can be represented in a string as a single rune ("\u00e9") or
// an "e" followed by an acute accent ("e\u0301").
// They should be treated as equal in text processing.
// Vietnamese text has an extra problem: diacritic position,
// example: old style: òa, óa, ỏa, õa, ọa; new style: oà, oá, oả, oã, oạ
func NormalizeText(text string) string {
	bs := norm.NFKC.Bytes([]byte(text))
	result := string(bs)
	// TODO: Vietnamese diacritic position
	return result
}
