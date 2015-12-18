package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	geo "github.com/frostoov/CtudcHandler/math"
	"github.com/frostoov/CtudcHandler/nevod"
	"github.com/frostoov/CtudcHandler/trek"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type appConfig struct {
	CtudcRoot string  `json:"ctudc_root"`
	Speed     float64 `json:"speed"`
	Offset    uint    `json:"offset"`
}

func readDecorTracks(filname string) (map[uint][]geo.Line3, error) {
	f, err := os.Open(filname)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	if !s.Scan() {
		return nil, errors.New("readDecorTracks no data")
	}

	events := make(map[uint][]geo.Line3, 10000)
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
		point := geo.Vec3{X: pts[0], Y: pts[1], Z: pts[2]}
		vector := geo.Vec3{X: pts[3], Y: pts[4], Z: pts[5]}
		track := geo.Line3{Point: point, Vector: vector}
		events[uint(eventNumber)] = append(events[uint(eventNumber)], track)
	}
	return events, nil
}

func ctudcReader(dirname string) (<-chan trek.Event, error) {
	fileList, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, err
	}
	c := make(chan trek.Event)
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
				c <- *s.Record()
			}
			f.Close()
		}
		close(c)
	}()
	return c, nil
}

func nevodReader(dirname string) (<-chan nevod.Event, error) {
	fileList, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, err
	}
	c := make(chan nevod.Event)
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
				c <- *s.Record()
			}
			f.Close()
		}
		close(c)
	}()
	return c, nil
}

func convertConfig(config []trek.ChamberDesc) {
	coor := geo.NewCoordSystem(
		geo.Vec3{X: 26891.4, Y: -10028.6, Z: -9572.1},
		geo.Vec3{X: 0, Y: 1, Z: 0},
		geo.Vec3{X: -1, Y: 0, Z: 0},
		geo.Vec3{X: 0, Y: 0, Z: 1})
	for i := range config {
		for p := range config[i].Points {
			config[i].Points[p].Y = -config[i].Points[p].Y
			config[i].Points[p] = coor.ConvertVector(config[i].Points[p])
		}
	}
}

func readChamberConfig(filename string) ([]trek.ChamberDesc, error) {
	var chamConfig []trek.ChamberDesc
	if data, err := ioutil.ReadFile(filename); err != nil {
		return nil, err
	} else if err := json.Unmarshal(data, &chamConfig); err != nil {
		return nil, err
	}
	convertConfig(chamConfig)
	return chamConfig, nil
}

func readChambers(filename string) (map[uint]*trek.Chamber, error) {
	chamConfig, err := readChamberConfig(filename)
	if err != nil {
		return nil, err
	}
	chambers := make(map[uint]*trek.Chamber)
	for i := range chamConfig {
		chambers[chamConfig[i].Number-1] = trek.NewChamber(chamConfig[i])
	}
	return chambers, nil
}

func merge() {
	if len(os.Args) < 2 {
		log.Fatalln("Specify runs")
	}
	var conf appConfig
	data, err := ioutil.ReadFile("CtudcReader.conf")
	if err != nil {
		log.Fatalln("Failed read CtudcReader.conf:", err)
	}
	if err := json.Unmarshal(data, &conf); err != nil {
		log.Fatalln("Failed unmarshal CtudcReader.conf:", err)
	}

	for _, run := range os.Args[1:] {
		num, err := strconv.ParseUint(run, 10, 32)
		if err != nil {
			log.Fatalln(err)
		}
		root := conf.CtudcRoot + fmt.Sprintf("/run_%03d", num)
		ctudc := root + "/ctudc"
		nevod := root + "/nevod"
		decor := root + "/decor.dat.bkp"
		decorShSh := root + "/decor.dat"
		extData := root + "/extctudc.tds"
		log.Println("Processing ", root)

		decorTracks, err := readDecorTracks(decor)
		if err != nil {
			log.Fatalln("Failed read shsh decor tracks:", err)
		}
		decorTracksShSh, err := readDecorTracks(decorShSh)
		if err != nil {
			log.Fatalln("Failed read decor tracks:", err)
		}
		ctudcStream, err := ctudcReader(ctudc)
		if err != nil {
			log.Fatalln("Failed open ctudc data:", err)
		}
		nevodStream, err := nevodReader(nevod)
		output, err := os.Create(extData)
		if err != nil {
			log.Fatalln("Failed create output file:", err)
		}
		w := bufio.NewWriter(output)
		w.WriteString("TDSext\n")
		for ctudcEvent, ok := <-ctudcStream; ok; ctudcEvent, ok = <-ctudcStream {
			nrun, nevent := ctudcEvent.Nrun(), ctudcEvent.Nevent()
			var decor []trek.DecorTrack

			if allTracks, ok := decorTracks[nevent]; ok {
				shTracks, _ := decorTracksShSh[nevent]
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

			for event, ok := <-nevodStream; ok; event, ok = <-nevodStream {
				if nrun != uint(event.Meta.Nrun) {
					panic("OOps!! invalid run number!! data is corrupted")
				}
				if nevent == uint(event.Meta.Nevent) {
					extEvent := trek.ExtEvent{
						Ctudc: ctudcEvent,
						Nevod: event.Meta,
						Decor: decor,
					}
					extEvent.Marshal(w)
					break
				} else if uint(event.Meta.Nevent) > nevent {
					break
				}
			}
		}
		output.Close()
	}
}

func print() {
	if len(os.Args) != 2 {
		log.Fatalln("Specify file")
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	reader := bufio.NewReader(f)
	header, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Print(header)
	var record trek.ExtEvent
	var ShSh uint
	var long uint
	for record.Unmarshal(reader) == nil {
		for i := range record.Decor {
			switch record.Decor[i].Type {
			case 0:
				long++
			case 1:
				ShSh++
			}
		}
	}
	fmt.Println("ShSh:", ShSh)
	fmt.Println("long:", long)
}

func toAng(rad float64) float64 {
	return rad / math.Pi * 180
}

func handle() {
	if len(os.Args) < 2 {
		log.Fatalln("Specify runs")
	}
	var conf appConfig
	data, err := ioutil.ReadFile("CtudcReader.conf")
	if err != nil {
		log.Fatalln("Failed read CtudcReader.conf:", err)
	}
	if err := json.Unmarshal(data, &conf); err != nil {
		log.Fatalln("Failed unmarshal CtudcReader.conf:", err)
	}

	if err := os.MkdirAll("output/tracks", 0777); err != nil {
		log.Fatalln("Failed create output dir:", err)
	}
	tracksFiles := make(map[uint]*os.File)
	load, err := os.Create("output/load.dat")
	if err != nil {
		log.Fatalln(err)
	}
	defer load.Close()
	defer func() {
		for i := range tracksFiles {
			tracksFiles[i].Close()
		}
	}()
	//	table, err := os.Create("output/table.dat") !!!!!!Файл для 3 пунтка
	for _, run := range os.Args[1:] {
		num, err := strconv.ParseUint(run, 10, 32)
		if err != nil {
			log.Fatalln("Invalid run:", err)
		}
		root := conf.CtudcRoot + fmt.Sprintf("/run_%03d", num)
		chambers, err := readChambers(root + "/chambers.conf.new")
		if err != nil {
			log.Fatalln("Failed read chamber config:", err)
		}
		f, err := os.Open(root + "/extctudc.tds")
		if err != nil {
			log.Println("Failed open extctudc.tds:", err)
			continue
		}
		r := bufio.NewReader(f)
		if header, err := r.ReadString('\n'); err != nil || header != "TDSext\n" {
			log.Println("Invalid header of extctudc.tds")
			continue
		}
		var record trek.ExtEvent
		for record.Unmarshal(r) == nil {
			// 1. Загрузка
			var loadChams uint
			var muons uint
			for _, times := range record.Ctudc.TrekTimes() {
				depth := getDepth(times)
				muons += uint(depth)
				if depth > 0 {
					loadChams++
				}
			}
			if muons > 1 {
				fmt.Fprintf(load, "%d\t%d\t%d\t%d\t%d\n", loadChams, muons, record.Ctudc.Nevent(), len(record.Decor), record.Nevod.NfifoC)
			}
			// 2. Углы
			for cham, times := range record.Ctudc.TrekTimes() {
				for _, dEvent := range record.Decor {
					if !chambers[cham].Hexahendron().Crossing(dEvent.Track) {
						continue
					}
					cTrack := chambers[cham].CreateTrack(times)
					if cTrack == nil {
						continue
					}
					if tracksFiles[cham] == nil {
						f, err := os.Create(fmt.Sprintf("output/tracks/chamber_%03d.dat", cham))
						if err != nil {
							log.Fatalln("Failed create track file:", err)
						}
						for i := 0; i < 4; i++ {
							fmt.Fprintf(f, "WIRE_%03d\t", i+1)
						}
						// fmt.Fprintf(f, "%8s\t",a ...interface{})

						fmt.Fprintln(f, "k1	      k2	     dev	     ang	       b	  ang[D]	    b[D]	 dang[D]	      db")
						tracksFiles[cham] = f
					}
					f := tracksFiles[cham]
					dTrack := chambers[cham].LineProjection(dEvent.Track)
					k1 := int(cTrack.Times[0] - cTrack.Times[1] - cTrack.Times[2] + cTrack.Times[3])
					k2 := int(cTrack.Times[0] - 3*cTrack.Times[1] + 3*cTrack.Times[2] - cTrack.Times[3])
					dAng := toAng(math.Atan(dTrack.K))
					cAng := toAng(math.Atan(cTrack.Line.K))
					fmt.Fprintf(f, "%8d\t%8d\t%8d\t%8d\t", cTrack.Times[0], cTrack.Times[1], cTrack.Times[2], cTrack.Times[3])
					fmt.Fprintf(f, "%8d\t%8d\t", k1, k2)
					fmt.Fprintf(f, "%8f\t%8f\t%8f\t", cTrack.Deviation, cAng, cTrack.Line.B)
					fmt.Fprintf(f, "%8f\t%8f\t", dAng, dTrack.B)
					fmt.Fprintf(f, "%8f\t%8f\n", cAng-dAng, cTrack.Line.B-dTrack.B)
				}
			}
			// 3. Таблица
		}
		f.Close()
	}
}

func getDepth(times *trek.ChamTimes) uint {
	depth := uint(math.MaxUint64)
	for _, wireHits := range times {
		if uint(len(wireHits)) < depth {
			depth = uint(len(wireHits))
		}
	}
	return depth
}

func main() {
	defer func() {
		fmt.Println("Press any key to exit")
		bufio.NewReader(os.Stdin).ReadString('\n')
	}()
	print()
}
