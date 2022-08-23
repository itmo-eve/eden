package expect

import (
	"fmt"

	"github.com/lf-edge/eve/api/go/config"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

//checkContentTree checks if provided img match expectation
func (exp *AppExpectation) checkContentTree(img *config.ContentTree, dsID string) bool {
	if img == nil {
		return false
	}
	switch exp.appType {
	case dockerApp:
		return exp.checkContentTreeDocker(img, dsID)
	case httpApp, httpsApp, fileApp:
		return exp.checkContentTreeHTTP(img, dsID)
	}
	return false
}

//createContentTree creates ContentTree with provided dsID for AppExpectation
func (exp *AppExpectation) createContentTree(dsID string) (*config.ContentTree, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	switch exp.appType {
	case dockerApp:
		return exp.createContentTreeDocker(id, dsID), nil
	case httpApp, httpsApp:
		return exp.createContentTreeHTTP(id, dsID), nil
	case fileApp:
		return exp.createContentTreeFile(id, dsID), nil
	case directoryApp:
		return exp.createContentTreeDirectory(id, dsID), nil
	default:
		return nil, fmt.Errorf("not supported appType")
	}
}

// imageFormatEnum return the correct enum for the image format
func (exp *AppExpectation) imageFormatEnum() config.Format {
	var defaultFormat, actual config.Format
	switch exp.appType {
	case dockerApp:
		defaultFormat = config.Format_CONTAINER
	case httpApp, httpsApp, fileApp:
		defaultFormat = config.Format_QCOW2
	default:
		defaultFormat = config.Format_QCOW2
	}
	switch exp.imageFormat {
	case "container", "oci":
		actual = config.Format_CONTAINER
	case "qcow2":
		actual = config.Format_QCOW2
	case "raw":
		actual = config.Format_RAW
	case "qcow":
		actual = config.Format_QCOW
	case "vmdk":
		actual = config.Format_VMDK
	case "vhdx":
		actual = config.Format_VHDX
	default:
		actual = defaultFormat
	}
	return actual
}

//ContentTree expects ContentTree in controller
//it gets ContentTree with defined in AppExpectation params, or creates new one, if not exists
func (exp *AppExpectation) ContentTree() (ct *config.ContentTree) {
	datastore := exp.DataStore()
	var err error
	for _, appID := range exp.device.GetApplicationInstances() {
		app, err := exp.ctrl.GetApplicationInstanceConfig(appID)
		if err != nil {
			log.Fatalf("no app %s found in controller: %s", appID, err)
		}
		for _, volumeRef := range app.VolumeRefList {
			vol, err := exp.ctrl.GetVolume(volumeRef.GetUuid())
			if err != nil {
				log.Fatalf("no volume %s found in controller: %s", volumeRef.GetUuid(), err)
			}
			if vol.Origin == nil || vol.Origin.DownloadContentTreeID == "" {
				continue
			}
			currentContentTree, err := exp.ctrl.GetContentTree(vol.Origin.DownloadContentTreeID)
			if err != nil {
				log.Fatalf("no ContentTree %s found in controller: %s", vol.Origin.DownloadContentTreeID, err)
			}
			if exp.checkContentTree(currentContentTree, datastore.Id) {
				ct = currentContentTree
				break
			}
		}
	}
	for _, baseID := range exp.device.GetBaseOSConfigs() {
		base, err := exp.ctrl.GetBaseOSConfig(baseID)
		if err != nil {
			log.Fatalf("no baseOS %s found in controller: %s", baseID, err)
		}
		vol, err := exp.ctrl.GetVolume(base.GetVolumeID())
		if err != nil {
			log.Fatalf("no volume %s found in controller: %s", base.GetVolumeID(), err)
		}
		if vol.Origin == nil || vol.Origin.DownloadContentTreeID == "" {
			continue
		}
		currentContentTree, err := exp.ctrl.GetContentTree(vol.Origin.DownloadContentTreeID)
		if err != nil {
			log.Fatalf("no ContentTree %s found in controller: %s", vol.Origin.DownloadContentTreeID, err)
		}
		if exp.checkContentTree(currentContentTree, datastore.Id) {
			ct = currentContentTree
			break
		}
	}
	if ct == nil { //if image not exists, create it
		if ct, err = exp.createContentTree(datastore.Id); err != nil {
			log.Fatalf("cannot create image: %s", err)
		}
		if err = exp.ctrl.AddContentTree(ct); err != nil {
			log.Fatalf("AddContentTree: %s", err)
		}
		log.Debugf("new ContentTree created %s", ct.Uuid)
	}
	return
}
