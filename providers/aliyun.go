package providers

import (
	"autossl/conf"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/providers/dns/alidns"
)

type AliyunProvider struct {
	chp       challenge.Provider
	appConfig *conf.AppConfig
}

func NewAliyunProvider(pro conf.Provider, appConfig *conf.AppConfig) (*AliyunProvider, error) {
	cfg := alidns.NewDefaultConfig()
	cfg.APIKey = pro.AK
	cfg.SecretKey = pro.SK
	p, err := alidns.NewDNSProviderConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create aliyun provider: %v", err)
	}
	return &AliyunProvider{chp: p, appConfig: appConfig}, nil
}

func (p *AliyunProvider) UpdateSSL(ctx context.Context, host *conf.Host) error {
	return p.ApplySSL(ctx, host)
}

// 第一步申请证书
func (p *AliyunProvider) ApplySSL(ctx context.Context, host *conf.Host) (err error) {
	// 申请证书
	cert, key, err := ApplySSL(ctx, p.chp, p.appConfig.Email, host)
	if err != nil {
		return err
	}
	// 过期时间是当前日期加上90天
	host.Exptime = time.Now().AddDate(0, 0, 90).Format("2006-01-02")
	expTime := host.Exptime
	log.Println("证书过期时间为" + expTime)
	err = DeployCertificates(ctx, host, p.appConfig, key, cert)
	return err
}

// 第二步更新证书 todo
func (p *AliyunProvider) DeployToCloud(ctx context.Context, host conf.Host, priKey, pubKey *string) error {
	// 部署到阿里云 阿里云有点复杂，
	//https://help.aliyun.com/zh/ssl-certificate/developer-reference/api-cas-2020-04-07-listcloudresources?spm=a2c4g.11186623.help-menu-28533.d_4_3_2_2_9.bb8d46edOQLXAA
	// 创建证书部署任务需要提供云产品资源 ID。多个资源 ID 用半角逗号（,）
	// 查询所有的云产品资源 ID，再创建证书部署任务
	//1. 上传证书

	//2. 部署证书
	return nil
}
