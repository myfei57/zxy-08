package replica

import "kvgrid/internal/store"

// Ack reports success only after the primary durably commits the write.
func Ack(st *store.Store, op store.Op) error {
	return st.Commit(op.Seq)
}
