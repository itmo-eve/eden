package expect

import (
	"fmt"
	"github.com/lf-edge/eden/pkg/defaults"
	"github.com/lf-edge/eden/pkg/utils"
	"github.com/lf-edge/eve/api/go/config"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

//parse file or url name and returns Base OS Version
func (exp *AppExpectation) getBaseOSVersion() string {
	if exp.baseOSVersion != "" {
		return exp.baseOSVersion
	}
	if exp.appType == dockerApp {
		return exp.appVersion
	}

	correctionFileName := fmt.Sprintf("%s.ver", exp.appURL)
	if rootFSFromCorrectionFile, err := ioutil.ReadFile(correctionFileName); err == nil {
		return string(rootFSFromCorrectionFile)
	}
	rootFSName := path.Base(exp.appURL)
	rootFSName = strings.TrimSuffix(rootFSName, filepath.Ext(rootFSName))
	rootFSName = strings.TrimPrefix(rootFSName, "rootfs-")
	if re := regexp.MustCompile(defaults.DefaultRootFSVersionPattern); !re.MatchString(rootFSName) {
		log.Warnf("Filename of rootfs %s does not match pattern %s", rootFSName, defaults.DefaultRootFSVersionPattern)
		// check for eve_version file
		if v, err := ioutil.ReadFile(filepath.Join(filepath.Dir(exp.appURL), "eve_version")); err == nil {
			baseOSVersion := strings.TrimSpace(string(v))
			log.Warnf("Will use version from eve_version file: %s", baseOSVersion)
			return baseOSVersion
		}
		log.Fatalf("Cannot use provided file: version unknown, please provide it with --os-version flag")
	}
	return rootFSName
}

//checkBaseOSConfig checks if provided BaseOSConfig match expectation
func (exp *AppExpectation) checkBaseOSConfig(baseOS *config.BaseOSConfig) bool {
	if baseOS == nil {
		return false
	}
	if baseOS.BaseOSVersion == exp.getBaseOSVersion() {
		return true
	}
	return false
}

//createBaseOSConfig creates BaseOSConfig with provided img
func (exp *AppExpectation) createBaseOSConfig(volume *config.Volume) (*config.BaseOSConfig, error) {
	baseOSConfig := &config.BaseOSConfig{
		Uuidandversion: &config.UUIDandVersion{
			Uuid:    volume.Uuid,
			Version: "4",
		},
		VolumeID:      volume.Uuid,
		Activate:      true,
		BaseOSVersion: exp.getBaseOSVersion(),
	}
	return baseOSConfig, nil
}

//BaseOSImage expectation gets or creates ContentTree definition,
//gets BaseOSConfig and returns it or creates BaseOSConfig, adds it into internal controller and returns it
//if version is not empty will use it as BaseOSVersion
func (exp *AppExpectation) BaseOSImage(baseOSVersion string) (baseOSConfig *config.BaseOSConfig) {
	exp.baseOSVersion = baseOSVersion
	var err error
	if exp.appType == fileApp {
		if exp.appURL, err = utils.GetFileFollowLinks(exp.appURL); err != nil {
			log.Fatalf("GetFileFollowLinks: %s", err)
		}
	}
	contentTree := exp.ContentTree()
	for _, baseOS := range exp.ctrl.ListBaseOSConfig() {
		if exp.checkBaseOSConfig(baseOS) {
			baseOSConfig = baseOS
			break
		}
	}
	if baseOSConfig == nil { //if baseOSConfig not exists, create it
		for _, baseOS := range exp.ctrl.ListBaseOSConfig() {
			baseOS.Activate = false
		}
		_ = exp.ctrl.AddContentTree(contentTree)
		volume := exp.contentTreeToVolume(0, 0, contentTree)
		_ = exp.ctrl.AddVolume(volume)
		exp.device.SetVolumeConfigs(append(exp.device.GetVolumes(), volume.Uuid))
		if baseOSConfig, err = exp.createBaseOSConfig(volume); err != nil {
			log.Fatalf("cannot create baseOS: %s", err)
		}
		if err = exp.ctrl.AddBaseOsConfig(baseOSConfig); err != nil {
			log.Fatalf("AddBaseOsConfig: %s", err)
		}
		log.Infof("new base os created %s", baseOSConfig.Uuidandversion.Uuid)
	}

	return
}
