package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"perf-moblie/entity"
	"perf-moblie/util"

	"github.com/gin-gonic/gin"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/spf13/cobra"
)

var androidCmd = &cobra.Command{
	Use:   "android",
	Short: "Get android device performance",
	Long:  "Get android device performance",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		device := util.AndroidInit(&aOpts)
		aOpts.Addr = fmt.Sprintf("127.0.0.1:%d", port)
		r := gin.New()
		r.Use(util.Cors())
		r.StaticFS("/statics", http.Dir("./statics"))
		//r.StaticFS("/statics", http.Dir("./statics"))
		done := make(chan os.Signal, 1)
		signal.Notify(done, os.Interrupt, os.Kill)
		exitCtx, exitCancel := context.WithCancel(context.TODO())
		go func() {
			<-done
			exitCancel()
			os.Exit(0)
		}()
		page := components.NewPage()
		util.SetPageInit(aOpts.Addr, page)
		util.RegisterAndroidChart(&device, page, r, exitCtx)
		r.GET("/", func(c *gin.Context) {
			page.Render(c.Writer)
		})
		r.Run(fmt.Sprintf(aOpts.Addr))

		return nil
	},
}

var aOpts entity.AndroidOptions

func init() {
	rootCmd.AddCommand(androidCmd)
	androidCmd.Flags().IntVar(&port, "port", 8081, "service port")
	androidCmd.Flags().StringVarP(&aOpts.AndroidSerial, "serial", "s", "", "device serial (default first device)")
	androidCmd.Flags().IntVarP(&aOpts.Pid, "pid", "d", -1, "get PID data")
	androidCmd.Flags().StringVarP(&aOpts.AndroidPackageName, "package", "p", "", "app package name")
	androidCmd.Flags().BoolVar(&aOpts.AndroidOptions.SystemCPU, "sys-cpu", false, "get system cpu data")
	androidCmd.Flags().BoolVar(&aOpts.AndroidOptions.SystemMem, "sys-mem", false, "get system memory data")

	androidCmd.Flags().BoolVar(&aOpts.AndroidOptions.SystemNetWorking, "sys-network", false, "get system networking data")
	androidCmd.Flags().BoolVar(&aOpts.AndroidOptions.ProcFPS, "proc-fps", false, "get fps data")
	androidCmd.Flags().BoolVar(&aOpts.AndroidOptions.ProcThreads, "proc-threads", false, "get process threads")
	androidCmd.Flags().IntVarP(&aOpts.AndroidOptions.RefreshTime, "refresh", "r", 1000, "data refresh time (millisecond)")
	androidCmd.Flags().BoolVar(&aOpts.AndroidOptions.ProcCPU, "proc-cpu", false, "get process cpu data")
	androidCmd.Flags().BoolVar(&aOpts.AndroidOptions.ProcMem, "proc-mem", false, "get process mem data")
}
