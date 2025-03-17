package bot

import (
	"sync"
	"time"
)

type requestPool struct {
	mu        sync.Mutex
	cond      *sync.Cond
	windowEnd int64
	requests  int
}

func newRequestPool() *requestPool {
	q := &requestPool{}
	q.cond = sync.NewCond(&q.mu)
	return q
}

func (p *requestPool) queue() {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now().UnixMilli()

	// If the window has expired, reset the queue
	if now >= p.windowEnd {
		p.requests = 0
		p.windowEnd = now + 1000
	}

	// If the request cap (5) is reached, wait until the window expires
	for p.requests >= 5 {
		waitTime := p.windowEnd - now
		if waitTime > 0 {
			// Start a goroutine that will wake up all waiting goroutines when the window expires
			go p.waitForWindowReset(waitTime)
			p.cond.Wait() // Wait for the reset signal
		}

		// Re-check if the window has expired after waiting
		now = time.Now().UnixMilli()
		if now >= p.windowEnd {
			p.requests = 0
			p.windowEnd = now + 1000
		}
	}

	// Process the request
	p.requests++
	p.windowEnd += 1000 // Extend the window by 1s
}

// Waits until the window expires and then resets the queue
func (q *requestPool) waitForWindowReset(waitTime int64) {
	time.Sleep(time.Duration(waitTime) * time.Millisecond) // Wait for window to expire
	q.mu.Lock()
	defer q.mu.Unlock()
	now := time.Now().UnixMilli()
	if now >= q.windowEnd {
		q.requests = 0
		q.windowEnd = now + 1000
		q.cond.Broadcast()
	}
}
