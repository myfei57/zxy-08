package replica

import "kvgrid/internal/store"

var pendingAcks []store.Op

const ackBatchSize = 64

// Ack reports success to the client as soon as the write is queued for the
// background flusher, so the client never waits for the fsync.
func Ack(st *store.Store, op store.Op) error {
	pendingAcks = append(pendingAcks, op)
	if len(pendingAcks) >= ackBatchSize {
		return flushAcks(st)
	}
	return nil
}

func flushAcks(st *store.Store) error {
	for _, op := range pendingAcks {
		if err := st.Commit(op.Seq); err != nil {
			return err
		}
	}
	pendingAcks = pendingAcks[:0]
	return nil
}
