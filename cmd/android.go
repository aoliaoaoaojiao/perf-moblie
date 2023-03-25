package cmd

import (
	"fmt"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/spf13/cobra"
	"net/http"
	"perf-moblie/entity"
)

var androidCmd = &cobra.Command{
	Use:   "android",
	Short: "Get android device performance",
	Long:  "Get android device performance",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		generateLineItems()
		//go func() {
		//	tickerTest()
		//}()
		js, data := RegisterChartsWS(1000, line.ChartID)
		line.AddJSFuncs(js)
		go func() {
			tickerTest(data)
		}()
		RegisterJS()
		http.HandleFunc("/", httpserver)
		//http.HandleFunc("/test", httpserver)
		http.ListenAndServe(":8081", nil)
		return nil
	},
}

func httpserver(w http.ResponseWriter, r *http.Request) {
	setCors(w, r)
	page := components.NewPage()
	page.Initialization.AssetsHost = fmt.Sprintf("http://%s/statics/", r.Host)
	page.Assets.JSAssets.Add("jquery.min.js")

	//lineMulti()
	page.AddCharts(
		line,
		barBasic(),
		barTitle(),
		barSize(),
	)
	//line := lineDemo()
	//line.Render(w)

	page.Render(w)
}

var options entity.Options

func init() {
	rootCmd.AddCommand(androidCmd)
	androidCmd.Flags().StringVarP(&options.Serial, "serial", "s", "", "device serial (default first device)")
	androidCmd.Flags().IntVarP(&options.Pid, "pid", "d", -1, "get PID data")
	androidCmd.Flags().StringVarP(&options.PackageName, "package", "p", "", "app package name")
	androidCmd.Flags().BoolVar(&options.SystemCPU, "sys-cpu", false, "get system cpu data")
	androidCmd.Flags().BoolVar(&options.SystemMem, "sys-mem", false, "get system memory data")

	androidCmd.Flags().BoolVar(&options.SystemNetWorking, "sys-network", false, "get system networking data")
	androidCmd.Flags().BoolVar(&options.ProcFPS, "proc-fps", false, "get fps data")
	androidCmd.Flags().BoolVar(&options.ProcThreads, "proc-threads", false, "get process threads")

	androidCmd.Flags().BoolVar(&options.ProcCPU, "proc-cpu", false, "get process cpu data")
	androidCmd.Flags().BoolVar(&options.ProcMem, "proc-mem", false, "get process mem data")
}
