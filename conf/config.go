package conf

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gookit/config/v2"
)

type AppConfig struct {
	Providers    []Provider `json:"providers"`
	Email        string     `json:"email"` //注册let's encrypt的邮箱
	Hosts        []Host     `json:"hosts"`
	WxNotify     *string    `mapstructure:"wx_notify" json:"wx_notify"` //企业微信通知URL
	ReloadScript *string    `mapstructure:"reload_script" json:"reload_script"`
	ExpireDays   int        `mapstructure:"expire_days" json:"expire_days"` //证书还剩多少天自动更新 默认15天
}

type Host struct {
	Name      string   `json:"name"`                                   //域名
	Provider  string   `json:"provider"`                               //云服务商
	AliCertId string   `mapstructure:"ali_cert_id" json:"ali_cert_id"` //证书ID 当前证书在阿里云的ID 如果不存在则不做更新，只上传到阿里云
	TxCertId  *string  `mapstructure:"tx_cert_id" json:"tx_cert_id"`   //证书ID 当前证书在腾讯云的ID 如果不存在则不做更新，只上传到腾讯云
	Exptime   string   `mapstructure:"exptime" json:"exptime"`         //证书过期时间
	SavePath  string   `json:"savepath"`                               //SSL证书保存路径
	DeployTo  []string `mapstructure:"deploy_to" json:"deploy_to"`     //部署到哪些服务器 tenxunCloud,aliyun
}

type Provider struct {
	Name string `json:"name"`
	AK   string `json:"ak"`
	SK   string `json:"sk"`
}

func InitConfig() *AppConfig {
	// 设置选项支持ENV变量解析：当获取的值为string类型时，会尝试解析其中的ENV变量
	config.WithOptions(config.ParseEnv)
	config.WithOptions(config.ParseTime)

	// 加载配置，可以同时传入多个文件
	err := config.LoadFiles("conf/config.json")
	if err != nil {
		panic(err)
	}

	var appConfig = &AppConfig{}
	err = config.BindStruct("", appConfig)
	if err != nil {
		panic(fmt.Errorf("unable to decode into struct, %v", err))
	}
	log.Println("load config success")
	return appConfig
}

func WriteConfig(appConfig *AppConfig) {
	data, err := json.MarshalIndent(appConfig, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling config: %v", err)
	}

	err = os.WriteFile("conf/config.json", data, 0644)
	if err != nil {
		log.Fatalf("Error writing config file: %v", err)
	}

	log.Println("Config written to conf/config.json")
}
