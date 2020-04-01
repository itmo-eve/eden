package device

import (
	"github.com/satori/go.uuid"
)

//Ctx is base struct for device
type Ctx struct {
	id                uuid.UUID
	baseOSConfigs     []string
	networkInstances  []string
	adaptersForSwitch []string
	networks          []string
	physicalIO        []string
	systemAdapters    []string
}

//CreateWithBaseConfig generate base config for device with id and associate with cloudCtx
func CreateWithBaseConfig(id uuid.UUID) *Ctx {
	return &Ctx{
		id: id,
	}
}

//GetID return id of device
func (cfg *Ctx) GetID() uuid.UUID { return cfg.id }

//GetBaseOSConfigs return baseOSConfigs of device
func (cfg *Ctx) GetBaseOSConfigs() []string { return cfg.baseOSConfigs }

//GetNetworkInstances return networkInstances of device
func (cfg *Ctx) GetNetworkInstances() []string { return cfg.networkInstances }

//GetNetworks return networks of device
func (cfg *Ctx) GetNetworks() []string { return cfg.networks }

//GetPhysicalIOs return physicalIO of device
func (cfg *Ctx) GetPhysicalIOs() []string { return cfg.physicalIO }

//GetSystemAdapters return systemAdapters of device
func (cfg *Ctx) GetSystemAdapters() []string { return cfg.systemAdapters }

//GetAdaptersForSwitch return adaptersForSwitch of device
func (cfg *Ctx) GetAdaptersForSwitch() []string {
	return cfg.adaptersForSwitch
}

//SetAdaptersForSwitch set adaptersForSwitch of device
func (cfg *Ctx) SetAdaptersForSwitch(adaptersForSwitch []string) {
	cfg.adaptersForSwitch = adaptersForSwitch
}

//SetBaseOSConfig set BaseOSConfig by configIDs from cloud
func (cfg *Ctx) SetBaseOSConfig(configIDs []string) *Ctx {
	cfg.baseOSConfigs = configIDs
	return cfg
}

//SetNetworkInstanceConfig set NetworkInstanceConfig by configIDs from cloud
func (cfg *Ctx) SetNetworkInstanceConfig(configIDs []string) *Ctx {
	cfg.networkInstances = configIDs
	return cfg
}

//SetNetworkConfig set networks by configIDs from cloud
func (cfg *Ctx) SetNetworkConfig(configIDs []string) *Ctx {
	cfg.networks = configIDs
	return cfg
}

//SetPhysicalIOConfig set physicalIO by configIDs from cloud
func (cfg *Ctx) SetPhysicalIOConfig(configIDs []string) *Ctx {
	cfg.physicalIO = configIDs
	return cfg
}

//SetSystemAdaptersConfig set systemAdapters by configIDs from cloud
func (cfg *Ctx) SetSystemAdaptersConfig(configIDs []string) *Ctx {
	cfg.systemAdapters = configIDs
	return cfg
}
