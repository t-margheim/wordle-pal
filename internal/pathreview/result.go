package pathreview

type result [5]letterFeedback

func (r result) String() string {
	out := make([]byte, 5, 5)
	for i := 0; i < 5; i++ {
		switch r[i] {
		case notInWord:
			out[i] = 'X'
		case wrongPosition:
			out[i] = '@'
		case correct:
			out[i] = '$'
		default:
			out[i] = '?'
		}
	}
	return string(out)
}
