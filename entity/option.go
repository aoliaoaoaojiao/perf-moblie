package entity

type Options struct {
	IsServer         bool
	BundleID         string
	UDID             string
	Serial           string
	Pid              int
	PackageName      string
	SystemCPU        bool
	SystemMem        bool
	SystemGPU        bool
	SystemNetWorking bool
	ProcCPU          bool
	ProcFPS          bool
	ProcMem          bool
	ProcThreads      bool
	RefreshTime      int
}
