package projects

import (
	"fmt"
	"github.com/lf-edge/eden/pkg/controller"
	"github.com/lf-edge/eden/pkg/controller/adam"
	"github.com/lf-edge/eden/pkg/defaults"
	"github.com/lf-edge/eden/pkg/device"
	"github.com/lf-edge/eden/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"strings"
)

//GetControllerMode parse url with controller
func GetControllerMode(controllerMode string) (modeType, modeURL string, err error) {
	params := utils.GetParams(controllerMode, defaults.DefaultControllerModePattern)
	if len(params) == 0 {
		return "", "", fmt.Errorf("cannot parse mode (not [file|proto|adam|zedcloud]://<URL>): %s", controllerMode)
	}
	ok := false
	if modeType, ok = params["Type"]; !ok {
		return "", "", fmt.Errorf("cannot parse modeType (not [file|proto|adam|zedcloud]://<URL>): %s", controllerMode)
	}
	if modeURL, ok = params["URL"]; !ok {
		return "", "", fmt.Errorf("cannot parse modeURL (not [file|proto|adam|zedcloud]://<URL>): %s", controllerMode)
	}
	return
}

//TestContext is main structure for running tests
type TestContext struct {
	cloud   controller.Cloud
	project *Project
	nodes   []*device.Ctx
}

//NewTestContext creates new TestContext
func NewTestContext() *TestContext {
	viperLoaded, err := utils.LoadConfigFile("")
	if err != nil {
		log.Fatalf("LoadConfigFile %s", err)
	}
	if viperLoaded {
		modeType, modeURL, err := GetControllerMode(viper.GetString("test.controller"))
		if err != nil {
			log.Fatal(err)
		}
		if modeType != "" {
			if modeType != "adam" {
				log.Fatalf("Not implemented controller type %s", modeType)
			}
		}
		if modeURL != "" { //overwrite config only if url defined
			ipPort := strings.Split(modeURL, ":")
			ip := ipPort[0]
			if ip == "" {
				log.Fatalf("cannot get ip/hostname from %s", modeURL)
			}
			port := "80"
			if len(ipPort) > 1 {
				port = ipPort[1]
			}
			viper.Set("adam.ip", ip)
			viper.Set("adam.port", port)
		}
	}
	vars, err := utils.InitVars()
	if err != nil {
		log.Fatalf("utils.InitVars: %s", err)
	}
	ctx := &controller.CloudCtx{Controller: &adam.Ctx{}}
	ctx.SetVars(vars)
	if err := ctx.InitWithVars(vars); err != nil {
		log.Fatalf("cloud.InitWithVars: %s", err)
	}
	ctx.GetAllNodes()
	return &TestContext{cloud: ctx}
}

//GetNodeDescriptions returns list of nodes from config
func (ctx *TestContext) GetNodeDescriptions() (nodes []*EdgeNodeDescription) {
	eveList := viper.GetStringMap("test.eve")
	for name := range eveList {
		eveKey := viper.GetString(fmt.Sprintf("test.eve.%s.onboard-cert", name))
		eveSerial := viper.GetString(fmt.Sprintf("test.eve.%s.serial", name))
		eveModel := viper.GetString(fmt.Sprintf("test.eve.%s.model", name))
		nodes = append(nodes, &EdgeNodeDescription{Name: name, Key: eveKey, Serial: eveSerial, Model: eveModel})
	}
	return
}

//GetController returns current controller
func (ctx *TestContext) GetController() controller.Cloud {
	if ctx.cloud == nil {
		log.Fatal("Controller not initialized")
	}
	return ctx.cloud
}

//InitProject init project object with defined name
func (ctx *TestContext) InitProject(name string) {
	ctx.project = &Project{name: name}
}

//GetEdgeNode return node from context
func (ctx *TestContext) GetEdgeNode(name string) *device.Ctx {
	for _, el := range ctx.nodes {
		if name == "" {
			return el
		}
		if el.GetName() == name {
			return el
		}
	}
	return nil
}

//AddNode add node to test context
func (ctx *TestContext) AddNode(node *device.Ctx) {
	ctx.nodes = append(ctx.nodes, node)
}

//NewEdgeNode creates edge node
func (ctx *TestContext) NewEdgeNode(opts ...EdgeNodeOption) *device.Ctx {
	d := device.CreateEdgeNode()
	for _, opt := range opts {
		opt(d)
	}
	if ctx.project == nil {
		log.Fatal("You must setup project before add node")
	}
	return d
}

//ConfigSync send config to controller
func (ctx *TestContext) ConfigSync(edgeNode *device.Ctx) {
	if edgeNode.GetState() == device.NotOnboarded {
		if err := ctx.GetController().OnBoardDev(edgeNode); err != nil {
			log.Fatalf("OnBoardDev %s", err)
		}
	} else {
		log.Debug("Device %s onboarded", edgeNode.GetID().String())
	}
	if err := ctx.GetController().ConfigSync(edgeNode); err != nil {
		log.Fatalf("Cannot send config of %s", edgeNode.GetName())
	}
}
