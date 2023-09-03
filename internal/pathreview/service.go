package pathreview

import (
	"bytes"
	"log/slog"
	"slices"
	"strings"
)

type PathRequest struct {
	Target string   `json:"target_word"`
	Path   []string `json:"path"`
}

type PathResponse struct {
	GuessResults []guessResult
}
type guessResult struct {
	Guess             string `json:"guess"`
	Result            result `json:"result"`
	PreviousWordCount int    `json:"previous_count"`
	NewWordCount      int    `json:"new_count"`
}

type Servicer interface {
	ReviewPath(req PathRequest) (PathResponse, error)
}

type Service struct{}

func (s *Service) ReviewPath(req PathRequest) (PathResponse, error) {
	wordlist := words
	var resp PathResponse
	for _, guess := range req.Path {
		if guess == req.Target {
			break
		}

		result := scoreWord(guess, req.Target)

		preCount := len(wordlist)
		wordlist = filterWordList(guess, result, wordlist)
		postCount := len(wordlist)

		resp.GuessResults = append(resp.GuessResults, guessResult{
			Guess:             guess,
			Result:            result,
			PreviousWordCount: preCount,
			NewWordCount:      postCount,
		})
	}
	return resp, nil
}

func scoreWord(guess, target string) result {
	guessAttr, targetAttr := slog.String("guess", guess), slog.String("target", target)
	slog.Debug("scoring word",
		guessAttr,
		targetAttr,
	)
	var res result
	for i := 0; i < 5; i++ {
		guessLetter, targetLetter := string(guess[i]), string(target[i])
		guessLetterAttr := slog.String("guess_letter", guessLetter)
		targetLetterAttr := slog.String("target_letter", targetLetter)
		posAttr := slog.Int("position", i)

		if guessLetter == targetLetter {
			slog.Debug("letter is correct",
				guessLetterAttr,
				targetLetterAttr,
				posAttr,
				guessAttr,
				targetAttr,
			)
			res[i] = Correct
			continue
		}

		if strings.Contains(target, guessLetter) {
			slog.Debug("letter is wrong position",
				guessLetterAttr,
				targetLetterAttr,
				posAttr,
				guessAttr,
				targetAttr,
			)
			res[i] = WrongPosition
			continue
		}
		slog.Debug("letter is not in word",
			guessLetterAttr,
			targetLetterAttr,
			posAttr,
			guessAttr,
			targetAttr,
		)
		res[i] = NotInWord
	}
	slog.Debug("ScoreWord finished",
		guessAttr,
		targetAttr,
		slog.String("result", res.String()),
	)
	return res
}

func filterWordList(guess string, res result, wordlist []string) []string {
	var newList []string
	mustMatch := map[int]byte{}
	mustContainButNotMatch := map[byte]int{}
	mustNotContain := []byte{}
	for i, resChar := range res {
		guessLetter := guess[i]
		switch resChar {
		case Correct:
			mustMatch[i] = guessLetter
		case WrongPosition:
			mustContainButNotMatch[guessLetter] = i
		case NotInWord:
			mustNotContain = append(mustNotContain, guessLetter)
		}
	}

	slog.Debug("filtering rules prepared",
		"mustMatch", mustMatch,
		"mustContainButNotMatch", mustContainButNotMatch,
		"mustNotContain", mustNotContain,
	)

	oldCount := len(wordlist)

WordLoop:
	for _, word := range wordlist {
		// check correct letters first
		for idx, char := range mustMatch {
			// if we are missing a correct letter, skip this word
			if word[idx] != char {
				continue WordLoop
			}
		}

		// check that letters which were in the incorrect position
		// on the previous guess are in the word in a different position now
		for char, idx := range mustContainButNotMatch {
			if !bytes.Contains([]byte(word), []byte{char}) || word[idx] == char {
				continue WordLoop
			}
		}

		// check that the word does not contain any of the letters that we know
		// are not in the target word
		for _, char := range mustNotContain {
			if slices.Contains([]byte(word), char) {
				continue WordLoop
			}
		}

		newList = append(newList, word)

	}
	newCount := len(newList)
	slog.Debug("filtering finished",
		slog.Int("old_count", oldCount),
		slog.Int("new_count", newCount),
	)
	return newList
}
