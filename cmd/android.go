package cmd

import (
	"bytes"
	"context"
	"fmt"
	sasentity "github.com/SonicCloudOrg/sonic-android-supply/src/entity"
	"github.com/gin-gonic/gin"
	"github.com/go-echarts/go-echarts/v2/components"
	"net/http"

	"github.com/spf13/cobra"
	"html/template"
)

var androidCmd = &cobra.Command{
	Use:   "android",
	Short: "Get android device performance",
	Long:  "Get android device performance",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		device, pidStr := androidInit()
		addr = "127.0.0.1:8081"
		r := gin.Default()
		r.Use(cors())
		r.StaticFS("/statics", http.Dir("./statics"))
		//r.StaticFS("/statics", http.Dir("./statics"))
		r.GET("/", func(c *gin.Context) {
			page := components.NewPage()
			setPageInit(addr, page)
			RegisterAndroidChart(&device, page, pidStr, r, context.TODO())
			page.Render(c.Writer)
		})
		r.Run(fmt.Sprintf(addr))

		//http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		//	page := components.NewPage()
		//	page.Initialization.AssetsHost = "http://127.0.0.1:8081/statics/"
		//	page.JSAssets.Add("jquery.min.js")
		//	line := charts.NewLine()
		//	line.AddJSFuncs(test(3*1000,line.ChartID))
		//	page.AddCharts(line)
		//	page.SetLayout(components.PageFlexLayout)
		//	page.Render(writer)
		//})
		////http.HandleFunc("/test", httpserver)
		//http.ListenAndServe(":8081", nil)
		return nil
	},
}

func test(Interval int, chartID string) string {
	tpl, err := template.New("view").Parse(JsTpl)
	if err != nil {
		panic("statsview: failed to parse template " + err.Error())
	}
	var c = struct {
		Interval int
		ViewID   string
	}{
		Interval: Interval,
		ViewID:   chartID,
	}

	buf := bytes.Buffer{}
	if err := tpl.Execute(&buf, c); err != nil {
		panic("statsview: failed to execute template " + err.Error())
	}

	return buf.String()
}

var (
	addr               string
	pid                int
	androidSerial      string
	androidPackageName string
	refreshTime        int
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
	androidCmd.Flags().IntVarP(&refreshTime, "refresh", "r", 1000, "data refresh time (millisecond)")
	androidCmd.Flags().BoolVar(&androidOptions.ProcCPU, "proc-cpu", false, "get process cpu data")
	androidCmd.Flags().BoolVar(&androidOptions.ProcMem, "proc-mem", false, "get process mem data")
}
