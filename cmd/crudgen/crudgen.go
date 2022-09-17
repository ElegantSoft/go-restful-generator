package main

import (
	"fmt"
	"github.com/ElegantSoft/go-crud-starter/common"
	"github.com/ElegantSoft/go-crud-starter/generators"
	"github.com/spf13/cobra"
	"log"
	"os"
)

func main() {
	//promptGetServiceName := promptui.Prompt{
	//	Label: "service name",
	//}
	moduleName := common.GetModuleName()

	rootCmd := &cobra.Command{
		Use:   "crudgen",
		Short: "crudgen gen cli tool",
		Run: func(cmd *cobra.Command, args []string) {
			log.Println("Please use crudgen init or crudgen service")
		},
	}

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "init new project structure",
		Run: func(cmd *cobra.Command, args []string) {
			generators.InitNewProject(moduleName)
		},
	}

	var serviceName string
	var servicePath string

	generateServiceCmd := &cobra.Command{
		Use:   "service",
		Short: "generate new service",
		Run: func(cmd *cobra.Command, args []string) {
			if serviceName == "" {
				log.Fatal("you must set service name ex: --name posts")
				return
			}
			generators.GenerateService(moduleName, serviceName, servicePath)
		},
	}

	rootCmd.PersistentFlags().StringVar(&serviceName, "name", "", "service name ex: posts")
	rootCmd.PersistentFlags().StringVar(&servicePath, "path", "", "service path default lib/service-name ex: services/service-name")

	rootCmd.AddCommand(initCmd, generateServiceCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
