package pathreview

type letterFeedback int

const (
	notInWord letterFeedback = iota + 1
	wrongPosition
	correct
)
