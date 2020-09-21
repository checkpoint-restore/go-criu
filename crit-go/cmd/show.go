package cmd

import (
	"bufio"
	"crit-go/gocrit"
	log "crit-go/logging"
	"github.com/spf13/cobra"
	"os"
)

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Converts Binary to Json and Displays full output",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if inloc == "" {
			reader := bufio.NewReader(os.Stdin)
			stdininp, err := reader.ReadString('\n')
			log.Fatal(err)
			gocrit.Show(stdininp)
		} else {
			gocrit.Show(inloc)
		}
	},
}

func init() {
	showCmd.Flags().StringVarP(&inloc, "in", "i", "", "criu image in binary format to be decoded (stdin by default)")
}
