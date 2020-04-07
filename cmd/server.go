package cmd

import (
	"log"
	"net/http"

	"github.com/spf13/cobra"
)

const (
	defaultPort      = "8888"
	defaultDirectory = "."
)

var (
	serverPort string
	serverDir  string
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "start a server",
	Long:  `Start a server.`,
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		http.Handle("/", http.FileServer(http.Dir(serverDir)))

		log.Printf("Serving %s on HTTP port: %s\n", serverDir, serverPort)
		log.Fatal(http.ListenAndServe(":"+serverPort, nil))
	},
}

func serverInit() {
	serverCmd.Flags().StringVarP(&serverPort, "port", "p", defaultPort, "port to serve on")
	serverCmd.Flags().StringVarP(&serverDir, "directory", "d", defaultDirectory, "location of static root for server with files")
}
