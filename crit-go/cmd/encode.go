package cmd

import (
	"bufio"
	"crit-go/gocrit"
	"github.com/spf13/cobra"
	"os"
)

// encodeCmd represents the encode command
var encodeCmd = &cobra.Command{
	Use:   "encode",
	Short: "Convert Json To Binary",
	Long: `Converts the Json file to Binary and writes to given location for
	more info visit https://www.criu.org/CRIT#Functionality`,
	Run: func(cmd *cobra.Command, args []string) {
		if inloc == "" {
			reader := bufio.NewReader(os.Stdin)
			stdininp, err := reader.ReadString('\n')
			gocrit.Check(err)
			gocrit.Encode(stdininp, outloc)
		} else {
			gocrit.Encode(inloc, outloc)
		}
	},
}

func init() {
	encodeCmd.Flags().StringVarP(&inloc, "in", "i", "", "criu image in binary format to be decoded (stdin by default)")
	encodeCmd.Flags().StringVarP(&outloc, "out", "o", "", "output loc of the file")
}
