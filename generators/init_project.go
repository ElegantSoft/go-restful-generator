package generators

import (
	"fmt"
	"github.com/ElegantSoft/go-crud-starter/pkg/writetemplate"
	"github.com/manifoldco/promptui"
	"log"
	"os"
	"path/filepath"
)

var wd, _ = os.Getwd()
var mainTemplate = filepath.Join(wd, "templates", "main.tmpl")

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

	writetemplate.ProcessTemplate("../../templates/main.tmpl", "main.tmpl", filepath.Join("_example", "main.go"), data)
}
