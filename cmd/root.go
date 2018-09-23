package cmd

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/cezkuj/trends-analyzer/server"
)

var (
	dbUser             string
	dbPass             string
	dbHost             string
	dbPort             int
	dbName             string
	dispatcherInterval int
	readOnly           bool
	twitterAPIKey      string
	newsAPIKey         string
	stocksAPIKey       string
	salt               string
	verbose            bool
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
	if verbose {
		log.SetLevel(log.DebugLevel)
	}
	dbCfg := server.NewDbCfg(dbUser, dbPass, dbHost, dbPort, dbName)
	server.StartServer(dbCfg, twitterAPIKey, newsAPIKey, stocksAPIKey, salt, dispatcherInterval, readOnly)

}
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Error(fmt.Errorf("Execute failed on %v", err))
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&twitterAPIKey, "twitter-api-key", "t", "", "Twitter API key.")
	rootCmd.MarkFlagRequired("twitter-api-key")
	rootCmd.Flags().StringVarP(&newsAPIKey, "news-api-key", "n", "", "News API key.")
	rootCmd.MarkFlagRequired("news-api-key")
	rootCmd.Flags().StringVarP(&stocksAPIKey, "stocks-api-key", "s", "", "Stocks API key.")
	rootCmd.MarkFlagRequired("stocks-api-key")
	rootCmd.Flags().StringVarP(&salt, "salt", "a", "", "Salt for encrypting users' passwords.")
	rootCmd.MarkFlagRequired("salt")
	rootCmd.Flags().StringVarP(&dbUser, "user", "u", "ta", "Sets user for database conneciton. Default value is ta.")
	rootCmd.Flags().StringVarP(&dbPass, "pass", "p", "", "Sets password for database conneciton, required")
	rootCmd.MarkFlagRequired("pass")
	rootCmd.Flags().StringVarP(&dbHost, "host", "o", "localhost", "Sets host for database conneciton. Default value is localhost")
	rootCmd.Flags().IntVarP(&dbPort, "port", "q", 3306, "Sets port for database conneciton. Default value is 3306")
	rootCmd.Flags().StringVarP(&dbName, "name", "d", "trends", "Sets name for database conneciton. Default value is trends")
	rootCmd.Flags().BoolVarP(&readOnly, "read-only", "e", false, "Sets read only mode. Default value is false.")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Sets logs to DEBUG level.")
	rootCmd.Flags().IntVarP(&dispatcherInterval, "dispatcher-interval", "b", 20, "Interval in minutes. Default value is 20.")
}
