// conv3to2 project main.go
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"runtime"
	"strings"
	//"time"
)

var workers = runtime.NumCPU()

func main() {
	//start := time.Now()
	runtime.GOMAXPROCS(workers)

	fileSourse := flag.String("s", "mission.sqf", "SOURCE FILE")
	fileDestination := flag.String("d", "local.sqf", "DESTINATION FILE")
	flag.Parse()

	bytes, err := ioutil.ReadFile(*fileSourse)
	if err != nil {
		fmt.Println("Error opening file:", err)
		os.Exit(1)
	}

	rx := regexp.MustCompile(`_vehicle_([\d]*) [^}]+};`)
	//rx0 := regexp.MustCompile(`_vehicle_([\d]*) = objNull;`)
	rx1 := regexp.MustCompile(`_this = createVehicle \["([A-Za-z0-9_]*)", \[[ -.,0-9e]*\], \[\], 0, "[A-Za-z0-9_]*"\];`)
	rx2 := regexp.MustCompile(`_this setDir ([ -.,0-9e]*;)`)
	rx3 := regexp.MustCompile(`_this setPos \[([ -.,0-9e]*)\];`)
	rx4 := regexp.MustCompile(`_this setVehicleInit "this setVectorUp \[([ -.,0-9e]*)\];";`)

	matches := rx.FindAllString(string(bytes), -1)
	if matches == nil {
		fmt.Println("No fount matches")
		os.Exit(1)
	}

	file, err := os.Create(*fileDestination)
	if err != nil {
		fmt.Println("Error creating file:", err)
		os.Exit(1)
	}
	defer file.Close()

	for _, m := range matches {
		s := bufio.NewScanner(strings.NewReader(m))
		var line string
		for s.Scan() {
			if rx1.MatchString(s.Text()) == true {
				line += strings.Trim((rx1.ReplaceAllString(s.Text(), `_this = "$1" createVehicleLocal [0,0,0];`)), " ") + "\n"
			} else if rx2.MatchString(s.Text()) == true {
				line += strings.Trim((rx2.ReplaceAllString(s.Text(), `_this setDir $1`)), " ") + "\n"
			} else if rx3.MatchString(s.Text()) == true {
				line += strings.Trim((rx3.ReplaceAllString(s.Text(), `_this setPos [$1];`)), " ") + "\n"
			} else if rx4.MatchString(s.Text()) == true {
				line += strings.Trim((rx4.ReplaceAllString(s.Text(), `_this setVectorUp [$1];`)), " ") + "\n"
			}
		}
		if len(line) != 0 {
			line += "_this allowDamage false;\n\n"
		}
		file.WriteString(line)
	}

	//elapsed := time.Since(start)
	//fmt.Printf("\nTime taken %s\n", elapsed)
}
