package webserver

import (
	"fmt"

	"github.com/t-margheim/wordle-pal/internal/pathreview"
)

type analyzeResponse struct {
	TargetLetters []string
	Guesses       []analyzeGuess
}

func newAnalyzeResponse(target string, pathResponse pathreview.PathResponse) analyzeResponse {
	var guesses []analyzeGuess
	for _, result := range pathResponse.GuessResults {
		var guess analyzeGuess
		for i, char := range result.Guess {
			guess.Guess[i] = analyzeGuessLetter{
				Value: string(char),
				Class: convertFeedbackToClass(result.Result[i]),
			}
		}

		guess.NewCount = result.NewWordCount
		guess.PrevCount = result.PreviousWordCount
		guess.Percentage = calculatePercentage(guess.PrevCount, guess.NewCount)

		guesses = append(guesses, guess)
	}

	var letters []string
	for _, char := range target {
		letters = append(letters, string(char))
	}
	return analyzeResponse{
		TargetLetters: letters,
		Guesses:       guesses,
	}
}

func convertFeedbackToClass(r pathreview.LetterFeedback) string {
	switch r {
	case pathreview.Correct:
		return "correct"
	case pathreview.NotInWord:
		return "wrong"
	case pathreview.WrongPosition:
		return "wrong_position"
	default:
		return "invalid"
	}
}

func calculatePercentage(before, after int) string {
	rawPct := (float64(before-after) / float64(before)) * 100
	return fmt.Sprintf("%.2f", rawPct)
}

type analyzeGuess struct {
	Guess      [5]analyzeGuessLetter
	PrevCount  int
	NewCount   int
	Percentage string
}

type analyzeGuessLetter struct {
	Value string
	Class string
}
