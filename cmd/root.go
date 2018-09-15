package cmd

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/cezkuj/trends-analyzer/server"
)

var (
	dbUser        string
	dbPass        string
	dbHost        string
	dbPort        int
	dbName        string
	prod          bool
	twitterAPIKey string
	newsAPIKey    string
	verbose       bool
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
	server.StartServer(dbCfg, twitterAPIKey, newsAPIKey, prod)

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
	rootCmd.Flags().StringVarP(&dbUser, "user", "u", "ta", "Sets user for database conneciton. Default value is ta.")
	rootCmd.Flags().StringVarP(&dbPass, "pass", "p", "", "Sets password for database conneciton, required")
	rootCmd.MarkFlagRequired("pass")
	rootCmd.Flags().StringVarP(&dbHost, "host", "o", "localhost", "Sets host for database conneciton. Default value is localhost")
	rootCmd.Flags().IntVarP(&dbPort, "port", "r", 3306, "Sets port for database conneciton. Default value is 3306")
	rootCmd.Flags().StringVarP(&dbName, "name", "d", "trends-analyzer", "Sets name for database conneciton. Default value is trends-analyzer")
	rootCmd.Flags().BoolVarP(&prod, "prod", "r", false, "Sets production mode with tls enabled. Default value is false.")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Sets logs to DEBUG level.")
}
