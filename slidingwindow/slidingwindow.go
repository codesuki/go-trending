package slidingwindow

import "time"

// Clock specifies the needed time related functions used by the time series.
// To use a custom clock implement the interface and pass it to the time series constructor.
// The default clock uses time.Now()
type Clock interface {
	Now() time.Time
}

// defaultClock is used in case no clock is provided to the constructor.
type defaultClock struct{}

func (c *defaultClock) Now() time.Time {
	return time.Now()
}

type slidingWindow struct {
	buffer []float64
	length int

	end   time.Time
	start time.Time

	oldest int
	newest int

	step     time.Duration
	duration time.Duration

	clock Clock
}

var defaultStep = time.Hour * 24
var defaultDuration = time.Hour * 24 * 7

type options struct {
	clock Clock

	step     time.Duration
	duration time.Duration
}

type option func(*options)

func WithStep(step time.Duration) option {
	return func(o *options) {
		o.step = step
	}
}

func WithDuration(duration time.Duration) option {
	return func(o *options) {
		o.duration = duration
	}
}

func WithClock(clock Clock) option {
	return func(o *options) {
		o.clock = clock
	}
}

func NewSlidingWindow(os ...option) *slidingWindow {
	opts := options{}
	for _, o := range os {
		o(&opts)
	}
	if opts.clock == nil {
		opts.clock = &defaultClock{}
	}
	if opts.step.Nanoseconds() == 0 {
		opts.step = defaultStep
	}
	if opts.duration.Nanoseconds() == 0 {
		opts.duration = defaultDuration
	}
	return newSlidingWindow(opts.step, opts.duration, opts.clock)
}

func newSlidingWindow(step time.Duration, duration time.Duration, clock Clock) *slidingWindow {
	length := int(duration / step)
	now := clock.Now()
	return &slidingWindow{
		buffer:   make([]float64, length),
		length:   length,
		end:      now.Truncate(step).Add(-duration),
		start:    now,
		step:     step,
		duration: duration,
		oldest:   1,
		clock:    clock,
	}
}

func (sw *slidingWindow) Insert(score float64) {
	sw.advance()
	if score > sw.buffer[sw.newest] {
		sw.buffer[sw.newest] = score
	}
}

func (sw *slidingWindow) Max() float64 {
	sw.advance()
	max := 0.0
	for i := range sw.buffer {
		if sw.buffer[i] > max {
			max = sw.buffer[i]
		}
	}
	return max
}

func (sw *slidingWindow) advance() {
	newEnd := sw.clock.Now().Truncate(sw.step).Add(-sw.duration)
	for newEnd.After(sw.end) {
		sw.end = sw.end.Add(sw.step)
		sw.buffer[sw.oldest] = 0.0
		sw.newest = sw.oldest
		sw.oldest = (sw.oldest + 1) % sw.length
	}
}
