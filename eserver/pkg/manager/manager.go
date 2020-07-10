package manager

import (
	"crypto/tls"
	"github.com/lf-edge/eden/eserver/api"
	"github.com/lf-edge/eden/pkg/utils"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

//EServerManager for process files
type EServerManager struct {
	Dir string
}

//Init directories for EServerManager
func (mgr *EServerManager) Init() {
	if _, err := os.Stat(mgr.Dir); err != nil {
		if err = os.MkdirAll(mgr.Dir, 0755); err != nil {
			log.Fatal(err)
		}
	}
}

//ListFileNames list downloaded files
func (mgr *EServerManager) ListFileNames() (result []string) {
	files, err := ioutil.ReadDir(mgr.Dir)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		result = append(result, f.Name())
	}
	return
}

//AddFile starts file download and return name of file for fileinfo requests
func (mgr *EServerManager) AddFile(url string) (string, error) {
	log.Println("Starting download of image from ", url)
	filePath := filepath.Join(mgr.Dir, path.Base(url))
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		log.Println("file already exists ", filePath)
	} else {
		go func() {
			http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
			if err := utils.DownloadFile(filePath, url); err != nil {
				log.Fatal(err)
			}
			log.Println("Download done for ", url)
		}()
	}
	return path.Base(url), nil
}

//GetFileInfo checks status of file and returns information
func (mgr *EServerManager) GetFileInfo(name string) (*api.FileInfo, error) {
	filePath := filepath.Join(mgr.Dir, name)
	filePathTMP := filePath + ".tmp"
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if _, err := os.Stat(filePathTMP); os.IsNotExist(err) {
			return nil, err
		} else {
			fileSize := utils.GetFileSize(filePathTMP)
			return &api.FileInfo{
				Size:    fileSize,
				ISReady: false,
			}, nil
		}
	} else {
		fileSize := utils.GetFileSize(filePath)
		sha256, err := utils.SHA256SUM(filePath)
		if err != nil {
			return nil, err
		}
		return &api.FileInfo{
			Sha256:   sha256,
			Size:     fileSize,
			FileName: path.Join("eserver", name),
			ISReady:  true,
		}, nil
	}
}

//GetFilePath returns path to file for serve
func (mgr *EServerManager) GetFilePath(name string) (string, error) {
	filePath := filepath.Join(mgr.Dir, name)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", err
	} else {
		return filePath, nil
	}
}
