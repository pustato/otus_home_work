package hw03frequencyanalysis

import (
	"regexp"
	"sort"
	"strings"
)

var (
	trimRx = regexp.MustCompile(`^[^\p{L}]*([\p{L}-]+)[^\p{L}]*$`)
	wordRx = regexp.MustCompile(`^[\p{L}-]+$`)
)

const emptyWord = ""

type WordCounter struct {
	counter map[string]int
	words   []string
}

func NewWordCounter() WordCounter {
	wc := WordCounter{
		counter: make(map[string]int),
	}

	return wc
}

func (wc *WordCounter) Add(word string) {
	word = wc.normalizeWord(word)
	if word == emptyWord {
		return
	}

	if count, exists := wc.counter[word]; exists {
		wc.counter[word] = count + 1
	} else {
		wc.counter[word] = 1
		wc.words = append(wc.words, word)
	}
}

func (wc *WordCounter) GetSorted() []string {
	sort.Slice(wc.words, func(i, j int) bool {
		w1, w2 := wc.words[i], wc.words[j]
		c1, c2 := wc.counter[w1], wc.counter[w2]

		if c1 != c2 {
			return c1 > c2
		}

		return w1 < w2
	})

	return wc.words
}

func (wc *WordCounter) Count() int {
	return len(wc.words)
}

func (wc *WordCounter) normalizeWord(w string) string {
	w = strings.ToLower(w)
	w = trimRx.ReplaceAllString(w, `${1}`)
	w = strings.Trim(w, "-")

	if !wordRx.MatchString(w) {
		return emptyWord
	}

	return w
}

func Top10(input string) []string {
	wc := NewWordCounter()
	n := 10

	for _, w := range strings.Fields(input) {
		wc.Add(w)
	}

	if count := wc.Count(); count == 0 {
		return []string{}
	} else if count < n {
		n = count
	}

	return wc.GetSorted()[:n]
}
