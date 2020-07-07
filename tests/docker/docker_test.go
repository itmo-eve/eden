package lim

import (
	"flag"
	"fmt"
	"github.com/docker/docker/pkg/namesgenerator"
	"github.com/lf-edge/eden/pkg/device"
	"github.com/lf-edge/eden/pkg/expect"
	"github.com/lf-edge/eden/pkg/projects"
	"github.com/lf-edge/eden/pkg/utils"
	"github.com/lf-edge/eve/api/go/info"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"net"
	"os"
	"testing"
	"time"
)

// This test deploys the docker://nginx app into EVE with port forwarding 8028->80
// wait for the RUNNING state and checks access to HTTP endpoint
// and removes app from EVE

var (
	timewait     = flag.Int("timewait", 600, "Timewait for items waiting in seconds")
	tc           *projects.TestContext
	externalIP   string
	appName      string
	externalPort = 8028
	appLink      = "docker://nginx"
	portPublish  = []string{fmt.Sprintf("%d:80", externalPort)}
)

// TestMain is used to provide setup and teardown for the rest of the
// tests. As part of setup we make sure that context has a slice of
// EVE instances that we can operate on. For any action, if the instance
// is not specified explicitly it is assumed to be the first one in the slice
func TestMain(m *testing.M) {
	fmt.Println("Docker app deployment Test")

	tc = projects.NewTestContext()

	projectName := fmt.Sprintf("%s_%s", "TestDockerDeploy", time.Now())

	tc.InitProject(projectName)

	tc.AddEdgeNodesFromDescription()

	tc.StartTrackingState(false)

	res := m.Run()

	os.Exit(res)
}

//checkAppDeployStarted wait for info of ZInfoApp type with mention of deployed AppName
func checkAppDeployStarted(appName string) projects.ProcInfoFunc {
	return func(msg *info.ZInfoMsg) error {
		if msg.Ztype == info.ZInfoTypes_ZiApp {
			if msg.GetAinfo().AppName == appName {
				return fmt.Errorf("app found with name %s", appName)
			}
		}
		return nil
	}
}

//checkAppRunning wait for info of ZInfoApp type with mention of deployed AppName and ZSwState_RUNNING state
func checkAppRunning(appName string) projects.ProcInfoFunc {
	return func(msg *info.ZInfoMsg) error {
		if msg.Ztype == info.ZInfoTypes_ZiApp {
			if msg.GetAinfo().AppName == appName {
				if msg.GetAinfo().State == info.ZSwState_RUNNING {
					return fmt.Errorf("app RUNNING with name %s", appName)
				}
			}
		}
		return nil
	}
}

//getEVEIP wait for IPs of EVE and returns them
func getEVEIP(edgeNode *device.Ctx) projects.ProcTimerFunc {
	return func() error {
		if edgeNode.GetRemoteAddr() == "" { //no eve.remote-addr defined
			if eveIPCIDR, err := tc.GetState(edgeNode).LookUp("Dinfo.Network[0].IPAddrs[0]"); err != nil {
				return nil
			} else {
				if ip, _, err := net.ParseCIDR(eveIPCIDR.String()); err != nil {
					return nil
				} else {
					externalIP = ip.To4().String()
					return fmt.Errorf("external ip is: %s", externalIP)
				}
			}
		} else {
			externalIP = edgeNode.GetRemoteAddr()
			return fmt.Errorf("external ip is: %s", externalIP)
		}
	}
}

//checkAppAccess try to access APP with timer
func checkAppAccess() projects.ProcTimerFunc {
	return func() error {
		if externalIP == "" {
			return nil
		}
		res, err := utils.RequestHTTPWithTimeout(fmt.Sprintf("http://%s:%d", externalIP, externalPort), time.Second)
		if err != nil {
			return nil
		} else {
			return fmt.Errorf(res)
		}
	}
}

//checkAppAbsent check if APP undefined in EVE
func checkAppAbsent(appName string) projects.ProcInfoFunc {
	return func(msg *info.ZInfoMsg) error {
		if msg.Ztype == info.ZInfoTypes_ZiDevice {
			for _, app := range msg.GetDinfo().AppInstances {
				if app.Name == appName {
					return nil
				}
			}
			return fmt.Errorf("no app with %s found", appName)
		}
		return nil
	}
}

//TestDockerStart gets EdgeNode and deploys app, defined in appLink
//it generates random appName and adds processing functions
//it checks if app processed by EVE, app in RUNNING state, app is accessible by HTTP get
//it uses timewait for processing all events
func TestDockerStart(t *testing.T) {
	edgeNode := tc.GetEdgeNode(tc.WithTest(t))

	rand.Seed(time.Now().UnixNano())

	appName = namesgenerator.GetRandomName(0)

	expectation := expect.AppExpectationFromUrl(tc.GetController(), appLink, appName, portPublish, nil, "")

	appInstanceConfig := expectation.Application()

	t.Log("Add app to list")

	edgeNode.SetApplicationInstanceConfig(append(edgeNode.GetApplicationInstances(), appInstanceConfig.Uuidandversion.Uuid))

	tc.ConfigSync(edgeNode)

	t.Log("Add processing of app messages")

	tc.AddProcInfo(edgeNode, checkAppDeployStarted(appName))

	t.Log("Add processing of app running messages")

	tc.AddProcInfo(edgeNode, checkAppRunning(appName))

	t.Log("Add function to obtain EVE IP")

	tc.AddProcTimer(edgeNode, getEVEIP(edgeNode))

	t.Log("Add trying to access app via http")

	tc.AddProcTimer(edgeNode, checkAppAccess())

	tc.WaitForProc(*timewait)
}

//TestDockerDelete gets EdgeNode and deletes previously deployed app, defined in appName
//it checks if app absent in EVE
//it uses timewait for processing all events
func TestDockerDelete(t *testing.T) {
	edgeNode := tc.GetEdgeNode(tc.WithTest(t))

	t.Logf("Add waiting for app %s absent", appName)

	tc.AddProcInfo(edgeNode, checkAppAbsent(appName))

	for id, appUUID := range edgeNode.GetApplicationInstances() {
		appConfig, _ := tc.GetController().GetApplicationInstanceConfig(appUUID)
		if appConfig.Displayname == appName {
			configs := edgeNode.GetApplicationInstances()
			t.Log("Remove app from list")
			utils.DelEleInSlice(&configs, id)
			edgeNode.SetApplicationInstanceConfig(configs)
			if err := tc.GetController().RemoveApplicationInstanceConfig(appUUID); err != nil {
				log.Fatal(err)
			}
			tc.ConfigSync(edgeNode)
			break
		}
	}

	tc.WaitForProc(*timewait)
}
