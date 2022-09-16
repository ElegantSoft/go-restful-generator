package generators

import (
	_ "embed"
	"fmt"
	"github.com/ElegantSoft/go-crud-starter/pkg/writetemplate"
	"github.com/manifoldco/promptui"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

var wd, _ = os.Getwd()

//go:embed templates/main.tmpl
var mainTemplate string

func Execute() {

	prompt := promptui.Prompt{
		Label: "package name",
	}

	result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}
	type Data struct {
		PackageName string
	}
	data := Data{PackageName: result}
	log.Printf("Prompt %v", filepath.Join("$GOPATH", "src", "github.com"))
	_, filename, _, _ := runtime.Caller(-1)
	dirname := filepath.Dir(filename)
	log.Printf("Dir name is %s\n", dirname)

	writetemplate.ProcessTemplate(mainTemplate, "main.tmpl", filepath.Join("_example", "main.go"), data)
}
