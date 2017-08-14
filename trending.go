package trending

import (
	"sort"
	"time"

	timeseries "github.com/codesuki/go-time-series"
	"github.com/codesuki/go-trending/slidingwindow"
)

// Algorithm:
// 1. Divide one week into 5 minutes bins
// The algorithm uses expected probability to compute its ranking.
// By choosing a one week span to compute the expectation the algorithm will forget old trends.
// 2. For every play event increase the counter in the current bin
// 3. Compute the KL Divergence with the following steps
//    - Compute the probability of the last full bin (this should be the current 5 minutes sliding window)
//    - Compute the expected probability over the past bins including the current bin
//    - Compute KL Divergence (kld = p * ln(p/e))
// 4. Keep the highest KL Divergence score together with its timestamp
// 5. Compute exponential decay multiplier and multiply with highest KL Divergence
// 6. Blend current KL Divergence score with decayed high score

var defaultHalfLife = 2 * time.Hour
var defaultRecentDuration = 5 * time.Minute
var defaultStorageDuration = 7 * 24 * time.Hour
var defaultMaxResults = 100
var defaultBaseCount = 3
var defaultScoreThreshold = 0.01
var defaultCountThreshold = 3.0

type options struct {
	creator              TimeSeriesCreator
	slidingWindowCreator SlidingWindowCreator

	halfLife time.Duration

	recentDuration  time.Duration
	storageDuration time.Duration

	maxResults     int
	baseCount      int
	scoreThreshold float64
	countThreshold float64
}

type Option func(*options)

func WithTimeSeries(creator TimeSeriesCreator) Option {
	return func(o *options) {
		o.creator = creator
	}
}

func WithSlidingWindow(creator SlidingWindowCreator) Option {
	return func(o *options) {
		o.slidingWindowCreator = creator
	}
}

func WithHalfLife(halfLife time.Duration) Option {
	return func(o *options) {
		o.halfLife = halfLife
	}
}

func WithRecentDuration(recentDuration time.Duration) Option {
	return func(o *options) {
		o.recentDuration = recentDuration
	}
}

func WithStorageDuration(storageDuration time.Duration) Option {
	return func(o *options) {
		o.storageDuration = storageDuration
	}
}

func WithMaxResults(maxResults int) Option {
	return func(o *options) {
		o.maxResults = maxResults
	}
}

func WithScoreThreshold(threshold float64) Option {
	return func(o *options) {
		o.scoreThreshold = threshold
	}
}

func WithCountThreshold(threshold float64) Option {
	return func(o *options) {
		o.countThreshold = threshold
	}
}

type Scorer struct {
	options options
	items   map[string]*item
}

type SlidingWindow interface {
	Insert(score float64)
	Max() float64
}

type SlidingWindowCreator func(string) SlidingWindow

type TimeSeries interface {
	IncreaseAtTime(amount int, time time.Time)
	Range(start, end time.Time) (float64, error)
}

type TimeSeriesCreator func(string) TimeSeries

func NewMemoryTimeSeries(id string) TimeSeries {
	ts, _ := timeseries.NewTimeSeries(timeseries.WithGranularities(
		[]timeseries.Granularity{
			{Granularity: time.Second, Count: 60},
			{Granularity: time.Minute, Count: 10},
			{Granularity: time.Hour, Count: 24},
			{Granularity: time.Hour * 24, Count: 7},
		},
	))
	return ts
}

func NewScorer(options ...Option) Scorer {
	scorer := Scorer{items: make(map[string]*item)}
	for _, o := range options {
		o(&scorer.options)
	}
	if scorer.options.creator == nil {
		scorer.options.creator = NewMemoryTimeSeries
	}
	if scorer.options.halfLife == 0 {
		scorer.options.halfLife = defaultHalfLife
	}
	if scorer.options.recentDuration == 0 {
		scorer.options.recentDuration = defaultRecentDuration
	}
	if scorer.options.storageDuration == 0 {
		scorer.options.storageDuration = defaultStorageDuration
	}
	if scorer.options.maxResults == 0 {
		scorer.options.maxResults = defaultMaxResults
	}
	if scorer.options.scoreThreshold == 0.0 {
		scorer.options.scoreThreshold = defaultScoreThreshold
	}
	if scorer.options.countThreshold == 0.0 {
		scorer.options.countThreshold = defaultCountThreshold
	}
	if scorer.options.baseCount == 0.0 {
		scorer.options.baseCount = defaultBaseCount
	}
	if scorer.options.slidingWindowCreator == nil {
		scorer.options.slidingWindowCreator = func(id string) SlidingWindow {
			return slidingwindow.NewSlidingWindow(
				slidingwindow.WithStep(time.Hour*24),
				slidingwindow.WithDuration(scorer.options.storageDuration),
			)
		}
	}
	return scorer
}

func (s *Scorer) AddEvent(id string, time time.Time) {
	item := s.items[id]
	if item == nil {
		item = newItem(id, &s.options)
		s.items[id] = item
	}
	s.addToItem(item, time)
}

func (s *Scorer) addToItem(item *item, time time.Time) {
	item.eventSeries.IncreaseAtTime(1, time)
}

func (s *Scorer) Score() Scores {
	var scores Scores
	for id, item := range s.items {
		score := item.score()
		score.ID = id
		scores = append(scores, score)
	}
	sort.Sort(scores)
	if s.options.scoreThreshold > 0 {
		scores = scores.threshold(s.options.scoreThreshold)
	}
	return scores.take(s.options.maxResults)
}
