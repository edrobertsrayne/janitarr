package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage Radarr/Sonarr server configurations",
	Long:  `Add, list, edit, remove, and test media server configurations.`,
}

var serverAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new media server",
	Long:  `Interactively add a new Radarr or Sonarr server configuration.`,
	RunE:  runServerAdd,
}

var serverListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured servers",
	Long:  `Display all configured media servers in a table or JSON format.`,
	RunE:  runServerList,
}

var serverEditCmd = &cobra.Command{
	Use:   "edit <id-or-name>",
	Short: "Edit an existing server",
	Long:  `Modify the configuration of an existing media server.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runServerEdit,
}

var serverRemoveCmd = &cobra.Command{
	Use:   "remove <id-or-name>",
	Short: "Remove a server",
	Long:  `Delete a media server configuration from the database.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runServerRemove,
}

var serverTestCmd = &cobra.Command{
	Use:   "test <id-or-name>",
	Short: "Test server connection",
	Long:  `Verify connectivity to a configured media server.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runServerTest,
}

func init() {
	serverListCmd.Flags().Bool("json", false, "Output as JSON")
	serverRemoveCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	serverCmd.AddCommand(serverAddCmd)
	serverCmd.AddCommand(serverListCmd)
	serverCmd.AddCommand(serverEditCmd)
	serverCmd.AddCommand(serverRemoveCmd)
	serverCmd.AddCommand(serverTestCmd)
}

func runServerAdd(cmd *cobra.Command, args []string) error {
	fmt.Println("not implemented")
	return nil
}

func runServerList(cmd *cobra.Command, args []string) error {
	fmt.Println("not implemented")
	return nil
}

func runServerEdit(cmd *cobra.Command, args []string) error {
	fmt.Println("not implemented")
	return nil
}

func runServerRemove(cmd *cobra.Command, args []string) error {
	fmt.Println("not implemented")
	return nil
}

func runServerTest(cmd *cobra.Command, args []string) error {
	fmt.Println("not implemented")
	return nil
}
