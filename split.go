package main

import (
	"bufio"
	"github.com/frostoov/CtudcHandler/trek"
	"io/ioutil"
	"log"
	"os"
	"path"
)

type RunData struct {
	file       *os.File
	writer     *bufio.Writer
	eventCount int
	fileCount  int
	lastRecord int
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
			if runWriter == nil {
				rundir := formatRunDir(run)
				if err := os.MkdirAll(rundir, 0777); err != nil {
					return err
				}
				log.Println("Created: ", formatRunDir(run))
				filename := formatFileName(run, 0)
				f, err := os.Open(filename)
				if err != nil {
					return err
				}
				w := bufio.NewWriter(f)
				w.WriteString("TDSa\n")
				runWriter = &RunData{
					file:       f,
					writer:     w,
					eventCount: 0,
					fileCount:  0,
				}
				runWriters[run] = runWriter
			} else if runWriter.eventCount > 10000 {
				runWriter.fileCount++
				runWriter.eventCount = 0
				filename := formatFileName(run, runWriter.fileCount)
				f, err := os.Open(filename)
				if err != nil {
					return err
				}
				runWriter.file = f
				runWriter.writer = bufio.NewWriter(f)
			}
			record.Marshal(runWriter.writer)
			if runWriter.lastRecord >= int(record.Nevent()) {
				panic("split runWriter.lastRecord >= record.Nevent()")
			}
			runWriter.lastRecord = int(record.Nevent())
			runWriter.eventCount++
		}
		return nil
	}

	for _, dirname := range dirnames {
		files, err := ioutil.ReadDir(dirname)
		if err != nil {
			return err
		}
		for _, filestat := range files {
			if path.Ext(filestat.Name()) != ".tds" {
				continue
			}
			splitFile(path.Join(dirname, filestat.Name()))
		}
	}

	return nil
}
