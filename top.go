package ps

import (
	"cmp"
	"slices"
	"sync"
	"time"
)

type TopProcess struct {
	User       string  `json:"user"`
	Pid        int32   `json:"pid"`
	Ppid       int32   `json:"ppid"`
	Status     string  `json:"status"`
	CmdLine    string  `json:"cmd_line"`
	CpuPercent float64 `json:"cpu_percent"`
	RSSBytes   uint64  `json:"rss_bytes"`
	p          *Process
}

func (t *TopProcess) Process() *Process {
	return t.p
}

type SortFunc func(*TopProcess, *TopProcess) int

func SortByCpu(i, j *TopProcess) int {
	return -cmp.Compare(i.CpuPercent, j.CpuPercent)
}

func SortByMemory(i, j *TopProcess) int {
	return -cmp.Compare(i.RSSBytes, j.RSSBytes)
}

func TopNProcess(n int, f SortFunc) []*TopProcess {
	var top []*TopProcess
	for _, p := range All() {
		top = append(
			top,
			&TopProcess{
				User:     p.User,
				Pid:      p.Pid,
				Ppid:     p.Ppid,
				Status:   p.Status,
				CmdLine:  p.CmdLine,
				RSSBytes: p.RSSBytes(),
				p:        p,
			},
		)
	}
	wg := sync.WaitGroup{}
	wg.Add(len(top))
	for _, i := range top {
		go func(p *TopProcess) {
			p.CpuPercent = p.p.CpuPercent(time.Second)
			wg.Done()
		}(i)
	}
	wg.Wait()
	slices.SortStableFunc(top, f)
	if n >= len(top) {
		return top
	}
	return top[0:n]
}
