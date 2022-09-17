package main

import (
	"github.com/ElegantSoft/go-crud-starter/common"
	"github.com/ElegantSoft/go-crud-starter/generators"
	"github.com/manifoldco/promptui"
)

func main() {
	//promptGetServiceName := promptui.Prompt{
	//	Label: "service name",
	//}

	promptSelectGenerator := promptui.Select{
		Label: "choose generator",
		Items: []string{"create new service", "init new project"},
	}

	index, _, err := promptSelectGenerator.Run()
	if err != nil {
		panic(err)
	}
	switch index {
	case 0:
		panic("not implemented")
	case 1:
		moduleName := common.GetModuleName()
		generators.InitNewProject(moduleName)
	default:
		panic("not implemented")
	}

	//result, err := prompt.Run()
}
