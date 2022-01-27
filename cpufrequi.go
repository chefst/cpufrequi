package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

var timeInStateSample = `
600000 8606271
700000 15348
800000 10935
900000 17106
1000000 9476
1100000 12240
1200000 103422
1300000 61198
1400000 19632
1500000 20460
1600000 18667
1700000 7363
1800000 135070
`

var timeInStateSample2 = `
600000 8606271
700000 15348
800000 10935
900000 17106
1000000 9476
1100000 12240
1200000 103422
1300000 61198
1400000 19632
1500000 20460
1600000 28667
1700000 7763
1800000 545070
`

var timeInStateSample3 = `
600000 9606271
700000 15348
800000 20935
900000 37106
1000000 19476
1100000 12240
1200000 103422
1300000 61198
1400000 19632
1500000 20460
1600000 28667
1700000 7763
1800000 545070
`

var timeInStateSample4 = `
600000 9606271
700000 35348
800000 20935
900000 37106
1000000 59476
1100000 12240
1200000 103422
1300000 61198
1400000 79632
1500000 20460
1600000 28667
1700000 7763
1800000 545070
`

var gauges []*widgets.Gauge
var interval, historySize, windowSize int

func main() {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	interval = 1000
	historySize = 1000
	windowSize = 5
	flag.IntVar(&interval, "i", 1000, "interval in ms")
	flag.IntVar(&historySize, "s", 1000, "size of history")
	flag.IntVar(&windowSize, "w", 1000, "size of avg window")

	gauges = make([]*widgets.Gauge, 0)
	totalTime := 0

	history := make([][][]int, 0)

	states := []string{timeInStateSample, timeInStateSample2, timeInStateSample3, timeInStateSample4}

	go func() {
		for range time.Tick(time.Millisecond * time.Duration(interval)) {
			if len(states) < 1 {
				return
			}
			values, totalTimeInInterval := getValuesFromDisk(states[0])
			states = states[1:]
			history = addToHistory(history, values)
			totalTime += totalTimeInInterval
			oldValues := history[0]
			populateGauges(oldValues, values, totalTimeInInterval)
			renderAll(gauges)
		}
	}()

	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		// fmt.Println(e.ID)
		// fmt.Println(e.Payload)
		// fmt.Println(e.Type)

		switch e.ID {
		case "<Resize>":
			renderAll(gauges)
		case "r":
			renderAll(gauges)
		case "q", "<C-c>":
			return
		}
	}
}

func addToHistory(history [][][]int, values [][]int) [][][]int {
	if len(history) >= historySize {
		history = append(history[1:], values)
	} else {
		history = append(history, values)
	}
	return history
}

func getValuesFromDisk(file string) ([][]int, int) {
	values := make([][]int, 0)
	totalTime := 0
	scanner := bufio.NewScanner(strings.NewReader(file))
	for scanner.Scan() {
		freqtime := strings.Split(scanner.Text(), " ")
		if len(freqtime) < 2 {
			continue
		}
		freq, err := strconv.Atoi(freqtime[0])
		if err != nil {
			fmt.Println("Lol", err)
			panic(err)
		}
		dur, err := strconv.Atoi(freqtime[1])
		if err != nil {
			fmt.Println("Lolll", err)
			panic(err)
		}
		val := []int{freq, dur}
		values = append(values, val)
		totalTime = totalTime + dur
	}

	return values, totalTime
}

func populateGauges(valuesOld [][]int, valuesNew [][]int, totalTimeInInterval int) {
	i := 0
	if len(gauges) < 1 {
		for range valuesNew {
			g := widgets.NewGauge()
			gauges = append(gauges, g)
		}
	}

	for _, v := range gauges {
		freq := valuesNew[i][0]
		dur := valuesNew[i][1]
		// g.Title = "Slim Gauge"
		// fmt.Println("freq dur ttii", freq, dur, totalTimeInInterval)
		v.Percent = dur / (totalTimeInInterval / 100)
		v.SetRect(0, i, 100, i+1)
		v.Label = fmt.Sprintf("%d MHz     %d%%", freq/1000, v.Percent)
		v.BarColor = ui.ColorYellow
		v.Border = false
		i++
	}
}

func renderAll(gauges []*widgets.Gauge) {
	for _, g := range gauges {
		ui.Render(g)
	}
}
