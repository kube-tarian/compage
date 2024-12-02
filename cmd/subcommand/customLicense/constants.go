package customlicense

const exampleCommand = `
# Convert XML file to JSON and YAML with the file path provided in the command line
compage customLicense path=https://raw.githubusercontent.com/licenses/license-templates/master/templates/apache.txt projectPath=/some/local/path/LICENSE
`
var (
	path = "https://raw.githubusercontent.com/licenses/license-templates/master/templates/apache.txt"
	projectPath = "LICENSE"
)
