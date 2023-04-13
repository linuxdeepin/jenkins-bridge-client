package cmd

import (
	"io"
	"jenkins-bridge-client/client"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var filename string

func init() {
	catCmd.Flags().StringVarP(&token, "token", "", defaultToken, "jenkins bridge token")
	catCmd.Flags().StringVarP(&server, "server", "", defaultServer, "jenkins bridge server address")
	catCmd.Flags().StringVarP(&filename, "file", "", "", "file path")
	catCmd.Flags().IntVarP(&runid, "runid", "", defaultRunid, "job runid")
	catCmd.MarkFlagRequired("runid")
	catCmd.MarkFlagRequired("file")

	// catFlag.IntVar(&runid, "runid", 0, "job runid")
	// catFlag.StringVar(&filename, "file", "", "file path")
	// catFlag.StringVar(&server, "server", "", "server")
	// catFlag.StringVar(&token, "token", "", "token (env BRIDGE_TOKEN)")

	rootCmd.AddCommand(catCmd)
}

var catCmd = &cobra.Command{
	Use:   "cat",
	Short: "print build result",
	Long:  `print build result such as warning, log`,
	Run: func(cmd *cobra.Command, args []string) {
		cl := client.NewClient()
		cl.SetHost(server)
		cl.SetToken(token)
		cl.SetID(runid)
		for _, f := range cl.GetApiJobArtifacts().Files {
			if f.Name == filename {
				resp, err := cl.R().SetDoNotParseResponse(true).Get(f.URL)
				if err != nil {
					log.Fatal(err)
				}
				defer func() {
					if err := resp.RawBody().Close(); err != nil {
						log.Fatal(err)
					}
				}()
				io.Copy(os.Stdout, resp.RawBody())
			}
		}
	},
}