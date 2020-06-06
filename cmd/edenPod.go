package cmd

import (
	"fmt"
	"github.com/lf-edge/eden/pkg/controller/einfo"
	"github.com/lf-edge/eden/pkg/expect"
	"github.com/lf-edge/eden/pkg/utils"
	"github.com/lf-edge/eve/api/go/info"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"text/tabwriter"
)

var podName = ""

var podCmd = &cobra.Command{
	Use: "pod",
}

//podDeployCmd is command for deploy application on EVE
var podDeployCmd = &cobra.Command{
	Use:   "deploy <docker>://<TAG>[:<VERSION>]",
	Short: "Deploy app in pod",
	Long:  `Deploy app in pod.`,
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		assingCobraToViper(cmd)
		_, err := utils.LoadConfigFile(configFile)
		if err != nil {
			return fmt.Errorf("error reading config: %s", err.Error())
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		appLink := args[0]
		changer := &adamChanger{}
		ctrl, dev, err := changer.getControllerAndDev()
		if err != nil {
			log.Fatalf("getControllerAndDev: %s", err)
		}
		expectation := expect.AppExpectationFromUrl(ctrl, appLink, podName)
		appInstanceConfig := expectation.Application()
		devModel, err := ctrl.GetDevModelByName(viper.GetString("eve.devmodel"))
		if err != nil {
			log.Fatalf("fail to get dev model %s: %s", viper.GetString("eve.devmodel"), err)
		}
		if err = ctrl.ApplyDevModel(dev, devModel); err != nil {
			log.Fatalf("ApplyDevModel: %s", err)
		}
		dev.SetApplicationInstanceConfig(append(dev.GetApplicationInstances(), appInstanceConfig.Uuidandversion.Uuid))
		if err = changer.setControllerAndDev(ctrl, dev); err != nil {
			log.Fatalf("setControllerAndDev: %s", err)
		}
		log.Infof("deploy pod %s with %s request sent", appInstanceConfig.Displayname, appLink)
	},
}

type appState struct {
	name      string
	image     string
	adamState string
	eveState  string
	ip        string
	deleted   bool
}

func appStateHeader() string {
	return "NAME\tIMAGE\tIP\tSTATE(ADAM)\tLAST_STATE(EVE)"
}

func (appStateObj *appState) toString() string {
	return fmt.Sprintf("%s\t%s\t%s\t%s\t%s", appStateObj.name, appStateObj.image, appStateObj.ip, appStateObj.adamState, appStateObj.eveState)
}

//podPsCmd is a command to list deployed apps
var podPsCmd = &cobra.Command{
	Use:   "ps",
	Short: "List pods",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		assingCobraToViper(cmd)
		_, err := utils.LoadConfigFile(configFile)
		if err != nil {
			return fmt.Errorf("error reading config: %s", err.Error())
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		changer := &adamChanger{}
		ctrl, dev, err := changer.getControllerAndDev()
		if err != nil {
			log.Fatalf("getControllerAndDev: %s", err)
		}
		appStates := make(map[string]*appState)
		for _, el := range dev.GetApplicationInstances() {
			app, err := ctrl.GetApplicationInstanceConfig(el)
			if err != nil {
				log.Fatalf("no app in cloud %s: %s", el, err)
			}
			imageName := ""
			if len(app.Drives) > 0 {
				imageName = app.Drives[0].Image.Name
			}
			appStateObj := &appState{name: app.Displayname, image: imageName, adamState: "IN_CONFIG", eveState: "UNKNOWN", ip: "-"}
			appStates[app.Uuidandversion.Uuid] = appStateObj
		}
		var handleInfo = func(im *info.ZInfoMsg, ds []*einfo.ZInfoMsgInterface, infoType einfo.ZInfoType) bool {
			appStateObj, ok := appStates[im.GetAinfo().AppID]
			if !ok {
				imageName := ""
				if len(im.GetAinfo().GetSoftwareList()) > 0 {
					imageName = im.GetAinfo().GetSoftwareList()[0].ImageName
				}
				appStateObj = &appState{name: im.GetAinfo().AppName, image: imageName, adamState: "NOT_IN_CONFIG"}
				appStates[im.GetAinfo().AppID] = appStateObj
			}
			appStateObj.eveState = im.GetAinfo().State.String()
			if len(im.GetAinfo().Network) > 0 && len(im.GetAinfo().Network[0].IPAddrs) > 0 {
				appStateObj.ip = im.GetAinfo().Network[0].IPAddrs[0]
			} else {
				appStateObj.ip = "-"
			}
			return false
		}
		if err = ctrl.InfoLastCallback(dev.GetID(), map[string]string{"devId": dev.GetID().String()}, einfo.ZInfoAppInstance, handleInfo); err != nil {
			log.Fatalf("Fail in get InfoLastCallback: %s", err)
		}
		var handleInfoDevice = func(im *info.ZInfoMsg, ds []*einfo.ZInfoMsgInterface, infoType einfo.ZInfoType) bool {
			for _, appStateObj := range appStates {
				if appStateObj.adamState == "NOT_IN_CONFIG" {
					appStateObj.deleted = true
					for _, el := range im.GetDinfo().AppInstances {
						if appStateObj.name == el.Name {
							appStateObj.deleted = false
						}
					}
				}
			}
			return false
		}
		if err = ctrl.InfoLastCallback(dev.GetID(), map[string]string{"devId": dev.GetID().String()}, einfo.ZInfoDinfo, handleInfoDevice); err != nil {
			log.Fatalf("Fail in get InfoLastCallback: %s", err)
		}
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 8, 1, '\t', 0)
		if _, err = fmt.Fprintln(w, appStateHeader()); err != nil {
			log.Fatal(err)
		}
		for _, el := range appStates {
			if el.deleted == false {
				if _, err = fmt.Fprintln(w, el.toString()); err != nil {
					log.Fatal(err)
				}
			}
		}
		if err = w.Flush(); err != nil {
			log.Fatal(err)
		}
	},
}

//podStopCmd is a command to stop app
var podStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop pod",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		assingCobraToViper(cmd)
		_, err := utils.LoadConfigFile(configFile)
		if err != nil {
			return fmt.Errorf("error reading config: %s", err.Error())
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		appName := args[0]
		changer := &adamChanger{}
		ctrl, dev, err := changer.getControllerAndDev()
		if err != nil {
			log.Fatalf("getControllerAndDev: %s", err)
		}
		for _, el := range dev.GetApplicationInstances() {
			app, err := ctrl.GetApplicationInstanceConfig(el)
			if err != nil {
				log.Fatalf("no app in cloud %s: %s", el, err)
			}
			if app.Displayname == appName {
				app.Activate = false
				if err = changer.setControllerAndDev(ctrl, dev); err != nil {
					log.Fatalf("setControllerAndDev: %s", err)
				}
				log.Infof("app %s stop done", appName)
				return
			}
		}
		log.Infof("not found app with name %s", appName)
	},
}

//podStopCmd is a command to start app
var podStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start pod",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		assingCobraToViper(cmd)
		_, err := utils.LoadConfigFile(configFile)
		if err != nil {
			return fmt.Errorf("error reading config: %s", err.Error())
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		appName := args[0]
		changer := &adamChanger{}
		ctrl, dev, err := changer.getControllerAndDev()
		if err != nil {
			log.Fatalf("getControllerAndDev: %s", err)
		}
		for _, el := range dev.GetApplicationInstances() {
			app, err := ctrl.GetApplicationInstanceConfig(el)
			if err != nil {
				log.Fatalf("no app in cloud %s: %s", el, err)
			}
			if app.Displayname == appName {
				app.Activate = true
				if err = changer.setControllerAndDev(ctrl, dev); err != nil {
					log.Fatalf("setControllerAndDev: %s", err)
				}
				log.Infof("app %s start done", appName)
				return
			}
		}
		log.Infof("not found app with name %s", appName)
	},
}

//podStopCmd is a command to delete app from EVE
var podDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete pod",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		assingCobraToViper(cmd)
		_, err := utils.LoadConfigFile(configFile)
		if err != nil {
			return fmt.Errorf("error reading config: %s", err.Error())
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		appName := args[0]
		changer := &adamChanger{}
		ctrl, dev, err := changer.getControllerAndDev()
		if err != nil {
			log.Fatalf("getControllerAndDev: %s", err)
		}
		for id, el := range dev.GetApplicationInstances() {
			app, err := ctrl.GetApplicationInstanceConfig(el)
			if err != nil {
				log.Fatalf("no app in cloud %s: %s", el, err)
			}
			if app.Displayname == appName {
				configs := dev.GetApplicationInstances()
				utils.DelEleInSlice(&configs, id)
				dev.SetApplicationInstanceConfig(configs)
				if err = changer.setControllerAndDev(ctrl, dev); err != nil {
					log.Fatalf("setControllerAndDev: %s", err)
				}
				log.Infof("app %s delete done", appName)
				return
			}
		}
		log.Infof("not found app with name %s", appName)
	},
}

func podInit() {
	podCmd.AddCommand(podDeployCmd)
	podDeployCmd.Flags().StringVarP(&podName, "name", "n", "", "name for pod")
	podCmd.AddCommand(podPsCmd)
	podCmd.AddCommand(podStopCmd)
	podCmd.AddCommand(podStartCmd)
	podCmd.AddCommand(podDeleteCmd)
}
