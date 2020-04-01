package controller

import (
	"encoding/json"
	"errors"
	"github.com/lf-edge/eden/pkg/device"
	"github.com/lf-edge/eve/api/go/config"
	uuid "github.com/satori/go.uuid"
)

//ConfigSync set config for devID
func (cloud *CloudCtx) ConfigSync(devUUID *uuid.UUID) (err error) {
	devConfig, err := cloud.GetConfigBytes(devUUID)
	if err != nil {
		return err
	}
	return cloud.ConfigSet(devUUID, devConfig)
}

//GetDeviceUUID return device object by devUUID
func (cloud *CloudCtx) GetDeviceUUID(devUUID *uuid.UUID) (dID *device.Ctx, err error) {
	for _, el := range cloud.devices {
		if devUUID.String() == el.GetID().String() {
			return el, nil
		}
	}
	return nil, errors.New("no device found")
}

//GetDeviceFirst return first device object
func (cloud *CloudCtx) GetDeviceFirst() (devUUID *device.Ctx, err error) {
	if len(cloud.devices) == 0 {
		return nil, errors.New("no device found")
	}
	return cloud.devices[0], nil
}

//AddDevice add device with specified devUUID
func (cloud *CloudCtx) AddDevice(devUUID *uuid.UUID) error {
	for _, el := range cloud.devices {
		if el.GetID().String() == devUUID.String() {
			return errors.New("already exists")
		}
	}
	cloud.devices = append(cloud.devices, device.CreateWithBaseConfig(devUUID))
	return nil
}

func (cloud *CloudCtx) ApplyDevModel(devUUID *uuid.UUID, devModel *DevModel) error {
	dev, err := cloud.GetDeviceUUID(devUUID)
	if err != nil {
		return err
	}
	dev.SetAdaptersForSwitch(devModel.adapterForSwitches)
	var adapters []string
	for _, el := range devModel.adapters {
		id := uuid.UUID{}.String()
		err = cloud.AddSystemAdapter(id, el)
		if err != nil {
			return err
		}
		adapters = append(adapters, id)
	}
	dev.SetSystemAdaptersConfig(adapters)
	var networks []string
	for _, el := range devModel.networks {
		err = cloud.AddNetworkConfig(el)
		if err != nil {
			return err
		}
		networks = append(networks, el.Id)
	}
	dev.SetNetworkConfig(networks)
	var physicalIOs []string
	for _, el := range devModel.physicalIOs {
		id := uuid.UUID{}.String()
		err = cloud.AddPhysicalIO(id, el)
		if err != nil {
			return err
		}
		physicalIOs = append(physicalIOs, id)
	}
	dev.SetPhysicalIOConfig(physicalIOs)
	return nil
}

func checkIfDatastoresContains(devUUID string, ds []*config.DatastoreConfig) bool {
	for _, el := range ds {
		if el.Id == devUUID {
			return true
		}
	}
	return false
}

//GetConfigBytes generate json representation of device config
func (cloud *CloudCtx) GetConfigBytes(devUUID *uuid.UUID) ([]byte, error) {
	dev, err := cloud.GetDeviceUUID(devUUID)
	if err != nil {
		return nil, err
	}
	var baseOS []*config.BaseOSConfig
	var dataStores []*config.DatastoreConfig
	for _, baseOSConfigID := range dev.GetBaseOSConfigs() {
		baseOSConfig, err := cloud.GetBaseOSConfig(baseOSConfigID)
		if err != nil {
			return nil, err
		}
		for _, drive := range baseOSConfig.Drives {
			if drive.Image == nil {
				return nil, errors.New("empty Image in Drive")
			}
			dataStore, err := cloud.GetDataStore(drive.Image.DsId)
			if err != nil {
				return nil, err
			}
			if !checkIfDatastoresContains(dataStore.Id, dataStores) {
				dataStores = append(dataStores, dataStore)
			}
		}
		baseOS = append(baseOS, baseOSConfig)
	}
	var networkInstanceConfigs []*config.NetworkInstanceConfig
	for _, networkInstanceConfigID := range dev.GetNetworkInstances() {
		networkInstanceConfig, err := cloud.GetNetworkInstanceConfig(networkInstanceConfigID)
		if err != nil {
			return nil, err
		}
		networkInstanceConfigs = append(networkInstanceConfigs, networkInstanceConfig)
	}
	var physicalIOs []*config.PhysicalIO
	for _, physicalIOID := range dev.GetPhysicalIOs() {
		physicalIOConfig, err := cloud.GetPhysicalIO(physicalIOID)
		if err != nil {
			return nil, err
		}
		physicalIOs = append(physicalIOs, physicalIOConfig)
	}
	var networkConfigs []*config.NetworkConfig
	for _, networkConfigID := range dev.GetNetworks() {
		networkConfig, err := cloud.GetNetworkConfig(networkConfigID)
		if err != nil {
			return nil, err
		}
		networkConfigs = append(networkConfigs, networkConfig)
	}
	var systemAdapterConfigs []*config.SystemAdapter
	for _, systemAdapterConfigID := range dev.GetSystemAdapters() {
		systemAdapterConfig, err := cloud.GetSystemAdapter(systemAdapterConfigID)
		if err != nil {
			return nil, err
		}
		systemAdapterConfigs = append(systemAdapterConfigs, systemAdapterConfig)
	}
	devConfig := &config.EdgeDevConfig{
		Id: &config.UUIDandVersion{
			Uuid:    dev.GetID().String(),
			Version: "4",
		},
		Apps:              nil,
		Networks:          networkConfigs,
		Datastores:        dataStores,
		LispInfo:          nil,
		Base:              baseOS,
		Reboot:            nil,
		Backup:            nil,
		ConfigItems:       nil,
		SystemAdapterList: systemAdapterConfigs,
		DeviceIoList:      physicalIOs,
		Manufacturer:      "",
		ProductName:       "",
		NetworkInstances:  networkInstanceConfigs,
		Enterprise:        "",
		Name:              "",
	}
	return json.Marshal(devConfig)
}
