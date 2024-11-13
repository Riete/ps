package ps

import (
	"cmp"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/procfs"
	"github.com/riete/convert/str"

	"github.com/shirou/gopsutil/v3/net"

	"github.com/shirou/gopsutil/v3/process"
)

type ProcessStat struct {
	User       string `json:"user"`
	Pid        int32  `json:"pid"`
	Ppid       int32  `json:"ppid"`
	Status     string `json:"status"`
	CmdLine    string `json:"cmd_line"`
	Cwd        string `json:"cwd"`
	CreateTime int64  `json:"create_time"`
	Nice       int32  `json:"nice"`
	process    *process.Process
}

func (s ProcessStat) ToString() string {
	b, _ := json.Marshal(s)
	return str.FromBytes(b)
}

func (s *ProcessStat) user() error {
	var err error
	s.User, err = s.process.Username()
	if err != nil {
		uids, _ := s.process.Uids()
		if len(uids) > 0 {
			s.User = fmt.Sprintf("%d", uids[0])
		}
	}
	return nil
}

func (s *ProcessStat) ppid() error {
	var err error
	s.Ppid, err = s.process.Ppid()
	return err
}

func (s *ProcessStat) status() error {
	status, err := s.process.Status()
	s.Status = strings.Join(status, ",")
	return err
}

func (s *ProcessStat) cmdLine() error {
	var err error
	s.CmdLine, err = s.process.Cmdline()
	return err
}

func (s *ProcessStat) cwd() error {
	var err error
	s.Cwd, err = s.process.Cwd()
	return err
}

func (s *ProcessStat) createTime() error {
	var err error
	s.CreateTime, err = s.process.CreateTime()
	return err
}

func (s *ProcessStat) nice() error {
	var err error
	s.Nice, err = s.process.Nice()
	return err
}

func (s *ProcessStat) RssBytes() (uint64, error) {
	memory, err := s.process.MemoryInfo()
	if err != nil {
		return 0, err
	}
	return memory.RSS, nil
}

func (s *ProcessStat) CpuPercent(interval time.Duration) (float64, error) {
	percent, err := s.process.Percent(interval)
	return percent, err
}

func (s *ProcessStat) MemoryPercent() (float32, error) {
	return s.process.MemoryPercent()
}

func (s *ProcessStat) NumFDs() (int32, error) {
	return s.process.NumFDs()
}

func (s *ProcessStat) NumThreads() (int32, error) {
	return s.process.NumThreads()
}

func (s *ProcessStat) NumContextSwitches() (voluntary, involuntary int64, err error) {
	var cs *process.NumCtxSwitchesStat
	cs, err = s.process.NumCtxSwitches()
	if err != nil {
		return
	}
	voluntary = cs.Voluntary
	involuntary = cs.Involuntary
	return
}

func (s *ProcessStat) NetTraffic() (in, out float64, err error) {
	var p procfs.Proc
	p, err = procfs.NewProc(int(s.Pid))
	if err != nil {
		return
	}
	var netstat procfs.ProcNetstat
	if netstat, err = p.Netstat(); err == nil {
		if netstat.IpExt.InOctets != nil {
			in = *netstat.IpExt.InOctets
		}
		if netstat.IpExt.OutOctets != nil {
			out = *netstat.IpExt.OutOctets
		}
	}
	return
}

func (s *ProcessStat) Fill() error {
	hf := func(f func() error, err error) error {
		if err != nil {
			return err
		}
		return f()
	}
	err := s.ppid()
	err = hf(s.user, err)
	err = hf(s.status, err)
	err = hf(s.cmdLine, err)
	err = hf(s.cwd, err)
	err = hf(s.createTime, err)
	err = hf(s.nice, err)
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
	User       string  `json:"user"`
	Pid        int32   `json:"pid"`
	Ppid       int32   `json:"ppid"`
	Status     string  `json:"status"`
	CmdLine    string  `json:"cmd_line"`
	CpuPercent float64 `json:"cpu_percent"`
	RssBytes   uint64  `json:"rss_bytes"`
	s          *ProcessStat
}

func (t TopProcess) ToString() string {
	b, _ := json.Marshal(t)
	return str.FromBytes(b)
}

type TopProcesses []*TopProcess

func (t TopProcesses) ToString() string {
	b, _ := json.Marshal(t)
	return str.FromBytes(b)
}

type SortFunc func(*TopProcess, *TopProcess) int

func SortByCpu(i, j *TopProcess) int {
	return -cmp.Compare(i.CpuPercent, j.CpuPercent)
}

func SortByMemory(i, j *TopProcess) int {
	return -cmp.Compare(i.RssBytes, j.RssBytes)
}

func TopNProcess(n int, f SortFunc) (TopProcesses, error) {
	p, err := Processes()
	if err != nil {
		return nil, err
	}
	var top TopProcesses
	for _, s := range p {
		rssBytes, _ := s.RssBytes()
		top = append(
			top,
			&TopProcess{
				User:     s.User,
				Pid:      s.Pid,
				Ppid:     s.Ppid,
				Status:   s.Status,
				CmdLine:  s.CmdLine,
				RssBytes: rssBytes,
				s:        s,
			},
		)
	}
	wg := sync.WaitGroup{}
	wg.Add(len(top))
	for _, i := range top {
		go func(p *TopProcess) {
			p.CpuPercent, _ = p.s.CpuPercent(time.Second)
			wg.Done()
		}(i)
	}
	wg.Wait()
	slices.SortStableFunc(top, f)
	if n >= len(top) {
		return top, nil
	}
	return top[0:n], nil
}
