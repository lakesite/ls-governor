# ls-governor #

📦 Microservice Management API 📦

Governor governs the interaction between service APIs, configurations and models.
It provides a set of conventions to create web APIs that interact with 
datastores through a model.

## usage ##

Governor provides a standard way of setting up APIs.  First, create a manager
service that reads from a config.toml:

```
	gms := &governor.ManagerService{}
	if config == "" {
		config = "config.toml"
	}
```

Initialize the manager with the config file, use the configuration to initialize
the datastore (if needed), then create a governor API for this application:

```
	// setup manager and create api
	gms.InitManager(config)
	gms.InitDatastore("example_app")
	gapi := gms.CreateAPI("example_app")
```

Next, using your model package, run Migrate with the governor API for your
example application.

```
	model.Migrate(gapi, "example_app")
```

### pkg/model/migrations.go ###

Below is an example convention for how you might handle your app's migrations.
Under the hood, governor uses [ls-superbase](https://github.com/lakesite/ls-superbase),
which uses gorm and provides gorm's drivers for sqlite, mssql, postgresql, and 
mysql:

```
package model

import (
	"errors"
	"fmt"

	"github.com/lakesite/ls-governor"
)

// Migrate takes a governor API and app name and migrates models, returns error
func Migrate(gapi *governor.API, app string) error {
	if gapi == nil {
		return errors.New("Migrate: Governor API is not initialized.")
	}

	if app == "" {
		return errors.New("Migrate: App name cannot be empty.")
	}

	dbc := gapi.ManagerService.DBConfig[app]
	
	if dbc == nil {
		return fmt.Errorf("Migrate: Database configuration for '%s' does not exist.", app)
	}

	if dbc.Connection == nil {
		return fmt.Errorf("Migrate: Database connection for '%s' does not exist.", app)
	}

	dbc.Connection.AutoMigrate(&YourGormModel{})
	return nil
}
```

The AutoMigrate feature of gorm is called against YourGormModel:

### pkg/models/YourGormModel.go

```
package model

import (
	"github.com/jinzhu/gorm"
)

// Example YourGormModel
type YourGormModel struct {
	gorm.Model
	StringField    string
}

```

Next, setup your routes using a wrapper convention:

```
	api.SetupRoutes(gapi)
```

### pkg/api/routes.go ###

Use a wrapper convention to define routes, e.g., YGMHandler:

```
package api

import (
	"net/http"

	"github.com/lakesite/ls-governor"
)

// SetupRoutes defines and associates routes to handlers.
// Use a wrapper convention to pass a governor API to each handler.
func SetupRoutes(gapi *governor.API) {
	gapi.WebService.Router.HandleFunc(
		"/example_app/api/v1/yourgormmodel/", 
		func(w http.ResponseWriter, r *http.Request) {
			YGMHandler(w, r, gapi)
		},
	).Methods("POST")
}
```

### pkg/api/handlers.go ###

```
// api contains the handlers to manage API endpoints
package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/schema"
	"github.com/lakesite/ls-governor"

	"github.com/path/to/your/pkg/models"
)

// YGMHandler handles POST data for a YourGormModel.
func YGMHandler(w http.ResponseWriter, r *http.Request, gapi *governor.API) {
	// parse the form
	err := r.ParseForm()
	if err != nil {
		gapi.WebService.JsonStatusResponse(w, "Error parsing form data.", http.StatusBadRequest)
	}

	// create a new YourGormModel
	ygm := new(model.YourGormModel)

	// using a new decoder, decode the form and bind it to the ygm
	decoder := schema.NewDecoder()
	decoder.Decode(ygm, r.Form)

	// insert the ygm structure
	if dbc := gapi.ManagerService.DBConfig["example_app"].Connection.Create(ygm); dbc.Error != nil {
		gapi.WebService.JsonStatusResponse(w, fmt.Sprintf("Error: %s", dbc.Error.Error()), http.StatusInternalServerError)
		return
	}

	// Return StatusOK with ygm made:
	gapi.WebService.JsonStatusResponse(w, "YGM made", http.StatusOK)
}
```

Now you can daemonize the service so it listens for connections:

```
	// now daemonize the api
	gms.Daemonize(gapi)
```

## Example ##

For a full example application, see: [zefram](https://github.com/lakesite/zefram)

The following summarizes the above into an example main.go.

### config.toml ###

```
[example_app]
dbdriver = "sqlite3"
dbpath   = "example.db"
other    = "settings..."
```

### main.go ###

Replace github.com/path/to/your/pkg/models with the appropriate reference to
your models

```
package main

import (
	"github.com/lakesite/ls-governor"

	"github.com/path/to/your/pkg/models"
	"github.com/path/to/your/pkg/api"
)

func main() {
	gms := &governor.ManagerService{}
	if config == "" {
		config = "config.toml"
	}

	// setup manager and create api
	gms.InitManager(config)
	gms.InitDatastore("example_app")
	gapi := gms.CreateAPI("example_app")

	// bridge logic
	model.Migrate(gapi, "example_app")
	api.SetupRoutes(gapi)

	// now daemonize the api
	gms.Daemonize(gapi)
}

```

## dependencies ##

1. [ls-config](https://github.com/lakesite/ls-config)
2. [ls-fibre](https://github.com/lakesite/ls-fibre)
3. [ls-superbase](https://github.com/lakesite/ls-superbase)
4. [go-toml](https://github.com/pelletier/go-toml)

## license ##

MIT
