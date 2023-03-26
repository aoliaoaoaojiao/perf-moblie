package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/SonicCloudOrg/sonic-android-supply/src/adb"
	sentity "github.com/SonicCloudOrg/sonic-android-supply/src/entity"
	sasp "github.com/SonicCloudOrg/sonic-android-supply/src/perfmonUtil"
	sasutil "github.com/SonicCloudOrg/sonic-android-supply/src/util"
	giDevice "github.com/SonicCloudOrg/sonic-gidevice"
	"github.com/gin-gonic/gin"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/templates"
	"html/template"
	"net/http"
	"os"
	"perf-moblie/entity"
	"time"
)

const (
	JsTpl = `
		$(function () { setInterval({{ .ViewID }}_sync, {{ .Interval }}); });
		function {{ .ViewID }}_sync() {
			$.ajax({
				type: "GET",
				url: "{{ .Addr }}",
				dataType: "json",
				success: function (result) {
					goecharts_{{ .ViewID }}.setOption(result);
				}
			});
		}
`
	BaseTpl = `
		{{- define "base" }}
		<div class="container">
			<div class="item" id="{{ .ChartID }}" style="width:{{ .Initialization.Width }};height:{{ .Initialization.Height }};"></div>
		</div>
		<script type="text/javascript">
			"use strict";
			let goecharts_{{ .ChartID | safeJS }} = echarts.init(document.getElementById('{{ .ChartID | safeJS }}'));
			let action_{{ .ChartID | safeJS }} = {{ .JSONNotEscapedAction | safeJS }};
			goecharts_{{ .ChartID | safeJS }}.dispatchAction(action_{{ .ChartID | safeJS }});
			{{- range .JSFunctions.Fns }}
			{{ . | safeJS }}
			{{- end }}
		</script>
		{{ end }}
`
)

func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", "*") // 可将将 * 替换为指定的域名
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}

func setPageInit(addr string, page *components.Page) {
	templates.BaseTpl = BaseTpl
	page.Initialization.AssetsHost = fmt.Sprintf("http://%s/statics/", addr)
	page.JSAssets.Add("jquery.min.js")
}

func androidInit() (d adb.Device) {
	var err error
	device := sasutil.GetDevice(androidSerial)

	if pid == -1 && androidPackageName != "" {
		sasp.PackageName = androidPackageName
		sasp.Pid, err = sasp.GetPidOnPackageName(device, androidPackageName)
		if err != nil {
			fmt.Println("no corresponding application PID found")
			os.Exit(0)
		}
	} else if pid != -1 && androidPackageName == "" {
		sasp.Pid = fmt.Sprintf("%d", pid)
		sasp.PackageName, err = sasp.GetNameOnPid(device, sasp.Pid)
		if err != nil {
			androidPackageName = ""
		}
	}

	if (sasp.Pid == "" && androidPackageName == "") &&
		!androidOptions.SystemCPU &&
		!androidOptions.SystemGPU &&
		!androidOptions.SystemNetWorking &&
		!androidOptions.SystemMem {
		androidParamsSet()
	}
	if (sasp.Pid != "" || androidPackageName != "") &&
		!androidOptions.ProcMem &&
		!androidOptions.ProcCPU &&
		!androidOptions.ProcThreads &&
		!androidOptions.ProcFPS {
		androidOptions.ProcMem = true
		androidOptions.ProcCPU = true
		androidOptions.ProcThreads = true
		androidOptions.ProcFPS = true
	}
	return *device
}

func iOSInit() (d giDevice.Device) {
	device := GetDeviceByUdId(iOSOptions.UDID)
	if device == nil {
		fmt.Println("device is not found")
		os.Exit(0)
	}

	if (pid == -1 || iOSOptions.BundleID == "") && !iOSOptions.SystemCPU && !iOSOptions.SystemMem && !iOSOptions.SystemDisk && !iOSOptions.SystemNetWorking && !iOSOptions.SystemGPU && !iOSOptions.SystemFPS {
		sysAllParamsSet()
	}

	if (pid != -1 || iOSOptions.BundleID != "") && !iOSOptions.SystemCPU && !iOSOptions.SystemMem && !iOSOptions.SystemDisk && !iOSOptions.SystemNetWorking && !iOSOptions.SystemGPU && !iOSOptions.SystemFPS && !iOSOptions.ProcNetwork && !iOSOptions.ProcMem && !iOSOptions.ProcCPU {
		sysAllParamsSet()
		iOSOptions.ProcNetwork = true
		iOSOptions.ProcMem = true
		iOSOptions.ProcCPU = true
	}

	if iOSOptions.ProcCPU {
		addCpuAttr()
	}

	if iOSOptions.ProcMem {
		addMemAttr()
	}

	perfOpts = []giDevice.PerfOption{
		giDevice.WithPerfSystemCPU(iOSOptions.SystemCPU),
		giDevice.WithPerfSystemMem(iOSOptions.SystemMem),
		giDevice.WithPerfSystemDisk(iOSOptions.SystemDisk),
		giDevice.WithPerfSystemNetwork(iOSOptions.SystemNetWorking),
		giDevice.WithPerfNetwork(iOSOptions.ProcNetwork),
		giDevice.WithPerfFPS(iOSOptions.SystemFPS),
		giDevice.WithPerfGPU(iOSOptions.SystemGPU),
		giDevice.WithPerfOutputInterval(iOSOptions.RefreshTime),
	}

	if pid != -1 {
		perfOpts = append(perfOpts, giDevice.WithPerfPID(pid))
		perfOpts = append(perfOpts, giDevice.WithPerfProcessAttributes(processAttributes...))
	} else if iOSOptions.BundleID != "" {
		perfOpts = append(perfOpts, giDevice.WithPerfBundleID(iOSOptions.BundleID))
		perfOpts = append(perfOpts, giDevice.WithPerfProcessAttributes(processAttributes...))
	}
	return device
}

func androidParamsSet() {
	androidOptions.SystemCPU = true
	androidOptions.SystemMem = true
	androidOptions.SystemGPU = true
	androidOptions.SystemNetWorking = true
}

func getLineTemplate(title string) *charts.Line {
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: title,
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Show:    true,
			Trigger: "axis",
			//AxisPointer:
		}),
		charts.WithLegendOpts(opts.Legend{
			Show: true,
			Top:  "5%",
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Theme: "shine",
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:  "inside",
			Start: 0,
			End:   100,
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:  "slider",
			Start: 0,
			End:   100,
		}),
		charts.WithGridOpts(opts.Grid{Top: "20%"}),
		charts.WithToolboxOpts(opts.Toolbox{
			Show: true,
			//Right: "20%",
			Feature: &opts.ToolBoxFeature{
				SaveAsImage: &opts.ToolBoxFeatureSaveAsImage{
					Show:  true,
					Type:  "png",
					Title: "Anything you want",
				},
				DataView: &opts.ToolBoxFeatureDataView{
					Show:  true,
					Title: "DataView",
					// set the language
					// Chinese version: ["数据视图", "关闭", "刷新"]
					Lang: []string{"data view", "turn off", "refresh"},
				},
				Restore: &opts.ToolBoxFeatureRestore{
					Show:  true,
					Title: "Restore",
				},
				DataZoom: &opts.ToolBoxFeatureDataZoom{
					Show:  true,
					Title: map[string]string{"zoom": "AreaZoom", "back": "RecoveryZoom"},
				},
			}}),
	)
	return line
}

func RegisterAndroidChart(device *adb.Device, page *components.Page, r *gin.Engine, exitCtx context.Context) {
	if androidOptions.SystemCPU {
		line, eData, dataChan := setAndroid("sys cpu info")
		sasp.GetSystemCPU(device, androidOptions, dataChan, exitCtx)
		r.GET(addr+"/"+line.ChartID, func(c *gin.Context) {
			androidDataConversion("sys cpu info", dataChan, eData, c)
		})
		page.AddCharts(&line)
	}
	if androidOptions.SystemMem {
		line, eData, dataChan := setAndroid("sys mem info")
		sasp.GetSystemMem(device, androidOptions, dataChan, exitCtx)
		r.GET(addr+"/"+line.ChartID, func(c *gin.Context) {
			androidDataConversion("sys mem info", dataChan, eData, c)
		})
		page.AddCharts(&line)
	}
	if androidOptions.SystemNetWorking {
		line, eData, dataChan := setAndroid("sys networking info")
		sasp.GetSystemNetwork(device, androidOptions, dataChan, exitCtx)
		r.GET(addr+"/"+line.ChartID, func(c *gin.Context) {
			androidDataConversion("sys networking info", dataChan, eData, c)
		})
		page.AddCharts(&line)
	}
	if androidOptions.ProcCPU {
		line, eData, dataChan := setAndroid("process cpu info")
		sasp.GetProcCpu(device, androidOptions, dataChan, exitCtx)
		r.GET(addr+"/"+line.ChartID, func(c *gin.Context) {
			androidDataConversion("process cpu info", dataChan, eData, c)
		})
		page.AddCharts(&line)
	}
	if androidOptions.ProcMem {
		line, eData, dataChan := setAndroid("process mem info")
		sasp.GetProcMem(device, androidOptions, dataChan, exitCtx)
		r.GET(addr+"/"+line.ChartID, func(c *gin.Context) {
			androidDataConversion("process mem info", dataChan, eData, c)
		})
		page.AddCharts(&line)
	}
	if androidOptions.ProcFPS {
		line, eData, dataChan := setAndroid("process FPS info")
		sasp.GetProcFPS(device, androidOptions, dataChan, exitCtx)
		r.GET(addr+"/"+line.ChartID, func(c *gin.Context) {
			androidDataConversion("process FPS info", dataChan, eData, c)
		})
		page.AddCharts(&line)
	}
	if androidOptions.ProcThreads {
		line, eData, dataChan := setAndroid("process Thread info")
		sasp.GetProcThreads(device, androidOptions, dataChan, exitCtx)
		r.GET(addr+"/"+line.ChartID, func(c *gin.Context) {
			androidDataConversion("process Thread info", dataChan, eData, c)
		})
		page.AddCharts(&line)
	}
}

func setAndroid(title string) (charts.Line, *entity.EchartsData, chan *sentity.PerfmonData) {
	line := getLineTemplate(title)
	line.AddJSFuncs(registerJs(androidOptions.RefreshTime, line.ChartID))
	dataChan := make(chan *sentity.PerfmonData)
	eData := &entity.EchartsData{
		XAxis:  []string{},
		Series: map[string][]opts.LineData{},
	}
	return *line, eData, dataChan
}

func androidDataConversion(title string, dataChan chan *sentity.PerfmonData, eData *entity.EchartsData, c *gin.Context) {
	data, _ := <-dataChan
	line := getLineTemplate(title)
	if data.System != nil {
		if data.System.CPU != nil {
			eData.XAxis = append(eData.XAxis, time.Unix(data.System.CPU["cpu"].TimeStamp/1000, 0).Format("2006-01-02 15:04:05"))
			for key, value := range data.System.CPU {
				lineData := eData.Series[key]
				if lineData != nil {
					lineData = append(lineData, opts.LineData{Value: int(value.Usage)})
					eData.Series[key] = lineData
				} else {
					lineData = []opts.LineData{}
					lineData = append(lineData, opts.LineData{Value: int(value.Usage)})
					eData.Series[key] = lineData
				}
			}
		}
		if data.System.MemInfo != nil {
			eData.XAxis = append(eData.XAxis, time.Unix(data.System.MemInfo.TimeStamp/1000, 0).Format("2006-01-02 15:04:05"))
			if eData.Series["memBuffers"] == nil {
				eData.Series["memBuffers"] = []opts.LineData{}
			}
			if eData.Series["memCached"] == nil {
				eData.Series["memCached"] = []opts.LineData{}
			}
			if eData.Series["memFree"] == nil {
				eData.Series["memFree"] = []opts.LineData{}
			}
			if eData.Series["memTotal"] == nil {
				eData.Series["memTotal"] = []opts.LineData{}
			}
			if eData.Series["memUsage"] == nil {
				eData.Series["memUsage"] = []opts.LineData{}
			}
			if eData.Series["swapTotal"] == nil {
				eData.Series["swapTotal"] = []opts.LineData{}
			}
			if eData.Series["swapFree"] == nil {
				eData.Series["swapFree"] = []opts.LineData{}
			}
			eData.Series["memBuffers"] = append(eData.Series["memBuffers"], opts.LineData{Value: data.System.MemInfo.MemTotal})
			eData.Series["memCached"] = append(eData.Series["memCached"], opts.LineData{Value: data.System.MemInfo.MemCached})
			eData.Series["memFree"] = append(eData.Series["memFree"], opts.LineData{Value: data.System.MemInfo.MemFree})
			eData.Series["memTotal"] = append(eData.Series["memTotal"], opts.LineData{Value: data.System.MemInfo.MemTotal})
			eData.Series["memUsage"] = append(eData.Series["memUsage"], opts.LineData{Value: data.System.MemInfo.MemUsage})
			eData.Series["swapTotal"] = append(eData.Series["swapTotal"], opts.LineData{Value: data.System.MemInfo.SwapTotal})
			eData.Series["swapFree"] = append(eData.Series["swapFree"], opts.LineData{Value: data.System.MemInfo.SwapFree})
		}
		if data.System.NetworkInfo != nil {
			var count = 0
			for key, value := range data.System.NetworkInfo {
				if count < 1 {
					eData.XAxis = append(eData.XAxis, time.Unix(value.TimeStamp/1000, 0).Format("2006-01-02 15:04:05"))
				}
				count++
				rxLineData := eData.Series[key+"_rx"]
				if rxLineData != nil {
					eData.Series[key+"_rx"] = append(eData.Series[key+"_rx"], opts.LineData{Value: value.Rx})
				} else {
					rxLineData = []opts.LineData{}
					rxLineData = append(rxLineData, opts.LineData{Value: value.Rx})
					eData.Series[key+"_rx"] = rxLineData
				}
				txLineData := eData.Series[key+"_tx"]
				if txLineData != nil {
					eData.Series[key+"_tx"] = append(txLineData, opts.LineData{Value: value.Tx})
				} else {
					txLineData = []opts.LineData{}
					txLineData = append(txLineData, opts.LineData{Value: value.Tx})
					eData.Series[key+"_tx"] = txLineData
				}
			}
		}
	} else {
		if data.Process.MemInfo != nil {
			eData.XAxis = append(eData.XAxis, time.Unix(data.Process.MemInfo.TimeStamp/1000, 0).Format("2006-01-02 15:04:05"))
			if eData.Series["pss"] == nil {
				eData.Series["pss"] = []opts.LineData{}
			}
			if eData.Series["vss"] == nil {
				eData.Series["vss"] = []opts.LineData{}
			}
			if eData.Series["rss"] == nil {
				eData.Series["rss"] = []opts.LineData{}
			}
			eData.Series["pss"] = append(eData.Series["pss"], opts.LineData{Value: data.Process.MemInfo.TotalPSS})
			eData.Series["vss"] = append(eData.Series["vss"], opts.LineData{Value: data.Process.MemInfo.VmSize})
			eData.Series["rss"] = append(eData.Series["rss"], opts.LineData{Value: data.Process.MemInfo.PhyRSS})
		}
		if data.Process.CPUInfo != nil {
			eData.XAxis = append(eData.XAxis, time.Unix(data.Process.CPUInfo.TimeStamp/1000, 0).Format("2006-01-02 15:04:05"))
			if eData.Series["cpuUtilization"] == nil {
				eData.Series["cpuUtilization"] = []opts.LineData{}
			}
			eData.Series["cpuUtilization"] = append(eData.Series["cpuUtilization"], opts.LineData{Value: data.Process.CPUInfo.CpuUtilization})
		}
		if data.Process.FPSInfo != nil {
			eData.XAxis = append(eData.XAxis, time.Unix(data.Process.FPSInfo.TimeStamp/1000, 0).Format("2006-01-02 15:04:05"))
			if eData.Series["FPS"] == nil {
				eData.Series["FPS"] = []opts.LineData{}
			}
			eData.Series["FPS"] = append(eData.Series["FPS"], opts.LineData{Value: data.Process.FPSInfo.FPS})
		}
		if data.Process.ThreadInfo != nil {
			eData.XAxis = append(eData.XAxis, time.Unix(data.Process.ThreadInfo.TimeStamp/1000, 0).Format("2006-01-02 15:04:05"))
			if eData.Series["threadCount"] == nil {
				eData.Series["threadCount"] = []opts.LineData{}
			}
			eData.Series["threadCount"] = append(eData.Series["threadCount"], opts.LineData{Value: data.Process.ThreadInfo.Threads})
		}
	}
	for key, value := range eData.Series {
		line = line.SetXAxis(eData.XAxis).AddSeries(key, value)
	}
	line.Validate()
	c.Writer.Write([]byte(line.JSONNotEscaped()))
}

func iOSDataSplit(title string, data map[string]interface{}) {

}

type iOSDataChan struct {
	SysChanCPU      chan map[string]interface{}
	SysChanMem      chan map[string]interface{}
	SysChanDisk     chan map[string]interface{}
	SysChanNetwork  chan map[string]interface{}
	ChanFPS         chan map[string]interface{}
	ChanGPU         chan map[string]interface{}
	ThreadsChan     chan map[string]interface{}
	ProcChanCPU     chan map[string]interface{}
	ProcChanMem     chan map[string]interface{}
	ProcChanNetwork chan map[string]interface{}
}

func registerJs(Interval int, chartID string) string {
	tpl, err := template.New("view").Parse(JsTpl)
	if err != nil {
		panic("statsview: failed to parse template " + err.Error())
	}
	chartDataAddr := addr + "/" + chartID
	var c = struct {
		Interval int
		Addr     string
		ViewID   string
	}{
		Interval: Interval,
		Addr:     chartDataAddr,
		ViewID:   chartID,
	}
	buf := bytes.Buffer{}
	if err := tpl.Execute(&buf, c); err != nil {
		panic("statsview: failed to execute template " + err.Error())
	}

	return buf.String()
}

func GetDeviceByUdId(udId string) (device giDevice.Device) {
	usbMuxClient, err := giDevice.NewUsbmux()
	if err != nil {
		panic(errors.New("unable to connect to usbmux"))
		return nil
	}
	list, err1 := usbMuxClient.Devices()
	if err1 != nil {
		panic(errors.New("unable to get device list"))
		return nil
	}
	if len(list) != 0 {
		if len(udId) != 0 {
			for i, d := range list {
				if d.Properties().SerialNumber == udId {
					device = list[i]
					break
				}
			}
		} else {
			device = list[0]
		}
		if device == nil || device.Properties().SerialNumber == "" {
			fmt.Println("device no found")
			return nil
		}
	} else {
		fmt.Println("no device connected")
		return nil
	}
	return
}
