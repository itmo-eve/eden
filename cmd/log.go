package cmd

import (
	"fmt"
	"github.com/lf-edge/eden/pkg/controller"
	"github.com/lf-edge/eden/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"strings"

	"github.com/lf-edge/eden/pkg/controller/elog"
	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:   "log [field:regexp ...]",
	Short: "Get logs from a running EVE device",
	Long: `
Scans the ADAM logs for correspondence with regular expressions requests to json fields.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		assignCobraToViper(cmd)
		viperLoaded, err := utils.LoadConfigFile(configFile)
		if err != nil {
			return fmt.Errorf("error reading config: %s", err.Error())
		}
		if viperLoaded {
			certsIP = viper.GetString("adam.ip")
			adamPort = viper.GetInt("adam.port")
			adamDist = utils.ResolveAbsPath(viper.GetString("adam.dist"))
			adamCA = utils.ResolveAbsPath(viper.GetString("adam.ca"))
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctrl, err := controller.CloudPrepare()
		if err != nil {
			log.Fatalf("CloudPrepare: %s", err)
		}
		devFirst, err := ctrl.GetDeviceFirst()
		if err != nil {
			log.Fatalf("GetDeviceFirst error: %s", err)
		}
		devUUID := devFirst.GetID()
		follow, err := cmd.Flags().GetBool("follow")
		if err != nil {
			log.Fatalf("Error in get param 'follow'")
		}

		q := make(map[string]string)

		for _, a := range args[0:] {
			s := strings.Split(a, ":")
			q[s[0]] = s[1]
		}

		handleFunc := func(le *elog.LogItem) bool {
			if printFields == nil {
				elog.LogPrn(le)
			} else {
				elog.LogItemPrint(le, printFields).Print()
			}
			return false
		}

		if follow {
			// Monitoring of new files
			if err = ctrl.LogChecker(devUUID, q, handleFunc, elog.LogNew, 0); err != nil {
				log.Fatalf("LogChecker: %s", err)
			}
		} else {
			if err = ctrl.LogLastCallback(devUUID, q, handleFunc); err != nil {
				log.Fatalf("LogChecker: %s", err)
			}
		}
	},
}

func logInit() {
	logCmd.Flags().StringSliceVarP(&printFields, "out", "o", nil, "Fields to print. Whole message if empty.")
	logCmd.Flags().BoolP("follow", "f", false, "Monitor changes in selected directory")
}
