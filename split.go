package main

import (
	"bufio"
	"io/ioutil"
	"log"
	"os"
	path "path/filepath"

	"github.com/frostoov/CtudcHandler/trek"
)

type RunData struct {
	file       *os.File
	writer     *bufio.Writer
	eventCount int
	fileCount  int
	lastRecord uint
	badRecord  bool
}

func pathExists(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

func (r *RunData) Close() error {
	if err := r.writer.Flush(); err != nil {
		return err
	}
	if err := r.file.Close(); err != nil {
		return err
	}
	return nil
}

func split(patterns []string) error {
	header := "TDSa\n"
	runWriters := map[int]*RunData{}

	defer func() {
		for _, writer := range runWriters {
			if err := writer.Close(); err != nil {
				log.Printf("Warning failed close event writer %s\n", err)
			}
		}
	}()

	splitFile := func(filename string) error {
		f, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer f.Close()
		s, err := trek.NewScanner(f)
		if err != nil {
			return err
		} else if s.Header() == "TDSdrop" {
			log.Printf("Skipping drop\n")
			return nil
		}

		for s.Scan() {
			record := s.Record()
			run := int(record.Nrun())
			runWriter := runWriters[run]
			if runWriter != nil && runWriter.lastRecord >= record.Nevent() {
				log.Printf("split previousRecord(%v) >= currentRecord(%v) run #%v\n",
					runWriter.lastRecord, record.Nevent(), run)
				if err := runWriter.Close(); err != nil {
					return err
				}
				delete(runWriters, run)
				runWriter = nil
			}
			if runWriter == nil {
				ctudcdir := formatCtudcSubdir(run)
				if pathExists(ctudcdir) {
					if err := os.RemoveAll(ctudcdir); err != nil {
						return err
					}
				}
				if err := os.MkdirAll(ctudcdir, 0777); err != nil {
					return err
				}
				filename := formatCtudcFilename(run, 0)

				f, err := os.Create(filename)
				if err != nil {
					return err
				}
				log.Println("Created: ", filename)
				w := bufio.NewWriter(f)
				w.WriteString(header)
				runWriter = &RunData{
					file:   f,
					writer: w,
				}
				runWriters[run] = runWriter
			} else if runWriter.eventCount > 10000 {
				runWriter.fileCount++
				runWriter.eventCount = 0
				filename := formatCtudcFilename(run, runWriter.fileCount)
				f, err := os.Create(filename)
				if err != nil {
					return err
				}
				log.Println("Created: ", filename)
				w := bufio.NewWriter(f)
				w.WriteString(header)
				runWriter.file = f
				runWriter.writer = w
				runWriters[run] = runWriter
			}
			if err := record.Marshal(runWriter.writer); err != nil {
				return err
			}
			runWriter.eventCount++
			runWriter.lastRecord = record.Nevent()
		}
		return nil
	}

	for _, pattern := range patterns {
		dirnames, err := path.Glob(pattern)
		if err != nil {
			log.Printf("Failed handle pattern %s %s\n", pattern, err)
			continue
		}
		for _, dirname := range dirnames {
			log.Println("Processing: ", dirname)
			files, err := ioutil.ReadDir(dirname)
			if err != nil {
				return err
			}
			for _, filestat := range files {
				if path.Ext(filestat.Name()) != ".tds" {
					continue
				}
				if err := splitFile(path.Join(dirname, filestat.Name())); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
