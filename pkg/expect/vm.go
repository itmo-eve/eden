package expect

import (
	"github.com/lf-edge/eve/api/go/config"
	uuid "github.com/satori/go.uuid"
)

//createAppInstanceConfigVM creates appBundle for VM with provided img, netInstance, id and acls
//  it uses name of app and cpu/mem params from AppExpectation
//  it use ZArch param to choose VirtualizationMode
func (exp *AppExpectation) createAppInstanceConfigVM(contentTree *config.ContentTree, id uuid.UUID) *appBundle {
	app := &config.AppInstanceConfig{
		Uuidandversion: &config.UUIDandVersion{
			Uuid:    id.String(),
			Version: "1",
		},
		Fixedresources: &config.VmConfig{
			Memory: exp.mem,
			Maxmem: exp.mem,
			Vcpus:  exp.cpu,
		},
		Activate:    true,
		Displayname: exp.appName,
	}
	if exp.openStackMetadata {
		app.MetaDataType = config.MetaDataType_MetaDataOpenStack
	}
	exp.applyUserData(app)
	app.Fixedresources.VirtualizationMode = exp.virtualizationMode
	maxSizeBytes := int64(contentTree.MaxSizeBytes)
	if exp.diskSize > 0 {
		maxSizeBytes = exp.diskSize
	}
	contentTrees := []*config.ContentTree{contentTree}
	volume := exp.contentTreeToVolume(maxSizeBytes, 0, contentTree)
	volumes := []*config.Volume{volume}
	app.VolumeRefList = []*config.VolumeRef{{MountDir: "/", Uuid: volume.Uuid}}

	return &appBundle{
		appInstanceConfig: app,
		contentTrees:      contentTrees,
		volumes:           volumes,
	}
}
