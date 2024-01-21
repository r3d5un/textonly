package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"textonly.islandwind.me/internal/utils"
)

type HealthCheckResponse struct {
	Status      string `json:"status"`
	Environment string `json:"environment"`
	Version     string `json:"version"`
}

// getCmd represents the host command
var hostCmd = &cobra.Command{
	Use:   "host",
	Short: "Get the status of the configured Textonly host",
	Long: `Outputs the current status, environment and version of the configured
Textonly host. Uses the /v1/healthcheck endpoint.`,
	Run: func(cmd *cobra.Command, args []string) {
		host := viper.Get("host")
		url := fmt.Sprintf("%s/v1/healthcheck", host)

		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Add("User-Agent", "toctl (Textonly API client)")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer res.Body.Close()

		hcr := HealthCheckResponse{}
		err = utils.ReadJSON(res.Body, &hcr)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if !jsonOutput {
			fmt.Printf(
				"Status: %s, Environment: %s, Version: %s\n",
				hcr.Status,
				hcr.Environment,
				hcr.Version,
			)
			return
		}

		js, err := json.Marshal(hcr)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(string(js))
	},
}

func init() {
	hostCmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output to JSON")
	getCmd.AddCommand(hostCmd)
}
