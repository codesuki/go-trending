package trending

import (
	"log"
	"math"
	"time"
)

type item struct {
	timeSeries TimeSeries
	max        float64
	maxTime    time.Time
	options    *options
}

func (i *item) score(recentTotal float64, total float64) score {
	recentCount, count := i.computeCounts()
	probability := recentCount / recentTotal
	expectation := count / total
	klScore := computeKullbackLeibler(probability, expectation)
	if klScore > i.max {
		i.updateMax(klScore)
	}
	i.decayMax()
	mixedScore := 0.5 * (klScore + i.max)
	return score{
		Score:       mixedScore,
		Probability: probability,
		Expectation: expectation,
		Maximum:     i.max,
		KLScore:     klScore,
	}
}

func (i *item) computeCounts() (float64, float64) {
	now := time.Now()
	count, _ := i.timeSeries.Range(now.Add(-i.options.recentDuration), now)
	totalCount, _ := i.timeSeries.Range(now.Add(-i.options.storageDuration), now)
	if count == totalCount {
		log.Println("count and totalCount are the same")
		return count, totalCount
	}
	return count, totalCount - count
}

func (i *item) decayMax() {
	i.updateMax(i.max * i.computeExponentialDecayMultiplier())
}

func (i *item) updateMax(score float64) {
	i.max = score
	i.maxTime = time.Now()
}

func (i *item) computeExponentialDecayMultiplier() float64 {
	return math.Pow(0.5, float64(time.Now().Unix()-i.maxTime.Unix())/i.options.halfLife.Seconds())
}

func computeKullbackLeibler(probability float64, expectation float64) float64 {
	if probability == 0.0 {
		return 0.0
	}
	return probability * math.Log(probability/expectation)
}
