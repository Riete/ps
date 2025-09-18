package ps

import (
	"encoding/json"

	"github.com/riete/convert/str"

	"github.com/shirou/gopsutil/v4/mem"
)

type VirtualMemoryStat struct {
	Total        uint64  `json:"total"`
	Available    uint64  `json:"available"`
	Used         uint64  `json:"used"`
	UsedPercent  float64 `json:"used_percent"`
	Free         uint64  `json:"free"`
	Active       uint64  `json:"active"`
	Inactive     uint64  `json:"inactive"`
	Buffers      uint64  `json:"buffers"`
	Cached       uint64  `json:"cached"`
	Shared       uint64  `json:"shared"`
	Slab         uint64  `json:"slab"`
	SReclaimable uint64  `json:"s_reclaimable"`
}

func (v VirtualMemoryStat) ToString() string {
	b, _ := json.Marshal(v)
	return str.FromBytes(b)
}

func VirtualMemory() (*VirtualMemoryStat, error) {
	m, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	return &VirtualMemoryStat{
		Total:        m.Total,
		Available:    m.Available,
		Used:         m.Used,
		UsedPercent:  m.UsedPercent,
		Free:         m.Free,
		Active:       m.Active,
		Inactive:     m.Inactive,
		Buffers:      m.Buffers,
		Cached:       m.Cached,
		Shared:       m.Shared,
		Slab:         m.Slab,
		SReclaimable: m.Sreclaimable,
	}, nil
}

type SwapMemoryStat struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
	Sin         uint64  `json:"sin"`
	Sout        uint64  `json:"sout"`
	PgIn        uint64  `json:"pg_in"`
	PgOut       uint64  `json:"pg_out"`
	PgFault     uint64  `json:"pg_fault"`
}

func (s SwapMemoryStat) ToString() string {
	b, _ := json.Marshal(s)
	return str.FromBytes(b)
}

func SwapMemory() (*SwapMemoryStat, error) {
	m, err := mem.SwapMemory()
	if err != nil {
		return nil, err
	}
	return &SwapMemoryStat{
		Total:       m.Total,
		Used:        m.Used,
		Free:        m.Free,
		UsedPercent: m.UsedPercent,
		Sin:         m.Sin,
		Sout:        m.Sout,
		PgIn:        m.PgIn,
		PgOut:       m.PgOut,
		PgFault:     m.PgFault,
	}, nil
}
