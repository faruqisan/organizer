package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

type (
	moverError struct {
		err      error
		src, dst string
	}
)

var (
	targetFolderPath string
	mux              sync.Mutex
	listenerDelay    = time.Second * 2
)

func main() {

	inpTargetFolder := flag.String("path", "", "path to which folder will organized")
	flag.Parse()

	targetFolderPath = *inpTargetFolder

	if !isFolderExist(targetFolderPath) {
		log.Fatalf("path %s doesn't exist", targetFolderPath)
	}

	workerErrChan := make(chan moverError)
	newFileNotiferChan := make(chan bool)
	fileListenerErrChan := make(chan error)

	// worker listener
	go func() {
		for {
			select {
			case err := <-workerErrChan:
				log.Println("error from worker : ", err)
			case newFile := <-newFileNotiferChan:
				log.Println("new file found ! running worker")
				if newFile {
					organizeFiles(targetFolderPath, workerErrChan)
				}
			}
		}
	}()

	// run file listener
	newFileListener(newFileNotiferChan, fileListenerErrChan)
}

func getFilesFromFolder(folderPath string) ([]string, error) {
	var (
		fileNames []string
		err       error
	)

	files, err := ioutil.ReadDir(folderPath)
	if err != nil {
		return fileNames, err
	}

	for _, file := range files {
		if file.IsDir() {
			// TODO : SUB FOLDER
			// files, err := getFilesFromFolder(folderPath + "/" + file.Name())
			// if err != nil {
			// 	return fileNames, err
			// }

			// fileNames = append(fileNames, files...)
			continue
		}
		fileNames = append(fileNames, file.Name())
	}

	return fileNames, err
}

func organizeFiles(folderPath string, workerErrChan chan moverError) {
	files, err := getFilesFromFolder(targetFolderPath)
	if err != nil {
		workerErrChan <- moverError{
			err: err,
		}
		return
	}

	for _, file := range files {

		go func(fileName string) {
			// get file's extension
			splited := strings.Split(fileName, ".")
			// check for file that has no extension
			if len(splited) == 1 {
				log.Printf("file %s has no extension, is this a directory ?", fileName)
				return
			}

			fileExtension := splited[1]
			toCreateFolderPath := folderPath + "/" + fileExtension
			fileSRC := folderPath + "/" + fileName
			fileDST := toCreateFolderPath + "/" + fileName
			// check for folder named extension
			mux.Lock()
			if _, err := os.Stat(toCreateFolderPath); os.IsNotExist(err) {
				err = os.Mkdir(toCreateFolderPath, 0700)
				if err != nil {
					workerErrChan <- moverError{
						err: err,
						src: fileSRC,
						dst: fileDST,
					}
					return
				}
			}
			mux.Unlock()

			err = os.Rename(fileSRC, fileDST)
			if err != nil {
				workerErrChan <- moverError{
					err: err,
					src: fileSRC,
					dst: fileDST,
				}
				return
			}
		}(file)
	}
}

func newFileListener(newFileNotifier chan bool, listenerError chan error) {
	var (
		currentFiles []string
	)

	for {
		tmpFiles, err := getFilesFromFolder(targetFolderPath)
		if err != nil {
			listenerError <- err
			return
		}

		if equal(tmpFiles, currentFiles) {
			continue
		}

		newFileNotifier <- true

		time.Sleep(listenerDelay)
	}

}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func isFolderExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}
