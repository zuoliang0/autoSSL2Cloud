## autoSSL2Cloud
这是一款基于Go的自动证书监控以及上传并部署到云的程序。
### 功能介绍
程序可以基于配置文件去检测证书是否过期，目前支持 **腾讯云** **阿里云** **cloudflare**申请证书，并将其上传到阿里云，腾讯云，并且可以使用API一键部署到CLB或CDN上。
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
      "savepath": "/etc/ssl/test.example.com/",//证书要存放到本地的哪一个路径上，以/结尾
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

```