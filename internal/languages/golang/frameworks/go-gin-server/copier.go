package goginserver

import (
	"errors"
	"fmt"
	"github.com/intelops/compage/internal/languages/golang/frameworks"
	"text/template"

	"github.com/gertd/go-pluralize"
	"github.com/iancoleman/strcase"
	corenode "github.com/intelops/compage/internal/core/node"
	"github.com/intelops/compage/internal/languages/executor"
	commonUtils "github.com/intelops/compage/internal/languages/utils"
	log "github.com/sirupsen/logrus"

	"strings"

	"github.com/intelops/compage/internal/utils"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const RestServerPath = "/pkg/rest/server"
const RestClientPath = "/pkg/rest/client"
const ConfigPath = "/config"

const ControllersPath = RestServerPath + "/controllers"
const ServicesPath = RestServerPath + "/services"
const DaosPath = RestServerPath + "/daos"
const ModelsPath = RestServerPath + "/models"

const NoSQLDBClientsPath = DaosPath + "/clients/nosqls"
const NoSQLControllerFile = "nosqls-controller.go.tmpl"
const NoSQLColumnBasedControllerFile = "nosqls-column-based-controller.go.tmpl"
const NoSQLServiceFile = "nosqls-service.go.tmpl"
const MongoDBDaoFile = "mongodb-dao.go.tmpl"
const ScyllaDBDaoFile = "scylladb-dao.go.tmpl"
const MongoDBConfigFile = "mongodb.go.tmpl"
const ScyllaDBConfigFile = "scylladb.go.tmpl"
const NoSQLModelFile = "nosqls-model.go.tmpl"
const NoSQLColumnBasedModelFile = "nosqls-column-based-model.go.tmpl"

const SQLDBClientsPath = DaosPath + "/clients/sqls"
const SQLControllerFile = "sqls-controller.go.tmpl"
const SQLServiceFile = "sqls-service.go.tmpl"
const MySQLDaoFile = "mysql-dao.go.tmpl"
const SQLModelFile = "sqls-model.go.tmpl"

const DaoFile = "dao.go.tmpl"
const SQLiteDaoFile = "sqlite-dao.go.tmpl"
const MySQLDBConfigFile = "mysql.go.tmpl"
const SQLiteDBConfigFile = "sqlite.go.tmpl"

// SQLGORMModelFile GORM integration
const SQLGORMModelFile = "sqls-gorm-model.go.tmpl"
const MySQLGORMDaoFile = "mysql-gorm-dao.go.tmpl"
const SQLiteGORMDaoFile = "sqlite-gorm-dao.go.tmpl"
const MySQLGORMDBConfigFile = "mysql-gorm.go.tmpl"
const SQLiteGORMDBConfigFile = "sqlite-gorm.go.tmpl"
const MapDBConfigFile = "map.go.tmpl"

const ClientFile = "client.go.tmpl"
const ConfigFile = "rest-opentel-config.go.tmpl"

// MongoDB nosql databases
const MongoDB = "MongoDB"
const ScyllaDB = "ScyllaDB"

// SQLite sql databases
const SQLite = "SQLite"
const MySQL = "MySQL"
const Map = "Map"

const SQLiteGORM = "SQLite-GORM"
const MySQLGORM = "MySQL-GORM"

// Copier Language specific *Copier
type Copier struct {
	NodeDirectoryName string
	TemplatesRootPath string
	Data              map[string]interface{}
	IsRestServer      bool
	HasRestClients    bool
	SQLDB             string
	IsSQLDB           bool
	NoSQLDB           string
	IsNoSQLDB         bool
	RestServerPort    string
	Resources         []*corenode.Resource
	ResourceConfig    map[string]*frameworks.RestResourceData
	RestClients       []*corenode.RestClient
	PluralizeClient   *pluralize.Client
}

func NewCopier(gitPlatformURL, gitPlatformUserName, gitRepositoryName, nodeName, nodeDirectoryName, templatesRootPath string, isRestServer bool, restServerPort string, isSQLDB bool, sqlDB string, isNoSQLDB bool, noSQLDB string, resources []*corenode.Resource, restClients []*corenode.RestClient) *Copier {
	var restResourceConfig = make(map[string]*frameworks.RestResourceData)
	pluralizeClient := pluralize.NewClient()

	// populate map to replace templates
	data := map[string]interface{}{
		"GitRepositoryName":   gitRepositoryName,
		"NodeName":            strings.ToLower(nodeName),
		"GitPlatformUserName": gitPlatformUserName,
		"GitPlatformURL":      strings.Replace(gitPlatformURL, "https://", "", -1),
	}
	data["SQLDB"] = sqlDB
	data["IsSQLDB"] = isSQLDB
	data["NoSQLDB"] = noSQLDB
	data["IsNoSQLDB"] = isNoSQLDB
	if isRestServer {
		var restResourceData []*frameworks.RestResourceData
		for _, r := range resources {
			lowerCamelResourceName := strcase.ToLowerCamel(r.Name)
			resourceData := frameworks.RestResourceData{
				SmallKebabCaseResourceNameSingular: strcase.ToKebab(r.Name),
				SmallSnakeCaseResourceNameSingular: strcase.ToSnake(r.Name),
				SmallResourceNameSingular:          lowerCamelResourceName,
				SmallResourceNamePlural:            pluralizeClient.Plural(lowerCamelResourceName),
				CapsResourceNameSingular:           r.Name,
				CapsResourceNamePlural:             pluralizeClient.Plural(r.Name),
			}
			frameworks.AddRESTAllowedMethods(&resourceData, r.AllowedMethods)
			restResourceConfig[r.Name] = &resourceData
			restResourceData = append(restResourceData, &resourceData)
		}
		data["RestResources"] = restResourceData
		data["RestServerPort"] = restServerPort
		data["IsRestServer"] = isRestServer
	}
	// if restClients slice has elements
	hasRestClients := len(restClients) > 0

	return &Copier{
		TemplatesRootPath: templatesRootPath,
		NodeDirectoryName: nodeDirectoryName,
		Data:              data,
		IsRestServer:      isRestServer,
		RestServerPort:    restServerPort,
		HasRestClients:    hasRestClients,
		SQLDB:             sqlDB,
		IsSQLDB:           isSQLDB,
		NoSQLDB:           noSQLDB,
		IsNoSQLDB:         isNoSQLDB,
		Resources:         resources,
		ResourceConfig:    restResourceConfig,
		RestClients:       restClients,
		PluralizeClient:   pluralizeClient,
	}
}

// createRestClientDirectories creates rest client directories.
func (c *Copier) createRestClientDirectories() error {
	clientDirectory := c.NodeDirectoryName + RestClientPath
	if err := utils.CreateDirectories(clientDirectory); err != nil {
		log.Errorf("error creating client directory: %v", err)
		return err
	}

	return nil
}

// createRestServerDirectories creates rest server directories.
func (c *Copier) createRestServerDirectories() error {
	configDirectory := c.NodeDirectoryName + ConfigPath
	controllersDirectory := c.NodeDirectoryName + ControllersPath
	modelsDirectory := c.NodeDirectoryName + ModelsPath
	servicesDirectory := c.NodeDirectoryName + ServicesPath
	daosDirectory := c.NodeDirectoryName + DaosPath
	if err := utils.CreateDirectories(configDirectory); err != nil {
		log.Errorf("error creating config directory: %v", err)
		return err
	}
	if err := utils.CreateDirectories(controllersDirectory); err != nil {
		log.Errorf("error creating controllers directory: %v", err)
		return err
	}
	if err := utils.CreateDirectories(modelsDirectory); err != nil {
		log.Errorf("error creating models directory: %v", err)
		return err
	}
	if err := utils.CreateDirectories(servicesDirectory); err != nil {
		log.Errorf("error creating services directory: %v", err)
		return err
	}
	if err := utils.CreateDirectories(daosDirectory); err != nil {
		log.Errorf("error creating daos directory: %v", err)
		return err
	}
	if c.IsSQLDB {
		// create directories for every resource's db client.
		sqlDBClientsDirectory := c.NodeDirectoryName + SQLDBClientsPath
		if err := utils.CreateDirectories(sqlDBClientsDirectory); err != nil {
			log.Errorf("error creating sql db clients directory: %v", err)
			return err
		}
	} else if c.IsNoSQLDB {
		// create directories for every resource's db client.
		noSQLDBClientsDirectory := c.NodeDirectoryName + NoSQLDBClientsPath
		if err := utils.CreateDirectories(noSQLDBClientsDirectory); err != nil {
			log.Errorf("error creating nosql db clients directory: %v", err)
			return err
		}
	}
	return nil
}

// copyRestServerResourceFiles copies rest server resource files from template and renames them as per resource config.
func (c *Copier) copyRestServerResourceFiles(resource *corenode.Resource) error {
	var filePaths []*string
	var err error
	resourceName := strcase.ToKebab(resource.Name)

	// copy sql controller/service/dao/models files to generated project
	if c.IsSQLDB {
		filePaths, err = c.copySQLDBResourceFiles(resourceName, filePaths)
		if err != nil {
			log.Errorf("error copying sql db resources: %v", err)
			return err
		}
	}

	// copy nosql controller/service/dao/models to generated project
	if c.IsNoSQLDB {
		filePaths, err = c.copyNoSQLDBResourceFiles(resourceName, filePaths)
		if err != nil {
			log.Errorf("error copying sql db resources: %v", err)
			return err
		}
	}

	// add resource-specific data to map in c needed for templates.
	err = c.addResourceSpecificTemplateData(resource)
	if err != nil {
		log.Errorf("error adding resource specific template data: %v", err)
		return err
	}

	// get func map for template
	funcMap := c.getFuncMap(resource)

	// apply template
	return executor.ExecuteWithFuncs(filePaths, c.Data, funcMap)
}

func (c *Copier) getFuncMap(resource *corenode.Resource) template.FuncMap {
	funcMap := template.FuncMap{
		"ToLowerCamelCase": func(key string) string {
			return strcase.ToLowerCamel(key)
		},
		"AddPointerIfCompositeField": func(key string) string {
			fieldMetaData, ok := resource.Fields[key]
			if ok && fieldMetaData.IsComposite {
				return "*" + fieldMetaData.Type
			}
			return fieldMetaData.Type
		},
		// This function helps the template to add a foreignKey based on composite field
		"AddForeignKeyIfCompositeField": func(key, value string) string {
			if fieldMetaData, ok := resource.Fields[key]; ok && fieldMetaData.IsComposite {
				return fmt.Sprintf("%s %s `gorm:\"foreignKey:ID\" json:\"%s,omitempty\"`", key, value, strcase.ToLowerCamel(key))
			}
			return fmt.Sprintf("%s %s `json:\"%s,omitempty\"`", key, value, strcase.ToLowerCamel(key))
		},
		"GetCompositeFields": func(key string) string {
			fieldMetaData, ok := resource.Fields[key]
			if ok && fieldMetaData.IsComposite {
				return key + ": &models." + fieldMetaData.Type + "{},"
			}
			return ""
		},
	}
	return funcMap
}

// copyRestClientResourceFiles copies rest client files from template and renames them as per client config.
func (c *Copier) copyRestClientResourceFiles(restClient *corenode.RestClient) error {
	/// add resource-specific data to map in c needed for templates.
	c.Data["RestClientPort"] = restClient.Port
	c.Data["RestClientServiceName"] = restClient.SourceNodeName
	c.Data["RestClientSourceNodeID"] = strings.Replace(cases.Title(language.Und, cases.NoLower).String(restClient.SourceNodeID), "-", "", -1)

	// copy restClient files to a generated project.
	targetResourceClientFileName := c.NodeDirectoryName + RestClientPath + "/" + restClient.SourceNodeName + "-" + ClientFile
	_, err := utils.CopyFile(targetResourceClientFileName, c.TemplatesRootPath+RestClientPath+"/"+ClientFile)
	if err != nil {
		log.Errorf("error while copying rest client file %s", targetResourceClientFileName)
		return err
	}
	var filePaths []*string
	filePaths = append(filePaths, &targetResourceClientFileName)

	// apply template
	return executor.Execute(filePaths, c.Data)
}

func (c *Copier) addSQLDetails(resource *corenode.Resource) error {
	var createQueryColumns *string
	var insertQueryColumns *string
	var insertQueryParams *string
	var insertQueryExecColumns *string
	var updateQueryColumnsAndParams *string
	var updateQueryExecColumns *string
	var getQueryScanColumns *string

	for key, value := range resource.Fields {
		dbDataType, err := c.getSQLDBDataType(value.Type)
		if err != nil {
			log.Errorf("error while getting db data type for %s", value.Type)
			return err
		}
		createQueryColumns = c.getCreateQueryColumns(createQueryColumns, key, value, dbDataType)
		insertQueryColumns, insertQueryParams, insertQueryExecColumns = c.getQueryParamsNColumnsNExecColumns(insertQueryColumns, insertQueryParams, insertQueryExecColumns, key, value)
		updateQueryColumnsAndParams, updateQueryExecColumns = c.getUpdateQueryColumnsAndParamsNExecColumns(updateQueryColumnsAndParams, updateQueryExecColumns, key, value)
		getQueryScanColumns = c.getGetQueryScanColumns(getQueryScanColumns, key, value)
	}
	// create query columns
	c.Data["CreateQueryColumns"] = createQueryColumns

	// insert query columns and params
	c.Data["InsertQueryColumns"] = insertQueryColumns
	c.Data["InsertQueryParams"] = insertQueryParams
	c.Data["InsertQueryExecColumns"] = insertQueryExecColumns

	// update query columns and params, execColumns
	c.Data["UpdateQueryColumnsAndParams"] = updateQueryColumnsAndParams
	c.Data["UpdateQueryExecColumns"] = updateQueryExecColumns

	// get query columns and params
	c.Data["GetQueryExecColumns"] = getQueryScanColumns
	return nil
}

func (c *Copier) addNoSQLDetails(resource *corenode.Resource) error {
	var createQueryColumns *string

	for key, value := range resource.Fields {
		dbDataType, err := c.getNoSQLDBDataType(value.Type)
		if err != nil {
			log.Debugf("error while getting db data type for %s", value.Type)
			return err
		}
		createQueryColumns = c.getNoSQLCreateQueryColumns(createQueryColumns, key, value, dbDataType)
	}
	// create query columns
	c.Data["CreateQueryColumns"] = createQueryColumns
	return nil
}

func (c *Copier) getUpdateQueryColumnsAndParamsNExecColumns(updateQueryColumnsAndParams, updateQueryExecColumns *string, key string, value corenode.FieldMetadata) (*string, *string) {
	if updateQueryColumnsAndParams != nil {
		if value.IsComposite {
			*updateQueryColumnsAndParams += ", " + cases.Title(language.Und, cases.NoLower).String(key) + " = ?"
			// m here is a model's variable
			*updateQueryExecColumns += ", m." + cases.Title(language.Und, cases.NoLower).String(key) + ".Id"
		} else {
			*updateQueryColumnsAndParams += ", " + cases.Title(language.Und, cases.NoLower).String(key) + " = ?"
			// m here is a model's variable
			*updateQueryExecColumns += ", m." + cases.Title(language.Und, cases.NoLower).String(key)
		}
	} else {
		updateQueryColumnsAndParams = new(string)
		updateQueryExecColumns = new(string)
		if value.IsComposite {
			*updateQueryColumnsAndParams = cases.Title(language.Und, cases.NoLower).String(key) + " = ?"
			// m here is a model's variable
			*updateQueryExecColumns = "m." + cases.Title(language.Und, cases.NoLower).String(key) + ".Id"
		} else {
			*updateQueryColumnsAndParams = cases.Title(language.Und, cases.NoLower).String(key) + " = ?"
			// m here is a model's variable
			*updateQueryExecColumns = "m." + cases.Title(language.Und, cases.NoLower).String(key)
		}
	}
	return updateQueryColumnsAndParams, updateQueryExecColumns
}

func (c *Copier) getQueryParamsNColumnsNExecColumns(insertQueryColumns, insertQueryParams, insertQueryExecColumns *string, key string, value corenode.FieldMetadata) (*string, *string, *string) {
	if insertQueryColumns != nil {
		if value.IsComposite {
			*insertQueryColumns += ", " + cases.Title(language.Und, cases.NoLower).String(key)
			*insertQueryParams += ", ?"
			// m here is a model's variable
			*insertQueryExecColumns += ", m." + cases.Title(language.Und, cases.NoLower).String(key) + ".Id"
		} else {
			*insertQueryColumns += ", " + cases.Title(language.Und, cases.NoLower).String(key)
			*insertQueryParams += ", ?"
			// m here is a model's variable
			*insertQueryExecColumns += ", m." + cases.Title(language.Und, cases.NoLower).String(key)
		}
	} else {
		insertQueryParams = new(string)
		insertQueryColumns = new(string)
		insertQueryExecColumns = new(string)
		if value.IsComposite {
			*insertQueryColumns = cases.Title(language.Und, cases.NoLower).String(key)
			*insertQueryParams = "?"
			// m here is a model's variable
			*insertQueryExecColumns = "m." + cases.Title(language.Und, cases.NoLower).String(key) + ".Id"
		} else {
			*insertQueryColumns = cases.Title(language.Und, cases.NoLower).String(key)
			*insertQueryParams = "?"
			// m here is a model's variable
			*insertQueryExecColumns = "m." + cases.Title(language.Und, cases.NoLower).String(key)
		}
	}
	return insertQueryColumns, insertQueryParams, insertQueryExecColumns
}

func (c *Copier) getCreateQueryColumns(createQueryColumns *string, key string, value corenode.FieldMetadata, dbDataType string) *string {
	if createQueryColumns != nil {
		if value.IsComposite {
			*createQueryColumns += "\n\t\t" + cases.Title(language.Und, cases.NoLower).String(key) + " " + dbDataType + " NULL,"
		} else {
			*createQueryColumns += "\n\t\t" + cases.Title(language.Und, cases.NoLower).String(key) + " " + dbDataType + " NOT NULL,"
		}
	} else {
		createQueryColumns = new(string)
		if value.IsComposite {
			*createQueryColumns = "\n\t\t" + cases.Title(language.Und, cases.NoLower).String(key) + " " + dbDataType + " NULL,"
		} else {
			*createQueryColumns = "\n\t\t" + cases.Title(language.Und, cases.NoLower).String(key) + " " + dbDataType + " NOT NULL,"
		}
	}
	return createQueryColumns
}

// Query Columns for NoSQL Column Based DB
func (c *Copier) getNoSQLCreateQueryColumns(createQueryColumns *string, key string, value corenode.FieldMetadata, dbDataType string) *string {
	if createQueryColumns != nil {
		if value.IsComposite {
			*createQueryColumns += "\n\t\t" + cases.Lower(language.Und, cases.NoLower).String(key) + " " + dbDataType + ","
		} else {
			*createQueryColumns += "\n\t\t" + cases.Lower(language.Und, cases.NoLower).String(key) + " " + dbDataType + ","
		}
	} else {
		createQueryColumns = new(string)
		if value.IsComposite {
			*createQueryColumns = "\n\t\t" + cases.Lower(language.Und, cases.NoLower).String(key) + " " + dbDataType + ","
		} else {
			*createQueryColumns = "\n\t\t" + cases.Lower(language.Und, cases.NoLower).String(key) + " " + dbDataType + ","
		}
	}
	return createQueryColumns
}
func (c *Copier) getGetQueryScanColumns(getQueryScanColumns *string, key string, value corenode.FieldMetadata) *string {
	if getQueryScanColumns != nil {
		if value.IsComposite {
			// m here is a model's variable
			*getQueryScanColumns += ", &m." + cases.Title(language.Und, cases.NoLower).String(key) + ".Id"
		} else {
			// m here is a model's variable
			*getQueryScanColumns += ", &m." + cases.Title(language.Und, cases.NoLower).String(key)
		}
	} else {
		getQueryScanColumns = new(string)
		if value.IsComposite {
			// m here is a model's variable
			*getQueryScanColumns = "&m." + cases.Title(language.Und, cases.NoLower).String(key) + ".Id"
		} else {
			// m here is a model's variable
			*getQueryScanColumns = "&m." + cases.Title(language.Und, cases.NoLower).String(key)
		}
	}
	return getQueryScanColumns
}

func (c *Copier) addResourceSpecificTemplateData(resource *corenode.Resource) error {
	// make every field public by making its first character capital.
	fields := map[string]string{}
	for key, value := range resource.Fields {
		fields[key] = value.Type
	}
	c.Data["Fields"] = fields
	// db fields
	if c.IsSQLDB {
		if c.SQLDB == SQLite || c.SQLDB == MySQL {
			err := c.addSQLDetails(resource)
			if err != nil {
				log.Debug("error while adding sql details to resource specific template data", err)
				return err
			}
		}
	}

	if c.IsNoSQLDB {
		if c.NoSQLDB == ScyllaDB {
			err := c.addNoSQLDetails(resource)
			if err != nil {
				log.Debug("error while adding nosql details to resource specific template data", err)
				return err
			}
		}
	}

	// Add another map with the below keys at root level (this is for specific resource for this iteration)
	restResourceData := c.ResourceConfig[resource.Name]
	c.Data["SmallKebabCaseResourceNameSingular"] = restResourceData.SmallKebabCaseResourceNameSingular
	c.Data["SmallSnakeCaseResourceNameSingular"] = restResourceData.SmallSnakeCaseResourceNameSingular
	c.Data["SmallResourceNameSingular"] = restResourceData.SmallResourceNameSingular
	c.Data["SmallResourceNamePlural"] = restResourceData.SmallResourceNamePlural
	c.Data["CapsResourceNameSingular"] = restResourceData.CapsResourceNameSingular
	c.Data["CapsResourceNamePlural"] = restResourceData.CapsResourceNamePlural
	c.Data["IsRESTCreateAllowed"] = restResourceData.IsRESTCreateAllowed
	c.Data["IsRESTListAllowed"] = restResourceData.IsRESTListAllowed
	c.Data["IsRESTGetAllowed"] = restResourceData.IsRESTGetAllowed
	c.Data["IsRESTPutAllowed"] = restResourceData.IsRESTPutAllowed
	c.Data["IsRESTDeleteAllowed"] = restResourceData.IsRESTDeleteAllowed
	c.Data["IsRESTPatchAllowed"] = restResourceData.IsRESTPatchAllowed
	c.Data["IsRESTOptionsAllowed"] = restResourceData.IsRESTPatchAllowed
	c.Data["IsRESTHeadAllowed"] = restResourceData.IsRESTHeadAllowed

	return nil
}

func (c *Copier) copyNoSQLDBResourceFiles(resourceName string, filePaths []*string) ([]*string, error) {
	var err error
	// copy service files to a generated project
	targetResourceServiceFileName := c.NodeDirectoryName + ServicesPath + "/" + resourceName + "-" + strings.Replace(NoSQLServiceFile, "nosqls-", "", 1)
	_, err = utils.CopyFile(targetResourceServiceFileName, c.TemplatesRootPath+ServicesPath+"/"+NoSQLServiceFile)
	if err != nil {
		log.Debugf("error copying service file: %v", err)
		return nil, err
	}
	filePaths = append(filePaths, &targetResourceServiceFileName)

	var targetResourceDaoFileName string
	if c.NoSQLDB == MongoDB {
		// copy controller files to a generated project
		targetResourceControllerFileName := c.NodeDirectoryName + ControllersPath + "/" + resourceName + "-" + strings.Replace(NoSQLControllerFile, "nosqls-", "", 1)
		_, err := utils.CopyFile(targetResourceControllerFileName, c.TemplatesRootPath+ControllersPath+"/"+NoSQLControllerFile)
		if err != nil {
			log.Debugf("error copying controller file: %v", err)
			return nil, err
		}
		filePaths = append(filePaths, &targetResourceControllerFileName)

		// copy model files to a generated project
		targetResourceModelFileName := c.NodeDirectoryName + ModelsPath + "/" + resourceName + "-" + strings.Replace(NoSQLModelFile, "nosqls-", "", 1)
		_, err = utils.CopyFile(targetResourceModelFileName, c.TemplatesRootPath+ModelsPath+"/"+NoSQLModelFile)
		if err != nil {
			log.Debugf("error copying model file: %v", err)
			return nil, err
		}
		filePaths = append(filePaths, &targetResourceModelFileName)

		// dao files
		targetResourceDaoFileName = c.NodeDirectoryName + DaosPath + "/" + resourceName + "-" + MongoDBDaoFile
		_, err = utils.CopyFile(targetResourceDaoFileName, c.TemplatesRootPath+DaosPath+"/"+MongoDBDaoFile)
		if err != nil {
			log.Debugf("error copying mongodb dao file: %v", err)
			return nil, err
		}
		filePaths = append(filePaths, &targetResourceDaoFileName)
	}

	if c.NoSQLDB == ScyllaDB {
		// copy controller files to a generated project
		targetResourceControllerFileName := c.NodeDirectoryName + ControllersPath + "/" + resourceName + "-" + strings.Replace(NoSQLColumnBasedControllerFile, "nosqls-column-based-", "", 1)
		_, err := utils.CopyFile(targetResourceControllerFileName, c.TemplatesRootPath+ControllersPath+"/"+NoSQLColumnBasedControllerFile)
		if err != nil {
			log.Debugf("error copying controller file: %v", err)
			return nil, err
		}
		filePaths = append(filePaths, &targetResourceControllerFileName)

		// copy model files to a generated project
		targetResourceModelFileName := c.NodeDirectoryName + ModelsPath + "/" + resourceName + "-" + strings.Replace(NoSQLColumnBasedModelFile, "nosqls-column-based-", "", 1)
		_, err = utils.CopyFile(targetResourceModelFileName, c.TemplatesRootPath+ModelsPath+"/"+NoSQLColumnBasedModelFile)
		if err != nil {
			log.Debugf("error copying model file: %v", err)
			return nil, err
		}
		filePaths = append(filePaths, &targetResourceModelFileName)

		// dao files
		targetResourceDaoFileName = c.NodeDirectoryName + DaosPath + "/" + resourceName + "-" + ScyllaDBDaoFile
		_, err = utils.CopyFile(targetResourceDaoFileName, c.TemplatesRootPath+DaosPath+"/"+ScyllaDBDaoFile)
		if err != nil {
			log.Debugf("error copying mongodb dao file: %v", err)
			return nil, err
		}
		filePaths = append(filePaths, &targetResourceDaoFileName)
	}

	return filePaths, nil
}

func (c *Copier) copySQLDBResourceFiles(resourceName string, filePaths []*string) ([]*string, error) {
	// copy controller files to a generated project
	targetResourceControllerFileName := c.NodeDirectoryName + ControllersPath + "/" + resourceName + "-" + strings.Replace(SQLControllerFile, "sqls-", "", 1)
	_, err := utils.CopyFile(targetResourceControllerFileName, c.TemplatesRootPath+ControllersPath+"/"+SQLControllerFile)
	if err != nil {
		log.Debugf("error copying controller file: %v", err)
		return nil, err
	}
	filePaths = append(filePaths, &targetResourceControllerFileName)

	// copy service files to a generated project
	targetResourceServiceFileName := c.NodeDirectoryName + ServicesPath + "/" + resourceName + "-" + strings.Replace(SQLServiceFile, "sqls-", "", 1)
	_, err = utils.CopyFile(targetResourceServiceFileName, c.TemplatesRootPath+ServicesPath+"/"+SQLServiceFile)
	if err != nil {
		log.Debugf("error copying service file: %v", err)
		return nil, err
	}
	filePaths = append(filePaths, &targetResourceServiceFileName)

	var targetResourceDaoFileName string
	if c.SQLDB == SQLite {
		// model files
		// copy model files to a generated project
		targetResourceModelFileName := c.NodeDirectoryName + ModelsPath + "/" + resourceName + "-" + strings.Replace(SQLModelFile, "sqls-", "", 1)
		_, err = utils.CopyFile(targetResourceModelFileName, c.TemplatesRootPath+ModelsPath+"/"+SQLModelFile)
		if err != nil {
			log.Debugf("error copying model file: %v", err)
			return nil, err
		}
		filePaths = append(filePaths, &targetResourceModelFileName)

		// dao files
		targetResourceDaoFileName = c.NodeDirectoryName + DaosPath + "/" + resourceName + "-" + SQLiteDaoFile
		_, err := utils.CopyFile(targetResourceDaoFileName, c.TemplatesRootPath+DaosPath+"/"+SQLiteDaoFile)
		if err != nil {
			log.Debugf("error copying sqlite dao file: %v", err)
			return nil, err
		}
		filePaths = append(filePaths, &targetResourceDaoFileName)
	} else if c.SQLDB == MySQL {
		// model files
		targetResourceModelFileName := c.NodeDirectoryName + ModelsPath + "/" + resourceName + "-" + strings.Replace(SQLModelFile, "sqls-", "", 1)
		_, err = utils.CopyFile(targetResourceModelFileName, c.TemplatesRootPath+ModelsPath+"/"+SQLModelFile)
		if err != nil {
			log.Debugf("error copying model file: %v", err)
			return nil, err
		}
		filePaths = append(filePaths, &targetResourceModelFileName)

		// dao files
		targetResourceDaoFileName = c.NodeDirectoryName + DaosPath + "/" + resourceName + "-" + MySQLDaoFile
		_, err := utils.CopyFile(targetResourceDaoFileName, c.TemplatesRootPath+DaosPath+"/"+MySQLDaoFile)
		if err != nil {
			log.Debugf("error copying mysql dao file: %v", err)
			return nil, err
		}
		filePaths = append(filePaths, &targetResourceDaoFileName)
	} else if c.SQLDB == Map {
		// model files
		// copy model files to a generated project
		targetResourceModelFileName := c.NodeDirectoryName + ModelsPath + "/" + resourceName + "-" + strings.Replace(SQLModelFile, "sqls-", "", 1)
		_, err = utils.CopyFile(targetResourceModelFileName, c.TemplatesRootPath+ModelsPath+"/"+SQLModelFile)
		if err != nil {
			log.Debugf("error copying model file: %v", err)
			return nil, err
		}
		filePaths = append(filePaths, &targetResourceModelFileName)
		targetResourceDaoFileName = c.NodeDirectoryName + DaosPath + "/" + resourceName + "-" + DaoFile
		_, err := utils.CopyFile(targetResourceDaoFileName, c.TemplatesRootPath+DaosPath+"/"+DaoFile)
		if err != nil {
			log.Debugf("error copying dao file: %v", err)
			return nil, err
		}
		filePaths = append(filePaths, &targetResourceDaoFileName)
	} else if c.SQLDB == SQLiteGORM {
		// model files
		targetResourceModelFileName := c.NodeDirectoryName + ModelsPath + "/" + resourceName + "-" + strings.Replace(SQLGORMModelFile, "sqls-gorm-", "", 1)
		_, err = utils.CopyFile(targetResourceModelFileName, c.TemplatesRootPath+ModelsPath+"/"+SQLGORMModelFile)
		if err != nil {
			log.Debugf("error copying model file: %v", err)
			return nil, err
		}
		filePaths = append(filePaths, &targetResourceModelFileName)

		// dao files
		targetResourceDaoFileName = c.NodeDirectoryName + DaosPath + "/" + resourceName + "-" + strings.Replace(SQLiteGORMDaoFile, "sqlite-gorm-", "", 1)
		_, err := utils.CopyFile(targetResourceDaoFileName, c.TemplatesRootPath+DaosPath+"/"+SQLiteGORMDaoFile)
		if err != nil {
			log.Debugf("error copying sqlite gorm dao file: %v", err)
			return nil, err
		}
		filePaths = append(filePaths, &targetResourceDaoFileName)
	} else if c.SQLDB == MySQLGORM {
		// model files
		targetResourceModelFileName := c.NodeDirectoryName + ModelsPath + "/" + resourceName + "-" + strings.Replace(SQLGORMModelFile, "sqls-gorm-", "", 1)
		_, err = utils.CopyFile(targetResourceModelFileName, c.TemplatesRootPath+ModelsPath+"/"+SQLGORMModelFile)
		if err != nil {
			log.Debugf("error copying model file: %v", err)
			return nil, err
		}
		filePaths = append(filePaths, &targetResourceModelFileName)

		// dao files
		targetResourceDaoFileName = c.NodeDirectoryName + DaosPath + "/" + resourceName + "-" + strings.Replace(MySQLGORMDaoFile, "mysql-gorm-", "", 1)
		_, err := utils.CopyFile(targetResourceDaoFileName, c.TemplatesRootPath+DaosPath+"/"+MySQLGORMDaoFile)
		if err != nil {
			log.Debugf("error copying mysql gorm dao file: %v", err)
			return nil, err
		}
		filePaths = append(filePaths, &targetResourceDaoFileName)
	}
	return filePaths, nil
}

// CreateRestConfigs creates/copies relevant files to a generated project
func (c *Copier) CreateRestConfigs() error {
	if err := c.CreateRestServer(); err != nil {
		log.Debugf("error creating rest server: %v", err)
		return err
	}
	if err := c.CreateRestClients(); err != nil {
		log.Debugf("error creating rest clients: %v", err)
		return err
	}
	return nil
}

// CreateRestServer creates/copies relevant files to a generated project
func (c *Copier) CreateRestServer() error {
	// if the node is server, add server code
	if c.IsRestServer {
		// create directories for controller, service, dao, models
		if err := c.createRestServerDirectories(); err != nil {
			log.Debugf("error creating rest server directories: %v", err)
			return err
		}
		// copy files with respect to the names of resources
		for _, resource := range c.Resources {
			if err := c.copyRestServerResourceFiles(resource); err != nil {
				log.Debugf("error copying rest server resource files: %v", err)
				return err
			}
		}
		// copy open telemetry config file
		var filePaths []*string
		// client files
		targetOpenTelConfigFileName := c.NodeDirectoryName + ConfigPath + "/" + ConfigFile
		_, err := utils.CopyFile(targetOpenTelConfigFileName, c.TemplatesRootPath+ConfigPath+"/"+ConfigFile)
		if err != nil {
			log.Debugf("error copying opentel config file: %v", err)
			return err
		}
		filePaths = append(filePaths, &targetOpenTelConfigFileName)
		err = executor.Execute(filePaths, c.Data)
		if err != nil {
			log.Debugf("error executing opentel config file: %v", err)
			return err
		}

		if c.IsSQLDB {
			// create sql db config file (common to all resources for specific database)
			// No vars in config file as of now, but in future they may be there.
			if c.SQLDB == SQLite {
				var filePaths []*string
				// client files
				targetSQLiteConfigFileName := c.NodeDirectoryName + SQLDBClientsPath + "/" + SQLiteDBConfigFile
				_, err := utils.CopyFile(targetSQLiteConfigFileName, c.TemplatesRootPath+SQLDBClientsPath+"/"+SQLiteDBConfigFile)
				if err != nil {
					log.Debugf("error copying sqlite config file: %v", err)
					return err
				}
				filePaths = append(filePaths, &targetSQLiteConfigFileName)
				return executor.Execute(filePaths, c.Data)
			} else if c.SQLDB == MySQL {
				var filePaths []*string
				// client files
				targetMySQLConfigFileName := c.NodeDirectoryName + SQLDBClientsPath + "/" + MySQLDBConfigFile
				_, err := utils.CopyFile(targetMySQLConfigFileName, c.TemplatesRootPath+SQLDBClientsPath+"/"+MySQLDBConfigFile)
				if err != nil {
					log.Debugf("error copying mysql config file: %v", err)
					return err
				}
				filePaths = append(filePaths, &targetMySQLConfigFileName)
				return executor.Execute(filePaths, c.Data)
			} else if c.SQLDB == SQLiteGORM {
				var filePaths []*string
				// client files
				targetSQLiteConfigFileName := c.NodeDirectoryName + SQLDBClientsPath + "/" + SQLiteGORMDBConfigFile
				_, err := utils.CopyFile(targetSQLiteConfigFileName, c.TemplatesRootPath+SQLDBClientsPath+"/"+SQLiteGORMDBConfigFile)
				if err != nil {
					log.Debugf("error copying sqlite gorm config file: %v", err)
					return err
				}
				filePaths = append(filePaths, &targetSQLiteConfigFileName)
				return executor.Execute(filePaths, c.Data)
			} else if c.SQLDB == MySQLGORM {
				var filePaths []*string
				// client files
				targetMySQLConfigFileName := c.NodeDirectoryName + SQLDBClientsPath + "/" + MySQLGORMDBConfigFile
				_, err := utils.CopyFile(targetMySQLConfigFileName, c.TemplatesRootPath+SQLDBClientsPath+"/"+MySQLGORMDBConfigFile)
				if err != nil {
					log.Debugf("error copying mysql gorm config file: %v", err)
					return err
				}
				filePaths = append(filePaths, &targetMySQLConfigFileName)
				return executor.Execute(filePaths, c.Data)
			} else if c.SQLDB == Map {
				var filePaths []*string
				// client files
				targetMapConfigFileName := c.NodeDirectoryName + SQLDBClientsPath + "/" + MapDBConfigFile
				_, err := utils.CopyFile(targetMapConfigFileName, c.TemplatesRootPath+SQLDBClientsPath+"/"+MapDBConfigFile)
				if err != nil {
					log.Debugf("error copying map config file: %v", err)
					return err
				}
				filePaths = append(filePaths, &targetMapConfigFileName)
				return executor.Execute(filePaths, c.Data)
			}
		} else if c.IsNoSQLDB {
			// create nosql db config file (common to all resources for specific database)
			// No vars in config file as of now, but in future they may be there.
			if c.NoSQLDB == MongoDB {
				var filePaths []*string
				// client files
				targetMongoDBConfigFileName := c.NodeDirectoryName + NoSQLDBClientsPath + "/" + MongoDBConfigFile
				_, err := utils.CopyFile(targetMongoDBConfigFileName, c.TemplatesRootPath+NoSQLDBClientsPath+"/"+MongoDBConfigFile)
				if err != nil {
					log.Debugf("error copying mongodb config file: %v", err)
					return err
				}
				filePaths = append(filePaths, &targetMongoDBConfigFileName)
				return executor.Execute(filePaths, c.Data)
			}

			if c.NoSQLDB == ScyllaDB {
				var filePaths []*string
				// client files
				targetScyllaDBConfigFileName := c.NodeDirectoryName + NoSQLDBClientsPath + "/" + ScyllaDBConfigFile
				_, err := utils.CopyFile(targetScyllaDBConfigFileName, c.TemplatesRootPath+NoSQLDBClientsPath+"/"+ScyllaDBConfigFile)
				if err != nil {
					log.Debugf("error copying scylla config file: %v", err)
					return err
				}
				filePaths = append(filePaths, &targetScyllaDBConfigFileName)
				return executor.Execute(filePaths, c.Data)
			}
		}
	}
	return nil
}

// CreateRestClients creates/copies relevant files to a generated project
func (c *Copier) CreateRestClients() error {
	// if the node is client, add client code
	if c.HasRestClients {
		// create directories for a client
		if err := c.createRestClientDirectories(); err != nil {
			return err
		}

		// copy files with respect to the names of resources
		for _, client := range c.RestClients {
			if err := c.copyRestClientResourceFiles(client); err != nil {
				return err
			}
		}
	}
	return nil
}

// CreateRootLevelFiles copies all root level files at language template.
func (c *Copier) CreateRootLevelFiles() error {
	err := utils.CopyFiles(c.NodeDirectoryName, c.TemplatesRootPath)
	if err != nil {
		return err
	}
	_, files, err0 := utils.GetDirectoriesAndFilePaths(c.NodeDirectoryName)
	if err0 != nil {
		return err0
	}
	return executor.Execute(files, c.Data)
}

func (c *Copier) getSQLDBDataType(value string) (string, error) {
	if c.SQLDB == SQLite {
		return commonUtils.GetSqliteDataType(value), nil
	} else if c.SQLDB == MySQL {
		return commonUtils.GetMySQLDataType(value), nil
	}
	return "", errors.New("database not supported")
}

func (c *Copier) getNoSQLDBDataType(value string) (string, error) {
	if c.NoSQLDB == ScyllaDB {
		return commonUtils.GetColumnBasedDBDataType(value), nil
	}
	return "", errors.New("database not supported")
}
