package eve

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/lf-edge/eden/pkg/controller"
	"github.com/lf-edge/eden/pkg/device"
	"github.com/lf-edge/eve/api/go/info"
	"github.com/lf-edge/eve/api/go/metrics"
)

//SnapshotInstState stores state of snapshot
type SnapshotInstState struct {
	Name      string
	UUID      string
	Volume    string
	Size      string
	AdamState string
	EveState  string
	deleted   bool
}

func snapshotInstStateHeader() string {
	return "NAME\tUUID\tVOLUME\tSize\tSTATE(ADAM)\tLAST_STATE(EVE)"
}

func (snapshotState *SnapshotInstState) toString() string {
	return fmt.Sprintf("%s\t%s\t%s\t%s\t%v\t%s",
		snapshotState.Name, snapshotState.UUID, snapshotState.Volume, snapshotState.Size,
		snapshotState.AdamState, snapshotState.EveState)
}

func (ctx *State) initSnapshots(ctrl controller.Cloud, dev *device.Ctx) error {
	ctx.snapshots = make(map[string]*SnapshotInstState)
	for _, el := range dev.GetSnapshots() {
		si, err := ctrl.GetSnapshot(el)
		if err != nil {
			return fmt.Errorf("no Snapshot in cloud %s: %s", el, err)
		}
		snapInstStateObj := &SnapshotInstState{
			Name:      si.DisplayName,
			UUID:      si.GetUuid(),
			Volume:    si.GetVolumeUuid(),
			AdamState: "IN_CONFIG",
			EveState:  "UNKNOWN",
			Size:      "-",
		}
		ctx.snapshots[si.GetUuid()] = snapInstStateObj
	}
	return nil
}

func (ctx *State) processSnapshotsByInfo(im *info.ZInfoMsg) {
	switch im.GetZtype() {
	case info.ZInfoTypes_ZiSnapshot:
		infoObject := im.GetSnapinfo()
		snapInstStateObj, ok := ctx.snapshots[infoObject.GetId()]
		if !ok {
			snapInstStateObj = &SnapshotInstState{
				Name:      infoObject.GetDisplayName(),
				UUID:      infoObject.GetId(),
				AdamState: "NOT_IN_CONFIG",
			}
			ctx.snapshots[infoObject.GetId()] = snapInstStateObj
		}
		snapInstStateObj.Volume = infoObject.VolumeUuid
		snapInstStateObj.EveState = fmt.Sprintf("created: %t; deleted: %t", infoObject.Created, infoObject.Deleted)
		if infoObject.RollbackTimeLastOp > 0 {
			snapInstStateObj.EveState = fmt.Sprintf("created: %t; deleted: %t; lastRollback: %d", infoObject.Created, infoObject.Deleted, infoObject.RollbackTimeLastOp)
		}
		if infoObject.ErrorMsg != "" {
			snapInstStateObj.EveState = fmt.Sprintf("%s: %s", snapInstStateObj.EveState, infoObject.ErrorMsg)
		}
		snapInstStateObj.deleted = infoObject.GetDisplayName() == ""
	}
}

func (ctx *State) processSnapshotsByMetric(_ *metrics.ZMetricMsg) {
	return
}

//SnapshotList prints snapshots
func (ctx *State) SnapshotList() error {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 1, '\t', 0)
	if _, err := fmt.Fprintln(w, snapshotInstStateHeader()); err != nil {
		return err
	}
	snapInstStatesSlice := make([]*SnapshotInstState, 0, len(ctx.Snapshots()))
	snapInstStatesSlice = append(snapInstStatesSlice, ctx.Snapshots()...)
	sort.SliceStable(snapInstStatesSlice, func(i, j int) bool {
		return snapInstStatesSlice[i].Name < snapInstStatesSlice[j].Name
	})
	for _, el := range snapInstStatesSlice {
		if _, err := fmt.Fprintln(w, el.toString()); err != nil {
			return err
		}
	}
	return w.Flush()
}
