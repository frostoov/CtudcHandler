package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"strings"
)

type appConfig struct {
	CtudcRoot string  `json:"ctudc_root"`
	Speed     float64 `json:"speed"`
	Offset    uint    `json:"offset"`
}

var appConf = readAppConfig()

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

func parseRuns(runList string) ([]int, error) {
	re, err := regexp.Compile(`\d+-\d`)
	if err != nil {
		return nil, err
	}
	var runs []int
	hasItem := func(item int, array []int) bool {
		for i := range array {
			if item == array[i] {
				return false
			}
		}
		return true
	}
	for _, str := range strings.Split(runList, ",") {
		str = strings.TrimSpace(str)
		if re.FindString(str) == str {
			dash := strings.IndexRune(str, '-')
			val1, err := strconv.Atoi(str[:dash])
			if err != nil {
				return nil, err
			}
			val2, err := strconv.Atoi(str[dash+1:])
			if err != nil {
				return nil, err
			}
			for i := val1; i < val2; i++ {
				if !hasItem(i, runs) {
					runs = append(runs, i)
				}
			}
		} else {
			if val, err := strconv.Atoi(str); err != nil {
				return nil, err
			} else if !hasItem(val, runs) {
				runs = append(runs, val)
			}
		}
	}
	return runs, nil
}

var cmd = flag.String("cmd", "handle", "type of command: handle|merge")
var runs = flag.String("runs", "", `list of runs, e.g. "1, 2, 3, 4"`)

func main() {
	flag.Parse()
	runList, err := parseRuns(*runs)
	if err != nil {
		log.Fatalln("Failed parse runs list:", err)
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
