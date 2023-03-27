package entity

import giDevice "github.com/SonicCloudOrg/sonic-gidevice"

type IOSProcPerf struct {
	giDevice.PerfDataBase
	IOSProcPerf IOSProcData `json:"proc_perf"`
}

type IOSProcData struct {
	CPUUsage        *float64 `json:"cpuUsage"`
	MemAnon         *int     `json:"memAnon"`
	MemResidentSize *int     `json:"memResidentSize"`
	MemVirtualSize  *int     `json:"memVirtualSize"`
	PhysFootprint   *int     `json:"physFootprint"`
}
