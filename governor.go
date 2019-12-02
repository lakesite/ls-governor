// governor governs the interaction between service APIs, configurations and 
// model/datastore logic.
package governor

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/lakesite/ls-config"
	"github.com/lakesite/ls-fibre"
	"github.com/lakesite/ls-superbase"
	"github.com/pelletier/go-toml"
)

// API contains a fibre web service and our management service.
type API struct {
	WebService *fibre.WebService
	ManagerService *ManagerService
}

func NewAPI(ws *fibre.WebService, ms *ManagerService) *API {
	return &API{WebService: ws, ManagerService: ms}
}

// ManagerService contains the configuration settings required to manage the api.
type ManagerService struct {
	Config   *toml.Tree
	DBConfig map[string]*superbase.DBConfig
}

// GetAppProperty gets the property for app as a string, if property does not 
// exist return err.
func (ms *ManagerService) GetAppProperty(app string, property string) (string, error) {
	if ms.Config.Get(app+"."+property) != nil {
		return ms.Config.Get(app + "." + property).(string), nil
	} else {
		return "", fmt.Errorf("Configuration missing '%s' section under [%s] heading.\n", property, app)
	}
}

// InitDatastore initializes the datastore by app name
// return true if successful false otherwise
func (ms *ManagerService) InitDatastore(app string) bool {
	if ms.DBConfig[app] == nil {
		ms.DBConfig[app] = &superbase.DBConfig{}
	}

	success := true

	// pull in the database config to DBConfig struct
	ms.DBConfig[app].Server, _ = ms.GetAppProperty(app, "dbserver")
	ms.DBConfig[app].Port, _ = ms.GetAppProperty(app, "dbport")
	ms.DBConfig[app].Database, _ = ms.GetAppProperty(app, "database")
	ms.DBConfig[app].User, _ = ms.GetAppProperty(app, "dbuser")
	ms.DBConfig[app].Password, _ = ms.GetAppProperty(app, "dbpassword")
	ms.DBConfig[app].Driver, _ = ms.GetAppProperty(app, "dbdriver")
	ms.DBConfig[app].Path, _ = ms.GetAppProperty(app, "dbpath")

	// Init the DB, which pulls in our gorm DB struct;
	ms.DBConfig[app].Init()

	return success
}

// InitManager reads in configuration data and prepares the datastore config.
func (ms *ManagerService) InitManager(cfgfile string) {
	if _, err := os.Stat(cfgfile); os.IsNotExist(err) {
		log.Fatalf("File '%s' does not exist.\n", cfgfile)
	} else {
		ms.Config, _ = toml.LoadFile(cfgfile)
		ms.DBConfig = make(map[string]*superbase.DBConfig)
	}
}


// CreateAPI sets up the web service for app
func (ms *ManagerService) CreateAPI(app string) *API {
	// the env convention here is APPNAME_HOST and APPNAME_PORT
	ua := strings.ToUpper(app)
	address := config.Getenv(ua + "_HOST", "127.0.0.1") + ":" + config.Getenv(ua + "_PORT", "7990")
	ws := fibre.NewWebService(app, address)
	
	// Create a new API bridge
	api := NewAPI(
		ws, // web service
		ms,	// manager service
	)

	return api
}

// Daemonize the API.
func (ms *ManagerService) Daemonize(api *API) {
	api.WebService.RunWebServer()
}
