package main

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
	"path"
)

func dcrSplitFile(filename string) error {
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
				if f, err := os.Create(path.Join(formatRunDir(int(run)), "decor.dat.bkp")); err != nil {
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

func dcrsplit(filenames []string) error {
	for _, f := range filenames {
		if err := dcrSplitFile(f); err != nil {
			return err
		}
	}
	return nil
}
