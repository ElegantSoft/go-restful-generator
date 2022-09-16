package writetemplate

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/Masterminds/sprig"
	"go/format"
	"log"
	"os"
	"text/template"
)

func ProcessTemplate(templatePath string, fileName string, outputPath string, data interface{}) {
	tmpl := template.Must(template.New("").Funcs(sprig.FuncMap()).ParseFiles(templatePath))
	var processed bytes.Buffer
	err := tmpl.ExecuteTemplate(&processed, fileName, data)
	if err != nil {
		log.Fatalf("Unable to parse data into template: %v\n", err)
	}
	formatted, err := format.Source(processed.Bytes())
	if err != nil {
		log.Fatalf("Could not format processed template: %v\n", err)
	}
	fmt.Println("Writing file: ", outputPath)
	f, _ := os.Create(outputPath)
	w := bufio.NewWriter(f)
	w.WriteString(string(formatted))
	w.Flush()
}
