package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// getCmd represents the host command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Gets a resource or thing from the Textonly application",
	Long: `The get commands is used to retrieve a resource or thing from the Textonly
application. For example, you can use the get command to retrieve a blog post
or user.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("need subcommand")
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
