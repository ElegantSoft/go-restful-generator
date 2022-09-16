package generators

import (
	"fmt"
	"github.com/ElegantSoft/go-crud-starter/pkg/writetemplate"
	"github.com/manifoldco/promptui"
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
	writetemplate.ProcessTemplate(mainTemplate, "main.tmpl", filepath.Join("_example", "main.go"), data)
}
