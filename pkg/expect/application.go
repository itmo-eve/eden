package expect

import (
	"fmt"
	"strings"

	"github.com/lf-edge/eden/pkg/defaults"
	"github.com/lf-edge/eden/pkg/utils"
	"github.com/lf-edge/eve/api/go/config"
	"github.com/lf-edge/eve/api/go/evecommon"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

//appBundle type for aggregate objects, needed for application
type appBundle struct {
	appInstanceConfig *config.AppInstanceConfig
	contentTrees      []*config.ContentTree
	volumes           []*config.Volume
}

//checkAppInstanceConfig checks if provided app match expectation
func (exp *AppExpectation) checkAppInstanceConfig(app *config.AppInstanceConfig) bool {
	if app == nil {
		return false
	}
	if app.Displayname == exp.appName && app.Displayname != exp.oldAppName {
		return true
	}
	return false
}

//createAppInstanceConfig creates AppInstanceConfig with provided img and netInstances
//  it uses published ports info from AppExpectation to create ACE
func (exp *AppExpectation) createAppInstanceConfig(contentTree *config.ContentTree, netInstances map[*NetInstanceExpectation]*config.NetworkInstanceConfig) (*appBundle, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	var bundle *appBundle
	switch exp.appType {
	case dockerApp, directoryApp:
		bundle = exp.createAppInstanceConfigDocker(contentTree, id)
	case httpApp, httpsApp, fileApp:
		bundle = exp.createAppInstanceConfigVM(contentTree, id)
	default:
		return nil, fmt.Errorf("not supported appType")
	}
	bundle.appInstanceConfig.StartDelayInSeconds = exp.startDelay
	for _, d := range exp.disks {
		mountPoint := ""
		proccessedLink := d
		splitLink := strings.Split(d, ",")
		if len(splitLink) > 1 {
			for _, el := range splitLink {
				splitArgs := strings.SplitN(el, "=", 2)
				if len(splitArgs) < 2 {
					log.Fatalf("cannot parse volume (must have src and dst): %s", el)
				}
				switch splitArgs[0] {
				case "source", "src":
					proccessedLink = splitArgs[1]
				case "destination", "dst", "target":
					mountPoint = splitArgs[1]
				}
			}
			log.Printf("will use volume [%s] at mount point [%s]", proccessedLink, mountPoint)
			//remove existing elements with the same mount point to overwrite
			utils.DelEleInSliceByFunction(&bundle.appInstanceConfig.VolumeRefList, func(i interface{}) bool {
				if i.(*config.VolumeRef).MountDir == mountPoint {
					utils.DelEleInSliceByFunction(&bundle.volumes, func(v interface{}) bool {
						return v.(*config.Volume).Uuid == i.(*config.VolumeRef).Uuid
					})
					return true
				}
				return false
			})
		}
		tempExp := AppExpectationFromURL(exp.ctrl, exp.device, proccessedLink, "")
		if tempExp.appType != dockerApp {
			//we should not overwrite type for docker
			tempExp.imageFormat = string(exp.volumesType)
		}
		innerContentTree := tempExp.ContentTree()
		if innerContentTree != nil {
			ind := len(bundle.volumes)
			toAppend := true
			for _, ct := range bundle.contentTrees {
				if ct.URL == innerContentTree.DisplayName && ct.Sha256 == innerContentTree.Sha256 {
					//skip append of existent ContentTree
					toAppend = false
					innerContentTree = ct
				}
			}
			if toAppend {
				bundle.contentTrees = append(bundle.contentTrees, innerContentTree)
			}
			volume := exp.contentTreeToVolume(exp.volumeSize, ind, innerContentTree)
			bundle.volumes = append(bundle.volumes, volume)
			bundle.appInstanceConfig.VolumeRefList = append(bundle.appInstanceConfig.VolumeRefList, &config.VolumeRef{Uuid: volume.Uuid, MountDir: mountPoint})
		}
	}
	if exp.virtualizationMode == config.VmMode_PV {
		bundle.appInstanceConfig.Fixedresources.Rootdev = "/dev/xvda1"
		bundle.appInstanceConfig.Fixedresources.Bootloader = "/usr/lib/xen/boot/ovmf.bin"
	}
	bundle.appInstanceConfig.Interfaces = []*config.NetworkAdapter{}

	//keep order of exp.netInstances
	for _, k := range exp.netInstances {
		ni, ok := netInstances[k]
		if !ok {
			log.Fatalf("broken network instance pointer: %v", k)
		}
		bundle.appInstanceConfig.Interfaces = append(bundle.appInstanceConfig.Interfaces, &config.NetworkAdapter{
			Name:         "default",
			NetworkId:    ni.Uuidandversion.Uuid,
			Acls:         exp.getAcls(k),
			MacAddress:   k.mac,
			AccessVlanId: exp.getAccessVID(k),
		})
	}
	if exp.vncDisplay != 0 {
		bundle.appInstanceConfig.Fixedresources.EnableVnc = true
		bundle.appInstanceConfig.Fixedresources.VncDisplay = exp.vncDisplay
		bundle.appInstanceConfig.Fixedresources.VncPasswd = exp.vncPassword
	}
	var adapters []*config.Adapter
	for _, adapterName := range exp.appAdapters {
		adapterType := evecommon.PhyIoType_PhyIoUSB
		if strings.HasPrefix(adapterName, "eth") {
			adapterType = evecommon.PhyIoType_PhyIoNetEth
		}
		adapters = append(adapters, &config.Adapter{
			Type: adapterType,
			Name: adapterName,
		})
	}
	bundle.appInstanceConfig.Adapters = adapters
	bundle.appInstanceConfig.ProfileList = exp.profiles
	return bundle, nil
}

//Application expectation gets or creates ContentTree definition, gets or create NetworkInstance definition,
//gets AppInstanceConfig and returns it or creates AppInstanceConfig, adds it into internal controller and returns it
func (exp *AppExpectation) Application() *config.AppInstanceConfig {
	contentTree := exp.ContentTree()
	networkInstances := exp.NetworkInstances()
	for _, appID := range exp.device.GetApplicationInstances() {
		app, err := exp.ctrl.GetApplicationInstanceConfig(appID)
		if err != nil {
			log.Fatalf("no app %s found in controller: %s", appID, err)
		}
		if exp.checkAppInstanceConfig(app) {
			return app
		}
	}
	bundle, err := exp.createAppInstanceConfig(contentTree, networkInstances)
	if err != nil {
		log.Fatalf("cannot create app: %s", err)
	}
	if err = exp.ctrl.AddApplicationInstanceConfig(bundle.appInstanceConfig); err != nil {
		log.Fatalf("AddApplicationInstanceConfig: %s", err)
	}
	if exp.appLink == defaults.DefaultDummyExpect {
		log.Debug("skip modify of entities")
	} else {
		for _, el := range bundle.contentTrees {
			_ = exp.ctrl.AddContentTree(el)
			exp.device.SetContentTreeConfig(append(exp.device.GetContentTrees(), el.Uuid))
		}
		for _, el := range bundle.volumes {
			_ = exp.ctrl.AddVolume(el)
			exp.device.SetVolumeConfigs(append(exp.device.GetVolumes(), el.Uuid))
		}
	}
	return bundle.appInstanceConfig
}
