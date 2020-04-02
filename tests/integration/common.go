package integration

import (
	"fmt"
	"github.com/lf-edge/eden/pkg/controller"
	"github.com/lf-edge/eden/pkg/controller/adam"
	"github.com/lf-edge/eden/pkg/utils"
	uuid "github.com/satori/go.uuid"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

var (
	adamIP   string
	adamPort string
	adamDir  string
	adamCA   string
	sshKey   string
)

//envRead use environment variables for init controller
//environment variable ADAM_IP - IP of adam
//environment variable ADAM_PORT - PORT of adam
//environment variable ADAM_DIST - directory of adam (absolute path)
//environment variable ADAM_CA - CA file of adam for https
//environment variable SSH_KEY - ssh public key for integrate into eve
func envRead() error {
	currentPath, err := os.Getwd()
	adamIP = os.Getenv("ADAM_IP")
	if len(adamIP) == 0 {
		adamIP, err = utils.GetIPForDockerAccess()
		if err != nil {
			return err
		}
	}
	adamPort = os.Getenv("ADAM_PORT")
	if len(adamPort) == 0 {
		adamPort = "3333"
	}
	adamDir = os.Getenv("ADAM_DIST")
	if len(adamDir) == 0 {
		adamDir = path.Join(filepath.Dir(filepath.Dir(currentPath)), "dist", "adam")
		if stat, err := os.Stat(adamDir); err != nil || !stat.IsDir() {
			return err
		}
	}

	adamCA = os.Getenv("ADAM_CA")
	sshKey = os.Getenv("SSH_KEY")
	return nil
}

//controllerPrepare is for init controller connection and obtain device list
func controllerPrepare() (ctx controller.Cloud, err error) {
	err = envRead()
	if err != nil {
		return ctx, err
	}
	var ctrl controller.Cloud = &controller.CloudCtx{Controller: &adam.Ctx{
		Dir:         adamDir,
		URL:         fmt.Sprintf("https://%s:%s", adamIP, adamPort),
		InsecureTLS: true,
	}}
	if len(adamCA) != 0 {
		ctrl = &controller.CloudCtx{Controller: &adam.Ctx{
			Dir:         adamDir,
			URL:         fmt.Sprintf("https://%s:%s", adamIP, adamPort),
			InsecureTLS: false,
			ServerCA:    adamCA,
		}}
	}
	devices, err := ctrl.DeviceList()
	if err != nil {
		return ctrl, err
	}
	for _, devID := range devices {
		devUUID, err := uuid.FromString(devID)
		if err != nil {
			return ctrl, err
		}
		dev, err := ctrl.AddDevice(devUUID)
		if err != nil {
			return ctrl, err
		}
		if sshKey != "" {
			b, err := ioutil.ReadFile(sshKey)
			switch {
			case err != nil && os.IsNotExist(err):
				return nil, fmt.Errorf("sshKey file %s does not exist", sshKey)
			case err != nil:
				return nil, fmt.Errorf("error reading sshKey file %s: %v", sshKey, err)
			}
			dev.SetSSHKeys([]string{string(b)})
		}
		deviceModel, err := ctrl.GetDevModel(controller.DevModelTypeQemu)
		if err != nil {
			return ctrl, fmt.Errorf("fail in get deviceModel: %s", err)
		}
		err = ctrl.ApplyDevModel(dev, deviceModel)
		if err != nil {
			return ctrl, fmt.Errorf("fail in ApplyDevModel: %s", err)
		}
	}
	return ctrl, nil
}
