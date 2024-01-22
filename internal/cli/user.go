package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"textonly.islandwind.me/internal/data"
	"textonly.islandwind.me/internal/utils"
)

type UserResponse struct {
	Metadata data.Metadata `json:"metadata"`
	Data     data.User     `json:"data"`
}

// getCmd represents the host command
var getUserCmd = &cobra.Command{
	Use:   "user",
	Short: "Get the user of the configured Textonly host",
	Long: `Get the user for the configured host.

There is no need to specify an ID, as there is only one user.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Cleanup this mess
		host := viper.Get("host")
		url := fmt.Sprintf("%s/api/user/1", host) // TODO: Fix hardcoded ID

		if len(args) > 0 {
			fmt.Println("Too many arguments")
			os.Exit(1)
		}

		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Add("User-Agent", "toctl (Textonly API client)")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer res.Body.Close()

		if res.StatusCode != 200 {
			fmt.Printf("Unable to get user: %s\n", res.Status)
			os.Exit(1)
		}

		u := UserResponse{}
		err = utils.ReadJSON(res.Body, &u)
		if err != nil {
			fmt.Printf("Unable to get user, %s\n", err)
			os.Exit(1)
		}

		// TODO: Implement long output flag
		if !jsonOutput {
			fmt.Printf(
				"ID: %d, Name: %s\n",
				u.Data.ID,
				u.Data.Name,
				// u.Data.Summary,
				// u.Data.Content,
			)
			return
		}

		js, err := json.Marshal(u)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(string(js))
	},
}

func init() {
	// TODO: Add flag for long output
	getUserCmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output to JSON")
	getCmd.AddCommand(getUserCmd)
}
