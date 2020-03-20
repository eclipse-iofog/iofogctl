# Deploy applications Package

This package contains executors to deploy iofog applications and microservices using the `client` package.

## Usage

```go
import (
	deploytypes "github.com/eclipse-iofog/iofog-go-sdk/pkg/deployapps"
	deploy "github.com/eclipse-iofog/iofog-go-sdk/pkg/deployapps/application"
)

// Create your Controller access structure
controller := deploytypes.IofogController{
  Endpoint: "127.0.0.1:51121",
	Email:    "user@domain.com",
	Password: "kj2gh0ooiwbug",
}

// Create your application structure
application := deploytypes.Application{
  Name: "my-app"
  //...Rest of the fields
}

// OR, read it from a yaml file
yamlFile, err := ioutil.ReadFile(filename)
if err != nil {
  return err
}
err = yaml.Unmarshal(yamlFile, &application)
if err != nil {
  return err
}

// Deploy
err = deploy.Execute(controller, application)
```