package main

import (
	"os"
	"testing"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

var freqTableFile, temperatureFile, currentFreqFile *os.File

func init() {
	freqGauges = make([]*widgets.Gauge, 0)
	allDrawables = make([]ui.Drawable, 0)
	history = make([][][]int, 0)

	interval = 1000
	historySize = 1000
	windowSize = 5

	freqTableFile, _ = os.Open("/sys/devices/system/cpu/cpufreq/policy0/stats/time_in_state")
	temperatureFile, _ = os.Open("/sys/devices/platform/soc/soc:firmware/raspberrypi-hwmon/hwmon/hwmon1/device/hwmon/hwmon1/subsystem/hwmon0/temp1_input")
	currentFreqFile, _ = os.Open("/sys/devices/system/cpu/cpufreq/policy0/scaling_cur_freq")

	setupUIElements(freqTableFile)
}

func BenchmarkRenderEmpty(b *testing.B) {
	for n := 0; n < b.N; n++ {
		renderAll()
	}
}

func BenchmarkRenderFilled(b *testing.B) {
	populateCurrentFreqParagraph(currentFreqFile)
	populateTemperatureParagraph(temperatureFile)
	populateGaugeSettingsParagraph()
	populateGauges(freqTableFile)
	for n := 0; n < b.N; n++ {
		renderAll()
	}
}

func BenchmarkUITempPara(b *testing.B) {
	for n := 0; n < b.N; n++ {
		populateTemperatureParagraph(temperatureFile)
	}

}

func BenchmarkUITempCurFreqPara(b *testing.B) {
	for n := 0; n < b.N; n++ {
		populateCurrentFreqParagraph(currentFreqFile)
	}
}

func BenchmarkUISettingsPara(b *testing.B) {
	for n := 0; n < b.N; n++ {
		populateGaugeSettingsParagraph()
	}
}

func BenchmarkUIFreqGauges(b *testing.B) {
	for n := 0; n < b.N; n++ {
		populateGauges(freqTableFile)
	}
}

func BenchmarkFreqValuesFromFile(b *testing.B) {
	for n := 0; n < b.N; n++ {
		getFreqTableFromFile(freqTableFile)
	}
}

func BenchmarkLoop(b *testing.B) {
	for n := 0; n < b.N; n++ {
		populateUI(freqTableFile, temperatureFile, currentFreqFile)
	}
}
