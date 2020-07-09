package main

import (
	"fmt"

	"github.com/biosugar0/tele/pkg/util"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   `telename <deployment name>`,
		Short: "This command converts a string into a string that can be used as a telepresence deployment name.",
		Long: `This command converts a string into a string that can be used as a telepresence deployment name.
`,
		RunE: Run,
	}
)

func Run(cmd *cobra.Command, args []string) error {

	if len(args[0]) == 0 {
		fmt.Println("telename")
		return nil
	}
	deployment := util.ToValidName(args[0])
	fmt.Println(deployment)
	return nil
}

func main() {
	rootCmd.PersistentFlags().SortFlags = false
	rootCmd.Execute()
}
