package get

import (
	"flow/cmd/get/agents"
	"flow/cmd/get/flows"
	"flow/cmd/get/namespaces"

	"fmt"

	"github.com/spf13/cobra"
)

// createCmd represents the create command
var GetCmd = &cobra.Command{
	Use:   "get",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("get called")

	},
}

func init() {
	GetCmd.AddCommand(namespaces.NamespacesGetCmd)
	GetCmd.AddCommand(flows.FlowsGetCmd)

	GetCmd.AddCommand(agents.AgentsGetCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	GetCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	GetCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
