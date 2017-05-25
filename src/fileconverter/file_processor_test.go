package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

var (
	tmpDir string
)

func setUpFakeFiles() *csvFileProcessor {
	r := strconv.FormatInt(int64(rand.Intn(10000)), 10)
	tmpDir = fmt.Sprintf("/tmp/scoir_fproc_%s", r)

	tmpIn := filepath.Join(tmpDir, "in")
	tmpOut := filepath.Join(tmpDir, "out")
	tmpErr := filepath.Join(tmpDir, "err")
	tmpCompl := filepath.Join(tmpDir, "compl")

	os.MkdirAll(tmpIn, 0777)
	os.MkdirAll(tmpOut, 0777)
	os.MkdirAll(tmpErr, 0777)
	os.MkdirAll(tmpCompl, 0777)

	tmpInFile := filepath.Join(tmpIn, "in")
	os.Create(tmpInFile)
	c := csvFileProcessor{InputFilePath: tmpInFile, OutputFilePath: tmpOut, ErrorFilePath: tmpErr, CompletedPath: tmpCompl}

	return &c
}

func TestProcessLineSuccess(t *testing.T) {
	fp := setUpFakeFiles()
	defer os.RemoveAll(tmpDir)

	inputLine := "12345678,Bobby,,Tables,555-555-5555"
	fieldsFuncs, err := fp.parseHeader("INTERNAL_ID,FIRST_NAME,MIDDLE_NAME,LAST_NAME,PHONE_NUM")
	if err != nil {
		t.Errorf("%s: unexpected error: %s", t.Name(), err)
		t.Fail()
	}

	p := fp.processLine(fieldsFuncs, inputLine)
	if p.ID != int(12345678) {
		t.Errorf("%s: unexpected person.ID value in output of processLine(): %+v", t.Name(), p)
		t.Fail()
	}
	if p.Name.First != "Bobby" {
		t.Errorf("%s: unexpected person.Name.First value in output of processLine(): %+v", t.Name(), p)
		t.Fail()
	}
	if p.Name.Last != "Tables" {
		t.Errorf("%s: unexpected person.Name.Last value in output of processLine(): %+v", t.Name(), p)
		t.Fail()
	}
	if p.Phone != "555-555-5555" {
		t.Errorf("%s: unexpected person.Phone value in output of processLine(): %+v", t.Name(), p)
		t.Fail()
	}
}
func TestProcessLineSuccessNonStandardOrder(t *testing.T) {
	fp := setUpFakeFiles()
	defer os.RemoveAll(tmpDir)

	inputLine := "555-555-5555,12345678,Bobby,,Tables"
	fieldsFuncs, err := fp.parseHeader("PHONE_NUM,INTERNAL_ID,FIRST_NAME,MIDDLE_NAME,LAST_NAME")
	if err != nil {
		t.Errorf("%s: unexpected error: %s", t.Name(), err)
		t.Fail()
	}

	p := fp.processLine(fieldsFuncs, inputLine)
	if p.ID != int(12345678) {
		t.Errorf("%s: unexpected person.ID value in output of processLine(): %+v", t.Name(), p)
		t.Fail()
	}
	if p.Name.First != "Bobby" {
		t.Errorf("%s: unexpected person.Name.First value in output of processLine(): %+v", t.Name(), p)
		t.Fail()
	}
	if p.Name.Last != "Tables" {
		t.Errorf("%s: unexpected person.Name.Last value in output of processLine(): %+v", t.Name(), p)
		t.Fail()
	}
	if p.Phone != "555-555-5555" {
		t.Errorf("%s: unexpected person.Phone value in output of processLine(): %+v", t.Name(), p)
		t.Fail()
	}
}
