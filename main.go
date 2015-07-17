// conv3to2 project main.go
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"runtime"
	"sync"
	"time"
)

var workers = runtime.NumCPU()

func main() {
	start := time.Now()
	runtime.GOMAXPROCS(workers)

	fileSourse := flag.String("s", "mission.sqf", "SOURCE FILE")
	fileDestination := flag.String("d", "local.sqf", "DESTINATION FILE")
	flag.Parse()

	ioBytes, err := ioutil.ReadFile(*fileSourse)
	if err != nil {
		fmt.Println("Error opening file:", err)
		os.Exit(1)
	}

	var (
		wg         sync.WaitGroup
		matcheChan = make(chan []byte)
		rx0        = regexp.MustCompile(`_vehicle_[\d]* [^}]+};`)
		rx1        = regexp.MustCompile(`_this = createVehicle \["([A-Za-z0-9_]*)", \[[ -.,0-9e]*\], \[\], 0, "[A-Za-z0-9_]*"\];`)
		rx2        = regexp.MustCompile(`_this setDir ([ -.,0-9e]*;)`)
		rx3        = regexp.MustCompile(`_this setPos \[([ -.,0-9e]*)\];`)
		rx4        = regexp.MustCompile(`_this setVehicleInit "this setVectorUp \[([ -.,0-9e]*)\];";`)
		rx1r       = []byte(`_this = "$1" createVehicleLocal [0,0,0]; `)
		rx2r       = []byte(`_this setDir $1 `)
		rx3r       = []byte(`_this setPos [$1]; `)
		rx4r       = []byte(`_this setVectorUp [$1]; `)
		dmg        = []byte("_this allowDamage false;\n")
	)

	matches := rx0.FindAll(ioBytes, -1)
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

	for _, matche := range matches {
		go func(matche []byte) {
			wg.Add(1)
			defer wg.Done()
			var obj []byte

			crt := rx1.ReplaceAll(rx1.Find(matche), rx1r)
			dir := rx2.ReplaceAll(rx2.Find(matche), rx2r)
			pos := rx2.ReplaceAll(rx3.Find(matche), rx3r)
			vct := rx4.ReplaceAll(rx4.Find(matche), rx4r)

			if crt != nil && dir != nil {
				obj = append(crt, dir...)
			} else if crt != nil && dir == nil {
				obj = crt
			}
			if pos != nil {
				obj = append(obj, pos...)
			}
			if vct != nil {
				obj = append(obj, vct...)
			}
			if obj != nil {
				obj = append(obj, dmg...)
			}

			if obj != nil {
				matcheChan <- obj
			} else {
				wg.Done()
			}

		}(matche)
	}

	go func() {
		for data := range matcheChan {
			file.Write(data)
		}
	}()

	wg.Wait()

	elapsed := time.Since(start)
	fmt.Printf("\nTime taken %s\n", elapsed)

}
