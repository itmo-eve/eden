package cmd

import (
	"reflect"

	log "github.com/sirupsen/logrus"

	"github.com/lf-edge/eden/pkg/controller/types"
	"github.com/lf-edge/eden/pkg/openevec"

	"github.com/spf13/cobra"
	"github.com/thediveo/enumflag"
)

func newInfoCmd(configName, verbosity *string) *cobra.Command {
	cfg := &openevec.EdenSetupArgs{}
	var outputFormat types.OutputFormat
	var infoCmd = &cobra.Command{
		Use:   "info [field:regexp ...]",
		Short: "Get information reports from a running EVE device",
		Long: `
Scans the ADAM Info for correspondence with regular expressions requests to json fields.`,
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
			if err := openevec.EdenInfo(cfg, outputFormat, args); err != nil {
				log.Fatal("Eden info failed ", err)
			}
		},
	}

	infoCmd.Flags().UintVar(&cfg.Runtime.InfoTail, "tail", 0, "Show only last N lines")
	infoCmd.Flags().BoolVarP(&cfg.Runtime.Follow, "follow", "f", false, "Monitor changes in selected directory")
	infoCmd.Flags().StringSliceVarP(&cfg.Runtime.PrintFields, "out", "o", nil, "Fields to print. Whole message if empty.")

	infoCmd.Flags().Var(
		enumflag.New(&outputFormat, "format", outputFormatIds, enumflag.EnumCaseInsensitive),
		"format",
		"Format to print logs, supports: lines, json")
	return infoCmd
}
