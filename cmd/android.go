package cmd

import (
	"context"
	"fmt"
	sasentity "github.com/SonicCloudOrg/sonic-android-supply/src/entity"
	"github.com/gin-gonic/gin"
	"github.com/go-echarts/go-echarts/v2/components"
	"net/http"

	"github.com/spf13/cobra"
)

var androidCmd = &cobra.Command{
	Use:   "android",
	Short: "Get android device performance",
	Long:  "Get android device performance",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		device := androidInit()
		addr = "127.0.0.1:8081"
		r := gin.Default()
		r.Use(cors())
		r.StaticFS("/statics", http.Dir("./statics"))
		//r.StaticFS("/statics", http.Dir("./statics"))
		r.GET("/", func(c *gin.Context) {
			page := components.NewPage()
			setPageInit(addr, page)
			RegisterAndroidChart(&device, page, r, context.TODO())
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
	androidCmd.Flags().StringVarP(&androidSerial, "androidSerial", "s", "", "device androidSerial (default first device)")
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
