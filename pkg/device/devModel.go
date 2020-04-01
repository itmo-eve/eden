package device

import "github.com/lf-edge/eve/api/go/config"

//DevModel is type for determinate model of eve device
type DevModel struct {
	physIO           []*config.PhysicalIO
	networks         []*config.NetworkConfig
	adapters         []*config.SystemAdapter
	adapterForSwitch string
}

//GetPhysicalIOs return PhysicalIO for device model
func (ctx *DevModel) GetPhysicalIOs() []*config.PhysicalIO {
	return ctx.physIO
}

//GetNetworkConfigs return NetworkConfig for device model
func (ctx *DevModel) GetNetworkConfigs() []*config.NetworkConfig {
	return ctx.networks
}

//GetSystemAdapters return SystemAdapter for device model
func (ctx *DevModel) GetSystemAdapters() []*config.SystemAdapter {
	return ctx.adapters
}

//GetAdapterForSwitch return adapter name for use in switch for device model
func (ctx *DevModel) GetAdapterForSwitch() string {
	return ctx.adapterForSwitch
}

const (
	netDHCPID   = "6822e35f-c1b8-43ca-b344-0bbc0ece8cf1"
	netNoDHCPID = "6822e35f-c1b8-43ca-b344-0bbc0ece8cf2"
)

var (
	//DevModelQemu is qemu device
	DevModelQemu = &DevModel{adapterForSwitch: "eth1",
		physIO: []*config.PhysicalIO{{
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
				Uplink:         true,
				NetworkUUID:    netDHCPID,
				Addr:           "",
				Alias:          "",
				LowerLayerName: "eth0",
			},
			{
				Name:           "eth1",
				Uplink:         true,
				NetworkUUID:    netNoDHCPID,
				Addr:           "",
				Alias:          "",
				LowerLayerName: "eth1",
			},
		}}
)
