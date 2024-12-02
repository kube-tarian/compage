package cmd

import (
	"fmt"
	"os"

	"github.com/intelops/compage/internal/languages/executor"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	projectName         string
	platformName        string
	platformURL         string
	platformUserName    string
	repositoryName      string
	serverType          string
	language            string
	overwriteConfigFile bool
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes the compage configuration, generates the file and writes it down in the current directory",
	Long: `Compage configuration file is a yaml file that contains the configuration that guides the compage to generate the code. 

You can change the file as per your needs and then run the compage generate command to generate the code.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := createOrUpdateDefaultConfigFile()
		cobra.CheckErr(err)
	},
}

func createOrUpdateDefaultConfigFile() error {
	var err error
	// create default config file
	configFilePath := "config.yaml"
	_, err = os.Stat(configFilePath)
	if err == nil {
		log.Warnf("config file already exists at %s", configFilePath)
		if !overwriteConfigFile {
			log.Infof("skipping config file creation")
			return err
		}
		log.Infof("overwriting the config file")
		err = os.Remove(configFilePath)
		if err != nil {
			log.Errorf("error while removing the config file %s", err)
			cobra.CheckErr(err)
		}
	}

	if language != "go" && language != "dotnet" {
		message := fmt.Sprintf("language %s is not supported", language)
		log.Errorf("%s", message)
		return fmt.Errorf("%s", message)
	}

	_, err = os.Create(configFilePath)
	if err != nil {
		log.Errorf("error while creating the config file %s", err)
		return err
	}

	if language == "go" {
		var goContentData []byte
		goContentData, err = GoConfigContent.ReadFile("go-config.yaml.tmpl")
		if err != nil {
			log.Errorf("error while reading the config file %s", err)
			return err
		}
		// copy the default config file and use go template to replace the values
		err = os.WriteFile(configFilePath, goContentData, 0644)
		if err != nil {
			log.Errorf("error while creating the config file %s", err)
			return err
		}
		var filePaths []*string
		filePaths = append(filePaths, &configFilePath)
		data := make(map[string]interface{})
		data["ProjectName"] = projectName
		data["GitRepositoryName"] = repositoryName
		data["GitRepositoryURL"] = platformURL + "/" + platformUserName + "/" + repositoryName
		data["GitPlatformName"] = platformName
		data["GitPlatformURL"] = platformURL
		data["GitPlatformUserName"] = platformUserName
		if serverType == "grpc" {
			data["IsRestAndGrpcServer"] = false
			data["IsGrpcServer"] = true
			data["IsRestServer"] = false
		} else if serverType == "rest-grpc" {
			data["IsRestAndGrpcServer"] = true
			data["IsGrpcServer"] = false
			data["IsRestServer"] = false
		} else {
			data["IsRestAndGrpcServer"] = false
			data["IsGrpcServer"] = false
			data["IsRestServer"] = true
		}
		err = executor.Execute(filePaths, data)
		if err != nil {
			log.Errorf("error while creating the config file %s", err)
			return err
		}
		log.Infof("config file created at %s", configFilePath)
	} else if language == "dotnet" {
		var dotnetContentData []byte
		dotnetContentData, err = DotNetConfigContent.ReadFile("dotnet-config.yaml.tmpl")
		if err != nil {
			log.Errorf("error while reading the config file %s", err)
			return err
		}
		// copy the default config file and use go template to replace the values
		err = os.WriteFile(configFilePath, dotnetContentData, 0644)
		if err != nil {
			log.Errorf("error while creating the config file %s", err)
			return err
		}
		var filePaths []*string
		filePaths = append(filePaths, &configFilePath)
		data := make(map[string]interface{})
		data["ProjectName"] = projectName
		data["GitRepositoryName"] = repositoryName
		data["GitRepositoryURL"] = platformURL + "/" + platformUserName + "/" + repositoryName
		data["GitPlatformName"] = platformName
		data["GitPlatformURL"] = platformURL
		data["GitPlatformUserName"] = platformUserName
		if serverType == "rest" {
			data["IsRestAndGrpcServer"] = false
			data["IsGrpcServer"] = false
			data["IsRestServer"] = true
		}
		err = executor.Execute(filePaths, data)
		if err != nil {
			log.Errorf("error while creating the config file %s", err)
			return err
		}
		log.Infof("config file created at %s", configFilePath)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	initCmd.Flags().StringVar(&projectName, "projectName", "myproject", "Project Name")
	initCmd.Flags().StringVar(&platformName, "platformName", "github", "Git Platform Name")
	initCmd.Flags().StringVar(&platformURL, "platformURL", "https://github.com", "Git Platform URL")
	initCmd.Flags().StringVar(&platformUserName, "platformUserName", "myusername", "Git Platform User Name")
	initCmd.Flags().StringVar(&repositoryName, "repositoryName", "myproject", "Git Repository Name")
	initCmd.Flags().StringVar(&serverType, "serverType", "rest", "Server Type (rest, grpc, rest-grpc)")
	initCmd.Flags().BoolVar(&overwriteConfigFile, "overwriteConfigFile", false, "Overwrite the config file if it already exists")
	initCmd.Flags().StringVar(&language, "language", "go", "Language (go, dotnet)")
}
