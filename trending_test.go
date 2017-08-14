package trending

import (
	"fmt"
	"testing"
	"time"
)

func TestNewScorer(t *testing.T) {

}

func TestAddEvent(t *testing.T) {

}

func TestScore(t *testing.T) {
	// check that current count is not used in total

}

func TestMemory(t *testing.T) {
	scorer := NewScorer()

	now := time.Now()
	for i := 0; i < 1000000; i++ {
		scorer.AddEvent(fmt.Sprintf("%d", i), now)
	}

	fmt.Println("finished")

	start := time.Now()
	scorer.Score()
	fmt.Println("took", time.Since(start))

	time.Sleep(1000 * time.Second)
}
