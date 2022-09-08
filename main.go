package main

import (
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/ucloud/ucloud-sdk-go/services/ucdn"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
	"ucloud-exporter/collector"
	"ucloud-exporter/exporter"
)

const (
	CONFIG_URL = "https://api.ucloud.cn"
)

var (
	domainList []ucdn.DomainBaseInfo
)

func FetchDomainList(projectID string, client *ucdn.UCDNClient) {
	domainList = collector.RetrieveInfoList(projectID, client).DomainInfoList
	if len(domainList) == 0{
		log.Fatal("domainList is empty")
	}
}

func main() {
	publicKey := flag.String("pubkey", os.Getenv("UCLOUD_PUB_KEY"), "ucloud api public key")
	privateKey := flag.String("privatekey", os.Getenv("UCLOUD_PRIVATE_KEY"), "ucloud api private key")
	projectid := flag.String("projectId", os.Getenv("ucloud_project_id"), "confirm the project")
	host := flag.String("host", "0.0.0.0", "服务监听地址")
	port := flag.Int("port", 9200, "服务监听端口")
	rangeTime := flag.Int64("rangeTime", 3000, "rangeTime")
	delayTime := flag.Int64("delayTime", 600, "delayTime")
	tickerTime := flag.Int("tickerTime", 10, "tickerTime")
	metricsPath := flag.String("metricsPath", "/metrics", "default metrics path")
	flag.Parse()
	cfg := ucloud.NewConfig()
	cfg.BaseUrl = CONFIG_URL

	cred := auth.NewCredential()
	cred.PublicKey = *publicKey
	cred.PrivateKey = *privateKey
	projectId := *projectid

	uCdnClient := ucdn.NewClient(&cfg, &cred)
	FetchDomainList(projectId, uCdnClient)
	ticker := time.NewTicker(time.Duration(*tickerTime) * time.Second)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				FetchDomainList(projectId, uCdnClient)
			}
		}
	}()

	cdn := exporter.CdnCloudExporter(&domainList, projectId, *rangeTime, *delayTime, uCdnClient)
	prometheus.MustRegister(cdn)

	listenAddress := net.JoinHostPort(*host, strconv.Itoa(*port))
	log.Println(listenAddress)
	log.Println("Running on", listenAddress)
	http.Handle(*metricsPath, promhttp.Handler()) //注册

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html>
             <head><title>UCloud CDN Exporter</title></head>
             <body>
             <h1>UCloud cdn exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})

	log.Fatal(http.ListenAndServe(listenAddress, nil))

}

