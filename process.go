package ps

import (
	"encoding/json"
	"fmt"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/riete/convert/str"

	"github.com/shirou/gopsutil/v3/net"

	"github.com/shirou/gopsutil/v3/process"
)

type ProcessStat struct {
	User                        string  `json:"user"`
	Pid                         int32   `json:"pid"`
	Ppid                        int32   `json:"ppid"`
	Status                      string  `json:"status"`
	CmdLine                     string  `json:"cmd_line"`
	Cwd                         string  `json:"cwd"`
	CreateTime                  int64   `json:"create_time"`
	RssBytes                    uint64  `json:"rss_bytes"`
	CpuPercent                  float64 `json:"cpu_percent"`
	Nice                        int32   `json:"nice"`
	MemoryPercent               float32 `json:"memory_percent"`
	Fds                         int32   `json:"fds"`
	Threads                     int32   `json:"threads"`
	VoluntaryContextSwitches    int64   `json:"voluntary_context_switches"`
	NonVoluntaryContextSwitches int64   `json:"non_voluntary_context_switches"`
	process                     *process.Process
}

func (p ProcessStat) ToString() string {
	b, _ := json.Marshal(p)
	return str.FromBytes(b)
}

func (p *ProcessStat) user() error {
	var err error
	p.User, err = p.process.Username()
	if err != nil {
		uids, _ := p.process.Uids()
		if len(uids) > 0 {
			p.User = fmt.Sprintf("%d", uids[0])
		}
	}
	return nil
}

func (p *ProcessStat) ppid() error {
	var err error
	p.Ppid, err = p.process.Ppid()
	return err
}

func (p *ProcessStat) status() error {
	status, err := p.process.Status()
	p.Status = strings.Join(status, ",")
	return err
}

func (p *ProcessStat) cmdLine() error {
	var err error
	p.CmdLine, err = p.process.Cmdline()
	return err
}

func (p *ProcessStat) cwd() error {
	var err error
	p.Cwd, err = p.process.Cwd()
	return err
}

func (p *ProcessStat) createTime() error {
	var err error
	p.CreateTime, err = p.process.CreateTime()
	return err
}

func (p *ProcessStat) rss() error {
	memory, err := p.process.MemoryInfo()
	if err != nil {
		return err
	}
	p.RssBytes = memory.RSS
	return nil
}

func (p *ProcessStat) CpuPercentCurrent(interval time.Duration) (float64, error) {
	percent, err := p.process.Percent(interval)
	return percent, err
}

func (p *ProcessStat) cpuPercent() error {
	var err error
	p.CpuPercent, err = p.process.CPUPercent()
	return err
}

func (p *ProcessStat) nice() error {
	var err error
	p.Nice, err = p.process.Nice()
	return err
}

func (p *ProcessStat) memoryPercent() error {
	var err error
	p.MemoryPercent, err = p.process.MemoryPercent()
	return err
}

func (p *ProcessStat) numFds() error {
	var err error
	p.Fds, err = p.process.NumFDs()
	return err
}

func (p *ProcessStat) numThreads() error {
	var err error
	p.Threads, err = p.process.NumThreads()
	return err
}

func (p *ProcessStat) numContextSwitches() error {
	cs, err := p.process.NumCtxSwitches()
	if err != nil {
		return err
	}
	p.VoluntaryContextSwitches = cs.Voluntary
	p.NonVoluntaryContextSwitches = cs.Involuntary
	return nil
}

func (p *ProcessStat) Fill() error {
	hf := func(f func() error, err error) error {
		if err != nil {
			return err
		}
		return f()
	}
	err := p.ppid()
	err = hf(p.user, err)
	err = hf(p.status, err)
	err = hf(p.cmdLine, err)
	err = hf(p.cwd, err)
	err = hf(p.createTime, err)
	err = hf(p.rss, err)
	err = hf(p.cpuPercent, err)
	err = hf(p.nice, err)
	err = hf(p.memoryPercent, err)
	if runtime.GOOS != "darwin" {
		err = hf(p.numFds, err)
		err = hf(p.numThreads, err)
		err = hf(p.numContextSwitches, err)
	}
	return err
}

type ProcessStats []*ProcessStat

func (p ProcessStats) ToString() string {
	b, _ := json.Marshal(p)
	return str.FromBytes(b)
}

func Processes() (ProcessStats, error) {
	p, err := process.Processes()
	if err != nil {
		return nil, err
	}
	var stats ProcessStats
	for _, i := range p {
		cmdLine, _ := i.Cmdline()
		if cmdLine == "" {
			continue
		}
		stat := &ProcessStat{Pid: i.Pid, process: i}
		if err = stat.Fill(); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}
	return stats, nil
}

func Process(pid int32) (*ProcessStat, error) {
	p, err := process.NewProcess(pid)
	if err != nil {
		return nil, err
	}
	stat := &ProcessStat{Pid: pid, process: p}
	if err = stat.Fill(); err != nil {
		return nil, err
	}
	return stat, nil
}

func ProcessKill(pid int32) error {
	p, err := process.NewProcess(pid)
	if err != nil {
		return err
	}
	return p.Kill()
}

func ProcessTerminate(pid int32) error {
	p, err := process.NewProcess(pid)
	if err != nil {
		return err
	}
	return p.Terminate()
}

func ProcessNetConnections(pid int32) (NetTcpConnectionStats, error) {
	var tcpStat NetTcpConnectionStats
	connections, err := net.ConnectionsPid("tcp", pid)
	if err != nil {
		return nil, err
	}
	for _, c := range connections {
		name := fmt.Sprintf("%s:%d->%s:%d", c.Laddr.IP, c.Laddr.Port, c.Raddr.IP, c.Raddr.Port)
		if c.Status == "LISTEN" {
			name = fmt.Sprintf("%s:%d", c.Laddr.IP, c.Laddr.Port)
		}
		tcpStat = append(tcpStat, NetTcpConnectionStat{Fd: c.Fd, Node: "TCP", Name: name, Status: c.Status, Pid: c.Pid})
	}
	return tcpStat, nil
}

type TopProcess struct {
	User              string  `json:"user"`
	Pid               int32   `json:"pid"`
	Ppid              int32   `json:"ppid"`
	Status            string  `json:"status"`
	CmdLine           string  `json:"cmd_line"`
	CpuPercentCurrent float64 `json:"cpu_percent_current"`
	RssBytes          uint64  `json:"rss_bytes"`
	process           *process.Process
}

func (t *TopProcess) cpuPercentCurrent(interval time.Duration) {
	t.CpuPercentCurrent, _ = t.process.Percent(interval)
}

type CpuTopProcesses []*TopProcess

func (c CpuTopProcesses) ToString() string {
	b, _ := json.Marshal(c)
	return str.FromBytes(b)
}

func (c CpuTopProcesses) Len() int {
	return len(c)
}

func (c CpuTopProcesses) Less(i, j int) bool {
	return c[i].CpuPercentCurrent > c[j].CpuPercentCurrent
}

func (c CpuTopProcesses) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

type MemoryTopProcesses []*TopProcess

func (m MemoryTopProcesses) ToString() string {
	b, _ := json.Marshal(m)
	return str.FromBytes(b)
}

func (m MemoryTopProcesses) Len() int {
	return len(m)
}

func (m MemoryTopProcesses) Less(i, j int) bool {
	return m[i].RssBytes > m[j].RssBytes
}

func (m MemoryTopProcesses) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func MemoryTopN(n int) (MemoryTopProcesses, error) {
	p, err := Processes()
	if err != nil {
		return nil, err
	}
	var top MemoryTopProcesses
	for _, i := range p {
		top = append(
			top,
			&TopProcess{
				User:     i.User,
				Pid:      i.Pid,
				Ppid:     i.Ppid,
				Status:   i.Status,
				CmdLine:  i.CmdLine,
				RssBytes: i.RssBytes,
				process:  i.process,
			},
		)
	}
	wg := sync.WaitGroup{}
	wg.Add(len(top))
	for _, i := range top {
		go func(p *TopProcess) {
			p.cpuPercentCurrent(time.Second)
			wg.Done()
		}(i)
	}
	wg.Wait()
	sort.Sort(top)
	if n >= len(top) {
		return top, nil
	}
	return top[0:n], nil
}

func CpuTopN(interval time.Duration, n int) (CpuTopProcesses, error) {
	p, err := Processes()
	if err != nil {
		return nil, err
	}
	var top CpuTopProcesses
	for _, i := range p {
		top = append(
			top,
			&TopProcess{
				User:     i.User,
				Pid:      i.Pid,
				Ppid:     i.Ppid,
				Status:   i.Status,
				CmdLine:  i.CmdLine,
				RssBytes: i.RssBytes,
				process:  i.process,
			},
		)
	}
	wg := sync.WaitGroup{}
	wg.Add(len(top))
	for _, i := range top {
		go func(p *TopProcess) {
			p.cpuPercentCurrent(interval)
			wg.Done()
		}(i)
	}
	wg.Wait()
	sort.Sort(top)
	if n > len(top) {
		return top, nil
	}
	return top[0:n], nil
}
