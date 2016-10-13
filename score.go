package trending

type score struct {
	ID          string
	Score       float64
	Probability float64
	Expectation float64
	Maximum     float64
	KLScore     float64
}

type Scores []score

func (s Scores) Len() int {
	return len(s)
}

func (s Scores) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s Scores) Less(i, j int) bool {
	return s[i].Score > s[j].Score
}

func (s Scores) take(count int) Scores {
	if count >= len(s) {
		return s
	}
	return s[0 : count-1]
}
