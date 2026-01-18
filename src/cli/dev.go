package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Start Janitarr in development mode (verbose logging)",
	Long:  `Starts the automation scheduler and web server with verbose logging and debug output.`,
	RunE:  runDev,
}

func init() {
	devCmd.Flags().IntP("port", "p", 3434, "Web server port")
	devCmd.Flags().String("host", "localhost", "Web server host")
}

func runDev(cmd *cobra.Command, args []string) error {
	fmt.Println("not implemented")
	return nil
}
