package main

import (
	"os"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path"
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
	confpath := path.Join(os.Getenv("HOME"), ".config", "ctudc", "CtudcHandler.conf")
	data, err := ioutil.ReadFile(confpath)
	if err != nil {
		log.Fatalln("Failed read CtudcHandler.conf:", err)
	}
	if err := json.Unmarshal(data, &conf); err != nil {
		log.Fatalln("Failed unmarshal CtudcHandler.conf:", err)
	}
	return conf
}

func formatRunDir(run int) string {
	return path.Join(appConf.CtudcRoot, "data", fmt.Sprintf("run_%05d", run))
}

func parseRuns(runList string) ([]int, error) {
	re, err := regexp.Compile(`\d+-\d+`)
	if err != nil {
		return nil, err
	}
	var runs []int
	runSet := make(map[int]bool)
	addRun := func(run int) {
		if runSet[run] == false {
			runs = append(runs, run)
			runSet[run] = true
		}
	}
	for _, str := range strings.Split(runList, ",") {
		if str = strings.TrimSpace(str); len(str) == 0 {
			return nil, errors.New("parseRuns: empty string")
		}
		if re.FindString(str) == str && len(str) != 0 {
			dash := strings.IndexRune(str, '-')
			val1, err := strconv.Atoi(str[:dash])
			if err != nil {
				return nil, err
			}
			val2, err := strconv.Atoi(str[dash+1:])
			if err != nil {
				return nil, err
			}
			for i := val1; i <= val2; i++ {
				addRun(i)
			}
		} else {
			if run, err := strconv.Atoi(str); err != nil {
				return nil, err
			} else {
				addRun(run)
			}
		}
	}
	return runs, nil
}

var cmd = flag.String("cmd", "handle", "type of command: handle|merge")
var runs = flag.String("runs", "", `list of runs, e.g. "1, 2, 3, 4, 6-10"`)

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
	case "list":
		if err := list(runList); err != nil {
			log.Println("Failed list data: ", err)
		}
	default:
		log.Println("Invalid cmd")
	}
}
