// Copyright (c) 2020 Zededa, Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"reflect"

	"github.com/lf-edge/eden/pkg/defaults"
	"github.com/lf-edge/eden/pkg/openevec"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewEdenCommand() *cobra.Command {
	var configName, verbosity string

	rootCmd := &cobra.Command{
		Use: "eden",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return openevec.SetUpLogs(verbosity)
		},
	}

	baseCmd := baseCmd{
		configName: &configName,
		verbosity:  &verbosity,
	}

	groups := CommandGroups{
		{
			Message: "Basic Commands",
			Commands: []*cobra.Command{
				baseCmd.newSetupCmd(),
				baseCmd.newStartCmd(),
				baseCmd.newEveCmd(),
				baseCmd.newPodCmd(),
				baseCmd.newStatusCmd(),
				baseCmd.newStopCmd(),
				baseCmd.newCleanCmd(),
				baseCmd.newConfigCmd(),
				baseCmd.newSdnCmd(),
			},
		},
		{
			Message: "Advanced Commands",
			Commands: []*cobra.Command{
				baseCmd.newInfoCmd(),
				baseCmd.newLogCmd(),
				baseCmd.newNetStatCmd(),
				baseCmd.newMetricCmd(),
				baseCmd.newAdamCmd(),
				baseCmd.newRegistryCmd(),
				baseCmd.newRedisCmd(),
				baseCmd.newEserverCmd(),
				baseCmd.newTestCmd(),
				baseCmd.newUtilsCmd(),
				baseCmd.newControllerCmd(),
				baseCmd.newNetworkCmd(),
				baseCmd.newVolumeCmd(),
				baseCmd.newDisksCmd(),
				baseCmd.newPacketCmd(),
				baseCmd.newRolCmd(),
			},
		},
	}

	groups.AddTo(rootCmd)

	rootCmd.PersistentFlags().StringVar(&configName, "config", defaults.DefaultContext, "Name of config")
	rootCmd.PersistentFlags().StringVarP(&verbosity, "verbosity", "v", log.InfoLevel.String(), "Log level (debug, info, warn, error, fatal, panic")

	return rootCmd
}

type baseCmd struct {
	configName *string
	verbosity  *string
}

func (base baseCmd) preRunViperLoadFunction(cfg *openevec.EdenSetupArgs) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		viperCfg, err := openevec.FromViper(*base.configName, *base.verbosity)
		if err != nil {
			return err
		}
		openevec.Merge(reflect.ValueOf(viperCfg).Elem(), reflect.ValueOf(*cfg), cmd.Flags())
		*cfg = *viperCfg
		return nil
	}
}

// Execute primary function for cobra
func Execute() {
	rootCmd := NewEdenCommand()
	_ = rootCmd.Execute()
}
