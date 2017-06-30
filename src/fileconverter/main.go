package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/golang/glog"
)

// watchForNewInput watches the specified input path for new files.
// It uses the github.com/fsnotify/fsnotify package to tie into OS specific
// system calls to be alerted of new files moved into the directory.
// also handles signals
func watchForNewInput(inPath string, outPath string, errPath string, complPath string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		glog.Fatal(err)
	}
	defer watcher.Close()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan,
		syscall.SIGINT,
		syscall.SIGHUP,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	exitChan := make(chan int)
	go func() {
		for {
			s := <-signalChan
			switch s {
			// kill -SIGINT
			case syscall.SIGINT:
				glog.Info("exiting: SIGINT")
				exitChan <- 0
			// kill -SIGHUP
			case syscall.SIGHUP:
				glog.Info("exiting: SIGHUP")
				exitChan <- 0
			// kill -SIGTERM
			case syscall.SIGTERM:
				glog.Info("exiting: SIGTERM")
				exitChan <- 0

			// kill -SIGQUIT
			case syscall.SIGQUIT:
				glog.Info("exiting: SIGQUIT")
				exitChan <- 0

			default:
				glog.Error("Unknown signal.")
			}
		}
	}()

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Create == fsnotify.Create {
					glog.Infof("processing new file: %s", event.Name)
					processor, err := NewFileProcessor(event.Name, outPath, errPath, complPath)
					if err != nil {
						glog.Error(err)
						return
					}
					go processor.Process()
				}

			case err := <-watcher.Errors:
				glog.Fatal(err)
			}
		}
	}()

	// add the input path to the watcher
	err = watcher.Add(inPath)
	if err != nil {
		glog.Fatal(err)
	}
	code := <-exitChan
	glog.Flush()
	os.Exit(code)
}

// processExistingInput should process all files in the input
// directory when the application first starts
func processExistingInput(inPath string, outPath string, errPath string, complPath string) {
	files, err := ioutil.ReadDir(inPath)
	if err != nil {
		glog.Fatal(err)
	}

	for _, file := range files {
		if !strings.HasPrefix(file.Name(), ".") {
			inFile := filepath.Join(inPath, file.Name())
			glog.Infof("processing existing file: %s", inFile)
			processor, err := NewFileProcessor(inFile, outPath, errPath, complPath)
			if err != nil {
				glog.Error(err)
				continue
			}
			go processor.Process()
		}
	}
}

func main() {
	config, err := GetConfigFromFile("config.json")
	if err != nil {
		fmt.Printf("ERROR: cannot process config.json: %s", err)
		os.Exit(1)
	}

	flag.StringVar(&config.InputPath, "input", config.InputPath, "full path to input directory")
	flag.StringVar(&config.OutputPath, "output", config.OutputPath, "full path to output directory")
	flag.StringVar(&config.ErrorPath, "errors", config.ErrorPath, "full path to error directory")
	flag.StringVar(&config.CompletedPath, "completed", config.CompletedPath, "full path to completed directory")

	// NOTE: the glog package also sets some flags.
	flag.Parse()

	// required args
	if !config.IsValid() {
		glog.Error("4 options must be supplied in config.json or on command line: ")
		glog.Error("	see config.json or supply")
		glog.Fatal(" 	command line: -input -output -errors -completed")
		os.Exit(1)
	}

	glog.Info("using input directory: ", config.InputPath)
	processExistingInput(config.InputPath, config.OutputPath, config.ErrorPath, config.CompletedPath)
	// watch the input directory
	// blocks until process receives appropriate signal
	watchForNewInput(config.InputPath, config.OutputPath, config.ErrorPath, config.CompletedPath)
}
