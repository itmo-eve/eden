package controller

import (
	"errors"
	"github.com/lf-edge/eden/pkg/utils"
	"github.com/lf-edge/eve/api/go/config"
)

func (cloud *CloudCtx) getNetworkInd(id string) (networkConfigInd int, err error) {
	for ind, el := range cloud.networks {
		if el != nil && el.Id == id {
			return ind, nil
		}
	}
	return -1, errors.New("not found")
}

//GetNetworkConfig return NetworkConfig config from cloud by ID
func (cloud *CloudCtx) GetNetworkConfig(id string) (networkConfig *config.NetworkConfig, err error) {
	networkInstanceConfigInd, err := cloud.getNetworkInstanceInd(id)
	if err != nil {
		return nil, err
	}
	return cloud.networks[networkInstanceConfigInd], nil
}

//AddNetworkConfig add NetworkConfig config to cloud
func (cloud *CloudCtx) AddNetworkConfig(networkInstanceConfig *config.NetworkConfig) error {
	cloud.networks = append(cloud.networks, networkInstanceConfig)
	return nil
}

//RemoveNetworkConfig remove NetworkConfig config to cloud
func (cloud *CloudCtx) RemoveNetworkConfig(id string) error {
	networkConfigInd, err := cloud.getNetworkInd(id)
	if err != nil {
		return err
	}
	utils.DelEleInSlice(cloud.networks, networkConfigInd)
	return nil
}
