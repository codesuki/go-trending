package slidingwindow

import (
	"fmt"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
)

func TestSlidingWindow(t *testing.T) {
	c := clock.NewMock()
	startTime, _ := time.Parse(time.RFC3339, "2017-01-14T12:00:00Z")
	c.Set(startTime)
	sw := NewSlidingWindow(24*time.Hour, 7*24*time.Hour, c)

	sw.Insert(1.0)
	sw.Insert(2.0)
	fmt.Println(sw.Max())

	c.Add(24 * time.Hour)
	sw.Insert(1.2)

	c.Add(24 * time.Hour)
	sw.Insert(1.3)

	c.Add(24 * time.Hour)
	sw.Insert(1.4)

	c.Add(24 * time.Hour)
	sw.Insert(1.5)

	c.Add(24 * time.Hour)
	sw.Insert(1.6)

	c.Add(24 * time.Hour)
	sw.Insert(1.7)

	c.Add(24 * time.Hour)
	sw.Insert(1.8)

	c.Add(24 * time.Hour)
	sw.Insert(1.9)
}
