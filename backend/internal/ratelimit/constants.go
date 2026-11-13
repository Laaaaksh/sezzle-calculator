package ratelimit

import "time"

// minSweepInterval stops a very fast limiter from scanning its map on every call.
const minSweepInterval = time.Second
