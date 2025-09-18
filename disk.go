package ps

import (
	"encoding/json"
	"errors"
	"path/filepath"

	"github.com/riete/convert/str"

	"github.com/shirou/gopsutil/v4/disk"
)

type DiskPartitionStat struct {
	Device     string   `json:"device"`
	Mountpoint string   `json:"mountpoint"`
	Fstype     string   `json:"fstype"`
	Opts       []string `json:"opts"`
}

type DiskPartitionStats []DiskPartitionStat

func (p DiskPartitionStats) ToString() string {
	b, _ := json.Marshal(p)
	return str.FromBytes(b)
}

// Partitions if all is false, only return local, physical partitions (hard disk, USB, CD/DVD partitions).
// if true, return all filesystems.
func Partitions(all bool) (DiskPartitionStats, error) {
	var stats DiskPartitionStats
	p, err := disk.Partitions(all)
	if err != nil {
		return nil, err
	}
	for _, i := range p {
		stats = append(
			stats,
			DiskPartitionStat{
				Device:     i.Device,
				Mountpoint: i.Mountpoint,
				Fstype:     i.Fstype,
				Opts:       i.Opts,
			},
		)
	}
	return stats, nil
}

type DiskUsageStat struct {
	Path              string  `json:"path"`
	Fstype            string  `json:"fstype"`
	Total             uint64  `json:"total"`
	Free              uint64  `json:"free"`
	Used              uint64  `json:"used"`
	UsedPercent       float64 `json:"used_percent"`
	InodesTotal       uint64  `json:"inodes_total"`
	InodesUsed        uint64  `json:"inodes_used"`
	InodesFree        uint64  `json:"inodes_free"`
	InodesUsedPercent float64 `json:"inodes_used_percent"`
}

func (u DiskUsageStat) ToString() string {
	b, _ := json.Marshal(u)
	return str.FromBytes(b)
}

type UsageStats []DiskUsageStat

func (u UsageStats) ToString() string {
	b, _ := json.Marshal(u)
	return str.FromBytes(b)
}

// DiskUsage if all is false, only return local, physical partitions (hard disk, USB, CD/DVD partitions).
// if true, return all mountpoints.
func DiskUsage(all bool) (UsageStats, error) {
	partitions, err := Partitions(all)
	if err != nil {
		return nil, err
	}

	var stats UsageStats
	for _, p := range partitions {
		u, err := disk.Usage(p.Mountpoint)
		if err != nil {
			return nil, err
		}
		stats = append(
			stats,
			DiskUsageStat{
				Path:              u.Path,
				Fstype:            u.Fstype,
				Total:             u.Total,
				Free:              u.Free,
				Used:              u.Used,
				UsedPercent:       u.UsedPercent,
				InodesTotal:       u.InodesTotal,
				InodesUsed:        u.InodesUsed,
				InodesFree:        u.InodesFree,
				InodesUsedPercent: u.InodesUsedPercent,
			},
		)
	}
	return stats, nil
}

// UsageByMountpoint mountpoint is filesystem path, such as "/", not device path like "/dev/vda"
func UsageByMountpoint(mountpoint string) (*DiskUsageStat, error) {
	u, err := disk.Usage(mountpoint)
	if err != nil {
		return nil, err
	}
	return &DiskUsageStat{
		Path:              u.Path,
		Fstype:            u.Fstype,
		Total:             u.Total,
		Free:              u.Free,
		Used:              u.Used,
		UsedPercent:       u.UsedPercent,
		InodesTotal:       u.InodesTotal,
		InodesUsed:        u.InodesUsed,
		InodesFree:        u.InodesFree,
		InodesUsedPercent: u.InodesUsedPercent,
	}, nil
}

type DiskIOCountersStat struct {
	ReadCount        uint64 `json:"read_count"`
	MergedReadCount  uint64 `json:"merged_read_count"`
	WriteCount       uint64 `json:"write_count"`
	MergedWriteCount uint64 `json:"merged_write_count"`
	ReadBytes        uint64 `json:"read_bytes"`
	WriteBytes       uint64 `json:"write_bytes"`
	ReadTime         uint64 `json:"read_time"`
	WriteTime        uint64 `json:"write_time"`
	IopsInProgress   uint64 `json:"iops_in_progress"`
	IoTime           uint64 `json:"io_time"`
	WeightedIO       uint64 `json:"weighted_io"`
	Name             string `json:"name"`
	SerialNumber     string `json:"serial_number"`
	Label            string `json:"label"`
}

func (i DiskIOCountersStat) ToString() string {
	b, _ := json.Marshal(i)
	return str.FromBytes(b)
}

// IOCounters device is suck like vda, vda1 ..etc
func DiskIOCounters(device string) (*DiskIOCountersStat, error) {
	device = filepath.Base(device)
	u, err := disk.IOCounters(device)
	if err != nil {
		return nil, err
	}
	if s, ok := u[device]; !ok {
		return nil, errors.New("device is not exists")
	} else {
		return &DiskIOCountersStat{
			ReadCount:        s.ReadCount,
			MergedReadCount:  s.MergedReadCount,
			WriteCount:       s.WriteCount,
			MergedWriteCount: s.MergedWriteCount,
			ReadBytes:        s.ReadBytes,
			WriteBytes:       s.WriteBytes,
			ReadTime:         s.ReadTime,
			WriteTime:        s.WriteTime,
			IopsInProgress:   s.IopsInProgress,
			IoTime:           s.IoTime,
			WeightedIO:       s.WeightedIO,
			Name:             s.Name,
			SerialNumber:     s.SerialNumber,
			Label:            s.Label,
		}, nil
	}
}
