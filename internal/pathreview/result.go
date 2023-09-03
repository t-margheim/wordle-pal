package pathreview

type result [5]LetterFeedback

func (r result) String() string {
	out := make([]byte, 5, 5)
	for i := 0; i < 5; i++ {
		switch r[i] {
		case NotInWord:
			out[i] = 'X'
		case WrongPosition:
			out[i] = '@'
		case Correct:
			out[i] = '$'
		default:
			out[i] = '?'
		}
	}
	return string(out)
}
