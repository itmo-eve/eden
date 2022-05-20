package controller

import (
	"fmt"

	"github.com/lf-edge/eden/pkg/utils"
	"github.com/lf-edge/eve/api/go/config"
)

//GetSnapshot return Snapshot config from cloud by ID
func (cloud *CloudCtx) GetSnapshot(id string) (*config.SnapshotConfig, error) {
	for _, snapshot := range cloud.snapshots {
		if snapshot.Uuid == id {
			return snapshot, nil
		}
	}
	return nil, fmt.Errorf("not found SnapshotConfig with ID: %s", id)
}

//AddSnapshot add SnapshotConfig to cloud
func (cloud *CloudCtx) AddSnapshot(snapshot *config.SnapshotConfig) error {
	for _, snap := range cloud.snapshots {
		if snap.Uuid == snapshot.GetUuid() {
			return fmt.Errorf("SnapshotConfig already exists with ID: %s", snapshot.GetUuid())
		}
	}
	cloud.snapshots = append(cloud.snapshots, snapshot)
	return nil
}

//RemoveSnapshot remove SnapshotConfig config from cloud
func (cloud *CloudCtx) RemoveSnapshot(id string) error {
	for ind, snapshot := range cloud.snapshots {
		if snapshot.Uuid == id {
			utils.DelEleInSlice(&cloud.snapshots, ind)
			return nil
		}
	}
	return fmt.Errorf("not found SnapshotConfig with ID: %s", id)
}

//ListSnapshot return Volume configs from cloud
func (cloud *CloudCtx) ListSnapshot() []*config.SnapshotConfig {
	return cloud.snapshots
}
