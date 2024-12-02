package cmd

import (
	"github.com/intelops/compage/cmd/subcommand/customlicense"
	"github.com/sirupsen/logrus"
)

func init() {
	// Create the logger instance
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// Create the instance for customlicense
	customLicense := customlicense.NewCustomLicenseCmd(logger)

	// Add Subcommand for the root command
	rootCmd.AddCommand(customLicense.Execute())
}
