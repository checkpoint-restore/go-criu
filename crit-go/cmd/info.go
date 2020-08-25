package cmd

import (
	"bufio"
	"crit-go/gocrit"
	"github.com/spf13/cobra"
	"os"
)

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Gives info of the image",
	Long: `Shows detailed information about the image
	 such as its magic no`,
	Run: func(cmd *cobra.Command, args []string) {
		if inloc == "" {
			reader := bufio.NewReader(os.Stdin)
			stdininp, err := reader.ReadString('\n')
			gocrit.Check(err)
			gocrit.Info(stdininp)
			gocrit.Check(err)
		} else {
			gocrit.Info(inloc)
		}
	},
}

func init() {
	infoCmd.Flags().StringVarP(&inloc, "in", "i", "", "show info about criu image (stdin by default)")
}
