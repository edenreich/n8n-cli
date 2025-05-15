/*
Copyright © 2025 Eden Reich

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/edenreich/n8n-cli/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "n8n",
	Short: "Command line interface for managing n8n instances",
	Long: `n8n is a command line tool for managing n8n instances.

It allows you to synchronize JSON workflows between your local filesystem and n8n instances,
import workflows from n8n instances to your local directory, and manage your workflows 
through version control systems.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if versionFlag, _ := cmd.Flags().GetBool("version"); versionFlag {
			err := versionCmd.RunE(cmd, args)
			return err
		}

		cmd.SetOut(cmd.OutOrStdout())
		return cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// GetRootCmd returns the root command for testing purposes
func GetRootCmd() *cobra.Command {
	return rootCmd
}

func init() {
	cobra.OnInitialize(config.Initialize)

	rootCmd.PersistentFlags().StringP("api-key", "k", "", "n8n API Key (env: N8N_API_KEY)")
	rootCmd.PersistentFlags().StringP("url", "u", "http://localhost:5678", "n8n instance URL (env: N8N_INSTANCE_URL)")
	rootCmd.Flags().Bool("version", false, "Display the version information")

	if err := viper.BindPFlag("api_key", rootCmd.PersistentFlags().Lookup("api-key")); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding api-key flag: %v\n", err)
	}
	if err := viper.BindPFlag("instance_url", rootCmd.PersistentFlags().Lookup("url")); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding url flag: %v\n", err)
	}
	rootCmd.Flags().BoolP("verbose", "V", false, "Show detailed output during synchronization")
}
