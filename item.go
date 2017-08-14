package trending

import (
	"math"
	"time"
)

type item struct {
	eventSeries TimeSeries
	maxSeries   SlidingWindow

	max     float64
	maxTime time.Time
	options *options

	// TODO: move outside of item because it's the same for all items
	defaultExpectation float64
	defaultHourlyCount float64
}

func newItem(id string, options *options) *item {
	defaultHourlyCount := float64(options.baseCount) * float64(options.storageDuration/time.Hour)
	defaultExpectation := float64(options.baseCount) / float64(time.Hour/options.recentDuration)
	return &item{
		eventSeries: options.creator(id),
		maxSeries:   options.slidingWindowCreator(id),
		options:     options,

		defaultExpectation: defaultExpectation,
		defaultHourlyCount: defaultHourlyCount,
	}
}

func (i *item) score() score {
	recentCount, count := i.computeCounts()
	if recentCount < i.options.countThreshold {
		return score{}
	}
	if recentCount == count {
		// we see this for the first time so there is no historical data
		// use a sensible default like average/median over all items
		count = recentCount + i.defaultHourlyCount
	}
	probability := recentCount / count

	// order of those two lines is important.
	// if we insert before reading we might just get the same value.
	expectation := i.computeRecentMax()
	i.maxSeries.Insert(probability)

	if expectation == 0.0 {
		expectation = i.defaultExpectation
	}

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
	totalCount, _ := i.eventSeries.Range(now.Add(-i.options.storageDuration), now)
	count, _ := i.eventSeries.Range(now.Add(-i.options.recentDuration), now)
	return count, totalCount
}

func (i *item) computeRecentMax() float64 {
	return i.maxSeries.Max()
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
