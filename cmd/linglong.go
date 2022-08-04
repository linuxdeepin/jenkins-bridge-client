package cmd

import (
	"fmt"

	"jenkins-bridge-client/client"

	"github.com/spf13/cobra"
)

func init() {
	linglongBuildCmd.Flags().StringVarP(&token, "token", "", defaultToken, "jenkins bridge token")
	linglongBuildCmd.Flags().StringVarP(&server, "server", "", defaultServer, "jenkins bridge server address")
	rootCmd.AddCommand(linglongBuildCmd)
}

var linglongBuildCmd = &cobra.Command{
	Use:   "triggerlinglong",
	Short: "trigger linglong build",
	Long:  `trigger jenkins to build linglong`,
	Run: func(cmd *cobra.Command, args []string) {
		cl := client.NewClient()
		cl.SetHost(server)
		cl.SetToken(token)
		cl.PostLinglongBuildjob()
		fmt.Println(cl.GetID())
	},
}
