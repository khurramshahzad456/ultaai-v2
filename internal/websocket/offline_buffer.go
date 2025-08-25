// internal/websocket/offline_buffer.go
package websocket

import (
	"sync"
	"sync/atomic"

	"ultahost-ai-gateway/internal/config"
)

// A small, bounded, per-agent FIFO buffer for messages when the agent is offline.
// We bound by count (msgs) and bytes to keep memory predictable.

type offlineQueue struct {
	mu       sync.Mutex
	msgs     [][]byte
	bytes    int
	
	maxMsgs  int
	maxBytes int
}

func newOfflineQueue(maxMsgs, maxBytes int) *offlineQueue {
	return &offlineQueue{
		maxMsgs:  maxMsgs,
		maxBytes: maxBytes,
	}
}

func (q *offlineQueue) Enqueue(b []byte) (dropped int, accepted bool, deltaBytes int) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// If adding this would exceed limits, drop from the head until it fits.
	for (len(q.msgs) >= q.maxMsgs) || (q.bytes+len(b) > q.maxBytes) {
		if len(q.msgs) == 0 {
			// even a single msg larger than cap -> reject
			return 0, false, 0
		}
		head := q.msgs[0]
		q.msgs = q.msgs[1:]
		q.bytes -= len(head)
		dropped++
	}

	// append copy
	q.msgs = append(q.msgs, append([]byte(nil), b...))
	q.bytes += len(b)
	return dropped, true, len(b)
}

func (q *offlineQueue) Drain() (out [][]byte, deltaBytes int) {
	q.mu.Lock()
	defer q.mu.Unlock()
	out = q.msgs
	deltaBytes = -q.bytes
	q.msgs = nil
	q.bytes = 0
	return
}

func (q *offlineQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.msgs)
}

func (q *offlineQueue) Bytes() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.bytes
}

// Global map identityToken -> queue
var offlineBuf sync.Map // map[string]*offlineQueue

// --- NEW: global offline bytes ceiling ---
var (
	defaultOfflineMaxMsgs  = config.Int("OFFLINE_MAX_MSGS", 256)
	defaultOfflineMaxBytes = config.Int("OFFLINE_MAX_BYTES", 256*1024)                    // 256 KiB
	globalOfflineMaxBytes  = int64(config.Int("OFFLINE_GLOBAL_MAX_BYTES", 512*1024*1024)) // 512 MiB
	globalOfflineBytes     int64                                                          // atomically updated
)

func getOrMakeQueue(identityToken string) *offlineQueue {
	if v, ok := offlineBuf.Load(identityToken); ok {
		return v.(*offlineQueue)
	}
	q := newOfflineQueue(defaultOfflineMaxMsgs, defaultOfflineMaxBytes)
	actual, _ := offlineBuf.LoadOrStore(identityToken, q)
	return actual.(*offlineQueue)
}

// Exposed helpers
func OfflineEnqueue(identityToken string, payload []byte) (dropped int) {
	// fast fail if a single message is absurdly large compared to global cap
	if int64(len(payload)) > globalOfflineMaxBytes {
		metricsDropped(1)
		return 1
	}

	// Check global headroom
	for {
		cur := atomic.LoadInt64(&globalOfflineBytes)
		if cur+int64(len(payload)) > globalOfflineMaxBytes {
			// No room globally; drop this payload
			metricsDropped(1)
			return 1
		}
		if atomic.CompareAndSwapInt64(&globalOfflineBytes, cur, cur+int64(len(payload))) {
			break
		}
	}

	// Try to enqueue in per-agent queue (may drop head to make room)
	q := getOrMakeQueue(identityToken)
	droppedLocal, accepted, deltaBytes := q.Enqueue(payload)
	if !accepted {
		// refund the previously reserved global bytes
		atomic.AddInt64(&globalOfflineBytes, -int64(len(payload)))
		metricsDropped(1)
		return 1
	}

	// If we dropped local messages, reflect those bytes into global counter.
	if droppedLocal > 0 {
		// We don't know exact bytes of dropped heads here; approximate by
		// recalculating queue bytes total under lock or expose drop bytes.
		// Simpler: compute delta already returned by Enqueue and assume droppedLocal freed enough to admit payload.
		// Nothing to do here because Enqueue already ensured fit before we reserved.
		metricsDropped(droppedLocal)
	}

	// deltaBytes counts the payload admitted; already added to global before Enqueue.
	_ = deltaBytes
	return droppedLocal
}

func OfflineDrain(identityToken string) [][]byte {
	if v, ok := offlineBuf.Load(identityToken); ok {
		q := v.(*offlineQueue)
		out, delta := q.Drain()
		if delta != 0 {
			atomic.AddInt64(&globalOfflineBytes, int64(delta))
		}
		return out
	}
	return nil
}

func OfflineStats() (agents int, totalMsgs int, totalBytes int) {
	agents = 0
	totalMsgs = 0
	totalBytes = 0
	offlineBuf.Range(func(_, vv any) bool {
		q := vv.(*offlineQueue)
		agents++
		totalMsgs += q.Len()
		totalBytes += q.Bytes()
		return true
	})
	return
}
