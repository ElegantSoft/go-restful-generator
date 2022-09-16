package generators

import (
	"fmt"
	"github.com/ElegantSoft/go-crud-starter/pkg/writetemplate"
	"github.com/manifoldco/promptui"
	"os"
	"path/filepath"
)

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
	wd, _ := os.Getwd()

	writetemplate.ProcessTemplate(filepath.Join(wd, "templates", "main.tmpl"), "main.tmpl", filepath.Join("_example", "main.go"), data)
}
