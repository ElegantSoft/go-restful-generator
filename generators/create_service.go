package generators

import (
	_ "embed"
	"github.com/ElegantSoft/go-restful-generator/pkg/writetemplate"
	"github.com/iancoleman/strcase"
	"log"
	"os"
	"path/filepath"
)

//go:embed templates/routes.tmpl
var routerTemplate string

//go:embed templates/controller.tmpl
var controllerTemplate string

//go:embed templates/service.tmpl
var serviceTemplate string

//go:embed templates/repository.tmpl
var repositoryTemplate string

//go:embed templates/model.tmpl
var modelTemplate string

func GenerateService(packageName string, serviceName string, servicePath string) {

	type Data struct {
		PackageName string
		ServiceName string
	}
	data := Data{PackageName: packageName, ServiceName: serviceName}
	if servicePath == "" {
		servicePath = "lib/" + strcase.ToKebab(serviceName)
	}

	log.Printf("servicePath: %v", servicePath)

	err := os.MkdirAll(servicePath, os.ModePerm)
	if err != nil {
		log.Fatal(err)
		return
	}

	writetemplate.ProcessTemplate(routerTemplate, "router.tmpl", filepath.Join(servicePath, "router.go"), data)
	writetemplate.ProcessTemplate(controllerTemplate, "controller.tmpl", filepath.Join(servicePath, "controller.go"), data)
	writetemplate.ProcessTemplate(serviceTemplate, "service.tmpl", filepath.Join(servicePath, "service.go"), data)
	writetemplate.ProcessTemplate(repositoryTemplate, "repository.tmpl", filepath.Join(servicePath, "repository.go"), data)
	writetemplate.ProcessTemplate(modelTemplate, "model.tmpl", filepath.Join("db/models", strcase.ToSnake(serviceName)+".go"), data)

}
