package prof_timer

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"sort"
	"time"

	"github.com/dterei/gotsc"
)

// Usage:
// func someFunctionToMeasure() {
// 	tm := prof_timer.StartTimer("unique name, i.e. sqpSuggest.applySqp")
// 	defer prof_timer.EndTimer(tm)

// 	// other code
// }
// At some higher level func call prof_timer.GetTimersReport() to see measured times

var (
	staticTimers         = map[string]*timer{}
	tscOverhead          = int64(gotsc.TSCOverhead())
	cyclesPerMillisecond = calibrateTimer()
)

type timer struct {
	cycles     []int64
	curCycles  int64
	recursions int
}

func calibrateTimer() int64 {
	duration := 200 * time.Millisecond
	cycles := int64(gotsc.BenchStart())
	time.Sleep(duration)
	cycles = int64(gotsc.BenchEnd()) - cycles - tscOverhead
	cyclesPerMillisecond := cycles / int64(duration.Milliseconds())
	fmt.Printf("timer calibration: %d cyclesPerMillisecond, tscOverhead: %d cycles\n", cyclesPerMillisecond, tscOverhead)
	return cyclesPerMillisecond
}

func assert(condition bool) {
	if !condition {
		fmt.Printf("assertion failed\n")
	}
}

func ResetTimers() {
	staticTimers = map[string]*timer{}
}

func GetTimersReport() string {
	type measuredTimer struct {
		name        string
		totalCycles int64
		cycles95    int64
		cycles99    int64
		hits        int
	}
	measuredTimers := make([]measuredTimer, 0, len(staticTimers))
	for name, t := range staticTimers {
		sort.Slice(t.cycles, func(i, j int) bool { return t.cycles[i] > t.cycles[j] })

		mt := measuredTimer{
			name: name,
			hits: len(t.cycles),
		}
		proc95 := max(1, int(0.05*float64(mt.hits)))
		proc99 := max(1, int(0.01*float64(mt.hits)))
		for i := range t.cycles {
			mt.totalCycles += t.cycles[i] - int64(tscOverhead)
			if i == proc95 {
				mt.cycles95 = t.cycles[i] - int64(tscOverhead)
			}
			if i == proc99 {
				mt.cycles99 = t.cycles[i] - int64(tscOverhead)
			}
		}
		measuredTimers = append(measuredTimers, mt)
	}
	sort.Slice(measuredTimers, func(i, j int) bool {
		return measuredTimers[i].totalCycles > measuredTimers[j].totalCycles
	})
	output := bytes.NewBuffer(nil)
	writer := csv.NewWriter(output)
	writer.Write([]string{"name", "avg milliseconds", "perc95", "perc99", "hits"})

	getDuration := func(cycles int64, hits int) string {
		ms := float64(cycles) / float64(cyclesPerMillisecond)
		ms /= max(1, float64(hits))
		return fmt.Sprintf("%.2f", ms)
	}

	for _, mt := range measuredTimers {
		writer.Write([]string{
			mt.name,
			getDuration(mt.totalCycles, mt.hits),
			getDuration(mt.cycles95, 1),
			getDuration(mt.cycles99, 1),
			fmt.Sprint(mt.hits),
		})
	}
	writer.Flush()
	return output.String()
}

func StartTimer(name string) *timer {
	tm, ok := staticTimers[name]
	if !ok {
		tm = &timer{}
		staticTimers[name] = tm
	}
	assert(tm.recursions >= 0)
	tm.recursions += 1
	if tm.recursions == 1 {
		tm.curCycles = -int64(gotsc.BenchStart())
	}
	return tm
}

func EndTimer(tm *timer) {
	assert(tm.recursions > 0)
	tm.recursions -= 1
	if tm.recursions == 0 {
		tm.cycles = append(tm.cycles, tm.curCycles+int64(gotsc.BenchStart()))
		tm.curCycles = 0
	}
}
