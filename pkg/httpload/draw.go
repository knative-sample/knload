package httpload

import (
	"encoding/base64"
	"fmt"
	"os"
	"sync"
	"text/template"

	"github.com/knative-sample/knload/pkg/c3"
	"github.com/golang/glog"
)

func (hl *HttpLoad) Draw() {
	var wg sync.WaitGroup
	var concurrentList []int
	var responseTimeList []float64
	var podNumList []float64

	drawHtml := func() {
		concurrentStr := ""
		responseTimeStr := ""
		podNumStr := ""
		for _, c := range concurrentList {
			if concurrentStr == "" {
				concurrentStr = fmt.Sprintf("%d", c)
			} else {
				concurrentStr = fmt.Sprintf("%s,%d", concurrentStr, c)
			}
		}

		for _, r := range responseTimeList {
			if responseTimeStr == "" {
				responseTimeStr = fmt.Sprintf("%.4f", r)
			} else {
				responseTimeStr = fmt.Sprintf("%s,%.4f", responseTimeStr, r)
			}
		}

		for _, pm := range podNumList {

			if podNumStr == "" {
				podNumStr = fmt.Sprintf("%.2f", pm)
			} else {
				podNumStr = fmt.Sprintf("%s,%.2f", podNumStr, pm)
			}
		}

		jqueryjs, _ := base64.StdEncoding.DecodeString(c3.Jqueryjs)
		d3js, _ := base64.StdEncoding.DecodeString(c3.D3js)
		c3js, _ := base64.StdEncoding.DecodeString(c3.C3js)
		c3css, _ := base64.StdEncoding.DecodeString(c3.C3css)

		dh := DrawHtml{
			ConcurrentStr:   concurrentStr,
			ResponseTimeStr: responseTimeStr,
			PodNumStr:       podNumStr,
			JqueryJS:        string(jqueryjs),
			D3JS:            string(d3js),
			C3CSS:           string(c3css),
			C3JS:            string(c3js),
		}

		r, err := os.OpenFile(hl.SavePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err == nil {
			defer r.Close()

			// TMPLE
			tmpl, _ := template.New("index").Parse(c3.Index)
			_ = tmpl.Execute(r, dh)
		} else {
			glog.Errorf("write result to file error:%s ", err.Error())
		}
	}

	for result := range hl.ResultChan {
		concurrentList = append(concurrentList, result.Concurrent)
		responseTimeList = append(responseTimeList, result.ResponseTime)
		podNumList = append(podNumList, result.PodNum)

		wg.Add(1)
		go func() {
			defer wg.Done()
			drawHtml()
		}()
	}

	wg.Wait()
	drawHtml()

}
