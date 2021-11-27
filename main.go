package main

import (
	"strconv"
	"fmt"
	"time"
	"log"
	"os/exec"
	"strings"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/shirou/gopsutil/v3/mem"
)
const pageSize = 16384



func getTotal() int {
	value, err := mem.VirtualMemory()
	if err != nil {
		fmt.Println(err.Error())
		return 0
	}
	return int(value.Total)
}

func getMem() int {
	value, err := mem.VirtualMemory()
	if err != nil {
		fmt.Println(err.Error())
		return 0
	}
	// return int(value.Total - value.Free - value.Cached - value.Buffers)
	return int(value.Used)
}

// func getMem() int {
//	cmd := exec.Command("vm_stat")
//	stdout, err := cmd.Output()
//	if err != nil {
//		fmt.Println(err.Error())
//		return 0
//	}
//	lines := strings.Split(string(stdout), "\n")
//	matchedLine := ""
//	for _, line := range lines {
//		if strings.Contains(line, "File-backed pages") {
//			matchedLine = line
//		}
//	}
//	if matchedLine == "" {
//		return 0
//	}

//	words := strings.Fields(matchedLine)
//	number := strings.Trim(words[len(words)-1], ".")
//	fileCache, err := strconv.ParseInt(number, 10, 64)
//	if err != nil {
//		return 0
//	}
//	return int(fileCache * pageSize)
// }

func getSwapRam() int {
	value, err := mem.SwapMemory()
	if err != nil {
		fmt.Println(err.Error())
		return 0
	}
	return int(value.Used)
}

func getCompressedRam() int {
	cmd := exec.Command("vm_stat")
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error())
		return 0
	}
	lines := strings.Split(string(stdout), "\n")
	matchedLine := ""
	for _, line := range lines {
		if strings.Contains(line, "Pages occupied by compressor") {
			matchedLine = line
		}
	}
	if matchedLine == "" {
		return 0
	}

	words := strings.Fields(matchedLine)
	number := strings.Trim(words[len(words)-1], ".")
	compressed, err := strconv.ParseInt(number, 10, 64)
	if err != nil {
		return 0
	}
	return int(compressed * pageSize)
}

func main() {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	rawTotal := getTotal()
	total := float64(rawTotal / 1024 / 1024 / 1024)

	x, y := ui.TerminalDimensions()
	data := make([][]float64, 1)
	data[0] = make([]float64, 256)

	swapData := make([][]float64, 1)
	swapData[0] = make([]float64, 256)

	memData := make([][]float64, 1)
	memData[0] = make([]float64, 256)

	halfX, halfY := int(x / 2), int(y / 2)

	p0 := widgets.NewPlot()
	p0.MaxVal = float64(total)
	p0.Title = "Compressed memory | 0.00 GB"
	p0.Data = data
	p0.SetRect(0, 0, halfX, halfY)
	p0.AxesColor = ui.ColorGreen
	p0.LineColors[0] = ui.ColorGreen
	p0.ShowAxes = false

	p1 := widgets.NewPlot()
	p1.Title = "Swap memory | 0.00 GB"
	p1.Data = swapData
	p1.SetRect(halfX, 0, x, halfY)
	p1.AxesColor = ui.ColorGreen
	p1.LineColors[0] = ui.ColorGreen
	p1.ShowAxes = false

	// p2 := widgets.NewPlot()
	// p2.MaxVal = float64(total)
	// p2.Title = "Memory used | 0.00 GB"
	// p2.Data = memData
	// p2.SetRect(0, halfY, x, y)
	// p2.AxesColor = ui.ColorGreen
	// p2.LineColors[0] = ui.ColorGreen
	// p2.ShowAxes = false

	ui.Render(p0, p1)

	uiEvents := ui.PollEvents()
	ticker := time.NewTicker(time.Second).C

	i := 0

	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return
			}
		case <- ticker:
			swap := float64(getSwapRam()) / 1024 / 1024 / 1024
			// mem := float64(getMem() / 1024 / 1024 / 1024)
			ram := float64(getCompressedRam()) / 1024 / 1024 / 1024

			p0.Title = fmt.Sprintf("Compressed memory | %.2f GB", ram)
			p1.Title = fmt.Sprintf("Swap memory | %.2f GB", swap)
			p1.MaxVal = swap * 2
			// p2.Title = fmt.Sprintf("Memory used (total %.f GB) | %.2f GB", total, mem)

			data[0][i] = ram
			swapData[0][i] = swap
			// memData[0][i] = mem

			i += 1
			ui.Render(p0, p1)
		}

	}
}
