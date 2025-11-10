package repository

import (
	"context"
	"time"
)

// RateLimitItem represents a single counter to increment with an associated TTL.
// Implementations should increment the key by 1 and set its expiration to TTL.
type RateLimitItem struct {
	Key string
	TTL time.Duration
}

// RateLimitRepository abstracts rate limit counter operations behind a repository layer.
//
// The primary use case is simple fixed-window rate limiting where requests are counted
// per window (e.g., per-minute, per-hour) using time-bucketed keys:
//   - Example keys:
//     ratelimit:<ip>:minute:<bucket>
//     ratelimit:<ip>:hour:<bucket>
//
// Typical flow per request:
//  1. Compute time-bucketed keys for each window (minute, hour, etc.)
//  2. Call IncrBatchWithTTL with TTLs slightly larger than the bucket size (to avoid overlap gaps)
//  3. Compare returned counters with configured thresholds
//
// Notes for implementers:
//   - When setting TTL/expiration, it is acceptable to reset TTL on every increment.
//   - Batch operation should be atomic where possible (e.g., Redis pipeline/transaction) but
//     exact guarantees are implementation-defined.
//   - Return counts must align with the order of input items.
type RateLimitRepository interface {
	IncrWithTTL(ctx context.Context, key string, ttl time.Duration) (int64, error)
	IncrBatchWithTTL(ctx context.Context, items []RateLimitItem) ([]int64, error)
}
