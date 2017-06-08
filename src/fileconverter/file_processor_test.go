package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var (
	tmpDir string
)

func setUpFakeFiles(buf *bytes.Buffer) *csvFileProcessor {
	seed := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(seed)
	tmpDir = fmt.Sprintf("/tmp/scoir_fproc_%d_%d", os.Getpid(), r1.Intn(10000))

	// make all temp paths
	tmpIn := filepath.Join(tmpDir, "in")
	tmpOut := filepath.Join(tmpDir, "out")
	tmpErr := filepath.Join(tmpDir, "err")
	tmpCompl := filepath.Join(tmpDir, "compl")

	// make all temp directories on disk
	os.MkdirAll(tmpIn, 0777)
	os.MkdirAll(tmpOut, 0777)
	os.MkdirAll(tmpErr, 0777)
	os.MkdirAll(tmpCompl, 0777)

	// create input file
	tmpInFile := filepath.Join(tmpIn, "in.csv")
	iFile, _ := os.Create(tmpInFile)

	// write contents of buf to iFile and close file descriptor
	writer := bufio.NewWriter(iFile)
	writer.WriteString(buf.String())
	writer.Flush()

	iFile.Close()

	c, _ := NewFileProcessor(tmpInFile, tmpOut, tmpErr, tmpCompl)

	return c.(*csvFileProcessor)
}

func TestProcessInputFileMixedOrder(t *testing.T) {
	var buf bytes.Buffer

	buf.WriteString("PHONE_NUM,INTERNAL_ID,FIRST_NAME,MIDDLE_NAME,LAST_NAME\n")
	buf.WriteString("555-555-5555,12345678,Bobby,,Tables\n")

	fp := setUpFakeFiles(&buf)
	defer os.RemoveAll(tmpDir)
	fp.Process()

	outputFile, err := os.Open(fp.OutputFilePath)
	if err != nil {
		t.Errorf("%s: cannot open expected output file %s", t.Name(), fp.OutputFilePath)
		t.FailNow()
	}

	h := sha256.New()
	if _, err := io.Copy(h, outputFile); err != nil {
		t.Error("%s: cannot take sha256 hash of input file %s", t.Name(), fp.OutputFilePath)
		t.FailNow()
	}
	expected := "4b9f660892a53168a7e898eb24d063d1f0bdf844515a9f058d2bd373e19f7b0f"
	actual := fmt.Sprintf("%x", h.Sum(nil))
	if actual != expected {
		t.Errorf("%s: expected: %s\nactual:   %x\n", t.Name(), expected, h.Sum(nil))
		t.Fail()
	}
}

func TestProcessInputFileSuccess(t *testing.T) {
	var buf bytes.Buffer

	buf.WriteString("INTERNAL_ID,FIRST_NAME,MIDDLE_NAME,LAST_NAME,PHONE_NUM\n")
	buf.WriteString("12345678,Bobby,,Tables,555-555-5555\n")

	fp := setUpFakeFiles(&buf)
	defer os.RemoveAll(tmpDir)
	fp.Process()

	outputFile, err := os.Open(fp.OutputFilePath)
	if err != nil {
		t.Errorf("%s: cannot open expected output file %s", t.Name(), fp.OutputFilePath)
		t.FailNow()
	}

	h := sha256.New()
	if _, err := io.Copy(h, outputFile); err != nil {
		t.Error("%s: cannot take sha256 hash of input file %s", t.Name(), fp.OutputFilePath)
		t.FailNow()
	}
	expected := "4b9f660892a53168a7e898eb24d063d1f0bdf844515a9f058d2bd373e19f7b0f"
	actual := fmt.Sprintf("%x", h.Sum(nil))
	if actual != expected {
		t.Errorf("%s: expected: %s\nactual:   %x\n", t.Name(), expected, h.Sum(nil))
		t.Fail()
	}
}

func TestProcessLineSuccess(t *testing.T) {
	fp := &csvFileProcessor{}

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
	fp := &csvFileProcessor{}

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
