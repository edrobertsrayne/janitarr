package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start Janitarr in production mode (scheduler + web server)",
	Long:  `Starts the automation scheduler and web server for production use.`,
	RunE:  runStart,
}

func init() {
	startCmd.Flags().IntP("port", "p", 3434, "Web server port")
	startCmd.Flags().String("host", "localhost", "Web server host")
}

func runStart(cmd *cobra.Command, args []string) error {
	fmt.Println("not implemented")
	return nil
}
