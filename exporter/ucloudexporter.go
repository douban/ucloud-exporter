package exporter

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/ucloud/ucloud-sdk-go/services/ucdn"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"ucloud-exporter/collector"
)

const cdnNameSpace = "uCloud"

type CdnExporter struct {
	client              *ucdn.UCDNClient
	domainList          *[]ucdn.DomainBaseInfo
	rangeTime           int64
	delayTime           int64
	projectId           string
	cdnRequestHitRate   *prometheus.Desc
	cdnFlowHitRate      *prometheus.Desc
	cdnBandWidth        *prometheus.Desc
	cdn95bandwidth      *prometheus.Desc
	cdnOriginBandwidth  *prometheus.Desc
	cdnRequestNum       *prometheus.Desc
	cdnOriginRequestNum *prometheus.Desc
	cdnHttpCode         *prometheus.Desc
	cdnBackSourceCode   *prometheus.Desc
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

		cdnRequestNum: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "request_num"),
			"请求数(Count)",
			[]string{
				"instanceId",
			},
			nil,
		),

		cdnOriginRequestNum: prometheus.NewDesc(
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
			codeTotal               int
			resourceCodeTotal       int
			httpStatusCodes         map[string]float64
			backSourceStatusCodes   map[string]float64
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
		for _, point := range requestNum {
			requestNumSum += point.CdnRequest
		}
		requestNumAverage, err = strconv.ParseFloat(fmt.Sprintf("%.2f", requestNumSum/float64(len(requestNum))), 64)
		if err != nil {
			log.Fatal(err)
		}

		originRequestNum := collector.RetrieveOriginRequestNum(domain.DomainId, e.projectId, e.rangeTime, e.delayTime, e.client).RequestList
		for _, point := range originRequestNum {
			originRequestNumSum += point.CdnRequest
		}
		originRequestNumAverage, err = strconv.ParseFloat(fmt.Sprintf("%.2f", originRequestNumSum/float64(len(originRequestNum))), 64)
		if err != nil {
			log.Fatal(err)
		}

		httpData := collector.RetrieveHttpCode(domain.DomainId, e.projectId, "edge", e.rangeTime, e.delayTime, e.client).HttpCodeDetail
		for _, point := range httpData {
			codeTotal += point.Http2XX.Total + point.Http3XX.Total +
				point.Http4XX.Total + point.Http5XX.Total
			httpStatusCodes["2xx"] += float64(point.Http2XX.Total)
			httpStatusCodes["3xx"] += float64(point.Http3XX.Total)
			httpStatusCodes["4xx"] += float64(point.Http4XX.Total)
			httpStatusCodes["5xx"] += float64(point.Http5XX.Total)
			pointValue := reflect.ValueOf(point)
			for i := 0; i < pointValue.NumField(); i++ {
				fieldValue := pointValue.Field(i)
				if fieldValue.Kind() == reflect.Struct {
					for j := 0; j < fieldValue.NumField(); j++ {
						subField := fieldValue.Type().Field(j)
						subFieldValue := fieldValue.Field(j)
						if subFieldValue.Int() == 0 {
							continue
						}
						statusCode := regexp.MustCompile(`\d+`).FindString(subField.Name)
						if statusCode != "" {
							httpStatusCodes[statusCode] += float64(subFieldValue.Int())
						}
					}
				}
			}
		}
		for code, count := range httpStatusCodes {
			httpStatusCodes[code] = count / float64(codeTotal)
		}

		backSourceCodeData := collector.RetrieveHttpCode(domain.DomainId, e.projectId, "layer", e.rangeTime, e.delayTime, e.client).HttpCodeDetail
		for _, point := range backSourceCodeData {
			resourceCodeTotal += point.Http2XX.Total + point.Http3XX.Total +
				point.Http4XX.Total + point.Http5XX.Total
			backSourceStatusCodes["2xx"] += float64(point.Http2XX.Total)
			backSourceStatusCodes["3xx"] += float64(point.Http3XX.Total)
			backSourceStatusCodes["4xx"] += float64(point.Http4XX.Total)
			backSourceStatusCodes["5xx"] += float64(point.Http5XX.Total)
			pointValue := reflect.ValueOf(point)
			for i := 0; i < pointValue.NumField(); i++ {
				fieldValue := pointValue.Field(i)
				if fieldValue.Kind() == reflect.Struct {
					for j := 0; j < fieldValue.NumField(); j++ {
						subField := fieldValue.Type().Field(j)
						subFieldValue := fieldValue.Field(j)
						if subFieldValue.Int() == 0 {
							continue
						}
						statusCode := regexp.MustCompile(`\d+`).FindString(subField.Name)
						if statusCode != "" {
							backSourceStatusCodes[statusCode] += float64(subFieldValue.Int())
						}
					}
				}
			}
		}
		for code, count := range backSourceStatusCodes {
			backSourceStatusCodes[code] = count / float64(resourceCodeTotal)
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

		for status, rate := range httpStatusCodes {
			proportion, err := strconv.ParseFloat(fmt.Sprintf("%.3f", rate*100), 64)
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
			proportion, err := strconv.ParseFloat(fmt.Sprintf("%.3f", rate*100), 64)
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
