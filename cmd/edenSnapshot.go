package cmd

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/docker/docker/pkg/namesgenerator"
	"github.com/lf-edge/eden/pkg/eve"
	"github.com/lf-edge/eden/pkg/utils"
	"github.com/lf-edge/eve/api/go/config"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	snapshotName string
)

var snapshotCmd = &cobra.Command{
	Use: "snapshot",
}

//snapshotLsCmd is a command to list deployed snapshots
var snapshotLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List snapshots",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		assignCobraToViper(cmd)
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
		state := eve.Init(ctrl, dev)
		if err := ctrl.MetricLastCallback(dev.GetID(), nil, state.MetricCallback()); err != nil {
			log.Fatalf("fail in get InfoLastCallback: %s", err)
		}
		if err := ctrl.InfoLastCallback(dev.GetID(), nil, state.InfoCallback()); err != nil {
			log.Fatalf("fail in get InfoLastCallback: %s", err)
		}
		if err := state.SnapshotList(); err != nil {
			log.Fatal(err)
		}
	},
}

//snapshotCreateCmd is a command to create snapshot
var snapshotCreateCmd = &cobra.Command{
	Use:   "create <volume name>",
	Short: "Create snapshot for volume",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		assignCobraToViper(cmd)
		_, err := utils.LoadConfigFile(configFile)
		if err != nil {
			return fmt.Errorf("error reading config: %s", err.Error())
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		volumeName := args[0]
		changer := &adamChanger{}
		ctrl, dev, err := changer.getControllerAndDev()
		if err != nil {
			log.Fatalf("getControllerAndDev: %s", err)
		}
		id, err := uuid.NewV4()
		if err != nil {
			log.Fatal(err)
		}
		if snapshotName == "" {
			//generate random name
			rand.Seed(time.Now().UnixNano())
			snapshotName = namesgenerator.GetRandomName(0)
		}
		volumeToUse := ""
		for _, volumeId := range dev.GetVolumes() {
			if v, err := ctrl.GetVolume(volumeId); err == nil {
				if v.DisplayName == volumeName {
					volumeToUse = volumeId
					break
				}
			}
		}
		if volumeToUse == "" {
			log.Fatalf("no volume with name %s found", volumeName)
		}
		snapshot := &config.SnapshotConfig{
			Uuid:        id.String(),
			VolumeUuid:  volumeToUse,
			DisplayName: snapshotName,
		}
		_ = ctrl.AddSnapshot(snapshot)
		dev.SetSnapshotConfigs(append(dev.GetSnapshots(), id.String()))
		if err = changer.setControllerAndDev(ctrl, dev); err != nil {
			log.Fatalf("setControllerAndDev: %s", err)
		}
		log.Infof("snapshot %s create done", snapshotName)
	},
}

//snapshotDeleteCmd is a command to delete snapshot
var snapshotDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete snapshot",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		assignCobraToViper(cmd)
		_, err := utils.LoadConfigFile(configFile)
		if err != nil {
			return fmt.Errorf("error reading config: %s", err.Error())
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		snapshotName = args[0]
		changer := &adamChanger{}
		ctrl, dev, err := changer.getControllerAndDev()
		if err != nil {
			log.Fatalf("getControllerAndDev: %s", err)
		}
		for id, el := range dev.GetSnapshots() {
			snapshot, err := ctrl.GetSnapshot(el)
			if err != nil {
				log.Fatalf("no snapshot in cloud %s: %s", el, err)
			}
			if snapshot.DisplayName == snapshotName {
				configs := dev.GetSnapshots()
				utils.DelEleInSlice(&configs, id)
				dev.SetSnapshotConfigs(configs)
				if err = changer.setControllerAndDev(ctrl, dev); err != nil {
					log.Fatalf("setControllerAndDev: %s", err)
				}
				log.Infof("snapshot %s delete done", snapshotName)
				return
			}
		}
		log.Infof("not found snapshot with name %s", snapshotName)
	},
}

//snapshotRollbackCmd is a command to rollback to snapshot
var snapshotRollbackCmd = &cobra.Command{
	Use:   "rollback <snapshot name>",
	Short: "Rollback to snapshot",
	Args:  cobra.MinimumNArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		assignCobraToViper(cmd)
		_, err := utils.LoadConfigFile(configFile)
		if err != nil {
			return fmt.Errorf("error reading config: %s", err.Error())
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		snapshotName = args[0]
		changer := &adamChanger{}
		ctrl, dev, err := changer.getControllerAndDev()
		if err != nil {
			log.Fatalf("getControllerAndDev: %s", err)
		}
		for _, el := range dev.GetSnapshots() {
			snapshot, err := ctrl.GetSnapshot(el)
			if err != nil {
				log.Fatalf("no snapshot in cloud %s: %s", el, err)
			}
			if snapshot.DisplayName == snapshotName {
				rollbackCounter := uint32(1)
				if snapshot.Rollback != nil {
					rollbackCounter = snapshot.Rollback.Counter + 1
				}
				snapshot.Rollback = &config.DeviceOpsCmd{Counter: rollbackCounter}
				if err = changer.setControllerAndDev(ctrl, dev); err != nil {
					log.Fatalf("setControllerAndDev: %s", err)
				}
				log.Infof("rollback command send for name %s", snapshotName)
			}
		}
		log.Infof("not found snapshot with name %s", volumeName)
	},
}

func snapshotInit() {
	snapshotCmd.AddCommand(snapshotLsCmd)

	snapshotCmd.AddCommand(snapshotCreateCmd)
	snapshotCreateCmd.Flags().StringVarP(&snapshotName, "name", "n", "", "name of volume, random if empty")

	snapshotCmd.AddCommand(snapshotDeleteCmd)
	snapshotCmd.AddCommand(snapshotRollbackCmd)
}
