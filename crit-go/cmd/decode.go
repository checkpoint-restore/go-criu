package cmd

import (
	"crit-go/gocrit"
	"github.com/spf13/cobra"
)

// decodeCmd represents the decode command
var decodeCmd = &cobra.Command{
	Use:   "decode",
	Short: "Convert Binary to Json",
	Long: `Converts the Binary file to Json and writes to given location
	pretty printing of json is available aswell using the pretty flag for
	more info visit https://www.criu.org/CRIT#Functionality `,
	Run: func(cmd *cobra.Command, args []string) {
		gocrit.Decode(inloc, outloc, pretty, nopl)
	},
}

func init() {
	decodeCmd.Flags().StringVarP(&inloc, "in", "i", "", "criu image in binary format to be decoded (stdin by default)")
	decodeCmd.Flags().StringVarP(&outloc, "out", "o", "", "where to put the image in json format(Stdout by default)")
	decodeCmd.Flags().BoolVarP(&pretty, "pretty", "p", false, "Multiline with indents and some numerical fields in field-specific format")
	decodeCmd.Flags().BoolVarP(&nopl, "nopl", "", false, "comment")
}
