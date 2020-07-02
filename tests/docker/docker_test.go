package lim

import (
	"flag"
	"fmt"
	"github.com/docker/docker/pkg/namesgenerator"
	"github.com/lf-edge/eden/pkg/defaults"
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

// This context holds all the configuration items in the same
// way that Eden context works: the commands line options override
// YAML settings. In addition to that, context is polymorphic in
// a sense that it abstracts away a particular controller (currently
// Adam and Zedcloud are supported)
/*
tc *TestContext // TestContext is at least {
                //    controller *Controller
                //    project *Project
                //    nodes []EdgeNode
                //    ...
                // }
*/

var (
	timewait     = flag.Int("timewait", 300, "Timewait for items waiting in seconds")
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

	projectName := fmt.Sprintf("%s_%s", "TestLogInfoMetric", time.Now())

	// Registering our own project namespace with controller for easy cleanup
	tc.InitProject(projectName)

	// Create representation of EVE instances (based on the names
	// or UUIDs that were passed in) in the context. This is the first place
	// where we're using zcli-like API:
	for _, node := range tc.GetNodeDescriptions() {
		edgeNode := tc.GetController().GetEdgeNode(node.Name)
		if edgeNode == nil {
			// Couldn't find existing edgeNode record in the controller.
			// Need to create it from scratch now:
			// this is modeled after: zcli edge-node create <name>
			// --project=<project> --model=<model> [--title=<title>]
			// ([--edge-node-certificate=<certificate>] |
			// [--onboarding-certificate=<certificate>] |
			// [(--onboarding-key=<key> --serial=<serial-number>)])
			// [--network=<network>...]
			//
			// XXX: not sure if struct (giving us optional fields) would be better
			edgeNode = tc.NewEdgeNode(tc.WithNodeDescription(node), tc.WithCurrentProject())
		} else {
			// make sure to move EdgeNode to the project we created, again
			// this is modeled after zcli edge-node update <name> [--title=<title>]
			// [--lisp-mode=experimental|default] [--project=<project>]
			// [--clear-onboarding-certs] [--config=<key:value>...] [--network=<network>...]
			tc.UpdateEdgeNode(edgeNode, tc.WithCurrentProject(), tc.WithDeviceModel(node.Model))
		}

		tc.ConfigSync(edgeNode)

		// finally we need to make sure that the edgeNode is in a state that we need
		// it to be, before the test can run -- this could be multiple checks on its
		// status, but for example:
		if edgeNode.GetState() == device.NotOnboarded {
			log.Fatal("Node is not onboarded now")
		}

		// this is a good node -- lets add it to the test context
		tc.AddNode(edgeNode)
	}

	tc.StartTrackingState(false)

	// we now have a situation where TestContext has enough EVE nodes known
	// for the rest of the tests to run. So run them:
	res := m.Run()

	// Finally, we need to cleanup whatever objects may be in in the project we created
	// and then we can exit
	os.Exit(res)
}

func checkAppMsg(appName string) projects.ProcInfoFunc {
	return func(msg *info.ZInfoMsg) error {
		if msg.Ztype == info.ZInfoTypes_ZiApp {
			if msg.GetAinfo().AppName == appName {
				return fmt.Errorf("app found with name %s", appName)
			}
		}
		return nil
	}
}

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

func requestApp(ip string) error {
	res, err := utils.RequestHTTPWithTimeout(fmt.Sprintf("http://%s:%d", ip, externalPort), time.Second)
	if err != nil {
		return nil
	} else {
		return fmt.Errorf(res)
	}
}

func getEVEIP(edgeNode *device.Ctx) projects.ProcTimerFunc {
	return func() error {
		if edgeNode.GetDevModel() == defaults.DefaultRPIModel {
			dInfo := tc.GetState(edgeNode).GetDinfo()
			if dInfo != nil {
				if len(dInfo.Network) > 0 {
					if dInfo.Network[0] != nil {
						if len(dInfo.Network[0].IPAddrs) > 0 {
							ip, _, err := net.ParseCIDR(dInfo.Network[0].IPAddrs[0])
							if err != nil {
								return nil
							}
							externalIP = ip.To4().String()
							return fmt.Errorf("external ip is: %s", externalIP)
						}
					}
				}
			}
		} else {
			externalIP = "127.0.0.1"
			return fmt.Errorf("external ip is: %s", externalIP)
		}
		return nil
	}
}

func checkAppAccess() projects.ProcTimerFunc {
	return func() error {
		if externalIP == "" {
			return nil
		}
		return requestApp(externalIP)
	}
}

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

func TestDockerStart(t *testing.T) {
	edgeNode := tc.GetEdgeNode(tc.WithTest(t))

	tc.ConfigSync(edgeNode)

	rand.Seed(time.Now().UnixNano())

	appName = namesgenerator.GetRandomName(0)

	expectation := expect.AppExpectationFromUrl(tc.GetController(), appLink, appName, portPublish, nil, "")

	appInstanceConfig := expectation.Application()

	t.Log("Add app to list")

	edgeNode.SetApplicationInstanceConfig(append(edgeNode.GetApplicationInstances(), appInstanceConfig.Uuidandversion.Uuid))

	tc.ConfigSync(edgeNode)

	t.Log("Add processing of app messages")

	tc.AddProcInfo(edgeNode, checkAppMsg(appName))

	t.Log("Add processing of app running messages")

	tc.AddProcInfo(edgeNode, checkAppRunning(appName))

	t.Log("Add function to obtain EVE IP")

	tc.AddProcTimer(edgeNode, getEVEIP(edgeNode))

	t.Log("Add trying to access app via http")

	tc.AddProcTimer(edgeNode, checkAppAccess())

	tc.WaitForProc(*timewait)
}

func TestDockerDelete(t *testing.T) {
	edgeNode := tc.GetEdgeNode(tc.WithTest(t))

	tc.ConfigSync(edgeNode)

	t.Logf("Add waiting for app %s absent", appName)

	tc.AddProcInfo(edgeNode, checkAppAbsent(appName))

	for id, appUUID := range edgeNode.GetApplicationInstances() {
		appConfig, _ := tc.GetController().GetApplicationInstanceConfig(appUUID)
		if appConfig.Displayname == appName {
			configs := edgeNode.GetApplicationInstances()
			t.Log("Remove app from list")
			utils.DelEleInSlice(&configs, id)
			edgeNode.SetApplicationInstanceConfig(configs)
			tc.ConfigSync(edgeNode)
			break
		}
	}

	tc.WaitForProc(*timewait)
}
