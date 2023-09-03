package pathreview

type LetterFeedback int

const (
	NotInWord LetterFeedback = iota + 1
	WrongPosition
	Correct
)
