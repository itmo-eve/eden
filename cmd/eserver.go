package cmd

import (
	"fmt"
	"os"
	"reflect"

	"github.com/lf-edge/eden/pkg/defaults"
	"github.com/lf-edge/eden/pkg/eden"
	"github.com/lf-edge/eden/pkg/openevec"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newEserverCmd(configName, verbosity *string) *cobra.Command {
	cfg := &openevec.EdenSetupArgs{}
	var eserverCmd = &cobra.Command{
		Use: "eserver",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			viper_cfg, err := openevec.FromViper(*configName, *verbosity)
			if err != nil {
				return err
			}
			openevec.Merge(reflect.ValueOf(viper_cfg).Elem(), reflect.ValueOf(*cfg), cmd.Flags())
			cfg = viper_cfg
			return nil
		},
	}

	groups := CommandGroups{
		{
			Message: "Basic Commands",
			Commands: []*cobra.Command{
				newStartEserverCmd(cfg),
				newStopEserverCmd(cfg),
				newStatusEserverCmd(cfg),
			},
		},
	}

	groups.AddTo(eserverCmd)

	return eserverCmd
}

func newStartEserverCmd(cfg *openevec.EdenSetupArgs) *cobra.Command {

	var startEserverCmd = &cobra.Command{
		Use:   "start",
		Short: "start eserver",
		Long:  `Start eserver.`,
		Run: func(cmd *cobra.Command, args []string) {
			command, err := os.Executable()
			if err != nil {
				log.Fatalf("cannot obtain executable path: %s", err)
			}
			log.Infof("Executable path: %s", command)

			if err := eden.StartEServer(cfg.Eden.Eserver.Port, cfg.Eden.Images.EserverImageDist, cfg.Eden.Eserver.Force, cfg.Eden.Eserver.Tag); err != nil {
				log.Errorf("cannot start eserver: %s", err)
			} else {
				log.Infof("Eserver is running and accesible on port %d", cfg.Eden.Eserver.Port)
			}
		},
	}

	startEserverCmd.Flags().StringVarP(&cfg.Eden.Images.EserverImageDist, "image-dist", "", "", "image dist for eserver")
	startEserverCmd.Flags().IntVarP(&cfg.Eden.Eserver.Port, "eserver-port", "", defaults.DefaultEserverPort, "eserver port")
	startEserverCmd.Flags().StringVarP(&cfg.Eden.Eserver.Tag, "eserver-tag", "", defaults.DefaultEServerTag, "tag of eserver container to pull")
	startEserverCmd.Flags().BoolVarP(&cfg.Eden.Eserver.Force, "eserver-force", "", false, "eserver force rebuild")

	return startEserverCmd
}

func newStopEserverCmd(cfg *openevec.EdenSetupArgs) *cobra.Command {
	var stopEserverCmd = &cobra.Command{
		Use:   "stop",
		Short: "stop eserver",
		Long:  `Stop eserver.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := eden.StopEServer(cfg.Runtime.EserverRm); err != nil {
				log.Errorf("cannot stop eserver: %s", err)
			}
		},
	}

	stopEserverCmd.Flags().BoolVarP(&cfg.Runtime.EserverRm, "eserver-rm", "", false, "eserver rm on stop")

	return stopEserverCmd
}

func newStatusEserverCmd(cfg *openevec.EdenSetupArgs) *cobra.Command {
	var statusEserverCmd = &cobra.Command{
		Use:   "status",
		Short: "status of eserver",
		Long:  `Status of eserver.`,
		Run: func(cmd *cobra.Command, args []string) {
			statusEServer, err := eden.StatusEServer()
			if err != nil {
				log.Errorf("cannot obtain status of eserver: %s", err)
			} else {
				fmt.Printf("EServer status: %s\n", statusEServer)
			}
		},
	}
	return statusEserverCmd
}
