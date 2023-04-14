package cmd

import (
	"github.com/ftlynx/httpbash/internal/app"
	"github.com/ftlynx/httpbash/internal/config/file"
	"github.com/ftlynx/httpbash/internal/global"
	"github.com/spf13/cobra"
)

var confFile = "config.yaml"

var runCmd = &cobra.Command{
	Use: "run",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := file.NewFileConf(confFile).GetConf()
		if err != nil {
			return err
		}
		if err := c.Validate(); err != nil {
			return nil
		}
		global.Conf = c
		return app.MyRouter()
	},
}

func init() {
	runCmd.Flags().StringVarP(&confFile, "config", "f", confFile, "config file")
	RootCmd.AddCommand(runCmd)
}
