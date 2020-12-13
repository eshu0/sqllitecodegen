package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	anl "github.com/eshu0/pangu/pkg/analysers"
	sl "github.com/eshu0/simplelogger/pkg"
	sli "github.com/eshu0/simplelogger/pkg/interfaces"
)

func main() {

	dbname := flag.String("db", "", "Database defaults to searching the current working directoyr for .db files")
	outdir := flag.String("out", "", "output is ../Autogen/<Database>")
	flag.Parse()

	slog := sl.NewApplicationLogger()

	// lets open a flie log using the session
	slog.OpenAllChannels()

	outputdir := "../Autogen/"

	if outdir != nil && *outdir != "" {
		outputdir = *outdir
	}

	if dbname == nil || (dbname != nil && *dbname == "") {
		filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
				return err
			}
			if !info.IsDir() && filepath.Ext(path) == ".db" {
				fmt.Printf("Parsing database: %+v \n", info.Name())
				Parse(path, outputdir, slog)
				return nil
			}
			fmt.Printf("visited file or dir: %q\n", path)
			return nil
		})
	} else {
		Parse(*dbname, outputdir, slog)
	}
}

func CheckCreatePath(slog sli.ISimpleLogger, path string, panicif bool) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if panicif {
			panic(path + " not found!")
		} else {
			os.Mkdir(path, 0777)
			fmt.Println("Created: " + path)
		}

	} else {
		fmt.Println("Exists: " + path)
	}
}

func CreateAndExecute(slog sli.ISimpleLogger, filename string, templ *template.Template, data interface{}) {
	file, err := os.Create(filename)
	if err != nil {
		slog.LogError("CreateAndExecute", fmt.Sprintf("Cannot create file%s", err.Error()))
		return
	}

	err = templ.Execute(file, data)
	if err != nil {
		fmt.Println("executing template:", err)
	}

	file.Close()
}

func Parse(dbname string, odir string, slog sli.ISimpleLogger) {

	dbfolder := strings.Replace(filepath.Base(dbname), filepath.Ext(dbname), "", -1)

	outputdir := odir + strings.Title(dbfolder)
	fmt.Println("Outputting to: " + outputdir)

	datastoredir := outputdir + "/Datastore/"
	handlerdir := outputdir + "/Handlers/"
	modelsdir := outputdir + "/Models/"
	appdir := outputdir + "/TestApp/"
	restdir := outputdir + "/REST/"
	controllersdir := restdir + "/Controllers/"

	CheckCreatePath(slog, dbname, true)
	CheckCreatePath(slog, odir, false)
	CheckCreatePath(slog, outputdir, false)
	CheckCreatePath(slog, datastoredir, false)
	CheckCreatePath(slog, handlerdir, false)
	CheckCreatePath(slog, modelsdir, false)
	CheckCreatePath(slog, appdir, false)
	CheckCreatePath(slog, controllersdir, false)
	CheckCreatePath(slog, restdir, false)

	fds := &anl.DatabaseAnalyser{}
	fds.Filename = dbname
	fds.Create(slog)

	dbstruct := fds.GetDatabaseStructure()

	CodeTemplate := CreateTemplate("./Templates/CodeTemplate.txt", "code")
	DataTemplate := CreateTemplate("./Templates/DataTemplate.txt", "data")
	DLTemplate := CreateTemplate("./Templates/DLTemplate.txt", "dl")
	MainTemplate := CreateTemplate("./Templates/MainTemplate.txt", "main")
	ControllersTemplate := CreateTemplate("./Templates/Controllers.txt", "control")
	RESTServerTemplate := CreateTemplate("./Templates/RESTServer.txt", "control")

	// Execute the template for each recipient.
	ctemplates := GenerateFile(dbstruct, slog)

	for _, cs := range ctemplates {
		CreateAndExecute(slog, handlerdir+cs.GetHandlersName()+".go", CodeTemplate, cs)
		CreateAndExecute(slog, controllersdir+cs.GetHandlersName()+".go", ControllersTemplate, cs)
		CreateAndExecute(slog, modelsdir+cs.GetDataName()+".go", DataTemplate, cs)
	}

	dl := Datats{Database: ctemplates[0].Database, Templates: ctemplates}

	CreateAndExecute(slog, datastoredir+dl.Database.FilenameTrimmed+".go", DLTemplate, dl)
	CreateAndExecute(slog, appdir+"main.go", MainTemplate, ctemplates)
	CreateAndExecute(slog, restdir+"main.go", RESTServerTemplate, ctemplates)

}

func CreateTemplate(filepath string, name string) *template.Template {
	b1, err1 := ioutil.ReadFile(filepath) // just pass the file name
	if err1 != nil {
		fmt.Print(err1)
		return nil
	}
	str1 := string(b1) // convert content to a 'string'

	// Create a new template and parse the letter into it.
	return template.Must(template.New(name).Parse(str1))
}
