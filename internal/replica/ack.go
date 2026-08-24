package replica

import "kvgrid/internal/store"

// Ack durably persists op before reporting success to the client.
//
// The commit marker may advance only after the journal is fsynced, so that a
// crash after the caller returns success cannot lose data the client already
// believes is committed. Acknowledgement must therefore wait for the local
// write to reach stable storage; deferring the fsync to a background flusher
// would acknowledge writes that have not yet been persisted.
func Ack(st *store.Store, op store.Op) error {
	return st.Commit(op.Seq)
}
