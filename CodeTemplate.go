package main

import (
	"fmt"
	"strings"

	anl "github.com/eshu0/pangu/pkg/analysers"
	pangudata "github.com/eshu0/pangu/pkg/structures"
	sli "github.com/eshu0/simplelogger/pkg/interfaces"
)

// solution to having data changes
type Datats struct {
	Templates []*CodeTemplate
	Database  *anl.Database
}

// Confsing name -> should rename
type CodeTemplate struct {
	PackageName           string
	StorageHandlerName    string
	StorageControllerName string
	TableConstant         *pangudata.Constant
	IdConstant            *pangudata.Constant
	Constants             []*pangudata.Constant
	Table                 *anl.Table
	StructDetails         *pangudata.StructDetails
	InsertDBColumns       string
	UpdateDBColumns       string
	InsertGo              string
	UpdateGo              string
	SelectDBColumns       string
	ParametersColumns     string
	CreateTableSQL        string
	ScanRow               string
	Database              *anl.Database
}

func (cs *CodeTemplate) GetHandlersName() string {
	return strings.Title(cs.Table.Name)
}

func (cs *CodeTemplate) GetDataName() string {
	name := strings.Title(cs.Table.Name)
	if last := len(name) - 1; last >= 0 && name[last] == 's' {
		name = name[:last]
	}

	return name
}

func GenerateFile(dbstruct *anl.DatabaseStructure, slog sli.ISimpleLogger) []*CodeTemplate {

	var temps []*CodeTemplate

	for _, tbl := range dbstruct.Tables {

		cs := CodeTemplate{PackageName: "pguhandlers", Table: tbl, StorageHandlerName: strings.Title(tbl.Name + "Handler"), StorageControllerName: strings.Title(tbl.Name + "Controller"), Database: dbstruct.Database}
		cs.StructDetails = tbl.CreateStructDetails()
		consts, idconst := tbl.CreateConstants()

		cs.Constants = consts
		cs.IdConstant = idconst
		cs.CreateTableSQL = strings.Replace(tbl.Sql, "CREATE TABLE", "CREATE TABLE IF NOT EXISTS", -1)

		cnst := &pangudata.Constant{}
		cnst.Comment = fmt.Sprintf("%s", tbl.Name)
		cnst.Name = strings.ToLower(tbl.Name) + strings.Title("TName")
		cnst.Value = tbl.TableName
		cs.TableConstant = cnst

		insertdbcolumns := ""
		updatedbcolumns := ""

		goselect := ""
		goinsert := ""
		goupdate := ""

		parameterscolumns := ""
		for i := 0; i < len(cs.StructDetails.Properties); i++ {
			if i == 0 {
				goselect = fmt.Sprintf("&%s", cs.StructDetails.Properties[i].Name)
			} else {
				goselect = fmt.Sprintf("%s,&%s", goselect, cs.StructDetails.Properties[i].Name)
			}
		}

		startedadd := false
		for i := 0; i < len(cs.StructDetails.Properties); i++ {
			if i == 0 || !startedadd {
				if !cs.StructDetails.Properties[i].IsIdentifier {
					goinsert = fmt.Sprintf("data.%s", cs.StructDetails.Properties[i].Name)
					startedadd = true
				}
			} else {
				if !cs.StructDetails.Properties[i].IsIdentifier {
					goinsert = fmt.Sprintf("%s,data.%s", goinsert, cs.StructDetails.Properties[i].Name)
				}
			}
		}

		// update is different as we want to add the indentifier at the end
		goupdate = goinsert
		for i := 0; i < len(cs.StructDetails.Properties); i++ {
			if cs.StructDetails.Properties[i].IsIdentifier {
				goupdate = fmt.Sprintf("%s,data.%s", goupdate, cs.StructDetails.Properties[i].Name)
			}
		}

		for j := 0; j < len(cs.Constants); j++ {
			if j == 0 {
				insertdbcolumns = fmt.Sprintf("%s", "+ \"[\"+"+cs.Constants[j].Name+"+\"]\" + ")
				updatedbcolumns = fmt.Sprintf("%s", "+ \"[\"+"+cs.Constants[j].Name+"+\"] = ? \" + ")

				parameterscolumns = "?"

			} else {
				insertdbcolumns = insertdbcolumns + fmt.Sprintf("%s", " \",[\"+"+cs.Constants[j].Name+"+\"]\" +")
				updatedbcolumns = updatedbcolumns + fmt.Sprintf("%s", " \",[\"+"+cs.Constants[j].Name+"+\"] = ? \" +")
				parameterscolumns = parameterscolumns + ",?"
			}
		}

		cs.InsertDBColumns = insertdbcolumns
		cs.UpdateDBColumns = updatedbcolumns
		cs.InsertGo = goinsert
		cs.UpdateGo = goupdate

		cs.SelectDBColumns = fmt.Sprintf("%s", "+ \"[\"+"+cs.IdConstant.Name+"+\"],\" ") + insertdbcolumns
		cs.ParametersColumns = parameterscolumns
		cs.ScanRow = goselect

		temps = append(temps, &cs)

	}

	return temps
}
