package entity

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestIOSData(t *testing.T) {
	data := "{\"pid\":4253,\"proc_perf\":{\"cpuUsage\":0.1118021393889875,\"memAnon\":40386560,\"memResidentSize\":53362688,\"memVirtualSize\":418828779520,\"physFootprint\":96535592,\"pid\":4253},\"timestamp\":1668850295,\"type\":\"process\"}\n"
	iosData := IOSProcPerf{}
	json.Unmarshal([]byte(data), &iosData)
	fmt.Println(iosData)
}
