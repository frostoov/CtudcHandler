package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/frostoov/CtudcHandler/math"
	"github.com/frostoov/CtudcHandler/nevod"
	"github.com/frostoov/CtudcHandler/trek"
)

func readDecorTracks(filname string) (map[uint][]math.Line3, error) {
	f, err := os.Open(filname)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	if !s.Scan() {
		return nil, errors.New("readDecorTracks no data")
	}

	events := make(map[uint][]math.Line3, 10000)
	var pts [6]float64
	for s.Scan() {
		nums := strings.Split(s.Text(), "\t")
		eventNumber, err := strconv.ParseUint(nums[1], 10, 64)
		if err != nil {
			return nil, errors.New("readDecorTracks " + err.Error())
		}
		for i := 3; i < 9; i++ {
			pts[i-3], err = strconv.ParseFloat(nums[i], 64)
			if err != nil {
				return nil, errors.New("readDecorTracks " + err.Error())
			}
		}
		point := math.Vec3{X: pts[0], Y: pts[1], Z: pts[2]}
		vector := math.Vec3{X: pts[3], Y: pts[4], Z: pts[5]}
		track := math.Line3{Point: point, Vector: vector}
		events[uint(eventNumber)] = append(events[uint(eventNumber)], track)
	}
	return events, nil
}

func ctudcReader(dirname string) (<-chan trek.Event, error) {
	fileList, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, err
	}
	c := make(chan trek.Event, 100)
	go func() {
		for _, fileStat := range fileList {
			if filepath.Ext(fileStat.Name()) != ".tds" {
				continue
			}
			f, err := os.Open(dirname + "/" + fileStat.Name())
			if err != nil {
				continue
			}
			s, err := trek.NewScanner(f)
			for s.Scan() {
				c <- s.Record().Copy()
			}
			f.Close()
		}
		close(c)
	}()
	return c, nil
}

func nevodReader(dirname string) (<-chan nevod.EventMeta, error) {
	fileList, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, err
	}
	c := make(chan nevod.EventMeta, 100)
	go func() {
		for _, fileStat := range fileList {
			if filepath.Ext(fileStat.Name()) != ".nad" {
				continue
			}
			f, err := os.Open(dirname + "/" + fileStat.Name())
			if err != nil {
				continue
			}
			s := nevod.NewScanner(f)
			for s.Scan() {
				c <- s.Record().Meta
			}
			f.Close()
		}
		close(c)
	}()
	return c, nil
}

func mergeRun(root string) error {
	ctudc := root + "/ctudc"
	nevod := root + "/nevod"
	decor := root + "/decor.dat.bkp"
	decorShSh := root + "/decor.dat"
	extData := root + "/extctudc.tds"

	decorTracks, err := readDecorTracks(decor)
	if err != nil {
		return fmt.Errorf("Failed read decor tracks: %s", err)
	}
	decorTracksShSh, err := readDecorTracks(decorShSh)
	if err != nil {
		return fmt.Errorf("Failed read ShSh decor tracks: %s", err)
	}
	ctudcStream, err := ctudcReader(ctudc)
	if err != nil {
		return fmt.Errorf("Failed open ctudc data: %s", err)
	}
	nevodStream, err := nevodReader(nevod)
	if err != nil {
		return fmt.Errorf("Failed open nevod data: %s", err)
	}
	output, err := os.Create(extData)
	if err != nil {
		return fmt.Errorf("Failed create output file: %s", err)
	}
	defer output.Close()
	w := bufio.NewWriter(output)
	w.WriteString("TDSext\n")
	for ctudcEvent := range ctudcStream {
		nrun, nevent := ctudcEvent.Nrun(), ctudcEvent.Nevent()
		var decor []trek.DecorTrack

		if allTracks, ok := decorTracks[nevent]; ok {
			shTracks := decorTracksShSh[nevent]
			for i := range allTracks {
				var trackType int8
				for j := range shTracks {
					if shTracks[j] == allTracks[i] {
						trackType = int8(1)
						break
					}
				}
				decor = append(decor, trek.DecorTrack{
					Type:  trackType,
					Track: allTracks[i],
				})
			}
		}

		for nevodMeta := range nevodStream {
			if nrun != uint(nevodMeta.Nrun) {
				panic("OOps!! invalid run number!! data is corrupted")
			}
			if nevent == uint(nevodMeta.Nevent) {
				extEvent := trek.ExtEvent{
					Ctudc: ctudcEvent,
					Nevod: nevodMeta,
					Decor: decor,
				}
				extEvent.Marshal(w)
				break
			} else if uint(nevodMeta.Nevent) > nevent {
				break
			}
		}
	}
	return nil
}

func merge(runs []int) error {
	for _, run := range runs {
		root := appConf.CtudcRoot + fmt.Sprintf("/run_%03d", run)
		log.Print("Processing ", root)
		if err := mergeRun(root); err != nil {
			log.Println(" failed:", err)
		} else {
			log.Println(" success")
		}
	}
	return nil
}
