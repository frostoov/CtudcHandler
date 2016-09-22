package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path"

	"github.com/frostoov/CtudcHandler/trek"
)

type Statistics struct {
	nevents  int
	fevents  int
	sevents  int
	wireHits [4]int
}

func (s *Statistics) String() string {
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "nevents         = %v\n", s.nevents)
	fmt.Fprintf(buf, "fevents         = %v\n", s.fevents)
	fmt.Fprintf(buf, "fevents/nevents = %v\n", float64(s.fevents)/float64(s.nevents))
	fmt.Fprintf(buf, "sevents         = %v\n", s.sevents)
	fmt.Fprintf(buf, "sevents/nevents = %v\n", float64(s.sevents)/float64(s.nevents))
	for i, w := range s.wireHits {
		fmt.Fprintf(buf, "WIRE %02d = %v  %v\n", i, w, float64(w)/float64(s.nevents))
	}
	return buf.String()
}

func (s *Statistics) Print(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.WriteString(s.String()); err != nil {
		return err
	}
	return nil
}

type IhepHandler struct {
	tracksFiles map[int]*os.File
	listFiles   map[int]*os.File
	stats       map[int]*Statistics
}

func ihepHandle(runs []int) error {
	h, err := NewIhepHandler()
	if err != nil {
		return err
	}
	defer h.Close()
	if err := h.Handle(runs); err != nil {
		return nil
	}
	for cham, stats := range h.stats {
		filename := path.Join("ihep_output/statistics", fmt.Sprintf("chamber_%02d.txt", cham))
		if err := stats.Print(filename); err != nil {
			return err
		}
	}
	return nil
}

func NewIhepHandler() (*IhepHandler, error) {
	os.RemoveAll("ihep_output")
	if err := os.MkdirAll("ihep_output/tracks", 0777); err != nil {
		return nil, fmt.Errorf("Failed create output dir: %s", err)
	}
	if err := os.MkdirAll("ihep_output/listing", 0777); err != nil {
		return nil, fmt.Errorf("Failed create output dir: %s", err)
	}
	if err := os.MkdirAll("ihep_output/statistics", 0777); err != nil {
		return nil, fmt.Errorf("Failed create output dir: %s", err)
	}
	return &IhepHandler{
		tracksFiles: make(map[int]*os.File),
		listFiles:   make(map[int]*os.File),
		stats:       make(map[int]*Statistics),
	}, nil
}

func (h *IhepHandler) Close() {
	for _, f := range h.tracksFiles {
		f.Close()
	}
	for _, f := range h.listFiles {
		f.Close()
	}
}

func (h *IhepHandler) Handle(runs []int) error {
	for _, run := range runs {
		root := formatRunDir(run)
		log.Println("Processing ", root)
		if err := h.handleRun(root); err != nil {
			log.Println("Failed:", err)
		} else {
			log.Println("Success")
		}
	}
	return nil
}

func fullDepth(ds *[4]int) bool {
	for _, d := range ds {
		if d == 0 {
			return false
		}
	}
	return true
}

func singleDepth(ds *[4]int) bool {
	for _, d := range ds {
		if d != 1 {
			return false
		}
	}
	return true
}

func (h *IhepHandler) handleRun(root string) error {
	reader, err := ctudcReader(root)
	if err != nil {
		return err
	}
	for r := range reader {
		ds := r.ChamberDepths()
		for cham, times := range r.Times() {
			// Listing
			w := h.listFiles[cham]
			if w == nil {
				filename := path.Join("ihep_output/listing", fmt.Sprintf("chamber_%02d.txt", cham))
				if f, err := os.Create(filename); err != nil {
					return err
				} else {
					w = f
					h.listFiles[cham] = w
				}
			}
			printIhepEventListing(w, times)

			if h.stats[cham] == nil {
				h.stats[cham] = &Statistics{}
			}
			stats := h.stats[cham]
			stats.nevents++
			for i, d := range ds[cham] {
				stats.wireHits[i] += d
			}
			if fullDepth(ds[cham]) {
				stats.fevents++
			}

			if singleDepth(ds[cham]) {
				stats.sevents++

				t1, t2, t3, t4 := times[0][0], times[0][1], times[0][2], times[0][3]
				k1 := t1 - t2 - t3 + t4
				k2 := t1 - 3*t2 + 3*t3 - t4

				w := h.tracksFiles[cham]
				if w == nil {
					filename := path.Join("ihep_output/tracks", fmt.Sprintf("chamber_%02d.txt", cham))
					if f, err := os.Create(filename); err != nil {
						return err
					} else {
						w = f
						h.tracksFiles[cham] = w
					}
					fmt.Fprintln(w, "WIRE_1\tWIRE_2\tWIRE_3\tWIRE_4\tk1\tk2")
				}
				fmt.Fprintf(w, "%d\t%d\t%d\t%d\t%f\t%f\n", t1, t2, t3, t4, k1, k2)

			}
		}
	}
	return nil
}

func printIhepEventListing(w io.Writer, times *trek.ChamTimes) {
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
	fmt.Fprintln(w, "Event============")
}
