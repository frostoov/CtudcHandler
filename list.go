package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path"

	"github.com/frostoov/CtudcHandler/trek"
)

func list(runList []int) error {
	outdir := "./ctudc_listing"
	if err := os.MkdirAll(outdir, 0777); err != nil {
		return err
	}
	for _, run := range runList {
		root := formatRunDir(run)
		log.Printf("Processing run #%v\n", run)
		listRun(path.Join(root, "ctudc"), outdir)
	}
	return nil
}

func listRun(dirname, outdir string) error {
	files := map[string]*os.File{}
	writers := map[string]*bufio.Writer{}
	defer func() {
		for _, w := range writers {
			w.Flush()
		}
		for _, f := range files {
			f.Close()
		}
	}()
	r, err := ctudcReader(dirname)
	if err != nil {
		return err
	}

	for event := range r {
		for cham, times := range event.Times() {
			filename := path.Join(outdir, fmt.Sprintf("chamber_%02d.txt", cham+1))
			w := writers[filename]
			if w == nil {
				if f, err := os.Create(filename); err != nil {
					return err
				} else {
					w = bufio.NewWriter(f)
					//fmt.Fprintln(w, "#WIRE_1 WIRE_2 WIRE_3 WIRE_4")
					files[filename] = f
					writers[filename] = w
				}
			}
			printEventListing(w, times)
		}
	}
	return nil
}

func printEventListing(w io.Writer, times *trek.ChamTimes) {
	depth := 0
	for _, t := range times {
		if len(t) > depth {
			depth = len(t)
		}
	}

	for i := 0; i < depth; i++ {
		for j, t := range *times {
			if j > 0 {
				fmt.Fprint(w, " ")
			}
			if i < len(t) {
				fmt.Fprint(w, t[i])
			} else {
				fmt.Fprint(w, "-")
			}
		}
		fmt.Fprintln(w)
	}
}
