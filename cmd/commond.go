package cmd

import (
	"bytes"
	"context"
	"fmt"
	"github.com/SonicCloudOrg/sonic-android-supply/src/adb"
	sentity "github.com/SonicCloudOrg/sonic-android-supply/src/entity"
	sasp "github.com/SonicCloudOrg/sonic-android-supply/src/perfmonUtil"
	sasutil "github.com/SonicCloudOrg/sonic-android-supply/src/util"
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

func androidInit() (d adb.Device, pidStr string) {
	var err error
	device := sasutil.GetDevice(androidSerial)

	if pid == -1 && androidPackageName != "" {
		pidStr, err = sasp.GetPidOnPackageName(device, androidPackageName)
		if err != nil {
			fmt.Println("no corresponding application PID found")
			os.Exit(0)
		}
	} else if pid != -1 && androidPackageName == "" {
		pidStr = fmt.Sprintf("%d", pid)
		androidPackageName, err = sasp.GetNameOnPid(device, pidStr)
		if err != nil {
			androidPackageName = ""
		}
	}

	if (pidStr == "" && androidPackageName == "") &&
		!androidOptions.SystemCPU &&
		!androidOptions.SystemGPU &&
		!androidOptions.SystemNetWorking &&
		!androidOptions.SystemMem {
		androidParamsSet()
	}
	if (pidStr != "" || androidPackageName != "") &&
		!androidOptions.ProcMem &&
		!androidOptions.ProcCPU &&
		!androidOptions.ProcThreads &&
		!androidOptions.ProcFPS {
		androidOptions.ProcMem = true
		androidOptions.ProcCPU = true
		androidOptions.ProcThreads = true
		androidOptions.ProcFPS = true
	}
	return *device, pidStr
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

func RegisterAndroidChart(device *adb.Device, page *components.Page, pidStr string, r *gin.Engine, exitCtx context.Context) {
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
}

func setAndroid(title string) (charts.Line, *entity.EchartsData, chan *sentity.PerfmonData) {
	line := getLineTemplate(title)
	line.AddJSFuncs(registerJs(refreshTime, line.ChartID))
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

		}
		if data.Process.CPUInfo != nil {

		}
		if data.Process.FPSInfo != nil {

		}
		if data.Process.ThreadInfo != nil {

		}
	}
	for key, value := range eData.Series {
		line = line.SetXAxis(eData.XAxis).AddSeries(key, value)
	}
	line.Validate()
	c.Writer.Write([]byte(line.JSONNotEscaped()))
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
