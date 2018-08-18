package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/cezkuj/trends-analyzer/server"
)

var (
	dbUser        string
	dbPass        string
	dbHost        string
	dbName        string
	prod          bool
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
	dbCfg := server.NewDbCfg(dbUser, dbPass, dbHost, dbName)
	server.StartServer(dbCfg, twitterApiKey, newsApiKey, prod)

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
	rootCmd.Flags().StringVarP(&dbUser, "user", "u", "", "Sets user for database conneciton, required")
	rootCmd.MarkFlagRequired("user")
	rootCmd.Flags().StringVarP(&dbPass, "pass", "p", "", "Sets password for database conneciton, required")
	rootCmd.MarkFlagRequired("pass")
	rootCmd.Flags().StringVarP(&dbHost, "host", "o", "", "Sets host for database conneciton, required")
	rootCmd.MarkFlagRequired("host")
	rootCmd.Flags().StringVarP(&dbName, "name", "n", "", "Sets name for database conneciton, required")
	rootCmd.MarkFlagRequired("name")
	rootCmd.Flags().BoolVarP(&prod, "prod", "r", false, "Sets production mode with tls enabled. Default value is false.")
}
