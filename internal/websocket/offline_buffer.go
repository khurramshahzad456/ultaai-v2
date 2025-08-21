package websocket

import (
	"sync"
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

func (q *offlineQueue) Enqueue(b []byte) (dropped int) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// If adding this would exceed limits, drop from the head until it fits.
	for (len(q.msgs) >= q.maxMsgs) || (q.bytes+len(b) > q.maxBytes) {
		if len(q.msgs) == 0 {
			break
		}
		head := q.msgs[0]
		q.msgs = q.msgs[1:]
		q.bytes -= len(head)
		dropped++
	}
	q.msgs = append(q.msgs, append([]byte(nil), b...)) // copy
	q.bytes += len(b)
	return
}

func (q *offlineQueue) Drain() (out [][]byte) {
	q.mu.Lock()
	defer q.mu.Unlock()
	out = q.msgs
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

const (
	defaultOfflineMaxMsgs  = 256
	defaultOfflineMaxBytes = 256 * 1024 // 256 KiB
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
	q := getOrMakeQueue(identityToken)
	return q.Enqueue(payload)
}

func OfflineDrain(identityToken string) [][]byte {
	if v, ok := offlineBuf.Load(identityToken); ok {
		q := v.(*offlineQueue)
		return q.Drain()
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
