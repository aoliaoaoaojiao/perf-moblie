package util

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	giDevice "github.com/SonicCloudOrg/sonic-gidevice"
	"github.com/gin-gonic/gin"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/gorilla/websocket"
	"log"
	"os"
	"perf-moblie/entity"
	"time"
)

var iOpts *entity.IOSOptions

func IOSInit(opts *entity.IOSOptions) (d giDevice.Device, perfData *entity.IOSDataChan, perfOpts []giDevice.PerfOption) {
	iOpts = opts
	device := GetDeviceByUdId(iOpts.UDID)
	if perfData == nil {
		perfData = &entity.IOSDataChan{}
	}
	if device == nil {
		fmt.Println("device is not found")
		os.Exit(0)
	}

	if (opts.Pid == -1 || iOpts.BundleID == "") && !iOpts.SystemCPU && !iOpts.SystemMem && !iOpts.SystemDisk && !iOpts.SystemNetWorking && !iOpts.SystemGPU && !iOpts.SystemFPS {
		sysAllParamsSet()
	}

	if (opts.Pid != -1 || iOpts.BundleID != "") && !iOpts.SystemCPU && !iOpts.SystemMem && !iOpts.SystemDisk && !iOpts.SystemNetWorking && !iOpts.SystemGPU && !iOpts.SystemFPS && !iOpts.ProcNetwork && !iOpts.ProcMem && !iOpts.ProcCPU {
		sysAllParamsSet()
		iOpts.ProcNetwork = true
		iOpts.ProcMem = true
		iOpts.ProcCPU = true
	}

	if iOpts.ProcCPU {
		addCpuAttr()
	}

	if iOpts.ProcMem {
		addMemAttr()
	}

	if iOpts.SystemCPU {
		perfData.SysChanCPU = make(chan giDevice.SystemCPUData)
	}
	if iOpts.SystemMem {
		perfData.SysChanMem = make(chan giDevice.SystemMemData)
	}
	if iOpts.SystemDisk {
		perfData.SysChanDisk = make(chan giDevice.SystemDiskData)
	}
	if iOpts.SystemNetWorking {
		perfData.SysChanNetwork = make(chan giDevice.SystemNetworkData)
	}
	if iOpts.SystemGPU {
		perfData.ChanGPU = make(chan giDevice.GPUData)
	}
	if iOpts.SystemFPS {
		perfData.ChanFPS = make(chan giDevice.FPSData)
	}
	if iOpts.ProcCPU || iOpts.ProcMem {
		perfData.ProcChanProc = make(chan entity.IOSProcPerf)
	}

	perfOpts = []giDevice.PerfOption{
		giDevice.WithPerfSystemCPU(iOpts.SystemCPU),
		giDevice.WithPerfSystemMem(iOpts.SystemMem),
		giDevice.WithPerfSystemDisk(iOpts.SystemDisk),
		giDevice.WithPerfSystemNetwork(iOpts.SystemNetWorking),
		giDevice.WithPerfNetwork(iOpts.ProcNetwork),
		giDevice.WithPerfFPS(iOpts.SystemFPS),
		giDevice.WithPerfGPU(iOpts.SystemGPU),
		giDevice.WithPerfOutputInterval(iOpts.RefreshTime),
	}

	if opts.Pid != -1 {
		perfOpts = append(perfOpts, giDevice.WithPerfPID(opts.Pid))
		perfOpts = append(perfOpts, giDevice.WithPerfProcessAttributes(opts.ProcessAttributes...))
	} else if iOpts.BundleID != "" {
		perfOpts = append(perfOpts, giDevice.WithPerfBundleID(iOpts.BundleID))
		perfOpts = append(perfOpts, giDevice.WithPerfProcessAttributes(opts.ProcessAttributes...))
	}
	return device, perfData, perfOpts
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

func RegisterIOSChart(data <-chan []byte, iosChan *entity.IOSDataChan, page *components.Page, r *gin.Engine, exitCtx context.Context) {
	go func() {
		for {
			select {
			case <-exitCtx.Done():
				return
			default:
				d, _ := <-data
				iOSDataSplit(d, iosChan)
			}
		}
	}()
	if iOpts.SystemCPU {
		line, eData := setChart("sys cpu info", iOpts.Addr)
		r.GET("/"+line.ChartID, func(c *gin.Context) {
			conversionIOSSysCPU("sys cpu info", iosChan.SysChanCPU, eData, c, exitCtx)
		})
		page.AddCharts(&line)
	}
	if iOpts.SystemMem {
		line, eData := setChart("sys mem info", iOpts.Addr)
		r.GET("/"+line.ChartID, func(c *gin.Context) {
			conversionIOSSysMem("sys mem info", iosChan.SysChanMem, eData, c, exitCtx)
		})
		page.AddCharts(&line)
	}
	if iOpts.SystemDisk {
		line, eData := setChart("sys disk info", iOpts.Addr)
		r.GET("/"+line.ChartID, func(c *gin.Context) {
			conversionIOSSysDisk("sys disk info", iosChan.SysChanDisk, eData, c, exitCtx)
		})
		page.AddCharts(&line)
	}
	if iOpts.SystemNetWorking {
		line, eData := setChart("sys network info", iOpts.Addr)
		r.GET("/"+line.ChartID, func(c *gin.Context) {
			conversionIOSSysNetwork("sys network info", iosChan.SysChanNetwork, eData, c, exitCtx)
		})
		page.AddCharts(&line)
	}
	if iOpts.SystemGPU {
		line, eData := setChart("sys gpu info", iOpts.Addr)
		r.GET("/"+line.ChartID, func(c *gin.Context) {
			conversionIOSSysGPU("sys gpu info", iosChan.ChanGPU, eData, c, exitCtx)
		})
		page.AddCharts(&line)
	}
	if iOpts.SystemFPS {
		line, eData := setChart("sys fps info", iOpts.Addr)
		r.GET("/"+line.ChartID, func(c *gin.Context) {
			conversionIOSSysFPS("sys fps info", iosChan.ChanFPS, eData, c, exitCtx)
		})
		page.AddCharts(&line)
	}
	if iOpts.ProcCPU {
		line, eData := setChart("sys proc cpu info", iOpts.Addr)
		r.GET("/"+line.ChartID, func(c *gin.Context) {
			conversionIOSProcCPU("sys proc cpu info", iosChan.ProcChanProc, eData, c, exitCtx)
		})
		page.AddCharts(&line)
	}
	if iOpts.ProcCPU {
		line, eData := setChart("sys proc mem info", iOpts.Addr)
		r.GET("/"+line.ChartID, func(c *gin.Context) {
			conversionIOSProcMem("sys proc mem info", iosChan.ProcChanProc, eData, c, exitCtx)
		})
		page.AddCharts(&line)
	}
}

func iOSDataSplit(dataByte []byte, iosChan *entity.IOSDataChan) {
	dataMap := make(map[string]interface{})
	err := json.Unmarshal(dataByte, &dataMap)
	if err != nil {
		panic(err)
	}
	dataType := dataMap["type"]
	switch dataType {
	case "sys_cpu":
		data := giDevice.SystemCPUData{}
		err = json.Unmarshal(dataByte, &data)
		if err != nil {
			panic(err)
		}
		iosChan.SysChanCPU <- data
	case "sys_disk":
		data := giDevice.SystemDiskData{}
		err = json.Unmarshal(dataByte, &data)
		if err != nil {
			panic(err)
		}
		iosChan.SysChanDisk <- data
	case "sys_mem":
		data := giDevice.SystemMemData{}
		err = json.Unmarshal(dataByte, &data)
		if err != nil {
			panic(err)
		}
		iosChan.SysChanMem <- data
	case "sys_network":
		data := giDevice.SystemNetworkData{}
		err = json.Unmarshal(dataByte, &data)
		if err != nil {
			panic(err)
		}
		iosChan.SysChanNetwork <- data
	case "gpu":
		data := giDevice.GPUData{}
		err = json.Unmarshal(dataByte, &data)
		if err != nil {
			panic(err)
		}
		iosChan.ChanGPU <- data
	case "fps":
		data := giDevice.FPSData{}
		err = json.Unmarshal(dataByte, &data)
		if err != nil {
			panic(err)
		}
		iosChan.ChanFPS <- data
	case "process":
		data := entity.IOSProcPerf{}
		err = json.Unmarshal(dataByte, &data)
		if err != nil {
			panic(err)
		}
		iosChan.ProcChanProc <- data
	}
}

func conversionIOSSysCPU(title string, dataChan chan giDevice.SystemCPUData, eData *entity.EchartsData, c *gin.Context, exitCtx context.Context) {
	conn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}
	var pingTicker = time.NewTicker(pingPeriod)
	defer func() {
		conn.Close()
		pingTicker.Stop()
	}()
	for {
		select {
		case data := <-dataChan:
			line := getLineTemplate(title)
			eData.XAxis = append(eData.XAxis, time.Unix(data.TimeStamp, 0).Format("2006-01-02 15:04:05"))
			if eData.Series["SystemLoad"] == nil {
				eData.Series["SystemLoad"] = []opts.LineData{}
			}
			if eData.Series["NiceLoad"] == nil {
				eData.Series["NiceLoad"] = []opts.LineData{}
			}
			if eData.Series["TotalLoad"] == nil {
				eData.Series["TotalLoad"] = []opts.LineData{}
			}
			if eData.Series["UserLoad"] == nil {
				eData.Series["UserLoad"] = []opts.LineData{}
			}
			eData.Series["SystemLoad"] = append(eData.Series["SystemLoad"], opts.LineData{Value: data.SystemLoad})
			eData.Series["NiceLoad"] = append(eData.Series["NiceLoad"], opts.LineData{Value: data.NiceLoad})
			eData.Series["TotalLoad"] = append(eData.Series["TotalLoad"], opts.LineData{Value: data.TotalLoad})
			eData.Series["UserLoad"] = append(eData.Series["UserLoad"], opts.LineData{Value: data.UserLoad})
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

func conversionIOSSysMem(title string, dataChan chan giDevice.SystemMemData, eData *entity.EchartsData, c *gin.Context, exitCtx context.Context) {
	conn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}
	var pingTicker = time.NewTicker(pingPeriod)
	defer func() {
		conn.Close()
		pingTicker.Stop()
	}()
	for {
		select {
		case data := <-dataChan:
			line := getLineTemplate(title)
			eData.XAxis = append(eData.XAxis, time.Unix(data.TimeStamp, 0).Format("2006-01-02 15:04:05"))
			if eData.Series["AppMemory"] == nil {
				eData.Series["AppMemory"] = []opts.LineData{}
			}
			if eData.Series["CachedFiles"] == nil {
				eData.Series["CachedFiles"] = []opts.LineData{}
			}
			if eData.Series["Compressed"] == nil {
				eData.Series["Compressed"] = []opts.LineData{}
			}
			if eData.Series["FreeMemory"] == nil {
				eData.Series["FreeMemory"] = []opts.LineData{}
			}
			if eData.Series["SwapUsed"] == nil {
				eData.Series["SwapUsed"] = []opts.LineData{}
			}
			if eData.Series["UsedMemory"] == nil {
				eData.Series["UsedMemory"] = []opts.LineData{}
			}
			if eData.Series["WiredMemory"] == nil {
				eData.Series["WiredMemory"] = []opts.LineData{}
			}
			eData.Series["AppMemory"] = append(eData.Series["AppMemory"], opts.LineData{Value: data.AppMemory})
			eData.Series["CachedFiles"] = append(eData.Series["CachedFiles"], opts.LineData{Value: data.CachedFiles})
			eData.Series["Compressed"] = append(eData.Series["Compressed"], opts.LineData{Value: data.Compressed})
			eData.Series["FreeMemory"] = append(eData.Series["FreeMemory"], opts.LineData{Value: data.FreeMemory})
			eData.Series["SwapUsed"] = append(eData.Series["SwapUsed"], opts.LineData{Value: data.SwapUsed})
			eData.Series["UsedMemory"] = append(eData.Series["UsedMemory"], opts.LineData{Value: data.UsedMemory})
			eData.Series["WiredMemory"] = append(eData.Series["WiredMemory"], opts.LineData{Value: data.WiredMemory})
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

func conversionIOSSysDisk(title string, dataChan chan giDevice.SystemDiskData, eData *entity.EchartsData, c *gin.Context, exitCtx context.Context) {
	conn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}
	var pingTicker = time.NewTicker(pingPeriod)
	defer func() {
		conn.Close()
		pingTicker.Stop()
	}()
	for {
		select {
		case data := <-dataChan:
			line := getLineTemplate(title)
			eData.XAxis = append(eData.XAxis, time.Unix(data.TimeStamp, 0).Format("2006-01-02 15:04:05"))
			if eData.Series["DataRead"] == nil {
				eData.Series["DataRead"] = []opts.LineData{}
			}
			if eData.Series["DataWritten"] == nil {
				eData.Series["DataWritten"] = []opts.LineData{}
			}
			if eData.Series["ReadOps"] == nil {
				eData.Series["ReadOps"] = []opts.LineData{}
			}
			if eData.Series["WriteOps"] == nil {
				eData.Series["WriteOps"] = []opts.LineData{}
			}
			eData.Series["DataRead"] = append(eData.Series["DataRead"], opts.LineData{Value: data.DataRead})
			eData.Series["DataWritten"] = append(eData.Series["DataWritten"], opts.LineData{Value: data.DataWritten})
			eData.Series["ReadOps"] = append(eData.Series["ReadOps"], opts.LineData{Value: data.ReadOps})
			eData.Series["WriteOps"] = append(eData.Series["WriteOps"], opts.LineData{Value: data.WriteOps})
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

func conversionIOSSysNetwork(title string, dataChan chan giDevice.SystemNetworkData, eData *entity.EchartsData, c *gin.Context, exitCtx context.Context) {
	conn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}
	var pingTicker = time.NewTicker(pingPeriod)
	defer func() {
		conn.Close()
		pingTicker.Stop()
	}()
	for {
		select {
		case data := <-dataChan:
			line := getLineTemplate(title)
			eData.XAxis = append(eData.XAxis, time.Unix(data.TimeStamp, 0).Format("2006-01-02 15:04:05"))
			if eData.Series["BytesIn"] == nil {
				eData.Series["BytesIn"] = []opts.LineData{}
			}
			if eData.Series["BytesOut"] == nil {
				eData.Series["BytesOut"] = []opts.LineData{}
			}
			if eData.Series["PacketsIn"] == nil {
				eData.Series["PacketsIn"] = []opts.LineData{}
			}
			if eData.Series["PacketsOut"] == nil {
				eData.Series["PacketsOut"] = []opts.LineData{}
			}
			eData.Series["BytesIn"] = append(eData.Series["BytesIn"], opts.LineData{Value: data.BytesIn})
			eData.Series["BytesOut"] = append(eData.Series["BytesOut"], opts.LineData{Value: data.BytesOut})
			eData.Series["PacketsIn"] = append(eData.Series["PacketsIn"], opts.LineData{Value: data.PacketsIn})
			eData.Series["PacketsOut"] = append(eData.Series["PacketsOut"], opts.LineData{Value: data.PacketsOut})
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

func conversionIOSSysFPS(title string, dataChan chan giDevice.FPSData, eData *entity.EchartsData, c *gin.Context, exitCtx context.Context) {
	conn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}
	var pingTicker = time.NewTicker(pingPeriod)
	defer func() {
		conn.Close()
		pingTicker.Stop()
	}()
	for {
		select {
		case data := <-dataChan:
			line := getLineTemplate(title)
			eData.XAxis = append(eData.XAxis, time.Unix(data.TimeStamp, 0).Format("2006-01-02 15:04:05"))
			if eData.Series["FPS"] == nil {
				eData.Series["FPS"] = []opts.LineData{}
			}
			eData.Series["FPS"] = append(eData.Series["FPS"], opts.LineData{Value: data.FPS})
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

func conversionIOSSysGPU(title string, dataChan chan giDevice.GPUData, eData *entity.EchartsData, c *gin.Context, exitCtx context.Context) {
	conn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}
	var pingTicker = time.NewTicker(pingPeriod)
	defer func() {
		conn.Close()
		pingTicker.Stop()
	}()
	for {
		select {
		case data := <-dataChan:
			line := getLineTemplate(title)
			eData.XAxis = append(eData.XAxis, time.Unix(data.TimeStamp, 0).Format("2006-01-02 15:04:05"))
			if eData.Series["DeviceUtilization"] == nil {
				eData.Series["DeviceUtilization"] = []opts.LineData{}
			}
			if eData.Series["RendererUtilization"] == nil {
				eData.Series["RendererUtilization"] = []opts.LineData{}
			}
			if eData.Series["TilerUtilization"] == nil {
				eData.Series["TilerUtilization"] = []opts.LineData{}
			}
			eData.Series["DeviceUtilization"] = append(eData.Series["DeviceUtilization"], opts.LineData{Value: data.DeviceUtilization})
			eData.Series["RendererUtilization"] = append(eData.Series["RendererUtilization"], opts.LineData{Value: data.RendererUtilization})
			eData.Series["TilerUtilization"] = append(eData.Series["TilerUtilization"], opts.LineData{Value: data.TilerUtilization})
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

func conversionIOSProcCPU(title string, dataChan chan entity.IOSProcPerf, eData *entity.EchartsData, c *gin.Context, exitCtx context.Context) {
	conn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}
	var pingTicker = time.NewTicker(pingPeriod)
	defer func() {
		conn.Close()
		pingTicker.Stop()
	}()
	for {
		select {
		case data := <-dataChan:
			line := getLineTemplate(title)
			eData.XAxis = append(eData.XAxis, time.Unix(data.TimeStamp, 0).Format("2006-01-02 15:04:05"))
			if eData.Series["CPUUsage"] == nil {
				eData.Series["CPUUsage"] = []opts.LineData{}
			}
			eData.Series["CPUUsage"] = append(eData.Series["CPUUsage"], opts.LineData{Value: data.IOSProcPerf.CPUUsage})
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

func conversionIOSProcMem(title string, dataChan chan entity.IOSProcPerf, eData *entity.EchartsData, c *gin.Context, exitCtx context.Context) {
	conn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}
	var pingTicker = time.NewTicker(pingPeriod)
	defer func() {
		conn.Close()
		pingTicker.Stop()
	}()
	for {
		select {
		case data := <-dataChan:
			line := getLineTemplate(title)
			eData.XAxis = append(eData.XAxis, time.Unix(data.TimeStamp, 0).Format("2006-01-02 15:04:05"))
			if eData.Series["MemAnon"] == nil {
				eData.Series["MemAnon"] = []opts.LineData{}
			}
			if eData.Series["MemResidentSize"] == nil {
				eData.Series["MemResidentSize"] = []opts.LineData{}
			}
			if eData.Series["MemVirtualSize"] == nil {
				eData.Series["MemVirtualSize"] = []opts.LineData{}
			}
			if eData.Series["PhysFootprint"] == nil {
				eData.Series["PhysFootprint"] = []opts.LineData{}
			}
			eData.Series["MemAnon"] = append(eData.Series["MemAnon"], opts.LineData{Value: data.IOSProcPerf.MemAnon})
			eData.Series["MemResidentSize"] = append(eData.Series["MemResidentSize"], opts.LineData{Value: data.IOSProcPerf.MemResidentSize})
			eData.Series["MemVirtualSize"] = append(eData.Series["MemVirtualSize"], opts.LineData{Value: data.IOSProcPerf.MemVirtualSize})
			eData.Series["PhysFootprint"] = append(eData.Series["PhysFootprint"], opts.LineData{Value: data.IOSProcPerf.PhysFootprint})
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

func addCpuAttr() {
	iOpts.ProcessAttributes = append(iOpts.ProcessAttributes, "cpuUsage")
}

func addMemAttr() {
	iOpts.ProcessAttributes = append(iOpts.ProcessAttributes, "memVirtualSize", "physFootprint", "memResidentSize", "memAnon")
}

func sysAllParamsSet() {
	iOpts.SystemFPS = true
	iOpts.SystemGPU = true
	iOpts.SystemCPU = true
	iOpts.SystemMem = true
}
