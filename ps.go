package ps

import (
	"fmt"
	"strings"
	"time"

	"github.com/prometheus/procfs"
	"github.com/shirou/gopsutil/v4/process"
)

func valueOrZero[T int32 | uint64 | float32 | float64](v T, err error) T {
	if err != nil {
		return 0
	}
	return v
}

type Process struct {
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

func (s *Process) user() error {
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

func (s *Process) ppid() error {
	var err error
	s.Ppid, err = s.process.Ppid()
	return err
}

func (s *Process) status() error {
	status, err := s.process.Status()
	s.Status = strings.Join(status, ",")
	return err
}

func (s *Process) cmdLine() error {
	var err error
	s.CmdLine, err = s.process.Cmdline()
	return err
}

func (s *Process) cwd() error {
	var err error
	s.Cwd, err = s.process.Cwd()
	return err
}

func (s *Process) createTime() error {
	var err error
	s.CreateTime, err = s.process.CreateTime()
	return err
}

func (s *Process) nice() error {
	var err error
	s.Nice, err = s.process.Nice()
	return err
}

func (s *Process) fetch() error {
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

func (s *Process) RSSBytes() uint64 {
	if memory, err := s.process.MemoryInfo(); err != nil {
		return 0
	} else {
		return memory.RSS
	}
}

func (s *Process) CpuPercent(interval time.Duration) float64 {
	return valueOrZero(s.process.Percent(interval))
}

func (s *Process) MemoryPercent() float32 {
	return valueOrZero(s.process.MemoryPercent())
}

func (s *Process) NumFDs() int32 {
	return valueOrZero(s.process.NumFDs())
}

func (s *Process) NumThreads() int32 {
	return valueOrZero(s.process.NumThreads())
}

func (s *Process) NumContextSwitches() (voluntary, involuntary int64) {
	var cs *process.NumCtxSwitchesStat
	var err error
	cs, err = s.process.NumCtxSwitches()
	if err != nil {
		return
	}
	voluntary = cs.Voluntary
	involuntary = cs.Involuntary
	return
}

func (s *Process) NetTraffic() (in, out float64) {
	var p procfs.Proc
	var err error
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

func (s *Process) Process() *process.Process {
	return s.process
}

func All() []*Process {
	var all []*Process
	processes, err := process.Processes()
	if err != nil {
		return all
	}
	for _, i := range processes {
		cmdLine, _ := i.Cmdline()
		if cmdLine == "" {
			continue
		}
		p := &Process{Pid: i.Pid, process: i}
		if err = p.fetch(); err == nil {
			all = append(all, p)
		}
	}
	return all
}

func New(pid int32) (*Process, error) {
	p, err := process.NewProcess(pid)
	if err != nil {
		return nil, err
	}
	proc := &Process{Pid: pid, process: p}
	return proc, proc.fetch()
}
