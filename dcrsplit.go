package main

import (
	"bufio"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

func dcrSplitFile(filename, outname string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	currentRun := -1
	var currentFile *os.File
	defer func() {
		if currentFile != nil {
			currentFile.Close()
		}
	}()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		numbers := strings.Split(scanner.Text(), "\t")
		if run, err := strconv.ParseUint(numbers[0], 10, 32); err != nil {
			log.Println("Failed handle string:", scanner.Text(), ":", err)
		} else {
			if currentRun != int(run) {
				currentRun = int(run)
				if currentFile != nil {
					currentFile.Close()
				}
				if f, err := os.Create(path.Join(formatRunDir(int(run)), outname)); err != nil {
					return err
				} else {
					currentFile = f
				}
			}
			currentFile.WriteString(scanner.Text() + "\n")
		}
	}
	return nil
}

func dcrsplit(filenames []string, outname string) error {
	for _, f := range filenames {
		if err := dcrSplitFile(f, outname); err != nil {
			return err
		}
	}
	return nil
}
