package generators

import (
	_ "embed"
	"github.com/ElegantSoft/go-crud-starter/pkg/writetemplate"
	"log"
	"path/filepath"
	"runtime"
)

//go:embed templates/main.tmpl
var mainTemplate string

func InitNewProject(packageName string) {

	type Data struct {
		PackageName string
	}
	data := Data{PackageName: packageName}
	log.Printf("Prompt %v", filepath.Join("$GOPATH", "src", "github.com"))
	_, filename, _, _ := runtime.Caller(-1)
	dirname := filepath.Dir(filename)
	log.Printf("Dir name is %s\n", dirname)

	writetemplate.ProcessTemplate(mainTemplate, "main.tmpl", filepath.Join("_example", "main.go"), data)
}
