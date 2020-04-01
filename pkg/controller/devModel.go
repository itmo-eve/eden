package controller

import (
	"fmt"
	"github.com/lf-edge/eve/api/go/config"
)

//DevModelType is type of dev model
type DevModelType string

//DevModel is dev model fields
type DevModel struct {
	//physicalIOs is PhysicalIO slice for DevModel
	physicalIOs []*config.PhysicalIO
	//networks is NetworkConfig slice for DevModel
	networks []*config.NetworkConfig
	//adapters is SystemAdapter slice for DevModel
	adapters []*config.SystemAdapter
	//adapterForSwitches is name of adapter for use in switch
	adapterForSwitches []string
}

//GetFirstAdapterForSwitches return first adapter available for switch networkInstance
func (ctx *DevModel) GetFirstAdapterForSwitches() string {
	if len(ctx.adapterForSwitches) > 0 {
		return ctx.adapterForSwitches[0]
	}
	return "uplink"
}

const (
	netDHCPID   = "6822e35f-c1b8-43ca-b344-0bbc0ece8cf1"
	netNoDHCPID = "6822e35f-c1b8-43ca-b344-0bbc0ece8cf2"
)

//DevModelTypeEmpty is empty model type
const DevModelTypeEmpty DevModelType = "Empty"

//DevModelTypeQemu is model type for qemu
const DevModelTypeQemu DevModelType = "Qemu"

//CreateDevModel create manual DevModel with provided params
func (cloud *CloudCtx) CreateDevModel(PhysicalIOs []*config.PhysicalIO, Networks []*config.NetworkConfig, Adapters []*config.SystemAdapter, AdapterForSwitches []string) *DevModel {
	return &DevModel{adapterForSwitches: AdapterForSwitches, physicalIOs: PhysicalIOs, networks: Networks, adapters: Adapters}
}

//GetDevModel return DevModel object by DevModelType
func (cloud *CloudCtx) GetDevModel(devModelType DevModelType) (*DevModel, error) {
	switch devModelType {
	case DevModelTypeEmpty:
		return cloud.CreateDevModel(nil, nil, nil, nil), nil
	case DevModelTypeQemu:
		return cloud.CreateDevModel([]*config.PhysicalIO{{
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
			[]*config.NetworkConfig{
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
			[]*config.SystemAdapter{
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
			}, []string{"eth1"}), nil
	}
	return nil, fmt.Errorf("not implemented type: %s", devModelType)
}
