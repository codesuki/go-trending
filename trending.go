package trending

import (
	"log"
	"sort"
	"time"

	"github.com/codesuki/go-time-series"
)

// Algorithm:
// 1. Divide one week into 5 minutes bins
// The algorithm uses expected probability to compute its ranking.
// By choosing a one week span to compute the expectation the algorithm will forget old trends.
// 2. For every play event increase the counter in the current bin
// 3. Compute the KL Divergence with the following steps
//    - Compute the probability of the last full bin (this should be the current 5 minutes sliding window)
//    - Compute the expected probability over the past bins (excluding the last full bin and current bin)
//    - Compute KL Divergence (kld = p * ln(p/e))
// 4. Keep the highest KL Divergence score together with its timestamp
// 5. Compute exponential decay multiplier and multiply with highest KL Divergence
// 6. Blend current KL Divergence score with decayed high score

var defaultHalfLife = 2 * time.Hour
var defaultRecentDuration = 1 * time.Minute
var defaultStorageDuration = 7 * 24 * time.Hour
var defaultMaxResults = 10

type options struct {
	creator         TimeSeriesCreator
	halfLife        time.Duration
	recentDuration  time.Duration
	storageDuration time.Duration
	maxResults      int
}

type Option func(*options)

func WithTimeSeries(creator TimeSeriesCreator) Option {
	return func(o *options) {
		o.creator = creator
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

type Scorer struct {
	options options
	items   map[string]*item
	total   TimeSeries
}

type TimeSeries interface {
	IncreaseAtTime(amount int, insertTime time.Time)
	Range(start, end time.Time) (float64, error)
}

type TimeSeriesCreator func(string) TimeSeries

func NewMemoryTimeSeries(id string) TimeSeries {
	ts, _ := timeseries.NewTimeSeries(timeseries.WithGranularities(
		[]timeseries.Granularity{
			{Granularity: time.Second, Count: 60},
			{Granularity: time.Minute, Count: 60},
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
	scorer.total = scorer.options.creator("total")
	return scorer
}

func (s *Scorer) AddEvent(id string, time time.Time) {
	item := s.items[id]
	if item == nil {
		item = s.createItem(id)
		s.items[id] = item
	}
	s.addToTotal(time)
	s.addToItem(item, time)
}

func (s *Scorer) createItem(id string) *item {
	return &item{timeSeries: s.options.creator(id), options: &s.options}
}

func (s *Scorer) addToTotal(time time.Time) {
	s.total.IncreaseAtTime(1, time)
}

func (s *Scorer) addToItem(item *item, time time.Time) {
	item.timeSeries.IncreaseAtTime(1, time)
}

func (s *Scorer) Score() Scores {
	var scores Scores
	for id, item := range s.items {
		score := s.scoreOne(item)
		score.ID = id
		scores = append(scores, score)
	}
	sort.Sort(scores)
	return scores.take(s.options.maxResults)
}

func (s *Scorer) scoreOne(item *item) score {
	recentTotal, total := s.computeTotals()
	return item.score(recentTotal, total)
}

func (s *Scorer) computeTotals() (float64, float64) {
	now := time.Now()
	recentTotal, _ := s.total.Range(now.Add(-s.options.recentDuration), now)
	if recentTotal == 0 {
		recentTotal = 1
	}
	total, _ := s.total.Range(now.Add(-s.options.storageDuration), now)
	if total == 0 {
		total = 1
	}
	if recentTotal == total {
		log.Println("recentTotal and total are the same")
		return recentTotal, total
	}
	return recentTotal, total - recentTotal
}
