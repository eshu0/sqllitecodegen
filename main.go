package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	pangu "github.com/eshu0/pangu/pkg"
)

func main() {

	dbname := flag.String("db", "", "Database defaults to searching the current working directoyr for .db files")
	outdir := flag.String("out", "", "output is ../Autogen/<Database>")
	tdir := flag.String("tdir", "", "Template directory is ./template/")

	flag.Parse()

	App := pangu.PanguApp{}
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
				App.Parse(path, outputdir, templatedir)
				return nil
			}
			fmt.Printf("visited file or dir: %q\n", path)
			return nil
		})
	} else {
		App.Parse(*dbname, outputdir, templatedir)
	}
}
