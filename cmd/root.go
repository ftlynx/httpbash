package cmd

import (
	"fmt"
	"github.com/ftlynx/httpbash/version"
	"github.com/spf13/cobra"
	"os"
)

var ver bool

var RootCmd = &cobra.Command{
	Use: os.Args[0],
	RunE: func(cmd *cobra.Command, args []string) error {
		if ver {
			fmt.Println(version.FullVersion())
			return nil
		}
		return fmt.Errorf("no flags find")
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
func init() {
	RootCmd.PersistentFlags().BoolVarP(&ver, "version", "v", false, "the version")
}
