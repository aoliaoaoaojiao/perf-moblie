package cmd

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"os/signal"
	"perf-moblie/entity"
	"perf-moblie/util"
)

var iOSCmd = &cobra.Command{
	Use:   "ios",
	Short: "Get iOS device performance",
	Long:  "Get iOS device performance",
	RunE: func(cmd *cobra.Command, args []string) error {
		device, iOSChanPerf, perfOpts := util.IOSInit(&iOpts)
		iOpts.Addr = fmt.Sprintf("127.0.0.1:%d", port)
		r := gin.Default()
		r.Use(util.Cors())
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
			os.Exit(0)
		}()
		page := components.NewPage()
		util.SetPageInit(iOpts.Addr, page)
		util.RegisterIOSChart(data, iOSChanPerf, page, r, exitCtx)
		r.GET("/", func(c *gin.Context) {
			page.Render(c.Writer)
		})
		r.Run(fmt.Sprintf(iOpts.Addr))

		return nil
	},
}

var iOpts entity.IOSOptions

func init() {
	rootCmd.AddCommand(iOSCmd)
	iOSCmd.Flags().IntVar(&port, "port", 8081, "service port")
	iOSCmd.Flags().StringVarP(&iOpts.UDID, "udid", "u", "", "device's serialNumber ( default first device )")
	iOSCmd.Flags().IntVarP(&iOpts.Pid, "pid", "p", -1, "get PID data")
	iOSCmd.Flags().StringVarP(&iOpts.BundleID, "bundleId", "b", "", "target bundleId")
	iOSCmd.Flags().BoolVar(&iOpts.SystemCPU, "sys-cpu", false, "get system cpu data")
	iOSCmd.Flags().BoolVar(&iOpts.SystemMem, "sys-mem", false, "get system memory data")
	iOSCmd.Flags().BoolVar(&iOpts.SystemDisk, "sys-disk", false, "get system disk data")
	iOSCmd.Flags().BoolVar(&iOpts.SystemNetWorking, "sys-network", false, "get system networking data")
	iOSCmd.Flags().BoolVar(&iOpts.SystemGPU, "gpu", false, "get gpu data")
	iOSCmd.Flags().BoolVar(&iOpts.SystemFPS, "fps", false, "get fps data")
	iOSCmd.Flags().BoolVar(&iOpts.ProcCPU, "proc-cpu", false, "get process cpu data")
	iOSCmd.Flags().BoolVar(&iOpts.ProcMem, "proc-mem", false, "get process mem data")
	iOSCmd.Flags().IntVarP(&iOpts.RefreshTime, "refresh", "r", 1000, "data refresh time(millisecond)")
}
