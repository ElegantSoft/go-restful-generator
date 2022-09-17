package writetemplate

import (
	"github.com/Masterminds/sprig"
	"html/template"
	"log"
	"os"
)

func ProcessTemplate(templatePath string, fileName string, outputPath string, data interface{}) {
	tmpl := template.Must(template.New("").Funcs(sprig.FuncMap()).Parse(templatePath))
	f, err := os.Create(outputPath)
	if err != nil {
		log.Println("create file: ", err)
		return
	}
	err = tmpl.Execute(f, data)
	if err != nil {
		log.Print("execute: ", err)
		return
	}
	log.Printf("create file: %s", fileName)
	//var processed bytes.Buffer
	//err := tmpl.ExecuteTemplate(&processed, fileName, data)
	//if err != nil {
	//	log.Fatalf("Unable to parse data into template: %v\n", err)
	//}
	//formatted, err := format.Source(processed.Bytes())
	//if err != nil {
	//	log.Fatalf("Could not format processed template: %v\n", err)
	//}
	//fmt.Println("Writing file: ", outputPath)
	//f, _ := os.Create(outputPath)
	//w := bufio.NewWriter(f)
	//w.WriteString(string(formatted))
	//w.Flush()
}
