package ps

import (
	"encoding/json"

	"github.com/riete/convert/str"

	"github.com/shirou/gopsutil/v4/host"
)

type SystemInfo struct {
	Hostname        string `json:"hostname"`
	Uptime          uint64 `json:"uptime"`
	Procs           uint64 `json:"procs"`
	Os              string `json:"os"`
	Platform        string `json:"platform"`
	PlatformFamily  string `json:"platform_family"`
	PlatformVersion string `json:"platform_version"`
	KernelVersion   string `json:"kernel_version"`
	KernelArch      string `json:"kernel_arch"`
}

func (s SystemInfo) ToString() string {
	b, _ := json.Marshal(s)
	return str.FromBytes(b)
}

func Info() (*SystemInfo, error) {
	if info, err := host.Info(); err != nil {
		return nil, err
	} else {
		return &SystemInfo{
			Hostname:        info.Hostname,
			Uptime:          info.Uptime,
			Procs:           info.Procs,
			Os:              info.OS,
			Platform:        info.Platform,
			PlatformFamily:  info.PlatformFamily,
			PlatformVersion: info.PlatformVersion,
			KernelVersion:   info.KernelVersion,
			KernelArch:      info.KernelArch,
		}, nil
	}
}
