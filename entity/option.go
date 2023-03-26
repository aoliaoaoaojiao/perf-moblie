package entity

type Options struct {
	BundleID         string
	UDID             string
	Pid              int
	SystemCPU        bool
	SystemMem        bool
	SystemGPU        bool
	SystemFPS        bool
	SystemDisk       bool
	SystemNetWorking bool
	ProcCPU          bool
	ProcMem          bool
	ProcNetwork      bool
	ProcThreads      bool
	RefreshTime      int
}
