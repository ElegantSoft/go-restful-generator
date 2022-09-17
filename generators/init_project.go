package generators

import (
	_ "embed"
	"github.com/ElegantSoft/go-crud-starter/pkg/writetemplate"
	"os"
	"path/filepath"
)

//go:embed templates/main.tmpl
var mainTemplate string

//go:embed templates/db/database.tmpl
var databaseTemplate string

//go:embed templates/db/migrations.tmpl
var migrationsTemplate string

//go:embed templates/env.tmpl
var envTemplate string

//go:embed templates/gitignore.tmpl
var gitignoreTemplate string

func InitNewProject(packageName string) {

	type Data struct {
		PackageName string
	}
	data := Data{PackageName: packageName}

	writetemplate.ProcessTemplate(mainTemplate, "main.tmpl", filepath.Join("main.go"), data)
	err := os.Mkdir("db", os.ModePerm)
	if err != nil {
		panic(err)
	}
	err = os.Mkdir("lib", os.ModePerm)
	if err != nil {
		panic(err)
	}
	writetemplate.ProcessTemplate(databaseTemplate, "database.tmpl", filepath.Join("db/database.go"), data)
	writetemplate.ProcessTemplate(migrationsTemplate, "migrations.tmpl", filepath.Join("db/migrations.go"), data)
	writetemplate.ProcessTemplate(envTemplate, "env.tmpl", filepath.Join(".env"), data)
	writetemplate.ProcessTemplate(gitignoreTemplate, "gitignore.tmpl", filepath.Join(".gitignore"), data)

}
