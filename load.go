package ps

import (
	"encoding/json"

	"github.com/riete/convert/str"

	"github.com/shirou/gopsutil/v3/load"
)

type LoadStat struct {
	Load1  float64 `json:"load_1"`
	Load5  float64 `json:"load_5"`
	Load15 float64 `json:"load_15"`
}

func (l LoadStat) ToString() string {
	b, _ := json.Marshal(l)
	return str.FromBytes(b)
}

func Load() (*LoadStat, error) {
	l, err := load.Avg()
	if err != nil {
		return nil, err
	}
	return &LoadStat{Load1: l.Load1, Load5: l.Load5, Load15: l.Load15}, nil
}
