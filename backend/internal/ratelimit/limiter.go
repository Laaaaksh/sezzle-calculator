// Package ratelimit is a small in-memory token bucket limiter keyed by client.
package ratelimit

import (
	"math"
	"sync"
	"time"
)

// Clock returns the current time. Tests pass a fake one so refill needs no sleeping.
type Clock func() time.Time

type bucket struct {
	tokens float64
	last   time.Time
}

// Limiter gives each key a bucket that holds up to burst tokens and refills at rate tokens per second. A request spends
// one token. A bucket refilled to full is the same as no bucket, so Allow sweeps those away and memory stays bounded.
// State is per process: several instances need a shared store such as Redis.
type Limiter struct {
	mu        sync.Mutex
	rate      float64
	burst     float64
	now       Clock
	buckets   map[string]*bucket
	nextSweep time.Time
}

// New returns a limiter for rate tokens per second with room for burst tokens.
// The caller validates both: rate and burst must be positive.
func New(rate float64, burst int, now Clock) *Limiter {
	return &Limiter{
		rate:    rate,
		burst:   float64(burst),
		now:     now,
		buckets: make(map[string]*bucket),
	}
}

// Allow spends one token for key. When none is left it returns false and how
// long until one is available.
func (l *Limiter) Allow(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	l.sweep(now)

	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{tokens: l.burst, last: now}
		l.buckets[key] = b
	}
	b.tokens = l.refilled(b, now)
	b.last = now

	if b.tokens >= 1 {
		b.tokens--
		return true, 0
	}
	missing := 1 - b.tokens
	return false, time.Duration(math.Ceil(missing / l.rate * float64(time.Second)))
}

// Tracked is the number of keys currently held. It exists for tests of the sweep.
func (l *Limiter) Tracked() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.buckets)
}

func (l *Limiter) refilled(b *bucket, now time.Time) float64 {
	elapsed := now.Sub(b.last).Seconds()
	return math.Min(l.burst, b.tokens+elapsed*l.rate)
}

// sweep drops full buckets at most once per full refill time (and no more than
// once a second), so the cost of the scan is spread over many requests.
func (l *Limiter) sweep(now time.Time) {
	if now.Before(l.nextSweep) {
		return
	}
	for key, b := range l.buckets {
		if l.refilled(b, now) >= l.burst {
			delete(l.buckets, key)
		}
	}
	interval := time.Duration(l.burst / l.rate * float64(time.Second))
	l.nextSweep = now.Add(max(interval, minSweepInterval))
}
