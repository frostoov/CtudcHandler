package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/frostoov/CtudcHandler/trek"
)

type appConfig struct {
	CtudcRoot string  `json:"ctudc_root"`
	Speed     float64 `json:"speed"`
	Offset    uint    `json:"offset"`
}

func readAppConfig() appConfig {
	var conf appConfig
	data, err := ioutil.ReadFile("CtudcHandler.conf")
	if err != nil {
		log.Fatalln("Failed read CtudcHandler.conf:", err)
	}
	if err := json.Unmarshal(data, &conf); err != nil {
		log.Fatalln("Failed unmarshal CtudcHandler.conf:", err)
	}
	return conf
}

func formatRunDir(run int) string {
	return appConf.CtudcRoot + fmt.Sprintf("/run_%03d", run)
}

var appConf = readAppConfig()

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

var cmd = flag.String("cmd", "handle", "type of command: handle|merge")
var runs = flag.String("runs", "", `list of runs, e.g. "1, 2, 3, 4"`)

func main() {
	flag.Parse()
	var runList []int
	for _, s := range strings.Split(*runs, ",") {
		val, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
		if err != nil {
			log.Fatalln("Invalid run list")
		}
		runList = append(runList, int(val))
	}
	switch *cmd {
	case "handle":
		if err := handle(runList); err != nil {
			log.Println("Failed handle data:", err)
		}
	case "merge":
		if err := merge(runList); err != nil {
			log.Println("Failed merge data:", err)
		}
	default:
		log.Println("Invalid cmd")
	}
}
