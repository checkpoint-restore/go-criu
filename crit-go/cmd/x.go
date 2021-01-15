package cmd

import (
	"crit-go/gocrit"
	"github.com/spf13/cobra"
)

// xCmd represents the x command
var xCmd = &cobra.Command{
	Use:   "x",
	Short: "Explore image dir",
	Long:  `Explore image dir with option such as ps,fds,mems,rss`,
	Run: func(cmd *cobra.Command, args []string) {
		gocrit.Explore(dir, what)
	},
}

func init() {
	xCmd.Flags().StringVarP(&dir, "dir", "", "", "location/or the image")
	xCmd.MarkFlagRequired("dir")
	xCmd.Flags().StringVarP(&what, "what", "", "", "choose between {ps,fss,mems,rss}")
	xCmd.MarkFlagRequired("what")
}
