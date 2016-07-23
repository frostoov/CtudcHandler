package main

import (
	"bufio"
	"io/ioutil"
	"log"
	"os"
	"path"

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

func split(dirnames []string) error {
	runWriters := map[int]*RunData{}

	defer func() {
		for _, writer := range runWriters {
			writer.writer.Flush()
			writer.file.Close()
		}
	}()

	splitFile := func(filename string) error {
		f, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer f.Close()
		r := bufio.NewReader(f)
		if str, err := r.ReadString('\n'); err != nil {
			return err
		} else if str == "TDSdrop\n" {
			log.Printf("Skipping drop")
			return nil
		}

		var record trek.Event
		for record.Unmarshal(r) == nil {
			run := int(record.Nrun())
			runWriter := runWriters[run]
			if runWriter != nil && runWriter.lastRecord >= record.Nevent() {
				log.Printf("split previousRecord(%v) >= currentRecord(%v) run #%v\n",
					runWriter.lastRecord, record.Nevent(), run)
			} else {
				if runWriter == nil {
					rundir := formatRunDir(run)
					if err := os.MkdirAll(path.Join(rundir, "ctudc"), 0777); err != nil {
						return err
					}
					filename := formatCtudcFilename(run, 0)

					f, err := os.Create(filename)
					log.Println("Created: ", rundir)
					if err != nil {
						return err
					}
					w := bufio.NewWriter(f)
					w.WriteString("TDSa\n")
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
					runWriter.file = f
					w := bufio.NewWriter(f)
					w.WriteString("TDSa\n")
					runWriter.writer = w
				}
				record.Marshal(runWriter.writer)
				runWriter.eventCount++
			}
			runWriter.lastRecord = record.Nevent()
		}
		return nil
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

	return nil
}
