package integration

import (
	"fmt"
	"github.com/lf-edge/eden/pkg/controller"
	"github.com/lf-edge/eden/pkg/controller/einfo"
	"github.com/lf-edge/eve/api/go/config"
	"github.com/pkg/errors"
	"testing"
)

const cloudOConfig = `{ "VpnRole": "onPremClient",
  "VpnGatewayIpAddr": "192.168.254.51",
  "VpnSubnetBlock": "20.1.0.0/24",
  "ClientConfigList": [{"IpAddr": "%any", "PreSharedKey": "0sjVzONCF02ncsgiSlmIXeqhGN", "SubnetBlock": "30.1.0.0/24"}]
}`

func prepareNetworkInstance(ctx controller.Cloud, networkInstanceID string, networkInstanceName string, networkInstanceType config.ZNetworkInstType, model *controller.DevModel) error {
	uid := config.UUIDandVersion{
		Uuid:    networkInstanceID,
		Version: "4",
	}
	networkInstance := config.NetworkInstanceConfig{
		Uuidandversion: &uid,
		Displayname:    networkInstanceName,
		InstType:       networkInstanceType,
		Activate:       true,
		Port:           nil,
		Cfg:            nil,
		IpType:         config.AddressType_First,
		Ip:             nil,
	}
	switch networkInstanceType {
	case config.ZNetworkInstType_ZnetInstSwitch:
		networkInstance.Port = &config.Adapter{
			Name: model.GetFirstAdapterForSwitches(),
		}
		networkInstance.Cfg = &config.NetworkInstanceOpaqueConfig{}
	case config.ZNetworkInstType_ZnetInstLocal:
		networkInstance.IpType = config.AddressType_IPV4
		networkInstance.Port = &config.Adapter{
			Name: "uplink",
		}
		networkInstance.Ip = &config.Ipspec{
			Subnet:  "10.1.0.0/24",
			Gateway: "10.1.0.1",
			Dns:     []string{"10.1.0.1"},
			DhcpRange: &config.IpRange{
				Start: "10.1.0.2",
				End:   "10.1.0.254",
			},
		}
		networkInstance.Cfg = &config.NetworkInstanceOpaqueConfig{}
	case config.ZNetworkInstType_ZnetInstCloud:
		networkInstance.IpType = config.AddressType_IPV4
		networkInstance.Port = &config.Adapter{
			Type: config.PhyIoType_PhyIoNoop,
			Name: "uplink",
		}
		networkInstance.Ip = &config.Ipspec{
			Dhcp:    config.DHCPType_DHCPNone,
			Subnet:  "30.1.0.0/24",
			Gateway: "30.1.0.1",
			Domain:  "",
			Ntp:     "",
			Dns:     []string{"30.1.0.1"},
			DhcpRange: &config.IpRange{
				Start: "30.1.0.2",
				End:   "30.1.0.254",
			},
		}
		networkInstance.Cfg = &config.NetworkInstanceOpaqueConfig{
			Oconfig:    cloudOConfig,
			LispConfig: nil,
			Type:       config.ZNetworkOpaqueConfigType_ZNetOConfigVPN,
		}
	default:
		return errors.New("not implemented type")
	}
	return ctx.AddNetworkInstanceConfig(&networkInstance)
}

//TestNetworkInstance test network instances creation in EVE
func TestNetworkInstance(t *testing.T) {
	ctx, err := controllerPrepare()
	if err != nil {
		t.Fatal("Fail in controller prepare: ", err)
	}

	deviceCtx, err := ctx.GetDeviceFirst()
	if err != nil {
		t.Fatal("Fail in get first device: ", err)
	}
	devModel, err := ctx.GetDevModelByName(deviceCtx.GetDevModel())
	if err != nil {
		t.Fatal("Fail in get dev model: ", err)
	}

	var networkInstances []string
	var networkInstanceTests = []struct {
		networkInstanceID   string
		networkInstanceName string
		networkInstanceType config.ZNetworkInstType
	}{
		{"eab8761b-5f89-4e0b-b757-4b87a9fa93e1",

			"testLocal",

			config.ZNetworkInstType_ZnetInstLocal,
		},
		{"eab8761b-5f89-4e0b-b757-4b87a9fa93e2",

			"testSwitch",

			config.ZNetworkInstType_ZnetInstSwitch,
		},
		{"eab8761b-5f89-4e0b-b757-4b87a9fa93e3",

			"testCloud",

			config.ZNetworkInstType_ZnetInstCloud,
		},
	}
	for _, tt := range networkInstanceTests {
		t.Run(tt.networkInstanceName, func(t *testing.T) {
			err = prepareNetworkInstance(ctx, tt.networkInstanceID, tt.networkInstanceName, tt.networkInstanceType, devModel)
			if err != nil {
				t.Fatal("Fail in prepare network instance: ", err)
			}

			devUUID := deviceCtx.GetID()
			//append networkInstance for run all of them together
			networkInstances = append(networkInstances, tt.networkInstanceID)
			deviceCtx.SetNetworkInstanceConfig(networkInstances)
			err = ctx.ConfigSync(deviceCtx)
			if err != nil {
				t.Fatal("Fail in sync config with controller: ", err)
			}
			t.Run("Process", func(t *testing.T) {
				err = ctx.InfoChecker(devUUID, map[string]string{"devId": devUUID.String(), "networkID": tt.networkInstanceID}, einfo.ZInfoNetworkInstance, 300)
				if err != nil {
					t.Fatal("Fail in waiting for process start from info: ", err)
				}
			})
			t.Run("Handled", func(t *testing.T) {
				err = ctx.LogChecker(devUUID, map[string]string{"devId": devUUID.String(), "msg": fmt.Sprintf(".*handleNetworkInstanceModify\\(%s\\) done.*", tt.networkInstanceID), "level": "info"}, 600)
				if err != nil {
					t.Fatal("Fail in waiting for handleNetworkInstanceModify done from zedagent: ", err)
				}
			})
			t.Run("Active", func(t *testing.T) {
				err = ctx.InfoChecker(devUUID, map[string]string{"devId": devUUID.String(), "networkID": tt.networkInstanceID, "activated": "true"}, einfo.ZInfoNetworkInstance, 200)
				if err != nil {
					t.Fatal("Fail in waiting for activated state from info: ", err)
				}
			})
		})
	}
}
