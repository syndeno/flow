package create

import (
	"flow/cmd/create/flow"
	"flow/cmd/create/namespace"

	"fmt"

	"github.com/spf13/cobra"
)

// createCmd represents the create command
var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("create called")

	},
}

func init() {
	CreateCmd.AddCommand(namespace.NamespaceCreateCmd)
	CreateCmd.AddCommand(flow.FlowCreateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// CreateCmd.PersistentFlags().String("foo", "", "A help for foo")
	// CreateCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// CreateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
