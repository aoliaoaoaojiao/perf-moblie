package entity

import "github.com/go-echarts/go-echarts/v2/opts"

type EchartsData struct {
	XAxis  []string
	Series map[string][]opts.LineData
}
