## autoSSL2Cloud
这是一款基于Go的自动证书监控以及上传并部署到云的程序。
### 功能介绍
程序可以基于配置文件去检测证书是否过期，目前支持 **腾讯云** **阿里云** **cloudflare**申请证书，并将其上传到阿里云（还不支持），腾讯云，并且可以使用API一键部署（通过替换旧证书实现）到CLB，CDN，COS等上。
在部署完成后还可以将证书复制到本地文件夹，并触发脚本来更新本地的nginx的证书。
然后还可以通过企业微信机器人来通知一共更新了多少个证书

### 使用方法

1. 配置文件

``` json
{
  "email":"youremail@gmail.com", // let's encrypt的邮箱
  "providers": [
    {
      "name": "aliyun",
      "ak": "${ALIYUN_AK | AKID*********************}",//配置支持从环境变量中读取配置
      "sk":  "${ALIYUN_SK | qYzV*********************}"
    },
    {
      "name": "txcloud",
      "ak":  "${TENCENT_AK | AKID*********************}",
      "sk":  "${TENCENT_SK | qYzV*********************}"
    },
    {
      "name":"cloudflare", 
      "ak":"${CLOUD_FLARE_MAIL | xxx@xxx.com}",//等于 AuthEmail
      "sk":"${CLOUD_FLARE_KEY | qYzV*********************}" //等于 AuthKey
    }
  ],
  "hosts": [
    {
      "name": "test.example.com",//域名 可以是*.xxx.com
      "provider": "tencent",//服务商，名字和上面一致
      "exptime": "2024-12-31",//当前过期时间
      "savepath": "/etc/ssl/test.example.com/",//证书要存放到本地的哪一个路径上，以/结尾 以域名_server.key和域名_server.pem 来存储私钥和公钥文件
      "tx_cert_id":"xxxxxxxx",//当前证书在腾讯云上的ID，如果没有这个配置，第一次更新时只会上传到腾讯云，但是不会自动部署
      "deploy_to":[
        "txcloud" //目前只支持部署到腾讯云
      ]
    },
    {
      "name": "test.com",
      "provider": "aliyun",
      "exptime": "2021-12-12",
      "savepath": "/etc/ssl/test.com",
       "deploy_to":[
        "txcloud"
      ]
    },
    
  ],
  "wx_notify":"https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxxx-xxxx-xxxx-xxxx",
  "reload_script": "/etc/init.d/nginx reload",
  "expire_days": 15
}
}
### 腾讯云更新证书方案
会调用UpdateCertificateInstance API，将老证书的下面例举的功能的所有可用区镜像更新
``` go
// Define the regions array
	regions := []*string{
		common.StringPtr("ap-guangzhou"),
		common.StringPtr("ap-shanghai"),
		common.StringPtr("ap-beijing"),
		common.StringPtr("ap-hongkong"),
		common.StringPtr("ap-singapore"),
		common.StringPtr("na-siliconvalley"),
		common.StringPtr("eu-frankfurt"),
		common.StringPtr("na-ashburn"),
		common.StringPtr("ap-bangkok"),
		common.StringPtr("ap-tokyo"),
		common.StringPtr("ap-nanjing"),
		common.StringPtr("ap-jakarta"),
		common.StringPtr("sa-saopaulo"),
		common.StringPtr("ap-chongqing"),
		common.StringPtr("ap-seoul"),
	}

	// 更新常用产品的所有可用区的证书
	updateSSLRequest.ResourceTypesRegions = []*ssl.ResourceTypeRegions{
		{ResourceType: common.StringPtr("clb"), Regions: regions},
		{ResourceType: common.StringPtr("cdn"), Regions: regions},
		{ResourceType: common.StringPtr("waf"), Regions: regions},
		{ResourceType: common.StringPtr("live"), Regions: regions},
		{ResourceType: common.StringPtr("ddos"), Regions: regions},
		{ResourceType: common.StringPtr("teo"), Regions: regions},
		{ResourceType: common.StringPtr("apigateway"), Regions: regions},
		{ResourceType: common.StringPtr("vod"), Regions: regions},
		{ResourceType: common.StringPtr("tke"), Regions: regions},
		{ResourceType: common.StringPtr("tcb"), Regions: regions},
		{ResourceType: common.StringPtr("tse"), Regions: regions},
		{ResourceType: common.StringPtr("cos"), Regions: regions},
	}
  ```

### 运行方法

``` sh
# 每天上午9点执行 autossl
0 9 * * * /path/to/autossl >> /var/log/autossl.log 2>&1
```

## 打包命令
```sh
set GOARCH=amd64
go env -w GOARCH=amd64
set GOOS=linux
go env -w GOOS=linux

go build -o autossl
```


### todo
1. 阿里云证书部署支持
2. 定时任务支持