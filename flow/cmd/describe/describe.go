package describe

import (
	"flow/cmd/describe/flow"
	"fmt"

	"github.com/spf13/cobra"
)

// createCmd represents the create command
var DescribeCmd = &cobra.Command{
	Use:   "describe",
	Short: "A brief description of your command",
	Long:  `A longer description `,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("describe called")

	},
}

func init() {

	DescribeCmd.AddCommand(flow.FlowDescribeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// DescribeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// DescribeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
