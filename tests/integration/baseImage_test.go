package integration

import (
	"fmt"
	"github.com/lf-edge/eden/pkg/controller"
	"github.com/lf-edge/eden/pkg/controller/einfo"
	"github.com/lf-edge/eden/pkg/utils"
	"github.com/lf-edge/eve/api/go/config"
	"os"
	"path"
	"path/filepath"
	"testing"
)

func prepareBaseImageLocal(ctx controller.Cloud, dataStoreID string, imageID string, baseID string, imagePath string, baseOSVersion string) error {
	fi, err := os.Stat(imagePath)
	if err != nil {
		return err
	}
	size := fi.Size()

	sha256sum, err := utils.SHA256SUM(imagePath)
	if err != nil {
		return err
	}
	err = ctx.AddDataStore(&config.DatastoreConfig{
		Id:       dataStoreID,
		DType:    config.DsType_DsHttp,
		Fqdn:     "http://mydomain.adam:8888",
		ApiKey:   "",
		Password: "",
		Dpath:    "",
		Region:   "",
	})
	if err != nil {
		return err
	}
	img := &config.Image{
		Uuidandversion: &config.UUIDandVersion{
			Uuid:    imageID,
			Version: "4",
		},
		Name:      filepath.Base(imagePath),
		Sha256:    sha256sum,
		Iformat:   config.Format_QCOW2,
		DsId:      dataStoreID,
		SizeBytes: size,
		Siginfo: &config.SignatureInfo{
			Intercertsurl: "",
			Signercerturl: "",
			Signature:     nil,
		},
	}
	err = ctx.AddImage(img)
	if err != nil {
		return err
	}
	err = ctx.AddBaseOsConfig(&config.BaseOSConfig{
		Uuidandversion: &config.UUIDandVersion{
			Uuid:    baseID,
			Version: "4",
		},
		Drives: []*config.Drive{{
			Image:        img,
			Readonly:     false,
			Preserve:     false,
			Drvtype:      config.DriveType_Unclassified,
			Target:       config.Target_TgtUnknown,
			Maxsizebytes: size,
		}},
		Activate:      true,
		BaseOSVersion: baseOSVersion,
		BaseOSDetails: nil,
	})
	if err != nil {
		return err
	}
	return nil
}

//TestBaseImage test base image loading into eve
//environment variable EVE_BASE_REF - version of eve image
//environment variable ZARCH - architecture of eve image
func TestBaseImage(t *testing.T) {
	ctx, err := controllerPrepare()
	if err != nil {
		t.Fatal("Fail in controller prepare: ", err)
	}
	eveBaseRef := os.Getenv("EVE_BASE_REF")
	if len(eveBaseRef) == 0 {
		eveBaseRef = "4.10.0"
	}
	zArch := os.Getenv("ZARCH")
	if len(eveBaseRef) == 0 {
		zArch = "amd64"
	}
	var baseImageTests = []struct {
		dataStoreID string
		imageID     string
		baseID      string
		imagePath   string
		eveBaseRef  string
		zArch       string
	}{
		{"eab8761b-5f89-4e0b-b757-4b87a9fa93ec",

			"1ab8761b-5f89-4e0b-b757-4b87a9fa93ec",

			"22b8761b-5f89-4e0b-b757-4b87a9fa93ec",

			path.Join(filepath.Dir(ctx.GetDir()), "images", "baseos.qcow2"),
			eveBaseRef,
			zArch,
		},
	}
	for _, tt := range baseImageTests {
		baseOSVersion := fmt.Sprintf("%s-%s", tt.eveBaseRef, tt.zArch)
		t.Run(baseOSVersion, func(t *testing.T) {

			err = prepareBaseImageLocal(ctx, tt.dataStoreID, tt.imageID, tt.baseID, tt.imagePath, baseOSVersion)

			if err != nil {
				t.Fatal("Fail in prepare base image from local file: ", err)
			}
			deviceCtx, err := ctx.GetDeviceFirst()
			if err != nil {
				t.Fatal("Fail in get first device: ", err)
			}
			deviceCtx.SetBaseOSConfig([]string{tt.baseID})
			devUUID := deviceCtx.GetID()
			err = ctx.ConfigSync(deviceCtx)
			if err != nil {
				t.Fatal("Fail in sync config with controller: ", err)
			}
			t.Run("Started", func(t *testing.T) {
				err := ctx.InfoChecker(devUUID, map[string]string{"devId": devUUID.String(), "shortVersion": baseOSVersion}, einfo.ZInfoDevSW, 300)
				if err != nil {
					t.Fatal("Fail in waiting for base image update init: ", err)
				}
			})
			t.Run("Downloaded", func(t *testing.T) {
				err := ctx.InfoChecker(devUUID, map[string]string{"devId": devUUID.String(), "shortVersion": baseOSVersion, "downloadProgress": "100"}, einfo.ZInfoDevSW, 1500)
				if err != nil {
					t.Fatal("Fail in waiting for base image download progress: ", err)
				}
			})
			t.Run("Logs", func(t *testing.T) {
				err = ctx.LogChecker(devUUID, map[string]string{"devId": devUUID.String(), "eveVersion": baseOSVersion}, 1200)
				if err != nil {
					t.Fatal("Fail in waiting for base image logs: ", err)
				}
			})
			t.Run("Active", func(t *testing.T) {
				err = ctx.InfoChecker(devUUID, map[string]string{"devId": devUUID.String(), "shortVersion": baseOSVersion, "status": "INSTALLED"}, einfo.ZInfoDevSW, 1200)
				if err != nil {
					t.Fatal("Fail in waiting for base image installed status: ", err)
				}
			})
		})
	}

}
