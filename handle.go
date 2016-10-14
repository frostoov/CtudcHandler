package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	path "path/filepath"

	geo "github.com/frostoov/CtudcHandler/math"
	"github.com/frostoov/CtudcHandler/trek"
)

var validHandlers = map[string]bool{
	"TDS_ext\n":  true,
	"TDSext_m\n": true,
}

type Handler struct {
	chambers    map[int]*trek.Chamber
	tracksFiles map[int]*os.File
	loadFile    *os.File
}

func formatTracksHeader() string {
	var buf bytes.Buffer
	for i := 0; i < 4; i++ {
		fmt.Fprintf(&buf, "WIRE_%03d\t", i+1)
	}
	fmt.Fprintf(&buf, "%8s\t%8s\t%8s\t%8s\t%8s\t%8s\t%8s\t%8s\t%8s",
		"k1", "k2", "dev", "ang[C]", "b[C]", "ang[D]", "b[D]", "dang", "db")
	return buf.String()
}

func NewHandler() (*Handler, error) {
	if err := os.MkdirAll("output/tracks", 0777); err != nil {
		return nil, fmt.Errorf("Failed create output dir: %s", err)
	}
	loadFile, err := os.Create("output/load.dat")
	if err != nil {
		return nil, fmt.Errorf("Failed create load file: %s", err)
	}
	return &Handler{
		tracksFiles: make(map[int]*os.File),
		loadFile:    loadFile,
	}, nil
}

func (h *Handler) Close() {
	if h.loadFile != nil {
		h.loadFile.Close()
	}
	for _, f := range h.tracksFiles {
		f.Close()
	}
}

func (h *Handler) Handle(runs []int) error {
	for _, run := range runs {
		log.Println("Processing ", run)
		if err := h.handleRun(run); err != nil {
			log.Println("Failed:", err)
		} else {
			log.Println("Success")
		}
	}
	return nil
}

func (h *Handler) handleRun(run int) error {
	root := formatRunDir(run)
	chambers, err := readChambers(path.Join(root, "/chambers.conf.new"))
	if err != nil {
		return fmt.Errorf("Failed read chamber config: %s", err)
	}
	f, err := os.Open(path.Join(root, fmt.Sprintf("extctudc_%05d.tds", run)))
	if err != nil {
		return fmt.Errorf("Failed open extctudc.tds: %s", err)
	}
	defer f.Close()
	r := bufio.NewReader(f)
	if header, err := r.ReadString('\n'); err != nil || !validHandlers[header] {
		return fmt.Errorf("Invalid header of extctudc.tds %s", header)
	} else if header == "TDSext_m\n" {
		new(trek.ExtHeader).Unmarshal(r)
	}
	var record trek.ExtEvent
	for record.Unmarshal(r) == nil {
		times := record.Ctudc.Times()
		// 1. Загрузка
		var loadChams uint
		var muons uint
		for cham, times := range times {
			if chamber, ok := chambers[cham]; ok {
				depth := chamber.TimesDepth(times)
				muons += uint(depth)
				if depth > 0 {
					loadChams++
				}
			}
		}
		if muons > 1 {
			fmt.Fprintf(h.loadFile, "%d\t%d\t%d\t%d\t%d\n", loadChams, muons, record.Ctudc.Nevent(), len(record.Decor), record.Nevod.NfifoC)
		}
		// 2. Углы
		for cham, times := range times {
			chamber, ok := chambers[cham]
			if !ok {
				continue
			}
			for _, dEvent := range record.Decor {
				if !chamber.Hexahendron().Crossing(dEvent.Track) {
					continue
				}
				cTrack := chamber.CreateTrack(times)
				if cTrack == nil {
					continue
				}
				if h.tracksFiles[cham] == nil {
					f, err := os.Create(fmt.Sprintf("output/tracks/chamber_%03d.dat", cham+1))
					if err != nil {
						log.Fatalln("Failed create track file:", err)
					}
					if _, err := fmt.Fprintln(f, "#", formatTracksHeader()); err != nil {
						log.Fatalln("Failed write track header:", err)
					}

					h.tracksFiles[cham] = f
				}
				f := h.tracksFiles[cham]
				dTrack := chamber.LineProjection(dEvent.Track)
				k1 := int(cTrack.Times[0] - cTrack.Times[1] - cTrack.Times[2] + cTrack.Times[3])
				k2 := int(cTrack.Times[0] - 3*cTrack.Times[1] + 3*cTrack.Times[2] - cTrack.Times[3])
				dAng := toAng(math.Atan(dTrack.K()))
				cAng := toAng(math.Atan(cTrack.Line.K()))
				fmt.Fprintf(f, "%8d\t%8d\t%8d\t%8d\t", cTrack.Times[0], cTrack.Times[1], cTrack.Times[2], cTrack.Times[3])
				fmt.Fprintf(f, "%8d\t%8d\t", k1, k2)
				fmt.Fprintf(f, "%8f\t%8f\t%8f\t", cTrack.Deviation, cAng, cTrack.Line.B())
				fmt.Fprintf(f, "%8f\t%8f\t", dAng, dTrack.B())
				fmt.Fprintf(f, "%8f\t%8f\n", cAng-dAng, cTrack.Line.B()-dTrack.B())
			}
		}
	}
	return nil
}

func handle(runs []int) error {
	h, err := NewHandler()
	if err != nil {
		return err
	}
	defer h.Close()
	return h.Handle(runs)
}

func toAng(rad float64) float64 {
	return rad / math.Pi * 180
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
		config[i].Number--
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

func readChambers(filename string) (map[int]*trek.Chamber, error) {
	chamConfig, err := readChamberConfig(filename)
	if err != nil {
		return nil, err
	}
	chambers := make(map[int]*trek.Chamber)
	for i := range chamConfig {
		chambers[chamConfig[i].Number] = trek.NewChamber(chamConfig[i])
	}
	return chambers, nil
}
