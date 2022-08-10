package exporter

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/ucloud/ucloud-sdk-go/services/ucdn"
	"log"
	"strconv"
	"ucloud-exporter/collector"
)

const cdnNameSpace = "uCloud"

type CdnExporter struct {
	client               *ucdn.UCDNClient
	domainList           *[]ucdn.DomainBaseInfo
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

func CdnCloudExporter(domainList *[]ucdn.DomainBaseInfo, projectId string, rangeTime int64, delayTime int64, c *ucdn.UCDNClient) *CdnExporter {
	return &CdnExporter{
		client:     c,
		domainList: domainList,
		rangeTime:  rangeTime,
		delayTime:  delayTime,
		projectId:  projectId,
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

	for _, domain := range *e.domainList {

		var requestHitRateData float64
		var flowHitRateData float64
		var bandWidthData float64
		var bandWidthAverage float64
		var http4xxData int
		var http5xxData int
		var http4xxAverage int
		var http5xxAverage int

		hitRateList := collector.RetrieveHitRate(domain.DomainId, e.projectId, e.rangeTime, e.delayTime, e.client).HitRateList
		for _, point := range hitRateList {
			flowHitRateData += point.FlowHitRate
			requestHitRateData += point.RequestHitRate
		}
		flowHitRateAverage, err := strconv.ParseFloat(fmt.Sprintf("%.2f", flowHitRateData/float64(len(hitRateList))), 64)
		if err != nil {
			log.Fatal(err)
		}
		requestHitRateAverage, err := strconv.ParseFloat(fmt.Sprintf("%.2f", requestHitRateData/float64(len(hitRateList))), 64)
		if err != nil {
			log.Fatal(err)
		}
		bandWidthList := collector.RetrieveBandWidth(domain.DomainId, e.projectId, e.rangeTime, e.delayTime, e.client).BandwidthList
		for _, point := range bandWidthList {
			bandWidthData += point.CdnBandwidth
		}
		bandWidthAverage, err = strconv.ParseFloat(fmt.Sprintf("%.2f", bandWidthData/float64(len(bandWidthList))), 64)
		if err != nil {
			log.Fatal(err)
		}

		httpList := collector.RetrieveOriginHttpCode4xx(domain.DomainId, e.projectId, e.rangeTime, e.delayTime, e.client).HttpCodeDetail
		for _, point := range httpList {
			http4xxData += point.Http4XX.Total
			http5xxData += point.Http5XX.Total
		}
		http4xxAverage = http4xxData / len(httpList)
		http5xxAverage = http5xxData / len(httpList)

		ch <- prometheus.MustNewConstMetric(
			e.cdnRequestHitRate,
			prometheus.GaugeValue,
			requestHitRateAverage,
			domain.Domain,
		)

		ch <- prometheus.MustNewConstMetric(
			e.cdnBandWidth,
			prometheus.GaugeValue,
			bandWidthAverage,
			domain.Domain,
		)

		ch <- prometheus.MustNewConstMetric(
			e.cdnOriginHttpCode4xx,
			prometheus.GaugeValue,
			float64(http4xxAverage),
			domain.Domain,
		)

		ch <- prometheus.MustNewConstMetric(
			e.cdnOriginHttpCode5xx,
			prometheus.GaugeValue,
			float64(http5xxAverage),
			domain.Domain,
		)

		ch <- prometheus.MustNewConstMetric(
			e.cdn95bandwidth,
			prometheus.GaugeValue,
			collector.Retrieve95BandWidth(domain.DomainId, e.projectId, e.rangeTime, e.delayTime, e.client).CdnBandwidth,
			domain.Domain,
		)

		ch <- prometheus.MustNewConstMetric(
			e.cdnFlowHitRate,
			prometheus.GaugeValue,
			flowHitRateAverage,
			domain.Domain,
		)
	}

}

