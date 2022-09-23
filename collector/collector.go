package collector

import (
	"github.com/ucloud/ucloud-sdk-go/services/ucdn"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"log"
	"time"
)

func RetrieveHitRate(domainId string, projectId string, rangeTime int64, delayTime int64, client *ucdn.UCDNClient) (response *ucdn.GetUcdnDomainHitRateResponse) {
	req := client.NewGetUcdnDomainHitRateRequest()
	req.DomainId = []string{
		domainId,
	}
	req.ProjectId = ucloud.String(projectId)
	req.Type = ucloud.Int(0)
	req.BeginTime = ucloud.Int(int(time.Now().Unix() - rangeTime))
	req.EndTime = ucloud.Int(int(time.Now().Unix() - delayTime))
	req.Areacode = ucloud.String("cn")

	newUCdnHitRate, err := client.GetUcdnDomainHitRate(req)
	if err != nil {
		log.Fatal(err)
	}
	return newUCdnHitRate
}

func RetrieveBandWidth(domainId string, projectId string, rangeTime int64, delayTime int64, client *ucdn.UCDNClient) (response *ucdn.GetUcdnDomainBandwidthV2Response) {
	req := client.NewGetUcdnDomainBandwidthV2Request()
	req.ProjectId = ucloud.String(projectId)
	req.DomainId = []string{
		domainId,
	}
	req.Type = ucloud.Int(0)
	req.BeginTime = ucloud.Int(int(time.Now().Unix() - rangeTime))
	req.EndTime = ucloud.Int(int(time.Now().Unix() - delayTime))
	req.Areacode = ucloud.String("cn")

	newUCdnBandWidth, err := client.GetUcdnDomainBandwidthV2(req)
	if err != nil {
		log.Fatal(err)
	}

	return newUCdnBandWidth
}

func Retrieve95BandWidth(domainId string, projectId string, rangeTime int64, delayTime int64, client *ucdn.UCDNClient) (response *ucdn.GetUcdnDomain95BandwidthV2Response) {
	req := client.NewGetUcdnDomain95BandwidthV2Request()
	req.DomainId = []string{
		domainId,
	}
	req.ProjectId = ucloud.String(projectId)
	req.BeginTime = ucloud.Int(int(time.Now().Unix() - rangeTime))
	req.EndTime = ucloud.Int(int(time.Now().Unix() - delayTime))
	req.Areacode = ucloud.String("cn")

	newUCdn95BandWidth, err := client.GetUcdnDomain95BandwidthV2(req)

	if err != nil {
		log.Fatal(err)
	}

	return newUCdn95BandWidth
}

func RetrieveOriginBandWidth(domainId string, projectId string, rangeTime int64, delayTime int64, client *ucdn.UCDNClient) (response *ucdn.GetUcdnPassBandwidthV2Response) {
	req := client.NewGetUcdnPassBandwidthV2Request()
	req.DomainId = []string{domainId}
	req.ProjectId = ucloud.String(projectId)
	req.Type = ucloud.Int(0)
	req.BeginTime = ucloud.Int(int(time.Now().Unix() - rangeTime))
	req.EndTime = ucloud.Int(int(time.Now().Unix() - delayTime))
	req.Areacode = ucloud.String("cn")

	bindWidth, err := client.GetUcdnPassBandwidthV2(req)
	if err != nil {
		log.Fatal(err)
	}

	return bindWidth
}

func RetrieveRequestNum(domainId string, projectId string, rangeTime int64, delayTime int64, client *ucdn.UCDNClient) (response *ucdn.GetUcdnDomainRequestNumV3Response) {
	req := client.NewGetUcdnDomainRequestNumV3Request()
	req.DomainId = []string{domainId}
	req.ProjectId = ucloud.String(projectId)
	req.Type = ucloud.Int(0)
	req.BeginTime = ucloud.Int(int(time.Now().Unix() - rangeTime))
	req.EndTime = ucloud.Int(int(time.Now().Unix() - delayTime))
	req.Areacode = ucloud.String("cn")
	requestNum, err := client.GetUcdnDomainRequestNumV3(req)

	if err != nil {
		log.Fatal(err)
	}
	return requestNum
}

func RetrieveOriginRequestNum(domainId string, projectId string, rangeTime int64, delayTime int64, client *ucdn.UCDNClient) (response *ucdn.GetUcdnDomainOriginRequestNumResponse) {
	req := client.NewGetUcdnDomainOriginRequestNumRequest()
	req.DomainId = []string{domainId}
	req.ProjectId = ucloud.String(projectId)
	req.Type = ucloud.Int(0)
	req.BeginTime = ucloud.Int(int(time.Now().Unix() - rangeTime))
	req.EndTime = ucloud.Int(int(time.Now().Unix() - delayTime))
	req.Areacode = ucloud.String("cn")
	originRequestNum, err := client.GetUcdnDomainOriginRequestNum(req)

	if err != nil {
		log.Fatal(err)
	}
	return originRequestNum
}

func RetrieveHttpCode(domainId string, projectId string, layer string, rangeTime int64, delayTime int64, client *ucdn.UCDNClient) (response *ucdn.GetUcdnDomainHttpCodeV2Response) {
	// layer 指定获取的状态码是边缘还是上层 edge 表示边缘 layer 表示上层
	req := client.NewGetUcdnDomainHttpCodeV2Request()
	req.DomainId = []string{
		domainId,
	}
	req.ProjectId = ucloud.String(projectId)
	req.Type = ucloud.Int(0)
	req.BeginTime = ucloud.Int(int(time.Now().Unix() - rangeTime))
	req.EndTime = ucloud.Int(int(time.Now().Unix() - delayTime))
	req.Layer = ucloud.String(layer)
	req.Areacode = ucloud.String("cn")

	newUCdnRequestStatusNum, err := client.GetUcdnDomainHttpCodeV2(req)
	if err != nil {
		log.Fatal(err)
	}

	return newUCdnRequestStatusNum
}

func RetrieveInfoList(projectId string, client *ucdn.UCDNClient) (response *ucdn.GetUcdnDomainInfoListResponse) {
	req := client.NewGetUcdnDomainInfoListRequest()
	req.ProjectId = ucloud.String(projectId)
	newUCdnInfoList, err := client.GetUcdnDomainInfoList(req)

	if err != nil {
		log.Fatal(err)
	}

	return newUCdnInfoList
}
