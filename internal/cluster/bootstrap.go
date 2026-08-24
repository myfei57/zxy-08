package cluster

import "kvgrid/internal/config"

// Recover restores the cluster from the newest snapshot generation, if any.
func (c *Cluster) Recover() error {
	gen, err := c.Snapshot.CurrentGeneration()
	if err != nil {
		return err
	}
	if gen == 0 {
		return nil
	}
	return c.Snapshot.Restore(config.SnapshotDir(c.Cfg.DataDir))
}
