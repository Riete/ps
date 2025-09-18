package ps

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/riete/convert/str"

	"github.com/shirou/gopsutil/v4/cpu"
)

type CpuPercentStat struct {
	Cpu     string  `json:"cpu"`
	Percent float64 `json:"percent"`
}

type CpuPercentStats []CpuPercentStat

func (c CpuPercentStats) ToString() string {
	b, _ := json.Marshal(c)
	return str.FromBytes(b)
}

func CpuPercent() (CpuPercentStats, error) {
	var cpuStat CpuPercentStats
	perCpu, err := cpu.Percent(time.Second, true)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(perCpu); i++ {
		cpuStat = append(cpuStat, CpuPercentStat{Cpu: fmt.Sprintf("cpu-%d", i), Percent: perCpu[i]})
	}

	total, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, err
	}
	cpuStat = append(cpuStat, CpuPercentStat{Cpu: "cpu-total", Percent: total[0]})
	return cpuStat, nil
}

type CpuTimeStat struct {
	Cpu     string  `json:"cpu"`
	User    float64 `json:"user"`
	System  float64 `json:"system"`
	Idle    float64 `json:"idle"`
	Nice    float64 `json:"nice"`
	Iowait  float64 `json:"iowait"`
	Irq     float64 `json:"irq"`
	Softirq float64 `json:"softirq"`
	Steal   float64 `json:"steal"`
}

type CpuTimeStats []CpuTimeStat

func (c CpuTimeStats) ToString() string {
	b, _ := json.Marshal(c)
	return str.FromBytes(b)
}

func CpuTime() (CpuTimeStats, error) {
	var cpuStat CpuTimeStats
	perCpu, err := cpu.Times(true)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(perCpu); i++ {
		cpuStat = append(
			cpuStat,
			CpuTimeStat{
				Cpu:     fmt.Sprintf("cpu-%d", i),
				User:    perCpu[i].User,
				System:  perCpu[i].System,
				Idle:    perCpu[i].Idle,
				Nice:    perCpu[i].Nice,
				Iowait:  perCpu[i].Iowait,
				Irq:     perCpu[i].Irq,
				Softirq: perCpu[i].Softirq,
				Steal:   perCpu[i].Steal,
			},
		)
	}

	total, err := cpu.Times(false)
	if err != nil {
		return nil, err
	}
	cpuStat = append(
		cpuStat,
		CpuTimeStat{
			Cpu:     "cpu-total",
			User:    total[0].User,
			System:  total[0].System,
			Idle:    total[0].Idle,
			Nice:    total[0].Nice,
			Iowait:  total[0].Iowait,
			Irq:     total[0].Irq,
			Softirq: total[0].Softirq,
			Steal:   total[0].Steal,
		},
	)
	return cpuStat, nil
}
