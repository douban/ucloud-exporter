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
	cdn95bandwidth       *prometheus.Desc
	cdnOriginBandwidth   *prometheus.Desc
	cdnRequestNum        *prometheus.Desc
	cdnOriginRequestNum  *prometheus.Desc
	cdnHttpCode          *prometheus.Desc
}

func CdnCloudExporter(domainList *[]ucdn.DomainBaseInfo, projectId string, rangeTime int64, delayTime int64, c *ucdn.UCDNClient) *CdnExporter {
	return &CdnExporter{
		client:     c,
		domainList: domainList,
		rangeTime:  rangeTime,
		delayTime:  delayTime,
		projectId:  projectId,
		cdnRequestHitRate: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "hit_rate"),
			"总请求命中率(%)",
			[]string{
				"instanceId",
			},
			nil,
		),

		cdnFlowHitRate: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "flux_hit_rate"),
			"总流量命中率(%)",
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

		cdn95bandwidth: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "95_band_width"),
			"95带宽数据(Mbps)",
			[]string{
				"instanceId",
			},
			nil,
		),

		cdnOriginBandwidth: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "backsource_band_width"),
			"回源带宽数据(Mbps)",
			[]string{
				"instanceId",
			},
			nil,
		),

		cdnRequestNum:  prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "request_num"),
			"请求数(Count)",
			[]string{
				"instanceId",
			},
			nil,
		),

		cdnOriginRequestNum:  prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "backsource_request_num"),
			"回源请求数(Count)",
			[]string{
				"instanceId",
			},
			nil,
		),

		cdnHttpCode: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "http_code"),
			"http状态码占比(%)",
			[]string{
				"instanceId",
				"status",
			},
			nil,
		),
	}
}

func (e *CdnExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.cdnRequestHitRate
	ch <- e.cdnFlowHitRate
	ch <- e.cdnBandWidth
	ch <- e.cdn95bandwidth
	ch <- e.cdnOriginBandwidth
	ch <- e.cdnRequestNum
	ch <- e.cdnOriginRequestNum
	ch <- e.cdnHttpCode
}

func (e *CdnExporter) Collect(ch chan<- prometheus.Metric) {

	for _, domain := range *e.domainList {

		var (
			requestHitRateSum       float64
			flowHitRateSum          float64
			bandWidthSum            float64
			bandWidthAverage        float64
			originBandWidthSum      float64
			originBandWidthAverage  float64
			requestNumSum           float64
			requestNumAverage       float64
			originRequestNumSum     float64
			originRequestNumAverage float64
			http1xxSum              int
			http2xxSum              int
			http3xxSum              int
			http4xxSum              int
			http5xxSum              int
			http6xxSum              int
			httpAverage             int
		)

		hitRateData := collector.RetrieveHitRate(domain.DomainId, e.projectId, e.rangeTime, e.delayTime, e.client).HitRateList
		// hitRateList 去掉最后两组数值不准确的数据，影响平均值计算，其他指标的原因相同
		// 示例：[{97.77 98.16 1662441600} {97.66 98.13 1662441900} {97.51 98.13 1662442200} {97.45 98.13 1662442500} {97.67 98.14 1662442800} {97.45 98.06 1662443100} {97.49 98.09 1662443400} {73.24 83.58 1662443700} {0 0 1662444000}]
		for _, point := range hitRateData {
			flowHitRateSum += point.FlowHitRate
			requestHitRateSum += point.RequestHitRate
		}
		flowHitRateAverage, err := strconv.ParseFloat(fmt.Sprintf("%.2f", flowHitRateSum/float64(len(hitRateData))), 64)
		if err != nil {
			log.Fatal(err)
		}
		requestHitRateAverage, err := strconv.ParseFloat(fmt.Sprintf("%.2f", requestHitRateSum/float64(len(hitRateData))), 64)
		if err != nil {
			log.Fatal(err)
		}

		bandWidthData := collector.RetrieveBandWidth(domain.DomainId, e.projectId, e.rangeTime, e.delayTime, e.client).BandwidthTrafficList
		for _, point := range bandWidthData {
			bandWidthSum += point.CdnBandwidth
		}
		bandWidthAverage, err = strconv.ParseFloat(fmt.Sprintf("%.2f", bandWidthSum/float64(len(bandWidthData))), 64)
		if err != nil {
			log.Fatal(err)
		}

		originBandWidth := collector.RetrieveOriginBandWidth(domain.DomainId, e.projectId, e.rangeTime, e.delayTime, e.client).BandwidthList
		for _, point := range originBandWidth {
			originBandWidthSum += point.Bandwidth
		}
		originBandWidthAverage, err = strconv.ParseFloat(fmt.Sprintf("%.2f", originBandWidthSum/float64(len(originBandWidth))), 64)
		if err != nil {
			log.Fatal(err)
		}

		requestNum := collector.RetrieveRequestNum(domain.DomainId, e.projectId, e.rangeTime, e.delayTime, e.client).RequestList
		for _, point := range requestNum{
			requestNumSum += point.CdnRequest
		}
		requestNumAverage, err = strconv.ParseFloat(fmt.Sprintf("%.2f", requestNumSum/float64(len(requestNum))), 64)
		if err != nil {
			log.Fatal(err)
		}

		originRequestNum := collector.RetrieveOriginRequestNum(domain.DomainId, e.projectId, e.rangeTime, e.delayTime, e.client).RequestList
		for _, point := range originRequestNum{
			originRequestNumSum += point.CdnRequest
		}
		originRequestNumAverage, err = strconv.ParseFloat(fmt.Sprintf("%.2f", originRequestNumSum/float64(len(originRequestNum))), 64)
		if err != nil {
			log.Fatal(err)
		}

		httpData := collector.RetrieveHttpCode(domain.DomainId, e.projectId, e.rangeTime, e.delayTime, e.client).HttpCodeDetail
		for _, point := range httpData {
			http1xxSum += point.Http1XX.Total
			http2xxSum += point.Http2XX.Total
			http3xxSum += point.Http3XX.Total
			http4xxSum += point.Http4XX.Total
			http5xxSum += point.Http5XX.Total
			http6xxSum += point.Http6XX.Total
		}
		httpStatusCodes := make(map[string]int)
		httpStatusCodes["1xx"] = http1xxSum / len(httpData)
		httpStatusCodes["2xx"] = http2xxSum / len(httpData)
		httpStatusCodes["3xx"] = http3xxSum / len(httpData)
		httpStatusCodes["4xx"] = http4xxSum / len(httpData)
		httpStatusCodes["5xx"] = http5xxSum / len(httpData)
		httpStatusCodes["6xx"] = http6xxSum / len(httpData)
		for _, sum := range httpStatusCodes{
			httpAverage += sum
		}

		ch <- prometheus.MustNewConstMetric(
			e.cdnRequestHitRate,
			prometheus.GaugeValue,
			requestHitRateAverage,
			domain.Domain,
		)

		ch <- prometheus.MustNewConstMetric(
			e.cdnFlowHitRate,
			prometheus.GaugeValue,
			flowHitRateAverage,
			domain.Domain,
		)

		ch <- prometheus.MustNewConstMetric(
			e.cdnBandWidth,
			prometheus.GaugeValue,
			bandWidthAverage,
			domain.Domain,
		)

		ch <- prometheus.MustNewConstMetric(
			e.cdn95bandwidth,
			prometheus.GaugeValue,
			collector.Retrieve95BandWidth(domain.DomainId, e.projectId, e.rangeTime, e.delayTime, e.client).CdnBandwidth,
			domain.Domain,
		)

		ch <- prometheus.MustNewConstMetric(
			e.cdnOriginBandwidth,
			prometheus.GaugeValue,
			originBandWidthAverage,
			domain.Domain,
		)

		ch <- prometheus.MustNewConstMetric(
			e.cdnRequestNum,
			prometheus.GaugeValue,
			requestNumAverage,
			domain.Domain,
		)

		ch <- prometheus.MustNewConstMetric(
			e.cdnOriginRequestNum,
			prometheus.GaugeValue,
			originRequestNumAverage,
			domain.Domain,
		)

		for status, num := range httpStatusCodes {
			proportion, err := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(num) / float64(httpAverage) * 100), 64)
			if err != nil {
				log.Fatal(err)
			}
			ch <- prometheus.MustNewConstMetric(
				e.cdnHttpCode,
				prometheus.GaugeValue,
				proportion,
				domain.Domain,
				status,
			)
		}
	}

}

