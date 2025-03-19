package providers

import (
	"autossl/conf"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
)

type CFProvider struct {
	chp       challenge.Provider
	pro       conf.Provider
	appConfig *conf.AppConfig
}

func (p *CFProvider) DeployToCloud(ctx context.Context, host conf.Host, key, cert *string) error {
	//无需实现
	return nil
}

func NewCFProvider(pro conf.Provider, appConfig *conf.AppConfig) (*CFProvider, error) {
	cfg := cloudflare.NewDefaultConfig()
	cfg.AuthEmail = pro.AK
	cfg.AuthKey = pro.SK
	p, err := cloudflare.NewDNSProviderConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloudflare provider: %v", err)
	}
	return &CFProvider{pro: pro, chp: p, appConfig: appConfig}, nil
}

func (p *CFProvider) UpdateSSL(ctx context.Context, host *conf.Host) error {
	return p.ApplySSL(ctx, host)
}

// 第一步申请证书
func (p *CFProvider) ApplySSL(ctx context.Context, host *conf.Host) (err error) {
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
