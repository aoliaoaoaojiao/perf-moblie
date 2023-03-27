package cmd

import (
	"context"
	"fmt"
	sasentity "github.com/SonicCloudOrg/sonic-android-supply/src/entity"
	"github.com/gin-gonic/gin"
	"github.com/go-echarts/go-echarts/v2/components"
	"net/http"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
)

var androidCmd = &cobra.Command{
	Use:   "android",
	Short: "Get android device performance",
	Long:  "Get android device performance",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		device := androidInit()
		addr = fmt.Sprintf("127.0.0.1:%d", port)
		r := gin.Default()
		r.Use(cors())
		r.StaticFS("/statics", http.Dir("./statics"))
		//r.StaticFS("/statics", http.Dir("./statics"))
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
			RegisterAndroidChart(&device, page, r, exitCtx)
			page.Render(c.Writer)
		})
		r.Run(fmt.Sprintf(addr))

		return nil
	},
}

var (
	addr               string
	pid                int
	androidSerial      string
	androidPackageName string
	androidOptions     sasentity.PerfOption
)

func init() {
	rootCmd.AddCommand(androidCmd)
	androidCmd.Flags().IntVar(&port, "port", 8081, "service port")
	androidCmd.Flags().StringVarP(&androidSerial, "serial", "s", "", "device serial (default first device)")
	androidCmd.Flags().IntVarP(&pid, "pid", "d", -1, "get PID data")
	androidCmd.Flags().StringVarP(&androidPackageName, "package", "p", "", "app package name")
	androidCmd.Flags().BoolVar(&androidOptions.SystemCPU, "sys-cpu", false, "get system cpu data")
	androidCmd.Flags().BoolVar(&androidOptions.SystemMem, "sys-mem", false, "get system memory data")

	androidCmd.Flags().BoolVar(&androidOptions.SystemNetWorking, "sys-network", false, "get system networking data")
	androidCmd.Flags().BoolVar(&androidOptions.ProcFPS, "proc-fps", false, "get fps data")
	androidCmd.Flags().BoolVar(&androidOptions.ProcThreads, "proc-threads", false, "get process threads")
	androidCmd.Flags().IntVarP(&androidOptions.RefreshTime, "refresh", "r", 1000, "data refresh time (millisecond)")
	androidCmd.Flags().BoolVar(&androidOptions.ProcCPU, "proc-cpu", false, "get process cpu data")
	androidCmd.Flags().BoolVar(&androidOptions.ProcMem, "proc-mem", false, "get process mem data")
}
