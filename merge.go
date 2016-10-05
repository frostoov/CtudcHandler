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
	"time"

	"golang.org/x/text/encoding/charmap"

	"github.com/frostoov/CtudcHandler/math"
	"github.com/frostoov/CtudcHandler/nevod"
	"github.com/frostoov/CtudcHandler/trek"
)

func parseNADDate(datetime string, loc *time.Location) (time.Time, error) {
	var (
		day, month, year     int
		hour, min, sec, msec int
	)
	if _, err := fmt.Sscanf(datetime, "%02d-%02d-%02d %02d:%02d:%02d.%03d", &day, &month, &year, &hour, &min, &sec, &msec); err != nil {
		return time.Time{}, fmt.Errorf("failed parse NAD date %v", err)
	}
	return time.Date(year+2000, time.Month(month), day, hour, min, sec, msec*1000000, loc), nil
}

func parseStdat(data string, h *trek.ExtHeader) error {
	fmt.Printf("stdat: %v", data)
	parseMark := func(text, mark, trail string) (string, error) {
		front := strings.Index(text, mark)
		if front == -1 {
			return "", fmt.Errorf("failed find %q mark", mark)
		}
		front += len(mark)
		back := front + strings.Index(text[front+1:], trail)
		if back < front {
			return "", fmt.Errorf("failed find %q trail", trail)
		}
		return strings.TrimSpace(text[front:back]), nil
	}
	parseTime := func(text, mark string) (time.Time, error) {
		if str, err := parseMark(text, mark, "\n"); err != nil {
			return time.Time{}, err
		} else if t, err := parseNADDate(str, time.UTC); err != nil {
			return time.Time{}, err
		} else {
			return t, nil
		}
	}
	parseDur := func(text, mark string) (time.Duration, error) {
		if str, err := parseMark(text, mark, " сек"); err != nil {
			return time.Duration(0), err
		} else if dur, err := strconv.Atoi(str); err != nil {
			return time.Duration(0), err
		} else {
			return time.Duration(dur) * time.Second, nil
		}
	}

	if startTime, err := parseTime(data, "Cтарт"); err != nil {
		return err
	} else {
		h.StartTime = startTime
	}
	if stopTime, err := parseTime(data, "Cтоп"); err != nil {
		return err
	} else {
		h.StopTime = stopTime
	}
	if liveDur, err := parseDur(data, "Живое время="); err != nil {
		return err
	} else {
		h.LiveDur = liveDur
	}
	if fullDur, err := parseDur(data, "Полное время="); err != nil {
		return err
	} else {
		h.FullDur = fullDur
	}
	return nil
}

func parseGener(data string, h *trek.ExtHeader) error {
	lines := strings.Split(strings.TrimSpace(data), "\n")
	if len(lines) != 3 {
		return fmt.Errorf("parseGener invalid(%d) lines count", len(lines))
	}
	// Пропускаем хэдэр
	lines = lines[1:]
	var eventNo [2]uint64
	for i := range eventNo {
		words := strings.Split(strings.TrimSpace(lines[i]), "\t")
		if len(words) != 4 {
			return fmt.Errorf("parseGener invalid(%d) words count", len(words))
		}
		val, err := strconv.ParseUint(strings.TrimSpace(words[1]), 10, 64)
		if err != nil {
			return fmt.Errorf("parseGener failed parse event numbe %v", err)
		}
		eventNo[i] = val
	}
	h.FirstEvent = eventNo[0]
	h.LastEvent = eventNo[1]
	return nil
}

func readCP1251(filename string) ([]byte, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return charmap.Windows1251.NewDecoder().Bytes(data)
}

func readRunMeta(run int) (trek.ExtHeader, error) {
	stdat, err := readCP1251(filepath.Join(formatNevodRunDir(run), "stdat"))
	if err != nil {
		return trek.ExtHeader{}, err
	}

	gener, err := readCP1251(filepath.Join(formatNevodRunDir(run), "gener"))
	if err != nil {
		return trek.ExtHeader{}, err
	}
	var header trek.ExtHeader
	if err := parseStdat(string(stdat), &header); err != nil {
		return header, err
	}
	if err := parseGener(string(gener), &header); err != nil {
		return header, err
	}
	return header, nil
}

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
			f, err := os.Open(filepath.Join(dirname, fileStat.Name()))
			log.Println("Opening file: ", fileStat.Name())
			if err != nil {
				continue
			}
			s, err := trek.NewScanner(f)
			if err != nil {
				panic("failed create scanner: " + err.Error())
			}
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
			f, err := os.Open(filepath.Join(dirname, fileStat.Name()))
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

func mergeRun(run int) error {
	root := formatRunDir(run)
	ctudc := formatCtudcSubdir(run)
	nevod := formatNevodRunDir(run)
	extData := filepath.Join(root, fmt.Sprintf("extctudc_%05d.tds", run))
	decor := filepath.Join(root, "decor.dat")
	decorShSh := filepath.Join(root, "decor_shsh.dat")
	meta, err := readRunMeta(run)
	//fmt.Printf("Meta: \nevents[%v, %v]\ntime[%v, %v]\ndur[%v, %v]\n",
	//	meta.FirstEvent, meta.LastEvent, meta.StartTime, meta.StopTime, meta.LiveDur, meta.FullDur)
	if err != nil {
		return fmt.Errorf("failed read nevod run meta %v", err)
	}

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
	w.WriteString("TDSext_m\n")
	if err := meta.Marshal(w); err != nil {
		return fmt.Errorf("failed marshal file header %v", err)
	}
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
				log.Fatalln("OOps!! invalid run number!! data is corrupted")
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
		log.Print("Processing ", formatRunDir(run))
		if err := mergeRun(run); err != nil {
			log.Println(" failed:", err)
		} else {
			log.Println(" success")
		}
	}
	return nil
}
