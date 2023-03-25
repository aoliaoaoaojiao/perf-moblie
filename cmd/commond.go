package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"log"
	"math/rand"
	"net/http"
	"perf-moblie/statics"
	"text/template"
	"time"
)

var (
	itemCntLine = 6
	fruits      = []string{"Apple", "Banana", "Peach ", "Lemon", "Pear", "Cherry"}
	items       = []opts.LineData{}
	reasons     = []string{
		"test1",
		"test2",
		"test3",
		"test4"}
	line = charts.NewLine()
)

func generateLineItems() {
	for i := 0; i < itemCntLine; i++ {
		items = append(items, opts.LineData{Value: rand.Intn(300)})
	}
}

func tickerTest(data chan interface{}) {
	ticker1 := time.NewTicker(1 * time.Second)

	for {
		// 每1秒中从chan t.C 中读取一次
		<-ticker1.C
		fruits = append(fruits, reasons[rand.Intn(len(reasons))])
		items = append(items, opts.LineData{Value: rand.Intn(300)})
		line1 := charts.NewLine()
		line1.SetGlobalOptions(
			charts.WithTitleOpts(opts.Title{
				Title: "multi lines",
			}),
			charts.WithTooltipOpts(opts.Tooltip{
				Show:    true,
				Trigger: "axis",
				//AxisPointer:
			}),
			charts.WithLegendOpts(opts.Legend{
				Show: true,
				//Right: "50%",
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
		line1.SetXAxis(fruits).
			//AddSeries("Category  A", generateLineItems()).
			//AddSeries("Category  B", generateLineItems()).
			//AddSeries("Category  C", generateLineItems()).
			AddSeries("Category  D", items)
		data <- line1.JSON()
	}

}

func lineMulti() {

	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "multi lines",
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Show:    true,
			Trigger: "axis",
			//AxisPointer:
		}),
		charts.WithLegendOpts(opts.Legend{
			Show: true,
			//Right: "50%",
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
	line.SetXAxis(fruits).
		//AddSeries("Category  A", generateLineItems()).
		//AddSeries("Category  B", generateLineItems()).
		//AddSeries("Category  C", generateLineItems()).
		AddSeries("Category  D", items)
}

type LineExamples struct{}

var (
	itemCnt = 7
	weeks   = []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
)

func generateBarItems() []opts.BarData {
	items := make([]opts.BarData, 0)
	for i := 0; i < itemCnt; i++ {
		items = append(items, opts.BarData{Value: rand.Intn(300)})
	}
	return items
}

func barBasic() *charts.Bar {
	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "basic bar example", Subtitle: "This is the subtitle."}),
	)

	bar.SetXAxis(weeks).
		AddSeries("Category A", generateBarItems()).
		AddSeries("Category B", generateBarItems())
	return bar
}

func barTitle() *charts.Bar {
	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title:    "title and legend options",
			Subtitle: "go-echarts is an awesome chart library written in Golang",
			Link:     "https://github.com/go-echarts/go-echarts",
			Right:    "40%",
		}),
		charts.WithToolboxOpts(opts.Toolbox{Show: true}),
		charts.WithLegendOpts(opts.Legend{Right: "80%"}),
	)
	bar.SetXAxis(weeks).
		AddSeries("Category A", generateBarItems()).
		AddSeries("Category B", generateBarItems())
	return bar
}

func barSize() *charts.Bar {
	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title:    "adjust canvas size",
			Subtitle: "I want a bigger canvas size :)",
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "1200px",
			Height: "600px",
		}),
	)
	bar.SetXAxis(weeks).
		AddSeries("Category A", generateBarItems()).
		AddSeries("Category B", generateBarItems())
	return bar
}

type Exampler interface {
	Examples()
}

type BarExamples struct{}

var DefaultTemplate = `
$(function () { setInterval({{ .ViewID }}_sync, {{ .Interval }}); });
function {{ .ViewID }}_sync() {
    $.ajax({
        type: "GET",
        url: "http://{{ .Addr }}",
        dataType: "json",
        success: function (result) {
            console.log('Received data:', result);
            let opt = goecharts_{{ .ViewID }}.getOption();

            goecharts_{{ .ViewID }}.setOption(opt);
        }
    });
}`

func setCors(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")                                                            // 允许访问所有域，可以换成具体url，注意仅具体url才能带cookie信息
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token") //header的类型
	w.Header().Add("Access-Control-Allow-Credentials", "true")                                                    //设置为true，允许ajax异步请求带cookie信息
	w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")                             //允许请求方法
	//w.Header().Set("content-type", "application/json;charset=UTF-8")             //返回数据格式是json
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
}

func genViewTemplate(interval int, vid, addr string) string {
	tpl, err := template.New("view").Parse(DefaultTemplate)
	if err != nil {
		panic("statsview: failed to parse template " + err.Error())
	}

	var c = struct {
		Interval int
		Addr     string
		ViewID   string
	}{
		Interval: interval,
		Addr:     addr,
		ViewID:   vid,
	}

	buf := bytes.Buffer{}
	if err := tpl.Execute(&buf, c); err != nil {
		panic("statsview: failed to execute template " + err.Error())
	}

	return buf.String()
}

func RegisterChartsWS(interval int, chartID string) (jsTemplate string, data chan interface{}) {
	var url = fmt.Sprintf("/perf/%s", chartID)
	var dataC = make(chan interface{})

	http.HandleFunc(url, func(responseWriter http.ResponseWriter, request *http.Request) {
		select {
		case responseData, ok := <-dataC:
			if ok {
				jsonResp, err := json.Marshal(responseData)
				if err != nil {
					log.Fatalf("Error happened in JSON marshal. Err: %s", err)
				}
				responseWriter.Write(jsonResp)
			}
		default:
			responseWriter.Write([]byte("not data"))
		}

	})
	return genViewTemplate(interval, chartID, "localhost:8081"+url), dataC
}

func RegisterJS() {
	staticsPrev := "/statics/"

	http.HandleFunc(staticsPrev+"jquery.min.js", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(statics.JqueryJS))
	})
	http.HandleFunc(staticsPrev+"echarts.min.js", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(statics.EchartJS))
	})
}
