# go-trending
[![License](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](./LICENSE)
[![GoDoc](https://godoc.org/github.com/codesuki/go-trending?status.svg)](https://godoc.org/github.com/codesuki/go-trending)
[![Build Status](http://img.shields.io/travis/codesuki/go-trending.svg?style=flat)](https://travis-ci.org/codesuki/go-trending)
[![codecov](https://codecov.io/gh/codesuki/go-trending/branch/master/graph/badge.svg)](https://codecov.io/gh/codesuki/go-trending)

Trending algorithm based on the article [Trending at Instagram](http://instagram-engineering.tumblr.com/post/122961624217/trending-at-instagram). To detect trends an items current behavior is compared to its usual behavior. The more it differes the higher / lower the score. Items will start trending if the current usage is higher than its average usage. To avoid items quickly become non-trending again the scores are smoothed.

* Configurable and simple to use
* Use your own clock implementation, e.g. for testing or similar
* Use any time series implementation as backend that implements the TimeSeries interface

### Details
Uses a [time series](https://www.github.com/codesuki/go-time-series) for each item to keep track of its past behavior and get recent behavior with small granularity. Computes the [Kullback-Leibler divergence](https://en.wikipedia.org/wiki/Kullback%E2%80%93Leibler_divergence) between recent behavior and expected, i.e. past, bahavior. Then blends the current item score with its past [decayed](https://en.wikipedia.org/wiki/Exponential_decay) maximum score to get the final score.

## Examples

### Creating a default scorer
```go
import "github.com/codesuki/go-trending"

...

scorer := trending.NewScorer()
```

### Creating a customized scorer
**Parameters**
* **Time series:** is used for creating the backing `TimeSeries` objects
* **Half-life:** controls how long an item is trending after the activity went back to normal.
* **Recent duration:** controls how much data is used to compute the current state. If there is not much activity try looking at larger duration.
* **Storage duration:** controls how much historical data is used. Trends older than the storage duration won't have any effect on the computation. The time series in use should have at least as much storage duration as specified here.

```go
import "github.com/codesuki/go-trending"

...
func NewTimeSeries(id string) TimeSeries {
    // create time series that satisfies the TimeSeries interface
    return timeSeries
}

...

scorer := trending.NewScorer(
    WithTimeSeries(NewTimeSeries),
    WithHalflife(time.Hour),
    WithRecentDuration(time.Minute),
    WithStorageDuration(7 * 24 * time.Hour),
)
```


### Using the scorer
```go
import "github.com/codesuki/go-trending"

...

scorer := trending.NewScorer()

scorer.AddEvent("id", time)
// add more events. maybe using an event stream.

...

trendingItems := scorer.Score()
```

## Documentation
GoDoc is located [here](https://godoc.org/github.com/codesuki/go-trending)

## License
go-trending is [MIT licensed](./LICENSE).
