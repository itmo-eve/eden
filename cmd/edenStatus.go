package cmd

import (
	"os"
	"path/filepath"
	"reflect"

	"github.com/lf-edge/eden/pkg/defaults"
	"github.com/lf-edge/eden/pkg/openevec"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newStatusCmd(configName, verbosity *string) *cobra.Command {
	cfg := &openevec.EdenSetupArgs{}
	var statusCmd = &cobra.Command{
		Use:   "status",
		Short: "status of harness",
		Long:  `Status of harness.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			viper_cfg, err := openevec.FromViper(*configName, *verbosity)
			if err != nil {
				return err
			}
			openevec.Merge(reflect.ValueOf(viper_cfg).Elem(), reflect.ValueOf(*cfg), cmd.Flags())
			cfg = viper_cfg
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			if err := openevec.Status(cfg); err != nil {
				log.Fatal(err)
			}
		},
	}

	currentPath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	statusCmd.Flags().StringVarP(&cfg.Eve.Pid, "eve-pid", "", filepath.Join(currentPath, defaults.DefaultDist, "eve.pid"), "file with EVE pid")
	statusCmd.Flags().BoolVar(&cfg.Runtime.AllConfigs, "all", true, "show status for all configs")
	statusCmd.Flags().StringVarP(&cfg.Runtime.VmName, "vmname", "", defaults.DefaultVBoxVMName, "vbox vmname required to create vm")

	addSdnPidOpt(statusCmd, cfg)
	addSdnPortOpts(statusCmd, cfg)

	return statusCmd

}
