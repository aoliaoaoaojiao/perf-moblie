package entity

import sentity "github.com/SonicCloudOrg/sonic-android-supply/src/entity"

type IOSOptions struct {
	BundleID          string
	UDID              string
	Pid               int
	SystemCPU         bool
	SystemMem         bool
	SystemGPU         bool
	SystemFPS         bool
	SystemDisk        bool
	SystemNetWorking  bool
	ProcCPU           bool
	ProcMem           bool
	ProcNetwork       bool
	ProcThreads       bool
	RefreshTime       int
	ProcessAttributes []string
}

type AndroidOptions struct {
	Addr               string
	Pid                int
	AndroidSerial      string
	AndroidPackageName string
	AndroidOptions     sentity.PerfOption
}
