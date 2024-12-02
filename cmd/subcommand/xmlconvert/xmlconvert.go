package xmlconvert

import (
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// XMLConvertCmd represents the structure of the xmlconvert command
type XMLConvertCmd struct {
	logger *logrus.Logger
}

// NewXMLConvertCmd returns a new instance of XMLConvertCmd
func NewXMLConvertCmd(logger *logrus.Logger) *XMLConvertCmd {
	return &XMLConvertCmd{
		logger: logger,
	}
}

// Execute runs the xmlconvert command
func (xml *XMLConvertCmd) Execute() *cobra.Command {
	// Create a new cobra command for xmlconvert
	xmlConvertCmd := &cobra.Command{
		Use:     "xmlconvert",
		Short:   "Convert XML to JSON and YAML",
		Long:    "`xmlconvert` converts XML file provided in the command line arguments to JSON and YAML format. It can be used to convert XML file to JSON and YAML file.",
		Example: exampleCommand,
		PreRun:  xml.preRun, // Define a pre-run function
		Run:     xml.run,    // Define the run function
	}

	// Define the flags for the xmlconvert command
	xmlConvertCmd.Flags().StringVar(&xmlFile, "xmlFile", xmlFile, "provide xml file path to convert. The default path is `config.xml`")
	xmlConvertCmd.Flags().StringArrayVar(&outputFiles, "outputFiles", outputFiles, "returns converted output file. The default path is `config.yaml`")

	return xmlConvertCmd
}

// preRun is a pre-run function that logs a warning message
func (xml *XMLConvertCmd) preRun(cmd *cobra.Command, args []string) {
	yellow := "\033[33m"
	reset := "\033[0m"
	text := "WARNING: This command is in alpha version and may need some changes."
	xml.logger.Println(yellow + text + reset)
}

// run is the function that will be called when the xmlconvert command is executed
func (xml *XMLConvertCmd) run(cmd *cobra.Command, args []string) {
	xmlData, err := ReadXML(xmlFile)
	if err != nil {
		xml.logger.Fatal(err)
	}
	xml.logger.Info("output files: ", outputFiles)

	// Check if only two output files are supported
	if len(outputFiles) > 2 {
		xml.logger.Fatal("only two output file extensions are supported: json and yaml")
	}

	// Check if output files are provided
	if len(outputFiles) == 0 {
		xml.logger.Fatal("please provide output file")
	}

	// Check if only one output file is provided
	if len(outputFiles) == 1 {
		fileExtension := strings.Split(outputFiles[0], ".")[len(strings.Split(outputFiles[0], "."))-1]
		err := CreateFile(xmlData, outputFiles[0], fileExtension)
		if err != nil {
			xml.logger.Fatal(err)
		}
	} else {
		for _, file := range outputFiles {
			fileExtension := strings.Split(file, ".")[len(strings.Split(file, "."))-1]
			err := CreateFile(xmlData, file, fileExtension)
			if err != nil {
				xml.logger.Fatal(err)
			}
		}
	}
}
