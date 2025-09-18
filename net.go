package ps

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/riete/convert/str"

	"github.com/shirou/gopsutil/v4/net"
)

type NetTcpConnectionStat struct {
	Fd     uint32 `json:"fd"`
	Pid    int32  `json:"pid"`
	Node   string `json:"node"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type NetTcpConnectionStats []NetTcpConnectionStat

func (t NetTcpConnectionStats) ToString() string {
	b, _ := json.Marshal(t)
	return str.FromBytes(bytes.Replace(b, []byte(`\u003e`), []byte(">"), -1))
}

func NetTcpConnections() (NetTcpConnectionStats, error) {
	var tcpStat NetTcpConnectionStats
	connections, err := net.Connections("tcp")
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

type NetInterfaceStat struct {
	Name  string   `json:"name"`
	Addrs []string `json:"addrs"`
}

type NetInterfaceStats []NetInterfaceStat

func (i NetInterfaceStats) ToString() string {
	b, _ := json.Marshal(i)
	return str.FromBytes(b)
}

func NetInterfaces() (NetInterfaceStats, error) {
	var ifStat NetInterfaceStats
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, i := range interfaces {
		var addrs []string
		for _, addr := range i.Addrs {
			addrs = append(addrs, addr.Addr)
		}
		ifStat = append(ifStat, NetInterfaceStat{Name: i.Name, Addrs: addrs})
	}
	return ifStat, nil
}

type NetIOCountersStat struct {
	Name        string `json:"name"`         // interface name
	BytesSent   uint64 `json:"bytes_sent"`   // number of bytes sent
	BytesRecv   uint64 `json:"bytes_recv"`   // number of bytes received
	PacketsSent uint64 `json:"packets_sent"` // number of packets sent
	PacketsRecv uint64 `json:"packets_recv"` // number of packets received
	ErrIn       uint64 `json:"err_in"`       // total number of errors while receiving
	ErrOut      uint64 `json:"err_out"`      // total number of errors while sending
	DropIn      uint64 `json:"drop_in"`      // total number of incoming packets which were dropped
	DropOut     uint64 `json:"drop_out"`     // total number of outgoing packets which were dropped (always 0 on OSX and BSD)
	FifoIn      uint64 `json:"fifo_in"`      // total number of FIFO buffers errors while receiving
	FifoOut     uint64 `json:"fifo_out"`     // total number of FIFO buffers errors while sending
}

type IOCountersStats []NetIOCountersStat

func (i IOCountersStats) ToString() string {
	b, _ := json.Marshal(i)
	return str.FromBytes(b)
}

// NetIOCounters if inet is empty, return all interface data
func NetIOCounters(inet string) (IOCountersStats, error) {
	var s IOCountersStats
	stats, err := net.IOCounters(true)
	if err != nil {
		return nil, err
	}
	for _, i := range stats {
		s = append(
			s,
			NetIOCountersStat{
				Name:        i.Name,
				BytesSent:   i.BytesSent,
				BytesRecv:   i.BytesRecv,
				PacketsSent: i.PacketsSent,
				PacketsRecv: i.PacketsRecv,
				ErrIn:       i.Errin,
				ErrOut:      i.Errout,
				DropIn:      i.Dropin,
				DropOut:     i.Dropout,
				FifoIn:      i.Fifoin,
				FifoOut:     i.Fifoout,
			},
		)
	}
	if inet == "" {
		return s, nil
	}
	for _, i := range s {
		if i.Name == inet {
			return IOCountersStats{i}, nil
		}
	}
	return nil, fmt.Errorf("inet [%s] is not exists", inet)
}
