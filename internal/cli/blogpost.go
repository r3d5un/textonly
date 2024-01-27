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

type BlogPostListResponse struct {
	Metadata data.Metadata    `json:"metadata"`
	Data     []*data.BlogPost `json:"data"`
}

type BlogPostResponse struct {
	Metadata data.Metadata `json:"metadata"`
	Data     data.BlogPost `json:"data"`
}

// getCmd represents the host command
var getBlogPostCmd = &cobra.Command{
	Use:   "blogpost",
	Short: "Get the blog post of the configured Textonly host",
	Long: `Gets the blog post for the user of the Textonly host.

By default, it lists all blog posts. If you specify the ID (integer), it
only lists the for that ID.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Cleanup this mess
		host := viper.Get("host")
		url := fmt.Sprintf("%s/api/post", host)

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
			fmt.Printf("Unable to get blog post: %s\n", res.Status)
			os.Exit(1)
		}

		if len(args) > 0 {
			bp := BlogPostResponse{}
			err = utils.ReadJSON(res.Body, &bp)
			if err != nil {
				fmt.Println("Unable to get blog post")
				os.Exit(1)
			}

			if !jsonOutput {
				fmt.Printf(
					"ID: %d, Title: %s, Created: %s, Last update: %s\n",
					bp.Data.ID,
					bp.Data.Title,
					bp.Data.Created,
					bp.Data.LastUpdate,
				)
				fmt.Println(bp.Data.Lead)

				return
			}

			js, err := json.Marshal(bp)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println(string(js))
		} else {
			s := BlogPostListResponse{}
			err = utils.ReadJSON(res.Body, &s)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			if !jsonOutput {
				for _, bp := range s.Data {
					fmt.Printf(
						"ID: %d, Title: %s, Created: %s, Last update: %s\n",
						bp.ID,
						bp.Title,
						bp.Created,
						bp.LastUpdate,
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
	getBlogPostCmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output to JSON")
	// TODO: Add flags for markdown output
	// getBlogPostCmd.Flags().BoolVarP(&jsonOutput, "markdown", "md", false, "Write to Markdown file")
	// getBlogPostCmd.Flags().BoolVarP(&jsonOutput, "read", "r", false, "Read blog post in terminal")
	getCmd.AddCommand(getBlogPostCmd)
}
