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
	"math/rand"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

//This test deploys two apps into EVE and checks access between them trough network
//Inside apps there are nginx and curl services.
//    curl try to download page provided in $url variable and save it into received-data.html file.
//    Nginx serve $url variable content at /user-data.html endpoint and received-data.html at /received-data.html endpoint.
//    $url variable set from metadata information provided to EVE on application start
//For the first app we define url=TEST_SEQUENCE (or another text from 'metadata' flag) and we can get it from /user-data.html endpoint
//For the second app we define url=http://internalIP/user-data.html (where internalIP - is IP of first app, we obtain it from the running app)
//Curl service of second app follow $url variable, try to access first app, and save output into received-data.html
//We try to get content of received-data.html through HTTP GET to second app /received-data.html endpoint
//    and compare result with TEST_SEQUENCE (or another text from 'metadata' flag)
var (
	timewait      = flag.Int("timewait", 600, "Timewait for items waiting in seconds")
	name1         = flag.String("name1", "", "Name of app1, random if empty")
	name2         = flag.String("name2", "", "Name of app2, random if empty")
	externalPort2 = flag.Int("externalPort2", 8028, "Port for access app2 from outside of EVE")
	internalPort  = flag.Int("internalPort", 80, "Port for access app inside EVE")
	appLink       = flag.String("appLink", "docker://itmoeve/docker-test", "Link to get app")
	metadata      = flag.String("metadata", "TEST_SEQUENCE", "Metadata to test")
	networkType   = flag.String("net-type", "local", "Network type: local or switch")
	tc            *projects.TestContext
	externalIP    string
	internalIP    string
	appName1      string
	appName2      string
)

// TestMain is used to provide setup and teardown for the rest of the
// tests. As part of setup we make sure that context has a slice of
// EVE instances that we can operate on. For any action, if the instance
// is not specified explicitly it is assumed to be the first one in the slice
func TestMain(m *testing.M) {
	fmt.Println("Network connection between apps Test")

	tc = projects.NewTestContext()

	projectName := fmt.Sprintf("%s_%s", "TestNetwork", time.Now())

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
func checkAppAccess(port int) projects.ProcTimerFunc {
	return func() error {
		if externalIP == "" {
			return nil
		}
		res, err := utils.RequestHTTPWithTimeout(fmt.Sprintf("http://%s:%d/received-data.html", externalIP, port), time.Second)
		if err != nil {
			return nil
		} else {
			if strings.Contains(res, *metadata) {
				return fmt.Errorf(res)
			}
		}
		return nil
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

//getInternalIP gets internalIP of APP
func getInternalIP(appName string) projects.ProcInfoFunc {
	macs := make(map[string]string)
	return func(msg *info.ZInfoMsg) error {
		switch msg.Ztype {
		case info.ZInfoTypes_ZiApp:
			if msg.GetAinfo().AppName == appName && len(msg.GetAinfo().Network) > 0 && len(msg.GetAinfo().Network[0].IPAddrs) > 0 {
				internalIP = msg.GetAinfo().Network[0].IPAddrs[0]
				if strings.TrimSpace(internalIP) != "" {
					return fmt.Errorf("internalIP %s found", internalIP)
				} else {
					//it is switch network, need to get from NetworkInstances
					macs[appName] = msg.GetAinfo().Network[0].MacAddr
				}
			}
		case info.ZInfoTypes_ZiNetworkInstance: //try to find ips from NetworkInstances
			for _, el := range msg.GetNiinfo().IpAssignments {
				for macAppName, mac := range macs {
					if macAppName == appName && mac == el.MacAddress {
						internalIP = el.IpAddress[0]
						return fmt.Errorf("internalIP %s found", internalIP)
					}
				}
			}
		}
		return nil
	}
}

//TestNetwork gets EdgeNode
//it generates random appNames and adds processing functions
//it starts app1 with defined metadata and checks:
//   is app processed by EVE,
//   is app in RUNNING state,
//   get app`s IP address
//it starts app2 with metadata pointed to app1 and two network instances:
//   1) same as the first app
//   2) new one with port mapping
//it starts checks for app2:
//   is app processed by EVE,
//   is app in RUNNING state,
//   access endpoint of app through HTTP GET to obtain metadata (metadata go through sequence endpoint->app2->app1)
//it uses timewait for processing all events
func TestNetworkAccess(t *testing.T) {
	edgeNode := tc.GetEdgeNode(tc.WithTest(t))

	if *name1 == "" {
		rand.Seed(time.Now().UnixNano())
		appName1 = namesgenerator.GetRandomName(0) //generates new name if no flag set
	} else {
		appName1 = *name1
	}

	if *name2 == "" {
		rand.Seed(time.Now().UnixNano())
		appName2 = namesgenerator.GetRandomName(0) //generates new name if no flag set
	} else {
		appName2 = *name2
	}

	metadata1Formatted := fmt.Sprintf("url=%s", *metadata)

	expectation := expect.AppExpectationFromUrl(tc.GetController(), *appLink, appName1, expect.AddNetInstanceAndPortPublish("10.2.0.0/24", *networkType, nil), expect.WithMetadata(metadata1Formatted))

	appInstanceConfig := expectation.Application()

	t.Log("Add app1 to list")

	edgeNode.SetApplicationInstanceConfig(append(edgeNode.GetApplicationInstances(), appInstanceConfig.Uuidandversion.Uuid))

	tc.ConfigSync(edgeNode)

	t.Log("Add processing of app messages")

	tc.AddProcInfo(edgeNode, checkAppDeployStarted(appName1))

	t.Log("Add processing of app running messages")

	tc.AddProcInfo(edgeNode, checkAppRunning(appName1))

	t.Log("Add function to obtain EVE IP")

	tc.AddProcTimer(edgeNode, getEVEIP(edgeNode))

	t.Log("Add function to obtain Internal IP")

	tc.AddProcInfo(edgeNode, getInternalIP(appName1))

	tc.WaitForProc(*timewait)

	metadata2 := fmt.Sprintf("url=http://%s/user-data.html", internalIP)

	portPublish2 := []string{fmt.Sprintf("%d:%d", *externalPort2, *internalPort)}

	expectation2 := expect.AppExpectationFromUrl(tc.GetController(), *appLink, appName2, expect.AddNetInstanceAndPortPublish("10.3.0.0/24", "local", portPublish2), expect.AddNetInstanceAndPortPublish("10.2.0.0/24", *networkType, nil), expect.WithMetadata(metadata2))

	appInstanceConfig2 := expectation2.Application()

	t.Log("Add app2 to list")

	edgeNode.SetApplicationInstanceConfig(append(edgeNode.GetApplicationInstances(), appInstanceConfig2.Uuidandversion.Uuid))

	tc.ConfigSync(edgeNode)

	t.Log("Add processing of app messages")

	tc.AddProcInfo(edgeNode, checkAppDeployStarted(appName2))

	t.Log("Add processing of app running messages")

	tc.AddProcInfo(edgeNode, checkAppRunning(appName2))

	t.Log("Add trying to access app via http")

	tc.AddProcTimer(edgeNode, checkAppAccess(*externalPort2))

	tc.WaitForProc(*timewait)

}

//TestNetworkAccessDelete gets EdgeNode and deletes previously deployed app, defined in appName
//it checks if app absent in EVE
//it uses timewait for processing all events
func TestNetworkAccessDelete(t *testing.T) {

	if appName1 == "" { //if previous appName not defined
		if *name1 == "" {
			t.Fatal("No name of app, please set 'name' flag")
		} else {
			appName1 = *name1
		}
	}

	if appName2 == "" { //if previous appName not defined
		if *name2 == "" {
			t.Fatal("No name of app, please set 'name' flag")
		} else {
			appName2 = *name2
		}
	}

	edgeNode := tc.GetEdgeNode(tc.WithTest(t))

	t.Logf("Add waiting for app %s absent", appName1)

	tc.AddProcInfo(edgeNode, checkAppAbsent(appName1))

	t.Logf("Add waiting for app %s absent", appName2)

	tc.AddProcInfo(edgeNode, checkAppAbsent(appName2))

	tc.AppDelete(edgeNode, appName1)

	tc.AppDelete(edgeNode, appName2)

	tc.ConfigSync(edgeNode)

	tc.WaitForProc(*timewait)
}
