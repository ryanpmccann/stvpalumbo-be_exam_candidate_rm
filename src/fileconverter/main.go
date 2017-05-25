package main

import (
	"flag"
	"os"
	"os/signal"
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
func processExistingInput(inputPath string) {
	// I ran out of time.
	//

	// pseudocode:
	// for file in inputPath/*
	// 		processor := NewFileProcessor(...)
	//		go processor.Process()
}

func main() {
	var inputPath string
	var outputPath string
	var errorPath string
	var completedPath string
	flag.StringVar(&inputPath, "input", "", "full path to input directory")
	flag.StringVar(&outputPath, "output", "", "full path to output directory")
	flag.StringVar(&errorPath, "errors", "", "full path to error directory")
	flag.StringVar(&completedPath, "completed", "", "full path to completed directory")

	// NOTE: the glog package also sets some flags.
	flag.Parse()

	// required args
	if inputPath == "" || outputPath == "" || errorPath == "" || completedPath == "" {
		glog.Fatal("4 options must be supplied: -input -output -errors -completed")
		os.Exit(1)
	}

	glog.Info("using input directory: ", inputPath)
	processExistingInput(inputPath)
	// watch the input directory
	// blocks until process receives appropriate signal
	watchForNewInput(inputPath, outputPath, errorPath, completedPath)
}
