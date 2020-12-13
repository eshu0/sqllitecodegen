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
	outdir := flag.String("out", "", "output is ./Autogen/<Database>")
	tdir := flag.String("tdir", "", "Template directory is ./template/")
	flag.Parse()

	slog := sl.NewApplicationLogger()

	// lets open a flie log using the session
	slog.OpenAllChannels()

	outputdir := "./Autogen/"
	templatedir := "./templates"

	if outdir != nil && *outdir != "" {
		outputdir = *outdir
	}

	if tdir != nil && *tdir != "" {
		templatedir = *tdir
	}

	if dbname == nil || (dbname != nil && *dbname == "") {
		filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
				return err
			}
			if !info.IsDir() && filepath.Ext(path) == ".db" {
				fmt.Printf("Parsing database: %+v \n", info.Name())
				Parse(path, outputdir, templatedir, slog)
				return nil
			}
			fmt.Printf("visited file or dir: %q\n", path)
			return nil
		})
	} else {
		Parse(*dbname, outputdir, templatedir, slog)
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
		slog.LogErrorEf("CreateAndExecute", "Cannot create file %s", err)
		return
	}

	fmt.Printf("Processing template: %s \n", filename)

	err = templ.Execute(file, data)
	if err != nil {
		slog.LogErrorEf("CreateAndExecute", "Execute Template failed: %s", err)
		fmt.Println("executing template: ", err)
	}

	file.Close()
}

func Parse(dbname string, odir string, tdir string, slog sli.ISimpleLogger) {

	if tdir == "" {
		tdir = "./templates/"
	}

	dbfolder := strings.Replace(filepath.Base(dbname), filepath.Ext(dbname), "", -1)

	outputdir := odir + strings.ToLower(dbfolder) //strings.Title(dbfolder)
	fmt.Println("Outputting to: " + outputdir)

	pkgdir := outputdir + "/pkg"
	fmt.Println("Package directory is: " + outputdir)

	// packages
	datastoredir := pkgdir + "/datastore/"
	handlerdir := pkgdir + "/handlers/"
	modelsdir := pkgdir + "/models/"
	pkgrestdir := pkgdir + "/REST/"
	controllersdir := pkgrestdir + "controllers/"

	// apps
	cmddir := outputdir + "/cmd/"
	appdir := cmddir + "testapp/"
	restdir := cmddir + "restserver/"

	CheckCreatePath(slog, dbname, true)
	CheckCreatePath(slog, odir, false)
	CheckCreatePath(slog, pkgdir, false)
	CheckCreatePath(slog, cmddir, false)

	CheckCreatePath(slog, pkgrestdir, false)
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

	CodeTemplate := CreateTemplate(tdir+"handlers.txt", "code")
	DataTemplate := CreateTemplate(tdir+"models.txt", "data")
	DLTemplate := CreateTemplate(tdir+"datastore.txt", "dl")
	ControllersTemplate := CreateTemplate(tdir+"controllers.txt", "control")

	MainTemplate := CreateTemplate(tdir+"apps/testapp.txt", "main")
	RESTServerTemplate := CreateTemplate(tdir+"apps/restserver.txt", "control")

	rpfolder := strings.Replace(filepath.Base(dbstruct.Database.Filename), filepath.Ext(dbstruct.Database.Filename), "", -1)
	reponame := strings.ToLower(rpfolder)
	fullreponame := "eshu0/" + reponame
	targetrepohost := "github.com"

	// Execute the template for each recipient.
	ctemplates := GenerateFile(dbstruct, slog, false, targetrepohost, fullreponame)

	for _, cs := range ctemplates {
		CreateAndExecute(slog, handlerdir+cs.Filename+".go", CodeTemplate, cs)
		CreateAndExecute(slog, controllersdir+cs.Filename+".go", ControllersTemplate, cs)
	}

	ctemplates = GenerateFile(dbstruct, slog, true, targetrepohost, fullreponame)

	for _, cs := range ctemplates {
		CreateAndExecute(slog, modelsdir+cs.Filename+".go", DataTemplate, cs)
	}

	dl := Datats{Database: ctemplates[0].Database, Templates: ctemplates, TargetRepoHost: targetrepohost, RepoName: fullreponame}

	CreateAndExecute(slog, strings.ToLower(datastoredir+dl.Database.FilenameTrimmed)+".go", DLTemplate, dl)
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
