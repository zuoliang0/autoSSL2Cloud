package main

import (
	"autossl/conf"
	"autossl/providers"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

var appConfig *conf.AppConfig

func main() {
	appConfig = conf.InitConfig()
	// Check and update SSL certificates
	checkSSLUpdate()
}

func checkSSLUpdate() {
	var updateCount int = 0
	var updatedHosts []string
	for _, host := range appConfig.Hosts {
		// 检查证书是否过期
		if host.Exptime.Before(time.Now().AddDate(0, 0, appConfig.ExpireDays)) {
			// 证书过期，更新证书
			log.Printf("host %s ssl cert is expired, update it\n", host.Name)
			provider, err := providers.GetProvider(host.Provider, appConfig)
			if err != nil {
				log.Printf("get provider %s failed: %v\n", host.Provider, err)
				continue
			}
			var ctx = context.Background()
			provider.UpdateSSL(ctx, host)
			updateCount++
			updatedHosts = append(updatedHosts, host.Name)
		}
	}
	log.Printf("update %d hosts ssl cert\n", updateCount)
	if updateCount > 0 {
		// 触发更新Nginx配置
		reloadLocalNginxConfig()
		// 触发邮件通知
		sendQYWXAlert(updatedHosts)
		conf.WriteConfig(appConfig)
	} else {
		log.Println("no need to update")
	}
}

func sendQYWXAlert(updatedHosts []string) {
	if appConfig.WxNotify == nil || len(updatedHosts) == 0 {
		return
	}
	message := fmt.Sprintf("共更新<font color=\"warning\">%d</font>个域名。\n%s", len(updatedHosts), strings.Join(updatedHosts, "\n"))
	payload := fmt.Sprintf(`{"msgtype": "markdown", "markdown": {"content": "%s"}}`, message)
	req, err := http.NewRequest("POST", *appConfig.WxNotify, strings.NewReader(payload))
	if err != nil {
		log.Printf("failed to create request for WeChat notification: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("failed to send WeChat notification: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Printf("WeChat notification failed with status %d: %s\n", resp.StatusCode, body)
	} else {
		log.Println("WeChat notification sent successfully")
	}
}

func reloadLocalNginxConfig() {
	// 更新本地Nginx配置
	if appConfig.ReloadScript != nil {
		log.Println("reload nginx config" + *appConfig.ReloadScript)
		cmd := exec.Command("sh", *appConfig.ReloadScript)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("failed to reload nginx config: %v\n", err)
		} else {
			log.Printf("nginx config reloaded: %s\n", output)
		}
	}
}
