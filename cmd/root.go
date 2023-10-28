/*
Copyright © 2022 Chris Slamar chris@slamar.com
*/
package cmd

import (
	"fmt"
	"github.com/cslamar/audiobooker/audiobooker"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var dryRun bool
var Verbose = false
var notify bool
var alert bool
var enableCaller = false

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "audiobooker",
	Short: "Audiobook creation/manipulation application",
	Long:  `The audiobook creator`, // TODO
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		notifyError(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().BoolVar(&alert, "alert", false, "enable audible pop-up notifications")
	RootCmd.PersistentFlags().BoolVar(&enableCaller, "debug", false, "debugging verbose output")
	RootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Run parsing commands, without converting/binding, and display expected output")
	RootCmd.PersistentFlags().BoolVar(&notify, "notify", false, "enable pop-up notifications")
	RootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.audiobooker.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if enableCaller {
		log.SetReportCaller(true)
		log.SetLevel(log.DebugLevel)
	} else if Verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".audiobooker" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".audiobooker")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

// watchForTermSignals background function for capturing premature term signals
func watchForTermSignals(config *audiobooker.Config) {
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sigterm:
		log.Warnln("got early termination signal.  Exiting!")
		if err := config.Cleanup(); err != nil {
			log.Errorln("config cleanup errored, check for lingering scratch data!")
			os.Exit(1)
		}
		os.Exit(0)
	}
}
