package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/cezkuj/trends-analyzer/server"
)

var (
	twitterApiKey string
	newsApiKey    string
)

var rootCmd = &cobra.Command{
	Use:   "Trends analyzer",
	Short: "Analyzis trends on data found in the web.",
	Long: ` Analyzis trends on data found in the web.
        Examples:
        
        trends-analyzer -t ABC -n DEF`,
	Args: cobra.NoArgs,
	Run:  startServer,
}

func startServer(cmd *cobra.Command, args []string) {
	server.StartServer(twitterApiKey, newsApiKey)

}
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&twitterApiKey, "twitter-api-key", "t", "", "Twitter API key.")
	rootCmd.MarkFlagRequired("twitter-api-key")
	rootCmd.Flags().StringVarP(&newsApiKey, "news-api-key", "n", "", "News API key.")
	rootCmd.MarkFlagRequired("news-api-key")
}
