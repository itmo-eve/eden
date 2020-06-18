package projects

import (
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/lf-edge/eden/pkg/controller"
	"github.com/lf-edge/eden/pkg/controller/einfo"
	"github.com/lf-edge/eden/pkg/device"
	"github.com/lf-edge/eve/api/go/info"
	log "github.com/sirupsen/logrus"
	"testing"
)

type rebootInfo struct {
	time   *timestamp.Timestamp
	reason string
}

var (
	reboots  []rebootInfo
	rebooted int
)

func checkReboot(im *info.ZInfoMsg, ds []*einfo.ZInfoMsgInterface, infoType einfo.ZInfoType) bool {
	lrbt := im.GetDinfo().LastRebootTime
	lrbr := im.GetDinfo().LastRebootReason

	if len(reboots) == 0 {
		r := rebootInfo{time: lrbt, reason: lrbr}
		reboots = append(reboots, r)
	} else {
		last := len(reboots) - 1
		if !proto.Equal(reboots[last].time, lrbt) {
			r := rebootInfo{time: lrbt, reason: lrbr}
			reboots = append(reboots, r)
			rebooted += 1
			//t.Log("Rebooted #%d at %s with reason '%s'\n", rebooted, ptypes.TimestampString(r.time), r.reason)
			return true
		}
	}
	return false
}

func CheckRebootsInfo(t *testing.T, tc *TestContext, edgeNode *device.Ctx, number int) bool {
	log.Debug("CheckRebootInfo")

	ctx, err := controller.CloudPrepare()
	if err != nil {
		t.Log("Fail in CloudPrepare: %s", err)
	}

	for {
		err := ctx.InfoChecker(edgeNode.GetID(), map[string]string{"devId": edgeNode.GetID().String(), "rebootConfigCounter": ".*"}, einfo.ZInfoDinfo, checkReboot, einfo.InfoNew, 0)
		if err != nil {
			t.Log("Fail in waiting for info: ", err)
			return false
		}
		if rebooted >= number {
			for _, r := range reboots {
				t.Logf("Rebooted at %s with reason '%s'\n", ptypes.TimestampString(r.time), r.reason)
			}
			return true
		}
	}
}

func CheckReboots(t *testing.T, tc *TestContext, edgeNode *device.Ctx, number int) AssertFunc {
	t.Log("CheckReboots")
	return func() bool {
		return CheckRebootsInfo(t, tc, edgeNode, number)
	}
}
