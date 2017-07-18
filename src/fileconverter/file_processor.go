package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
)

const (
	COMMA             = ","
	ERROR_FILE_HEADER = "LINE_NUM,ERROR_MSG\n"
)

type FileProcessor interface {
	Process()
}

type csvFileProcessor struct {
	Errors         []ErrorRecord
	InputFilePath  string
	OutputFilePath string
	ErrorFilePath  string
	CompletedPath  string
	linesRead      int32
	linesWritten   int
}

type ErrorRecord interface {
	String() string
}

type CsvErrorRecord struct {
	LineNumber   int32
	ErrorMessage error
}

func (e *CsvErrorRecord) String() string {
	return fmt.Sprintf("%d,%s", e.LineNumber, e.ErrorMessage)
}

// NewFileProcessor creates a new FileProcessor
func NewFileProcessor(inputFilePath string, outputDirectory string, errorDirectory string, completedPath string) (FileProcessor, error) {
	fs, err := os.Stat(inputFilePath)
	if err != nil || fs.IsDir() {
		msg := fmt.Sprintf("not a valid input file: %s", inputFilePath)
		return &csvFileProcessor{}, errors.New(msg)
	}
	fs, err = os.Stat(outputDirectory)
	if err != nil || !fs.IsDir() {
		msg := fmt.Sprintf("not a valid output directory: %s", outputDirectory)
		return &csvFileProcessor{}, errors.New(msg)
	}
	fs, err = os.Stat(errorDirectory)
	if err != nil || !fs.IsDir() {
		msg := fmt.Sprintf("not a valid error directory: %s", errorDirectory)
		return &csvFileProcessor{}, errors.New(msg)
	}
	fs, err = os.Stat(completedPath)
	if err != nil || !fs.IsDir() {
		msg := fmt.Sprintf("not a valid error directory: %s", errorDirectory)
		return &csvFileProcessor{}, errors.New(msg)
	}

	fileName := filepath.Base(inputFilePath)
	out := filepath.Join(outputDirectory, fmt.Sprintf("%s.json", fileName))
	e := filepath.Join(errorDirectory, fileName)
	s := filepath.Join(completedPath, fileName)
	c := csvFileProcessor{InputFilePath: inputFilePath, OutputFilePath: out, ErrorFilePath: e, CompletedPath: s}

	return &c, nil
}

// isAlreadyProcessed returns true if output file already exists. In this case we assume that it was
// already processed.
func (c *csvFileProcessor) isAlreadyProcessed() bool {
	if _, err := os.Stat(c.OutputFilePath); os.IsNotExist(err) {
		return false
	}
	return true
}

// the Process method should be called in a separate go routine when a new file
// appears within the input directory.  This is the primary entry point into
// handling the input files and processing them according to specification.
func (c *csvFileProcessor) Process() {
	if c.isAlreadyProcessed() {
		glog.Errorf("file: %s is already processed.", c.InputFilePath)
		return
	}

	err := c.processFile()
	if err != nil {
		// file level errors indicate failure to process the file, so remove the output file
		glog.Error(err)
		os.Remove(c.OutputFilePath)
		// move input to failed directory?
		return
	} else {
		// move input to success directory
		os.Rename(c.InputFilePath, c.CompletedPath)
	}

	if err := c.processErrors(); err == nil {
		glog.Infof("successfully processed: %s: lines[%d] errors[%d].", c.InputFilePath, c.linesRead, len(c.Errors))
	} else {
		glog.Errorf("failed to produce error file. processed: %s: lines[%d] errors[%d].", c.InputFilePath, c.linesRead, len(c.Errors))
	}
}

// processFile processes its input file.  Returns an error for file level errors
// and these should be handled appropriately by the caller, ie. - deleting output
// when appropriate and logging error messages.
func (c *csvFileProcessor) processFile() error {
	infile, err := os.Open(c.InputFilePath)
	if err != nil {
		msg := fmt.Sprintf("cannot open input file: %s", c.InputFilePath)
		return errors.New(msg)
	}
	defer infile.Close()

	outfile, err := os.Create(c.OutputFilePath)
	if err != nil {
		msg := fmt.Sprintf("cannot open output file: %s", c.OutputFilePath)
		return errors.New(msg)
	}
	defer outfile.Close()

	scanner := bufio.NewScanner(infile)
	writer := bufio.NewWriter(outfile)
	jout := json.NewEncoder(writer)

	if scanner.Scan() {
		c.linesRead++
		fieldsFuncs, err := c.parseHeader(scanner.Text())
		if err != nil {
			return err
		}
		writer.WriteString("[\n")
		for scanner.Scan() {
			c.linesRead++
			p := c.processLine(fieldsFuncs, scanner.Text())
			if p != nil {
				if c.linesWritten > 0 {
					writer.WriteString(",\n")
				}
				jout.Encode(p)
				c.linesWritten++
			}
		}
	}
	if c.linesWritten > 0 {
		writer.WriteString("]\n")
		writer.Flush()
	}

	if err := scanner.Err(); err != nil {
		msg := fmt.Sprintf("an error occurred processing file: %s : %s", c.InputFilePath, err)
		return errors.New(msg)
	}
	return nil
}

// parseHeader parses the header line and returns a slice of columnValueFuncs
// in the order in which fields will appear on each line. the requirements do not guarantee
// the order of the columns, so we take the order from the header line.
func (c *csvFileProcessor) parseHeader(line string) (fieldsFuncs []columnValueFunc, err error) {

	fields := strings.Split(line, COMMA)
	for _, fname := range fields {
		f, ok := name2function[fname]
		if !ok {
			msg := fmt.Sprintf("file: %s contains an invalid header: %s", c.InputFilePath, fname)
			err = errors.New(msg)
			return
		} else {
			fieldsFuncs = append(fieldsFuncs, f)
		}
	}
	// we must have exactly the number of expected fields
	if len(fieldsFuncs) != len(name2function) {
		msg := fmt.Sprintf("file: %s is missing expected headers.", c.InputFilePath)
		err = errors.New(msg)
	}
	return
}

// return a *PersonRecord for each line passed in.
// if there is an error processing the input line, returns nil and adds an ErrorRecord to c.Errors
func (c *csvFileProcessor) processLine(fieldsFuncs []columnValueFunc, line string) *PersonRecord {
	p := PersonRecord{}
	fields := strings.Split(line, COMMA)
	for i, _ := range fields {
		f := fieldsFuncs[i]
		// call the appropriate function for each field in the delimited text
		err := f(&p, fields[i])
		if err != nil {
			c.Errors = append(c.Errors, &CsvErrorRecord{LineNumber: c.linesRead, ErrorMessage: err})
			return nil
		}
	}
	return &p
}

// if there are any line processing errors, then this method writes them
// in the prescribed format to c.ErrorFilePath/input_filename
func (c *csvFileProcessor) processErrors() error {
	if len(c.Errors) > 0 {
		errFile, err := os.Create(c.ErrorFilePath)
		if err != nil {
			return err
		}
		defer errFile.Close()

		writer := bufio.NewWriter(errFile)
		writer.WriteString(ERROR_FILE_HEADER)
		for i, _ := range c.Errors {
			writer.WriteString(c.Errors[i].String())
			writer.WriteString("\n")
		}
		writer.Flush()
	}
	return nil
}
