// Copyright © 2016 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

func main() {
	RootCmd.Execute()
}

var debugLogging bool
var registrySource, registryTarget RegistryInfo
var namespaces []string
var tagRegexp string
var pollingFrequency time.Duration
var port int

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "registryrsync",
	Short: "Simple tool to help keep registries in sync",
	Long: `

	registrysync --source-url <registrysource>  --target-url

	All settings can be overriden via enviornment variables prefixed with RR, e.g.
	RR_SOURCE_URL
	`,

	// Set logging for all commands
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if debugLogging {
			log.SetLevel(log.DebugLevel)
		}
	},
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		if registrySource.address == "" {
			log.Error("No source registry address specified")
			cmd.Usage()
		}
		if registryTarget.address == "" {
			log.Error("No target registry address specified")
		}
		var nameFilter, tagFilter Filter
		if len(namespaces) == 0 {
			nameFilter = matchEverything{}
		} else {
			nameFilter = NewNamespaceFilter(namespaces...)
		}
		if tagRegexp == "" {
			tagRegexp = ".*"
		}
		tagFilter, err := NewRegexTagFilter(tagRegexp)
		if err != nil {
			log.Errorf("Can't create filter from bad regular expression %s", tagRegexp)
			return
		}
		filter := DockerImageFilter{nameFilter, tagFilter}
		handler, err := NewDockerCLIHandler(registrySource, registryTarget, filter)
		if err != nil {
			log.Errorf("Couldn't not connect to registries %s %s",
				registrySource.Address(), registryTarget.Address())
			return
		}

		if pollingFrequency > 0 {
			go func() {
				c := time.Tick(pollingFrequency)
				for range c {
					// Note this purposfully runs the function
					// in the same goroutine so we make sure there is
					// only ever one. If it might take a long time and
					// it's safe to have several running just add "go" here.
					err := handler.RSync(filter)
					if err != nil {
						log.Errorf("Failure syncing between registries %s %s", registrySource.Address(),
							registryTarget.Address())
					}
				}
			}()
		}
		http.Handle("/", registryEventHandler(handler))
		http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// TODO
	RootCmd.Flags().StringVar(&registrySource.address, "source-url", "", "registry url to read images from")
	RootCmd.Flags().StringVar(&registryTarget.address, "target-url", "", "registry url to send images to")

	RootCmd.Flags().StringVar(&registryTarget.address, "source-user", "", "username for registry to read images from")
	RootCmd.Flags().StringVar(&registryTarget.address, "target-user", "", "username for registry to send images to")
	RootCmd.Flags().StringVar(&registryTarget.password, "source-password", "", "password for registry to read images from")
	RootCmd.Flags().StringVar(&registryTarget.password, "target-password", "", "password for registry to send images to")

	RootCmd.Flags().StringVar(&tagRegexp, "tag-regex", ".*", "regular expression of tags to match")
	// RootCmd.Flags().StringSliceP(&namespaces, "name", []string{}, "namespace to watch.  Can have multiple. Blank for all")
	// RootCmd.Flags().Duration(&pollingFrequency, "poll", "Set to have a cron job setup to converge")
	RootCmd.Flags().DurationVar(&pollingFrequency, "poll", 0, "How frequently should we check the registries")
	// RootCmd.Flags().IntVar(p, name, value, usage)
	RootCmd.Flags().IntVar(&port, "port", 8787, "Port to  listen to notifications on")
	RootCmd.PersistentFlags().BoolVarP(&debugLogging, "debug", "d", false, "turn on debug")

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "registryrsync.yml", "config file (default is $HOME/.registryrsync.yaml)")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viperStringFlags := []string{"source-url",
		"target-url", "source-user",
		"target-user", "source-password", "target-password",
		"tag-regex", "namespace-regex", "port",
	}

	for _, flag := range viperStringFlags {
		viper.BindPFlag(flag, RootCmd.PersistentFlags().Lookup(flag))
	}

	//Allow us to use RR environment variables
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix("rr")
	viper.AutomaticEnv()

	viper.SetConfigName("registryrsync.yml") // name of config file (without extension)
	viper.AddConfigPath("$HOME")             // adding home directory as first search path

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
