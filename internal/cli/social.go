package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"textonly.islandwind.me/internal/data"
	"textonly.islandwind.me/internal/utils"
)

type SocialListResponse struct {
	Metadata data.Metadata  `json:"metadata"`
	Data     []*data.Social `json:"data"`
}

type SocialResponse struct {
	Metadata data.Metadata `json:"metadata"`
	Data     data.Social   `json:"data"`
}

// getCmd represents the host command
var getSocialCmd = &cobra.Command{
	Use:   "social",
	Short: "Get the social data of the configured Textonly host",
	Long: `Gets the social data for the user of the Textonly host.

By default, it lists all social data for the user. If you specify the ID (integer), it
only lists the social data for that ID.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Cleanup this mess
		host := viper.Get("host")
		url := fmt.Sprintf("%s/api/social", host)

		if len(args) > 1 {
			fmt.Println("Too many arguments")
			os.Exit(1)
		}

		if len(args) > 0 {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				fmt.Println("ID must be an integer")
				os.Exit(1)
			}

			url = fmt.Sprintf("%s/%d", url, id)
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
			fmt.Printf("Unable to get social data: %s\n", res.Status)
			os.Exit(1)
		}

		if len(args) > 0 {
			social := SocialResponse{}
			err = utils.ReadJSON(res.Body, &social)
			if err != nil {
				fmt.Println("Unable to get social data")
				os.Exit(1)
			}

			if !jsonOutput {
				fmt.Printf(
					"ID: %d, User ID: %d, Plattform: %s, Link: %s\n",
					social.Data.ID,
					social.Data.UserID,
					social.Data.SocialPlatform,
					social.Data.Link,
				)
				return
			}

			js, err := json.Marshal(social)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println(string(js))
		} else {
			s := SocialListResponse{}
			err = utils.ReadJSON(res.Body, &s)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			if !jsonOutput {
				for _, social := range s.Data {
					fmt.Printf(
						"ID: %d, User ID: %d, Plattform: %s, Link: %s\n",
						social.ID,
						social.UserID,
						social.SocialPlatform,
						social.Link,
					)
				}
				return
			}
			js, err := json.Marshal(s)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println(string(js))
		}
	},
}

func init() {
	getSocialCmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output to JSON")
	getCmd.AddCommand(getSocialCmd)
}
