package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/templates"
	"github.com/gorilla/websocket"
	"html/template"
	"net/http"
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

func Cors() gin.HandlerFunc {
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

var upGrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// solve cross domain problems
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
} // use default options

func SetPageInit(addr string, page *components.Page) {
	templates.BaseTpl = BaseTpl
	page.Initialization.AssetsHost = fmt.Sprintf("http://%s/statics/", addr)
	page.JSAssets.Add("jquery.min.js")
}

func setChart(title, addr string) (charts.Line, *entity.EchartsData) {
	line := getLineTemplate(title)
	line.AddJSFuncs(registerJs(aOpts.AndroidOptions.RefreshTime, line.ChartID, addr))
	eData := &entity.EchartsData{
		XAxis:  []string{},
		Series: map[string][]opts.LineData{},
	}
	return *line, eData
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

func registerJs(Interval int, chartID, addr string) string {
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

const (
	// writeWait is the time allowed to write the file to the client.
	writeWait = 10 * time.Second
	// pongWait is the time allowed to read the next pong message from the client.
	pongWait = 60 * time.Second
	// pingPeriod is the interval between pings sent to client. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

func reader(ws *websocket.Conn) {
	defer ws.Close()
	ws.SetReadLimit(512)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}
}

func writer(ws *websocket.Conn, dataC <-chan interface{}) {
	var pingTicker = time.NewTicker(pingPeriod)
	defer func() {
		pingTicker.Stop()
		ws.Close()
	}()
	for {
		select {
		case changes := <-dataC:
			bytes, err := json.Marshal(changes)
			if err != nil {
				panic(fmt.Sprintf("unexpected err while marshalling changes: %v", changes))
			}
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.TextMessage, bytes); err != nil {
				return
			}
		case <-pingTicker.C:
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}
