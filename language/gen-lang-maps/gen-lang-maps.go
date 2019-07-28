package main

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"path"

	"github.com/ungerik/go-fs/fsimpl"
)

// https://iso639-3.sil.org/code_tables/download_tables
// http://www.loc.gov/standards/iso639-2/ISO-639-2_utf-8.txt

func main() {
	iso6393URL := "https://iso639-3.sil.org/sites/iso639-3/files/downloads/iso-639-3_Code_Tables_20190408.zip"

	log.Println("Downloading", iso6393URL)
	response, err := http.Get(iso6393URL)
	if err != nil {
		log.Fatal(err)
	}

	buf, err := fsimpl.NewReadonlyFileBufferReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	zipReader, err := zip.NewReader(buf, buf.Size())
	if err != nil {
		log.Fatal(err)
	}

	var (
		nameIndexFile *zip.File
	)
	for _, f := range zipReader.File {
		switch path.Base(f.Name) {
		case "iso-639-3_Name_Index_20190408.tab":
			nameIndexFile = f
		}
	}

	r, err := nameIndexFile.Open()
	if err != nil {
		log.Fatal(err)
	}
	csvReader := csv.NewReader(r)
	csvReader.Comma = '\t'
	rows, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	idWritten := make(map[string]bool)
	goFileBuf := bytes.NewBufferString(iso6393namesFileHeader)
	for _, row := range rows[1:] {
		id, name := row[0], row[1]
		if idWritten[id] {
			fmt.Fprint(goFileBuf, "//")
		}
		fmt.Fprintf(goFileBuf, "\t%q: %q,\n", id, name)
		idWritten[id] = true
	}
	fmt.Fprintln(goFileBuf, "}")

	goFilePath := "../iso6393names.go"
	err = ioutil.WriteFile(goFilePath, goFileBuf.Bytes(), 0664)
	if err != nil {
		log.Fatal(err)
	}

	err = exec.Command("gofmt", "-w", goFilePath).Run()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Written", goFilePath)
}

const iso6393namesFileHeader = `// GENERATED
package language

var iso6393Names = map[Code]string{
`
