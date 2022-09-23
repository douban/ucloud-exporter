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
	cdnBackSourceCode    *prometheus.Desc
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

		cdnBackSourceCode: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "backsource_code"),
			"http回源状态码占比(%)",
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
	ch <- e.cdnBackSourceCode
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
			codeTotal      int
			code200Total   int
			code206Total   int
			code301Total   int
			code302Total   int
			code304Total   int
			code400Total   int
			code403Total   int
			code404Total   int
			code500Total   int
			code502Total   int
			code503Total   int
			code504Total   int
			resourceCodeTotal      int
			resourceCode200Total   int
			resourceCode206Total   int
			resourceCode301Total   int
			resourceCode302Total   int
			resourceCode304Total   int
			resourceCode400Total   int
			resourceCode403Total   int
			resourceCode404Total   int
			resourceCode500Total   int
			resourceCode502Total   int
			resourceCode503Total   int
			resourceCode504Total   int
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

		httpData := collector.RetrieveHttpCode(domain.DomainId, e.projectId, "edge", e.rangeTime, e.delayTime, e.client).HttpCodeDetail
		for _, point := range httpData {
			code200Total += point.Http2XX.Http200
			code206Total += point.Http2XX.Http206
			code301Total += point.Http3XX.Http301
			code302Total += point.Http3XX.Http302
			code304Total += point.Http3XX.Http304
			code400Total += point.Http4XX.Http400
			code403Total += point.Http4XX.Http403
			code404Total += point.Http4XX.Http404
			code500Total += point.Http5XX.Http500
			code502Total += point.Http5XX.Http502
			code503Total += point.Http5XX.Http503
			code504Total += point.Http5XX.Http504
		}
		codeTotal = code200Total + code206Total + code301Total + code302Total + code304Total + code400Total +
			code403Total + code404Total + code500Total + code502Total + code503Total + code504Total
		httpStatusCodes := make(map[string]float64)
		httpStatusCodes["200"] = float64(code200Total) / float64(codeTotal)
		httpStatusCodes["206"] = float64(code206Total) / float64(codeTotal)
		httpStatusCodes["2xx"] = float64(code200Total + code206Total) / float64(codeTotal)
		httpStatusCodes["301"] = float64(code301Total) / float64(codeTotal)
		httpStatusCodes["302"] = float64(code302Total) / float64(codeTotal)
		httpStatusCodes["304"] = float64(code304Total) / float64(codeTotal)
		httpStatusCodes["3xx"] = float64(code301Total + code302Total + code304Total) / float64(codeTotal)
		httpStatusCodes["400"] = float64(code400Total) / float64(codeTotal)
		httpStatusCodes["403"] = float64(code403Total) / float64(codeTotal)
		httpStatusCodes["404"] = float64(code404Total) / float64(codeTotal)
		httpStatusCodes["4xx"] = float64(code400Total + code403Total + code404Total) / float64(codeTotal)
		httpStatusCodes["500"] = float64(code500Total) / float64(codeTotal)
		httpStatusCodes["502"] = float64(code502Total) / float64(codeTotal)
		httpStatusCodes["503"] = float64(code503Total) / float64(codeTotal)
		httpStatusCodes["504"] = float64(code504Total) / float64(codeTotal)
		httpStatusCodes["5xx"] = float64(code500Total + code502Total + code503Total + code504Total) / float64(codeTotal)

		backSourceCodeData := collector.RetrieveHttpCode(domain.DomainId, e.projectId, "layer", e.rangeTime, e.delayTime, e.client).HttpCodeDetail
		for _, point := range backSourceCodeData {
			resourceCode200Total += point.Http2XX.Http200
			resourceCode206Total += point.Http2XX.Http206
			resourceCode301Total += point.Http3XX.Http301
			resourceCode302Total += point.Http3XX.Http302
			resourceCode304Total += point.Http3XX.Http304
			resourceCode400Total += point.Http4XX.Http400
			resourceCode403Total += point.Http4XX.Http403
			resourceCode404Total += point.Http4XX.Http404
			resourceCode500Total += point.Http5XX.Http500
			resourceCode502Total += point.Http5XX.Http502
			resourceCode503Total += point.Http5XX.Http503
			resourceCode504Total += point.Http5XX.Http504
		}
		resourceCodeTotal = resourceCode200Total + resourceCode206Total + resourceCode301Total + resourceCode302Total +
			resourceCode304Total + resourceCode400Total + resourceCode403Total + resourceCode404Total +
			resourceCode500Total + resourceCode502Total + resourceCode503Total + resourceCode504Total
		backSourceStatusCodes := make(map[string]float64)
		backSourceStatusCodes["200"] = float64(resourceCode200Total) / float64(resourceCodeTotal)
		backSourceStatusCodes["206"] = float64(resourceCode206Total) / float64(resourceCodeTotal)
		backSourceStatusCodes["2xx"] = float64(resourceCode200Total + resourceCode206Total) / float64(resourceCodeTotal)
		backSourceStatusCodes["301"] = float64(resourceCode301Total) / float64(resourceCodeTotal)
		backSourceStatusCodes["302"] = float64(resourceCode302Total) / float64(resourceCodeTotal)
		backSourceStatusCodes["304"] = float64(resourceCode304Total) / float64(resourceCodeTotal)
		backSourceStatusCodes["3xx"] = float64(resourceCode301Total + resourceCode302Total + resourceCode304Total) / float64(resourceCodeTotal)
		backSourceStatusCodes["400"] = float64(resourceCode400Total) / float64(resourceCodeTotal)
		backSourceStatusCodes["403"] = float64(resourceCode403Total) / float64(resourceCodeTotal)
		backSourceStatusCodes["404"] = float64(resourceCode404Total) / float64(resourceCodeTotal)
		backSourceStatusCodes["4xx"] = float64(resourceCode400Total + resourceCode403Total + resourceCode404Total) / float64(resourceCodeTotal)
		backSourceStatusCodes["500"] = float64(resourceCode500Total) / float64(resourceCodeTotal)
		backSourceStatusCodes["502"] = float64(resourceCode502Total) / float64(resourceCodeTotal)
		backSourceStatusCodes["503"] = float64(resourceCode503Total) / float64(resourceCodeTotal)
		backSourceStatusCodes["504"] = float64(resourceCode504Total) / float64(resourceCodeTotal)
		backSourceStatusCodes["5xx"] = float64(resourceCode500Total + resourceCode502Total + resourceCode503Total + resourceCode504Total) / float64(resourceCodeTotal)

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

		for status, rate := range httpStatusCodes {
			proportion, err := strconv.ParseFloat(fmt.Sprintf("%.3f", rate * 100), 64)
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

		for status, rate := range backSourceStatusCodes {
			proportion, err := strconv.ParseFloat(fmt.Sprintf("%.3f", rate * 100), 64)
			if err != nil {
				log.Fatal(err)
			}
			ch <- prometheus.MustNewConstMetric(
				e.cdnBackSourceCode,
				prometheus.GaugeValue,
				proportion,
				domain.Domain,
				status,
			)
		}
	}

}

