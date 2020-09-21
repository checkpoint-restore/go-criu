package cmd

import (
	log "crit-go/logging"
	"fmt"
	"github.com/spf13/cobra"
)

var inloc, outloc, dir, what, cfgFile string
var pretty, nopl bool

// rootCmd represents the base command when called without any subcommands

var rootCmd = &cobra.Command{
	Use:   "crit-go",
	Short: "CRIT is a feature-rich replacement for existing criu show",
	Long: `CRIT is a feature-rich replacement for existing "criu show". 
	This version is written in Go for usage with Go codebase, crit is also available in python
	for more information visit https://criu.org/CRIT `,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("crit executed")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	rootCmd.AddCommand(decodeCmd)
	rootCmd.AddCommand(encodeCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(xCmd)
}
