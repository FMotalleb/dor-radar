/*
Copyright © 2025 Motalleb Fallahnezhad

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"net/http"
	"os"

	"github.com/fmotalleb/go-tools/git"
	"github.com/spf13/cobra"

	"github.com/fmotalleb/dor-radar/config"
	"github.com/fmotalleb/dor-radar/front"
	"github.com/fmotalleb/dor-radar/prometheus"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "dor-radar",
	Short: "Network status radar for dornica infrastructure",

	Version: git.String(),
	// Uncomment the following line if your bare application
	// has an action associated with it:
	RunE: func(cmd *cobra.Command, args []string) error {
		var configFile string
		var err error
		var cfg config.Config
		if configFile, err = cmd.Flags().GetString("config"); err != nil {
			return err
		}

		if err = config.Parse(&cfg, configFile, true); err != nil {
			return err
		}
		return serve(cfg)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringP("config", "c", "./config.toml", "config file path")
}

func serve(cfg config.Config) error {
	http.Handle("/status", prometheus.New(cfg.Core.Collector))
	return front.Serve(cfg.Core.Listen)
}
