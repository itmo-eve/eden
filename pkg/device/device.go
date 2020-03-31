package device

import (
	"github.com/lf-edge/eve/api/go/config"
	"github.com/satori/go.uuid"
)

//Ctx is base struct for device
type Ctx struct {
	id               *uuid.UUID
	devModel         *DevModel
	baseOSConfigs    []string
	networkInstances []string
}

//DevModel is type for determinate model of eve device
type DevModel struct {
	physIO   []*config.PhysicalIO
	networks []*config.NetworkConfig
	adapters []*config.SystemAdapter
}

//GetPhysicalIOs return PhysicalIO for model
func (ctx *DevModel) GetPhysicalIOs() []*config.PhysicalIO {
	return ctx.physIO
}

//GetNetworkConfigs return NetworkConfig for model
func (ctx *DevModel) GetNetworkConfigs() []*config.NetworkConfig {
	return ctx.networks
}

//GetSystemAdapters return SystemAdapter for model
func (ctx *DevModel) GetSystemAdapters() []*config.SystemAdapter {
	return ctx.adapters
}

const (
	netDHCPID   = "6822e35f-c1b8-43ca-b344-0bbc0ece8cf1"
	netNoDHCPID = "6822e35f-c1b8-43ca-b344-0bbc0ece8cf2"
)

var (
	//DevModelQemu is qemu device
	DevModelQemu = &DevModel{physIO: []*config.PhysicalIO{{
		Ptype:        config.PhyIoType_PhyIoNetEth,
		Phylabel:     "eth0",
		Logicallabel: "eth0",
		Usage:        config.PhyIoMemberUsage_PhyIoUsageMgmtAndApps,
		UsagePolicy: &config.PhyIOUsagePolicy{
			FreeUplink:       true,
			FallBackPriority: 0,
		},
	}, {
		Ptype:        config.PhyIoType_PhyIoNetEth,
		Phylabel:     "eth1",
		Logicallabel: "eth1",
		Usage:        config.PhyIoMemberUsage_PhyIoUsageMgmtAndApps,
		UsagePolicy: &config.PhyIOUsagePolicy{
			FreeUplink:       true,
			FallBackPriority: 0,
		},
	},
	},
		networks: []*config.NetworkConfig{
			{
				Id:   netDHCPID,
				Type: config.NetworkType_V4,
				Ip: &config.Ipspec{
					Dhcp:      config.DHCPType_Client,
					Subnet:    "",
					Gateway:   "",
					Domain:    "",
					Ntp:       "",
					Dns:       nil,
					DhcpRange: &config.IpRange{},
				},
				Dns:      nil,
				EntProxy: nil,
				Wireless: nil,
			},
			{
				Id:   netNoDHCPID,
				Type: config.NetworkType_V4,
				Ip: &config.Ipspec{
					Dhcp:      config.DHCPType_DHCPNone,
					Subnet:    "",
					Gateway:   "",
					Domain:    "",
					Ntp:       "",
					Dns:       nil,
					DhcpRange: &config.IpRange{},
				},
				Dns:      nil,
				EntProxy: nil,
				Wireless: nil,
			},
		},
		adapters: []*config.SystemAdapter{
			{
				Name:           "eth0",
				FreeUplink:     true,
				Uplink:         true,
				NetworkUUID:    netDHCPID,
				Addr:           "",
				Alias:          "",
				LowerLayerName: "eth0",
			},
			{
				Name:           "eth1",
				FreeUplink:     true,
				Uplink:         true,
				NetworkUUID:    netNoDHCPID,
				Addr:           "",
				Alias:          "",
				LowerLayerName: "eth1",
			},
		}}
)

//CreateWithBaseConfig generate base config for device with id and associate with cloudCtx
func CreateWithBaseConfig(id *uuid.UUID, devModel *DevModel) *Ctx {
	return &Ctx{
		id:       id,
		devModel: devModel,
	}
}

//GetID return id of device
func (cfg *Ctx) GetID() *uuid.UUID { return cfg.id }

//GetBaseOSConfigs return baseOSConfigs of device
func (cfg *Ctx) GetBaseOSConfigs() []string { return cfg.baseOSConfigs }

//GetNetworkInstances return networkInstances of device
func (cfg *Ctx) GetNetworkInstances() []string { return cfg.networkInstances }

//GetDevModel return devModel of device
func (cfg *Ctx) GetDevModel() *DevModel { return cfg.devModel }
