package util

import (
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
	"github.com/gorilla/websocket"
	"log"
	"os"
	"perf-moblie/entity"
	"time"
)

var aOpts *entity.AndroidOptions

func AndroidInit(opt *entity.AndroidOptions) (d adb.Device) {
	aOpts = opt
	var err error
	device := sasutil.GetDevice(aOpts.AndroidSerial)

	if aOpts.Pid == -1 && aOpts.AndroidPackageName != "" {
		sasp.PackageName = aOpts.AndroidPackageName
		sasp.Pid, err = sasp.GetPidOnPackageName(device, aOpts.AndroidPackageName)
		if err != nil {
			fmt.Println("no corresponding application PID found")
			os.Exit(0)
		}
	} else if aOpts.Pid != -1 && aOpts.AndroidPackageName == "" {
		sasp.Pid = fmt.Sprintf("%d", aOpts.Pid)
		sasp.PackageName, err = sasp.GetNameOnPid(device, sasp.Pid)
		if err != nil {
			aOpts.AndroidPackageName = ""
		}
	}

	if (sasp.Pid == "" && aOpts.AndroidPackageName == "") &&
		!aOpts.AndroidOptions.SystemCPU &&
		!aOpts.AndroidOptions.SystemGPU &&
		!aOpts.AndroidOptions.SystemNetWorking &&
		!aOpts.AndroidOptions.SystemMem {
		androidParamsSet()
	}
	if (sasp.Pid != "" || aOpts.AndroidPackageName != "") &&
		!aOpts.AndroidOptions.ProcMem &&
		!aOpts.AndroidOptions.ProcCPU &&
		!aOpts.AndroidOptions.ProcThreads &&
		!aOpts.AndroidOptions.ProcFPS {
		aOpts.AndroidOptions.ProcMem = true
		aOpts.AndroidOptions.ProcCPU = true
		aOpts.AndroidOptions.ProcThreads = true
		aOpts.AndroidOptions.ProcFPS = true
	}
	return *device
}

func androidParamsSet() {
	aOpts.AndroidOptions.SystemCPU = true
	aOpts.AndroidOptions.SystemMem = true
	aOpts.AndroidOptions.SystemGPU = true
	aOpts.AndroidOptions.SystemNetWorking = true
}

func RegisterAndroidRoute(device *adb.Device, page *components.Page, r *gin.Engine, exitCtx context.Context) {
	if aOpts.AndroidOptions.SystemCPU {
		line, eData, dataChan := setAndroid("sys cpu info", aOpts.Addr+"/android/sys/cpu")
		r.GET("/android/sys/cpu", func(c *gin.Context) {
			sasp.GetSystemCPU(device, aOpts.AndroidOptions, dataChan, exitCtx)
			androidSysCPU("sys cpu info", dataChan, eData, c, exitCtx)
		})
		page.AddCharts(&line)
	}
	if aOpts.AndroidOptions.SystemMem {
		line, eData, dataChan := setAndroid("sys mem info", aOpts.Addr)
		sasp.GetSystemMem(device, aOpts.AndroidOptions, dataChan, exitCtx)
		r.GET("/"+line.ChartID, func(c *gin.Context) {
			androidSysMem("sys mem info", dataChan, eData, c, exitCtx)
		})
		page.AddCharts(&line)
	}
	if aOpts.AndroidOptions.SystemNetWorking {
		line, eData, dataChan := setAndroid("sys networking info", aOpts.Addr)
		sasp.GetSystemNetwork(device, aOpts.AndroidOptions, dataChan, exitCtx)
		r.GET("/"+line.ChartID, func(c *gin.Context) {
			androidSysNetwork("sys networking info", dataChan, eData, c, exitCtx)
		})
		page.AddCharts(&line)
	}
	if aOpts.AndroidOptions.ProcCPU {
		line, eData, dataChan := setAndroid("process cpu info", aOpts.Addr)
		sasp.GetProcCpu(device, aOpts.AndroidOptions, dataChan, exitCtx)
		r.GET("/"+line.ChartID, func(c *gin.Context) {
			androidProcCPU("process cpu info", dataChan, eData, c, exitCtx)
		})
		page.AddCharts(&line)
	}
	if aOpts.AndroidOptions.ProcMem {
		line, eData, dataChan := setAndroid("process mem info", aOpts.Addr)
		sasp.GetProcMem(device, aOpts.AndroidOptions, dataChan, exitCtx)
		r.GET("/"+line.ChartID, func(c *gin.Context) {
			androidProcMem("process mem info", dataChan, eData, c, exitCtx)
		})
		page.AddCharts(&line)
	}
	if aOpts.AndroidOptions.ProcFPS {
		line, eData, dataChan := setAndroid("process FPS info", aOpts.Addr)
		sasp.GetProcFPS(device, aOpts.AndroidOptions, dataChan, exitCtx)
		r.GET("/"+line.ChartID, func(c *gin.Context) {
			androidProcFPS("process FPS info", dataChan, eData, c, exitCtx)
		})
		page.AddCharts(&line)
	}
	if aOpts.AndroidOptions.ProcThreads {
		line, eData, dataChan := setAndroid("process Thread info", aOpts.Addr)
		sasp.GetProcThreads(device, aOpts.AndroidOptions, dataChan, exitCtx)
		r.GET("/"+line.ChartID, func(c *gin.Context) {
			androidProcThreads("process Thread info", dataChan, eData, c, exitCtx)
		})
		page.AddCharts(&line)
	}
}

func RegisterAndroidChart(device *adb.Device, page *components.Page, r *gin.Engine, exitCtx context.Context) {
	if aOpts.AndroidOptions.SystemCPU {
		line, eData, dataChan := setAndroid("sys cpu info", aOpts.Addr)
		sasp.GetSystemCPU(device, aOpts.AndroidOptions, dataChan, exitCtx)
		r.GET("/"+line.ChartID, func(c *gin.Context) {
			androidSysCPU("sys cpu info", dataChan, eData, c, exitCtx)
		})
		page.AddCharts(&line)
	}
	if aOpts.AndroidOptions.SystemMem {
		line, eData, dataChan := setAndroid("sys mem info", aOpts.Addr)
		sasp.GetSystemMem(device, aOpts.AndroidOptions, dataChan, exitCtx)
		r.GET("/"+line.ChartID, func(c *gin.Context) {
			androidSysMem("sys mem info", dataChan, eData, c, exitCtx)
		})
		page.AddCharts(&line)
	}
	if aOpts.AndroidOptions.SystemNetWorking {
		line, eData, dataChan := setAndroid("sys networking info", aOpts.Addr)
		sasp.GetSystemNetwork(device, aOpts.AndroidOptions, dataChan, exitCtx)
		r.GET("/"+line.ChartID, func(c *gin.Context) {
			androidSysNetwork("sys networking info", dataChan, eData, c, exitCtx)
		})
		page.AddCharts(&line)
	}
	if aOpts.AndroidOptions.ProcCPU {
		line, eData, dataChan := setAndroid("process cpu info", aOpts.Addr)
		sasp.GetProcCpu(device, aOpts.AndroidOptions, dataChan, exitCtx)
		r.GET("/"+line.ChartID, func(c *gin.Context) {
			androidProcCPU("process cpu info", dataChan, eData, c, exitCtx)
		})
		page.AddCharts(&line)
	}
	if aOpts.AndroidOptions.ProcMem {
		line, eData, dataChan := setAndroid("process mem info", aOpts.Addr)
		sasp.GetProcMem(device, aOpts.AndroidOptions, dataChan, exitCtx)
		r.GET("/"+line.ChartID, func(c *gin.Context) {
			androidProcMem("process mem info", dataChan, eData, c, exitCtx)
		})
		page.AddCharts(&line)
	}
	if aOpts.AndroidOptions.ProcFPS {
		line, eData, dataChan := setAndroid("process FPS info", aOpts.Addr)
		sasp.GetProcFPS(device, aOpts.AndroidOptions, dataChan, exitCtx)
		r.GET("/"+line.ChartID, func(c *gin.Context) {
			androidProcFPS("process FPS info", dataChan, eData, c, exitCtx)
		})
		page.AddCharts(&line)
	}
	if aOpts.AndroidOptions.ProcThreads {
		line, eData, dataChan := setAndroid("process Thread info", aOpts.Addr)
		sasp.GetProcThreads(device, aOpts.AndroidOptions, dataChan, exitCtx)
		r.GET("/"+line.ChartID, func(c *gin.Context) {
			androidProcThreads("process Thread info", dataChan, eData, c, exitCtx)
		})
		page.AddCharts(&line)
	}
}

func setAndroid(title, addr string) (charts.Line, *entity.EchartsData, chan *sentity.PerfmonData) {
	dataChan := make(chan *sentity.PerfmonData)
	line, eData := setChart(title, addr)
	return line, eData, dataChan
}

func androidSysCPU(title string, dataChan chan *sentity.PerfmonData, eData *entity.EchartsData, c *gin.Context, exitCtx context.Context) {
	conn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	var pingTicker = time.NewTicker(pingPeriod)
	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}
	defer func() {
		conn.Close()
		pingTicker.Stop()
	}()
	for {
		select {
		case data := <-dataChan:
			line := getLineTemplate(title)
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
			for key, value := range eData.Series {
				line = line.SetXAxis(eData.XAxis).AddSeries(key, value)
			}
			line.Validate()
			err = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				panic(err)
			}
			err = conn.WriteJSON(line.JSON())
			if err != nil {
				panic(err)
			}
		case <-pingTicker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err = conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		case <-exitCtx.Done():
			break
		}
	}
}

func androidSysMem(title string, dataChan chan *sentity.PerfmonData, eData *entity.EchartsData, c *gin.Context, exitCtx context.Context) {
	conn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	var pingTicker = time.NewTicker(pingPeriod)
	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}
	defer func() {
		conn.Close()
		pingTicker.Stop()
	}()
	for {
		select {
		case data := <-dataChan:
			line := getLineTemplate(title)
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
			for key, value := range eData.Series {
				line = line.SetXAxis(eData.XAxis).AddSeries(key, value)
			}
			line.Validate()
			err = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				panic(err)
			}
			err = conn.WriteJSON(line.JSON())
			if err != nil {
				panic(err)
			}
		case <-pingTicker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err = conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		case <-exitCtx.Done():
			break
		}
	}
}

func androidSysNetwork(title string, dataChan chan *sentity.PerfmonData, eData *entity.EchartsData, c *gin.Context, exitCtx context.Context) {
	conn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	var pingTicker = time.NewTicker(pingPeriod)
	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}
	defer func() {
		conn.Close()
		pingTicker.Stop()
	}()
	for {
		select {
		case data := <-dataChan:
			line := getLineTemplate(title)
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
			for key, value := range eData.Series {
				line = line.SetXAxis(eData.XAxis).AddSeries(key, value)
			}
			line.Validate()
			err = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				panic(err)
			}
			err = conn.WriteJSON(line.JSON())
			if err != nil {
				panic(err)
			}
		case <-pingTicker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err = conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		case <-exitCtx.Done():
			break
		}
	}
}

func androidProcCPU(title string, dataChan chan *sentity.PerfmonData, eData *entity.EchartsData, c *gin.Context, exitCtx context.Context) {
	conn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	var pingTicker = time.NewTicker(pingPeriod)
	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}
	defer func() {
		conn.Close()
		pingTicker.Stop()
	}()
	for {
		select {
		case data := <-dataChan:
			line := getLineTemplate(title)
			if data.Process.CPUInfo != nil {
				eData.XAxis = append(eData.XAxis, time.Unix(data.Process.CPUInfo.TimeStamp/1000, 0).Format("2006-01-02 15:04:05"))
				if eData.Series["cpuUtilization"] == nil {
					eData.Series["cpuUtilization"] = []opts.LineData{}
				}
				eData.Series["cpuUtilization"] = append(eData.Series["cpuUtilization"], opts.LineData{Value: data.Process.CPUInfo.CpuUtilization})
			}
			for key, value := range eData.Series {
				line = line.SetXAxis(eData.XAxis).AddSeries(key, value)
			}
			line.Validate()
			err = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				panic(err)
			}
			err = conn.WriteJSON(line.JSON())
			if err != nil {
				panic(err)
			}
		case <-pingTicker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err = conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		case <-exitCtx.Done():
			break
		}
	}
}

func androidProcMem(title string, dataChan chan *sentity.PerfmonData, eData *entity.EchartsData, c *gin.Context, exitCtx context.Context) {
	conn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	var pingTicker = time.NewTicker(pingPeriod)
	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}
	defer func() {
		conn.Close()
		pingTicker.Stop()
	}()
	for {
		select {
		case data := <-dataChan:
			line := getLineTemplate(title)
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
			for key, value := range eData.Series {
				line = line.SetXAxis(eData.XAxis).AddSeries(key, value)
			}
			line.Validate()
			err = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				panic(err)
			}
			err = conn.WriteJSON(line.JSON())
			if err != nil {
				panic(err)
			}
		case <-pingTicker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err = conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		case <-exitCtx.Done():
			break
		}
	}
}

func androidProcFPS(title string, dataChan chan *sentity.PerfmonData, eData *entity.EchartsData, c *gin.Context, exitCtx context.Context) {
	conn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	var pingTicker = time.NewTicker(pingPeriod)
	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}
	defer func() {
		conn.Close()
		pingTicker.Stop()
	}()
	for {
		select {
		case data := <-dataChan:
			line := getLineTemplate(title)
			if data.Process.FPSInfo != nil {
				eData.XAxis = append(eData.XAxis, time.Unix(data.Process.FPSInfo.TimeStamp/1000, 0).Format("2006-01-02 15:04:05"))
				if eData.Series["FPS"] == nil {
					eData.Series["FPS"] = []opts.LineData{}
				}
				eData.Series["FPS"] = append(eData.Series["FPS"], opts.LineData{Value: data.Process.FPSInfo.FPS})
			}
			for key, value := range eData.Series {
				line = line.SetXAxis(eData.XAxis).AddSeries(key, value)
			}
			line.Validate()
			err = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				panic(err)
			}
			err = conn.WriteJSON(line.JSON())
			if err != nil {
				panic(err)
			}
		case <-pingTicker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err = conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		case <-exitCtx.Done():
			break
		}
	}
}

func androidProcThreads(title string, dataChan chan *sentity.PerfmonData, eData *entity.EchartsData, c *gin.Context, exitCtx context.Context) {
	conn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	var pingTicker = time.NewTicker(pingPeriod)
	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}
	defer func() {
		conn.Close()
		pingTicker.Stop()
	}()
	for {
		select {
		case data := <-dataChan:
			line := getLineTemplate(title)
			if data.Process.ThreadInfo != nil {
				eData.XAxis = append(eData.XAxis, time.Unix(data.Process.ThreadInfo.TimeStamp/1000, 0).Format("2006-01-02 15:04:05"))
				if eData.Series["threadCount"] == nil {
					eData.Series["threadCount"] = []opts.LineData{}
				}
				eData.Series["threadCount"] = append(eData.Series["threadCount"], opts.LineData{Value: data.Process.ThreadInfo.Threads})
			}
			for key, value := range eData.Series {
				line = line.SetXAxis(eData.XAxis).AddSeries(key, value)
			}
			line.Validate()
			err = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				panic(err)
			}
			err = conn.WriteJSON(line.JSON())
			if err != nil {
				panic(err)
			}
		case <-pingTicker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err = conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		case <-exitCtx.Done():
			break
		}
	}
}
