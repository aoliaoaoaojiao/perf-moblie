package cmd

import (
	"fmt"
	"github.com/SonicCloudOrg/sonic-android-supply/src/util"
	giDevice "github.com/SonicCloudOrg/sonic-gidevice"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"perf-moblie/entity"
)

var pefmonCmd = &cobra.Command{
	Use:   "perfmon",
	Short: "Get perfmon from your device.",
	Long:  "Get perfmon from your device.",
	RunE: func(cmd *cobra.Command, args []string) error {
		device := iOSInit()
		data, err := device.PerfStart(perfOpts...)

		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
		done := make(chan os.Signal, 1)
		signal.Notify(done, os.Interrupt, os.Kill)

		for {
			select {
			case <-done:
				device.PerfStop()
				fmt.Println("force end perfmon")
				os.Exit(0)
			case d := <-data:
				//p := &entity.PerfData{
				//	PerfDataBytes: d,
				//}
				//fmt.Println(util.Format(p, isFormat, isJson))
			}
		}
		return nil
	},
}

var (
	processAttributes []string
	iOSOptions        entity.Options
	perfOpts          []giDevice.PerfOption
)

func addCpuAttr() {
	processAttributes = append(processAttributes, "cpuUsage")
}

func addMemAttr() {
	processAttributes = append(processAttributes, "memVirtualSize", "physFootprint", "memResidentSize", "memAnon")
}

func sysAllParamsSet() {
	iOSOptions.SystemFPS = true
	iOSOptions.SystemGPU = true
	iOSOptions.SystemCPU = true
	iOSOptions.SystemMem = true
}

func init() {
	rootCmd.AddCommand(pefmonCmd)
	pefmonCmd.Flags().StringVarP(&iOSOptions.UDID, "udid", "u", "", "device's serialNumber ( default first device )")
	pefmonCmd.Flags().IntVarP(&iOSOptions.Pid, "pid", "p", -1, "get PID data")
	pefmonCmd.Flags().StringVarP(&iOSOptions.BundleID, "bundleId", "b", "", "target bundleId")
	pefmonCmd.Flags().BoolVar(&iOSOptions.SystemCPU, "sys-cpu", false, "get system cpu data")
	pefmonCmd.Flags().BoolVar(&iOSOptions.SystemMem, "sys-mem", false, "get system memory data")
	pefmonCmd.Flags().BoolVar(&iOSOptions.SystemDisk, "sys-disk", false, "get system disk data")
	pefmonCmd.Flags().BoolVar(&iOSOptions.SystemNetWorking, "sys-network", false, "get system networking data")
	pefmonCmd.Flags().BoolVar(&iOSOptions.SystemGPU, "gpu", false, "get gpu data")
	pefmonCmd.Flags().BoolVar(&iOSOptions.SystemFPS, "fps", false, "get fps data")
	pefmonCmd.Flags().BoolVar(&iOSOptions.ProcNetwork, "proc-network", false, "get process network data")
	pefmonCmd.Flags().BoolVar(&iOSOptions.ProcCPU, "proc-cpu", false, "get process cpu data")
	pefmonCmd.Flags().BoolVar(&iOSOptions.ProcMem, "proc-mem", false, "get process mem data")
	pefmonCmd.Flags().IntVarP(&iOSOptions.RefreshTime, "refresh", "r", 1000, "data refresh time(millisecond)")
}
