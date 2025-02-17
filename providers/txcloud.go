package providers

import (
	"autossl/conf"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/providers/dns/tencentcloud"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	ssl "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205"
)

type TXProvider struct {
	chp       challenge.Provider
	pro       conf.Provider
	appConfig *conf.AppConfig
}

func NewTXProvider(pro conf.Provider, appConfig *conf.AppConfig) (*TXProvider, error) {
	cfg := tencentcloud.NewDefaultConfig()
	cfg.SecretID = pro.AK
	cfg.SecretKey = pro.SK
	cfg.Region = "ap-guangzhou"
	p, err := tencentcloud.NewDNSProviderConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create aliyun provider: %v", err)
	}
	return &TXProvider{pro: pro, chp: p, appConfig: appConfig}, nil

}
func (p *TXProvider) UpdateSSL(ctx context.Context, host *conf.Host) error {
	return p.ApplySSL(ctx, host)
}

// 第一步申请证书
func (p *TXProvider) ApplySSL(ctx context.Context, host *conf.Host) (err error) {
	// 申请证书
	cert, key, err := ApplySSL(ctx, p.chp, p.appConfig.Email, host)
	if err != nil {
		return err
	}
	// 过期时间是当前日期加上90天
	host.Exptime = time.Now().AddDate(0, 0, 90).Format("2006-01-02")
	expTime := host.Exptime
	log.Println("证书过期时间为" + expTime)
	return DeployCertificates(ctx, host, p.appConfig, key, cert)
}

// 第二步更新证书 todo
func (p *TXProvider) DeployToCloud(ctx context.Context, host conf.Host, priKey, pubKey *string) error {
	// 部署到腾讯云
	//1. 上传证书
	// 实例化一个认证对象，入参需要传入腾讯云账户 SecretId 和 SecretKey，此处还需注意密钥对的保密
	// 代码泄露可能会导致 SecretId 和 SecretKey 泄露，并威胁账号下所有资源的安全性。以下代码示例仅供参考，建议采用更安全的方式来使用密钥，请参见：https://cloud.tencent.com/document/product/1278/85305
	// 密钥可前往官网控制台 https://console.cloud.tencent.com/cam/capi 进行获取
	credential := common.NewCredential(
		p.pro.AK,
		p.pro.SK,
	)
	// 实例化一个client选项，可选的，没有特殊需求可以跳过
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "ssl.tencentcloudapi.com"
	// 实例化要请求产品的client对象,clientProfile是可选的
	client, _ := ssl.NewClient(credential, "", cpf)

	var certType = "SVR"
	// 实例化一个请求对象,每个接口都会对应一个request对象
	request := ssl.NewUploadCertificateRequest()
	request.CertificatePublicKey = pubKey
	request.CertificatePrivateKey = priKey
	request.CertificateType = &certType
	//alias 名称 host + 今天日期+ 过期时间
	expTime := host.Exptime
	request.Alias = common.StringPtr(host.Name + "_Exp_" + expTime)
	log.Println("开始上传" + host.Name + "的SSL证书到腾讯云" + "有效期至" + expTime)
	// 返回的resp是一个UploadCertificateResponse的实例，与请求对象对应
	response, err := client.UploadCertificate(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Printf("An API error has returned: %s", err)
		return err
	}
	if err != nil {
		return err
	}
	// 输出json格式的字符串回包
	fmt.Printf("%s", response.ToJsonString())
	log.Println("上传成功 证书ID" + *response.Response.CertificateId)
	//2. 部署证书
	if host.TxCertId == nil {
		log.Println("host.TxCertId is nil, not update ssl")
		return nil
	}
	updateSSLRequest := ssl.NewUpdateCertificateInstanceRequest()
	updateSSLRequest.OldCertificateId = host.TxCertId
	updateSSLRequest.CertificateId = response.Response.CertificateId
	updateSSLRequest.ExpiringNotificationSwitch = common.Uint64Ptr(0) //0:不忽略通知。
	//更新所有可能的资源类型
	updateSSLRequest.ResourceTypes = []*string{
		common.StringPtr("clb"),
		common.StringPtr("cdn"),
		common.StringPtr("waf"),
		common.StringPtr("live"),
		common.StringPtr("ddos"),
		common.StringPtr("teo"),
		common.StringPtr("apigateway"),
		common.StringPtr("vod"),
		common.StringPtr("tke"),
		common.StringPtr("tcb"),
		common.StringPtr("tse"),
		common.StringPtr("cos"),
	}
	log.Printf("重新部署证书时参数: OldCertificateId=%s, CertificateId=%s, ResourceTypes=%v \n",
		*updateSSLRequest.OldCertificateId, *updateSSLRequest.CertificateId, updateSSLRequest.ResourceTypes)
	// 返回的resp是一个UpdateCertificateInstanceResponse的实例，与请求对象对应
	updateResponse, err := client.UpdateCertificateInstance(updateSSLRequest)
	host.TxCertId = response.Response.CertificateId //更新证书ID
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Printf("An API error has returned: %s", err)
		return err
	}
	if err != nil {
		panic(err)
	}
	// 输出json格式的字符串回包
	fmt.Printf("%s \n", updateResponse.ToJsonString())
	return nil
}
