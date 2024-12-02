package customlicense

import (	
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type CustomLicenseCmd struct {
	logger *logrus.Logger
}

func NewCustomLicenseCmd(logger *logrus.Logger) *CustomLicenseCmd {
	return &CustomLicenseCmd{
		logger: logger,
	}
}

// Execute runs the xmlconvert command
func (cl *CustomLicenseCmd) Execute() *cobra.Command {
	// Create a new cobra command for xmlconvert
	customLicenseCmd := &cobra.Command{
		Use:     "customLicense",
		Short:   "customLicense takes the public url of the license file and stores it in the specified path of the project",
		Long:    `The 'customLicense' command retrieves a license file from a specified public URL and stores it in a designated path within your project. By default, it uses 'config.xml' for the license URL and 'config.yaml' for the project path, ensuring easy integration and updates of license files. This command includes pre-run validation and customizable flags for flexible usage.`,
		Example: exampleCommand,
		PreRun:  cl.preRun, // Define a pre-run function
		Run:     cl.run,    // Define the run function
	}

	// Define the flags for the xmlconvert command
	customLicenseCmd.Flags().StringVar(&path, "path", path, "please specify the public url of the license file. The default path is ``")
	customLicenseCmd.Flags().StringVar(&projectPath, "projectPath", projectPath, "Provide the project path to store the license file. The default path is ``")

	return customLicenseCmd
}

func (cl *CustomLicenseCmd) preRun(cmd *cobra.Command, args []string) {
	// Do any pre-run setup here
	yellow := "\033[33m"
	reset := "\033[0m"
	text := "WARNING: This command is in alpha version and may need some changes."
	cl.logger.Println(yellow + text + reset)
}

func (cl *CustomLicenseCmd) run(cmd *cobra.Command, args []string) {
	// Ensure path is set
	if path == "" {
		cl.logger.Fatal("Path to the license file URL must be specified")
	}

	// Use the current working directory if projectPath is not provided
	if projectPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			cl.logger.Fatalf("Failed to get the current working directory: %v", err)
		}
		projectPath = filepath.Join(cwd, "LICENSE")
	}

	// Log the start of the process
	cl.logger.Println("Starting to download the license file from:", path)

	// Create the HTTP request to fetch the license file
	response, err := http.Get(path)
	if err != nil {
		cl.logger.Fatalf("Failed to download the license file: %v", err)
	}
	defer response.Body.Close()

	// Check if the HTTP request was successful
	if response.StatusCode != http.StatusOK {
		cl.logger.Fatalf("Failed to download the license file, HTTP Status: %s", response.Status)
	}

	// Create the destination directory if it does not exist
	err = os.MkdirAll(filepath.Dir(projectPath), os.ModePerm)
	if err != nil {
		cl.logger.Fatalf("Failed to create the destination directory: %v", err)
	}

	// Create the destination file
	outFile, err := os.Create(projectPath)
	if err != nil {
		cl.logger.Fatalf("Failed to create the destination file: %v", err)
	}
	defer outFile.Close()

	// Copy the content from the response to the file
	_, err = io.Copy(outFile, response.Body)
	if err != nil {
		cl.logger.Fatalf("Failed to save the license file: %v", err)
	}

	// Log the success of the operation
	cl.logger.Println("License file successfully downloaded and stored at:", projectPath)
}
