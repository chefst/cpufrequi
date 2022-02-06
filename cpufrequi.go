package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

var freqGauges []*widgets.Gauge
var temperatureParagraph, currentFreqParagraph, gaugeSettingsParagraph *widgets.Paragraph
var allDrawables []ui.Drawable
var interval, historySize, windowSize, temperature, termWidth, termHeight int
var history [][][]int

func main() {
	flag.IntVar(&interval, "i", 1000, "interval in ms")
	flag.IntVar(&historySize, "s", 1000, "size of history")
	flag.IntVar(&windowSize, "w", 5, "size of avg window")
	flag.Parse()

	if historySize < windowSize {
		fmt.Println("dont use history size < window size")
		os.Exit(1)
	}

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	freqGauges = make([]*widgets.Gauge, 0)
	allDrawables = make([]ui.Drawable, 0)
	history = make([][][]int, 0)

	freqTableFile, err := os.Open("/sys/devices/system/cpu/cpufreq/policy0/stats/time_in_state")
	if err != nil {
		panic(err)
	}

	temperatureFile, err := os.Open("/sys/devices/platform/soc/soc:firmware/raspberrypi-hwmon/hwmon/hwmon1/device/hwmon/hwmon1/subsystem/hwmon0/temp1_input")
	if err != nil {
		panic(err)
	}

	currentFreqFile, err := os.Open("/sys/devices/system/cpu/cpufreq/policy0/scaling_cur_freq")
	if err != nil {
		panic(err)
	}

	setupUIElements(freqTableFile)

	go func() {
		populateUI(freqTableFile, temperatureFile, currentFreqFile)
		for range time.Tick(time.Millisecond * time.Duration(interval)) {
			populateUI(freqTableFile, temperatureFile, currentFreqFile)
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
			terminalResized()
		case "q", "<C-c>":
			return
		}
	}
}

func terminalResized() {
	termWidth, termHeight = ui.TerminalDimensions()
	currentFreqParagraph.SetRect(0, 0, termWidth, 1)
	temperatureParagraph.SetRect(0, 1, termWidth, 2)
	gaugeSettingsParagraph.SetRect(0, 3, termWidth, 4)
	for _, g := range freqGauges {
		rect := g.GetRect()
		g.SetRect(rect.Min.X, rect.Min.Y, termWidth, rect.Max.Y)
	}
	renderAll()
}

func setupUIElements(freqTableFile *os.File) {

	termWidth, termHeight = ui.TerminalDimensions()

	// freq table gauges
	scanner := bufio.NewScanner(freqTableFile)
	for scanner.Scan() {
		g := widgets.NewGauge()
		freqGauges = append(freqGauges, g)
		allDrawables = append(allDrawables, g)
	}

	// current freq graph
	currentFreqParagraph = widgets.NewParagraph()
	currentFreqParagraph.SetRect(0, 0, termWidth, 1)
	currentFreqParagraph.Border = false
	allDrawables = append(allDrawables, currentFreqParagraph)

	// temperature text
	temperatureParagraph = widgets.NewParagraph()
	temperatureParagraph.SetRect(0, 1, termWidth, 2)
	temperatureParagraph.Border = false
	allDrawables = append(allDrawables, temperatureParagraph)

	// gauge settings
	gaugeSettingsParagraph = widgets.NewParagraph()
	gaugeSettingsParagraph.SetRect(0, 3, termWidth, 4)
	gaugeSettingsParagraph.Border = false
	allDrawables = append(allDrawables, gaugeSettingsParagraph)
}

func populateTemperatureParagraph(temperatureFile *os.File) {
	temperatureFile.Seek(0, 0)
	scanner := bufio.NewScanner(temperatureFile)
	for scanner.Scan() {
		tempStr := scanner.Text()
		tempFloat, _ := strconv.ParseFloat(tempStr, 32)
		tempFloat = tempFloat / 1000.0
		temperatureParagraph.Text = fmt.Sprintf("Current Temp. (°C): %.3f", tempFloat)
	}
}

func populateCurrentFreqParagraph(currentFreqFile *os.File) {
	currentFreqFile.Seek(0, 0)
	scanner := bufio.NewScanner(currentFreqFile)
	for scanner.Scan() {
		freqInt, _ := strconv.Atoi(scanner.Text())
		currentFreqParagraph.Text = "Current Frq. (MHz): " + strconv.Itoa(freqInt/1000)
	}
}

func populateGaugeSettingsParagraph() {
	windowLengthMs := windowSize * interval
	gaugeSettingsParagraph.Text = fmt.Sprintf("%dms avg. (window=%d * interval=%dms)", windowLengthMs, windowSize, interval)
}

func populateUI(freqTableFile, temperatureFile, currentFreqFile *os.File) {
	populateCurrentFreqParagraph(currentFreqFile)
	populateTemperatureParagraph(temperatureFile)
	populateGaugeSettingsParagraph()
	populateGauges(freqTableFile)
	renderAll()
}

func addToHistory(history [][][]int, values [][]int) [][][]int {
	if len(history) >= historySize {
		history = append(history[1:], values)
	} else {
		history = append(history, values)
	}
	return history
}

func getTotalTime(values [][]int) int {
	t := 0
	for _, v := range values {
		t += v[1]
	}
	return t
}

func getFreqTableFromFile(freqTableFile *os.File) ([][]int, int) {
	values := make([][]int, 0)
	val := make([]int, 0)
	totalTime := 0
	freqTableFile.Seek(0, 0)
	scanner := bufio.NewScanner(freqTableFile)
	for scanner.Scan() {
		freqtime := strings.Split(scanner.Text(), " ")
		if len(freqtime) < 2 {
			continue
		}
		freq, err := strconv.Atoi(freqtime[0])
		if err != nil {
			panic(err)
		}
		dur, err := strconv.Atoi(freqtime[1])
		if err != nil {
			panic(err)
		}
		val = []int{freq, dur}
		values = append(values, val)
		totalTime = totalTime + dur
	}

	return values, totalTime
}

func populateGauges(freqTableFile *os.File) {
	i := 0
	var freq, durNew, durOld, durWindow int

	valuesNew, totalTimeNew := getFreqTableFromFile(freqTableFile)
	history = addToHistory(history, valuesNew)
	historyIndex := len(history) - (windowSize + 1)
	if historyIndex < 0 || historyIndex+1 > len(history) {
		historyIndex = 0
	}
	valuesOld := history[historyIndex]
	totalTimeOld := getTotalTime(valuesOld)
	totalTimeInInterval := totalTimeNew - totalTimeOld

	for _, v := range freqGauges {
		freq = valuesNew[i][0]
		durNew = valuesNew[i][1]
		durOld = valuesOld[i][1]
		durWindow = durNew - durOld
		percentFloat := float64(durWindow) / float64(float64(totalTimeInInterval)/100.0)
		v.Percent = int(math.Round(percentFloat))
		// v.Title = strconv.Itoa(freqNew / 1000)
		// v.TitleStyle = v.LabelStyle
		v.SetRect(0, i+5, termWidth, i+1+5)
		v.Label = fmt.Sprintf("%4d MHz %6.2f%%", freq/1000, percentFloat)
		v.Border = false
		i++
	}
}

func renderAll() {
	ui.Render(allDrawables...)
}
