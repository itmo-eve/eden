package expect

import (
	"path/filepath"

	"github.com/dustin/go-humanize"
	"github.com/lf-edge/eden/pkg/defaults"
	"github.com/lf-edge/eden/pkg/eden"
	"github.com/lf-edge/eden/pkg/utils"
	"github.com/lf-edge/eve/api/go/config"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

//createContentTreeFile uploads image into EServer from file and calculates size and sha256 of image
func (exp *AppExpectation) createContentTreeFile(id uuid.UUID, dsID string) *config.ContentTree {
	server := &eden.EServer{
		EServerIP:   exp.ctrl.GetVars().EServerIP,
		EServerPort: exp.ctrl.GetVars().EServerPort,
	}
	var fileSize int64
	sha256 := ""
	filePath := ""
	status := server.EServerCheckStatus(filepath.Base(exp.appURL))
	if !status.ISReady || status.Size != utils.GetFileSize(exp.appURL) || status.Sha256 != utils.SHA256SUM(exp.appURL) {
		log.Infof("Start uploading into eserver of %s", exp.appLink)
		status = server.EServerAddFile(exp.appURL, "")
		if status.Error != "" {
			log.Error(status.Error)
		}
	}
	sha256 = status.Sha256
	fileSize = status.Size
	filePath = status.FileName
	log.Infof("ContentTree uploaded with size %s and sha256 %s", humanize.Bytes(uint64(status.Size)), status.Sha256)
	if filePath == "" {
		log.Fatal("Not uploaded")
	}
	if exp.sftpLoad {
		filePath = filepath.Join(defaults.DefaultSFTPDirPrefix, filePath)
	}
	return &config.ContentTree{
		Uuid:         id.String(),
		DisplayName:  filePath,
		URL:          filePath,
		Iformat:      exp.imageFormatEnum(),
		DsId:         dsID,
		MaxSizeBytes: uint64(fileSize),
		Sha256:       sha256,
	}
}
