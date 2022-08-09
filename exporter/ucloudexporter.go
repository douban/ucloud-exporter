package exporter

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/ucloud/ucloud-sdk-go/services/ucdn"
	"strconv"
	"ucloud-exporter/collector"
)

const cdnNameSpace = "uCloud"

type CdnExporter struct {
	client               *ucdn.UCDNClient
	infoList             [20]*string
	rangeTime            int64
	delayTime            int64
	projectId            string
	cdnRequestHitRate    *prometheus.Desc
	cdnFlowHitRate       *prometheus.Desc
	cdnBandWidth         *prometheus.Desc
	cdnOriginHttpCode4xx *prometheus.Desc
	cdnOriginHttpCode5xx *prometheus.Desc
	cdn95bandwidth       *prometheus.Desc
}

func CdnCloudExporter(infoList [20]*string, projectId string, rangeTime int64, delayTime int64, c *ucdn.UCDNClient) *CdnExporter {
	return &CdnExporter{
		client:    c,
		infoList:  infoList,
		rangeTime: rangeTime,
		delayTime: delayTime,
		projectId: projectId,
		cdnRequestHitRate: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "request_hit_rate"),
			"总请求命中率(%)",
			[]string{
				"instanceId",
			},
			nil,
		),

		cdnBandWidth: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "band_width"),
			"域名带宽(Mbps)",
			[]string{
				"instanceId",
			},
			nil,
		),

		cdnOriginHttpCode4xx: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "http_code_4XX"),
			"http4XX请求数(Count)",
			[]string{
				"instanceId",
			},
			nil,
		),

		cdnOriginHttpCode5xx: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "http_code_5XX"),
			"http5XX请求数(Count)",
			[]string{
				"instanceId",
			},
			nil,
		),

		cdn95bandwidth: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "95_band_width"),
			"95带宽数据(Mbps)",
			[]string{
				"instanceId",
			},
			nil,
		),

		cdnFlowHitRate: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "flow_hit_rate"),
			"总流量命中率(%)",
			[]string{
				"instanceId",
			},
			nil,
		),
	}
}

func (e *CdnExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.cdnRequestHitRate
	ch <- e.cdnFlowHitRate
	ch <- e.cdnBandWidth
	ch <- e.cdnOriginHttpCode4xx
	ch <- e.cdn95bandwidth
	ch <- e.cdnOriginHttpCode5xx
}

func (e *CdnExporter) Collect(ch chan<- prometheus.Metric) {
	var infoCount int
	infoCount = collector.RetrieveInfoList(e.projectId, e.client).TotalCount

	var requestHitRateData float64
	var requestHitRateCount float64
	var requestHitRateAverage float64
	var flowHitRateData float64
	var flowHitRateCount float64
	var flowHitRateAverage float64

	hitRateList := collector.RetrieveHitRate(e.projectId, e.rangeTime, e.delayTime, e.client).HitRateList
	flowHitRateCount = float64(len(hitRateList))
	requestHitRateCount = float64(len(hitRateList))

	for _, point := range hitRateList {
		flowHitRateData += point.FlowHitRate
		requestHitRateData += point.RequestHitRate
	}
	flowHitRateAverage = flowHitRateData / flowHitRateCount
	flowHitRateAverage, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", flowHitRateAverage), 64)
	requestHitRateAverage = requestHitRateData / requestHitRateCount
	requestHitRateAverage, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", requestHitRateAverage), 64)

	var bandWidthData float64
	var bandWidthCount float64
	var bandWidthAverage float64
	bandWidthList := collector.RetrieveBandWidth(e.projectId, e.rangeTime, e.delayTime, e.client).BandwidthList
	bandWidthCount = float64(len(bandWidthList))

	for _, point := range bandWidthList {
		bandWidthData += point.CdnBandwidth
	}

	bandWidthAverage = bandWidthData / bandWidthCount
	bandWidthAverage, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", bandWidthAverage), 64)

	var httpCount int
	var http4xxData int
	var http4xxAverage int
	var http5xxData int
	var http5xxAverage int

	httpList := collector.RetrieveOriginHttpCode4xx(e.projectId, e.rangeTime, e.delayTime, e.client).HttpCodeDetail
	httpCount = len(httpList)

	for _, point := range httpList {
		http4xxData += point.Http4XX.Total
		http5xxData += point.Http5XX.Total
	}
	http4xxAverage = http4xxData / httpCount
	http5xxAverage = http5xxData / httpCount

	for i := 0; i < infoCount; i++ {
		ch <- prometheus.MustNewConstMetric(
			e.cdnRequestHitRate,
			prometheus.GaugeValue,
			requestHitRateAverage,
			*e.infoList[i],
		)

		ch <- prometheus.MustNewConstMetric(
			e.cdnBandWidth,
			prometheus.GaugeValue,
			bandWidthAverage,
			*e.infoList[i],
		)

		ch <- prometheus.MustNewConstMetric(
			e.cdnOriginHttpCode4xx,
			prometheus.GaugeValue,
			float64(http4xxAverage),
			*e.infoList[i],
		)

		ch <- prometheus.MustNewConstMetric(
			e.cdnOriginHttpCode5xx,
			prometheus.GaugeValue,
			float64(http5xxAverage),
			*e.infoList[i],
		)

		ch <- prometheus.MustNewConstMetric(
			e.cdn95bandwidth,
			prometheus.GaugeValue,
			collector.Retrieve95BandWidth(e.projectId, e.rangeTime, e.delayTime, e.client).CdnBandwidth,
			*e.infoList[i],
		)

		ch <- prometheus.MustNewConstMetric(
			e.cdnFlowHitRate,
			prometheus.GaugeValue,
			flowHitRateAverage,
			*e.infoList[i],
		)
	}

}

