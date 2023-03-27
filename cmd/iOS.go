package cmd

import (
	"context"
	"fmt"
	giDevice "github.com/SonicCloudOrg/sonic-gidevice"
	"github.com/gin-gonic/gin"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"os/signal"
	"perf-moblie/entity"
)

var iOSCmd = &cobra.Command{
	Use:   "ios",
	Short: "Get iOS device performance",
	Long:  "Get iOS device performance",
	RunE: func(cmd *cobra.Command, args []string) error {
		device, iOSChanPerf := iOSInit()
		addr = fmt.Sprintf("127.0.0.1:%d", port)
		r := gin.Default()
		r.Use(cors())
		r.StaticFS("/statics", http.Dir("./statics"))
		//r.StaticFS("/statics", http.Dir("./statics"))
		data, err := device.PerfStart(perfOpts...)

		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
		done := make(chan os.Signal, 1)
		signal.Notify(done, os.Interrupt, os.Kill)
		exitCtx, exitCancel := context.WithCancel(context.TODO())
		go func() {
			<-done
			exitCancel()
		}()

		r.GET("/", func(c *gin.Context) {
			page := components.NewPage()
			setPageInit(addr, page)
			RegisterIOSChart(data, iOSChanPerf, page, r, exitCtx)
			page.Render(c.Writer)
		})
		r.Run(fmt.Sprintf(addr))

		return nil
	},
}

var (
	processAttributes []string
	iOSOptions        entity.Options
	perfOpts          []giDevice.PerfOption
	port              int
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
	rootCmd.AddCommand(iOSCmd)
	iOSCmd.Flags().IntVar(&port, "port", 8071, "service port")
	iOSCmd.Flags().StringVarP(&iOSOptions.UDID, "udid", "u", "", "device's serialNumber ( default first device )")
	iOSCmd.Flags().IntVarP(&iOSOptions.Pid, "pid", "p", -1, "get PID data")
	iOSCmd.Flags().StringVarP(&iOSOptions.BundleID, "bundleId", "b", "", "target bundleId")
	iOSCmd.Flags().BoolVar(&iOSOptions.SystemCPU, "sys-cpu", false, "get system cpu data")
	iOSCmd.Flags().BoolVar(&iOSOptions.SystemMem, "sys-mem", false, "get system memory data")
	iOSCmd.Flags().BoolVar(&iOSOptions.SystemDisk, "sys-disk", false, "get system disk data")
	iOSCmd.Flags().BoolVar(&iOSOptions.SystemNetWorking, "sys-network", false, "get system networking data")
	iOSCmd.Flags().BoolVar(&iOSOptions.SystemGPU, "gpu", false, "get gpu data")
	iOSCmd.Flags().BoolVar(&iOSOptions.SystemFPS, "fps", false, "get fps data")
	iOSCmd.Flags().BoolVar(&iOSOptions.ProcCPU, "proc-cpu", false, "get process cpu data")
	iOSCmd.Flags().BoolVar(&iOSOptions.ProcMem, "proc-mem", false, "get process mem data")
	iOSCmd.Flags().IntVarP(&iOSOptions.RefreshTime, "refresh", "r", 1000, "data refresh time(millisecond)")
}
