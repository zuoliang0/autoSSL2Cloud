package providers

import (
	"autossl/conf"
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"log"
	"os"
	"strings"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
)

var providMap = map[string]BaseProvider{}

func GetProvider(providerName string, appConfig *conf.AppConfig) (BaseProvider, error) {
	for _, p := range appConfig.Providers {
		if _, ok := providMap[p.Name]; !ok {
			if p.Name == providerName {
				switch providerName {
				case "aliyun":
					pp, err := NewAliyunProvider(p, appConfig)
					if err != nil {
						return nil, err
					}
					providMap[p.Name] = pp
					return pp, nil
				case "txcloud":
					pp, err := NewTXProvider(p, appConfig)
					if err != nil {
						return nil, err
					}
					providMap[p.Name] = pp
					return pp, nil
				}
			}
		} else {
			return providMap[p.Name], nil
		}
	}
	return nil, errors.New("provider not found")
}

type BaseProvider interface {
	UpdateSSL(ctx context.Context, host conf.Host) error
	DeployToCloud(ctx context.Context, host conf.Host, key, cert *string) error
}

func DeployCertificates(ctx context.Context, host conf.Host, appConfig *conf.AppConfig, key, cert string) error {
	for _, deployTo := range host.DeployTo {
		switch deployTo {
		case "aliyun":
			// 部署到阿里云 todo
			aly, _ := GetProvider("aliyun", appConfig)
			if aly != nil {
				return aly.DeployToCloud(ctx, host, &key, &cert)
			}

		case "txcloud":
			// 部署到腾讯云
			tx, _ := GetProvider("txcloud", appConfig)
			return tx.DeployToCloud(ctx, host, &key, &cert)
		}

	}
	return nil
}

func ApplySSL(ctx context.Context, chp challenge.Provider, email string, host conf.Host) (cert, key string, err error) {
	// 申请证书
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}
	myUser := MyUser{
		Email: email,
		key:   privateKey,
	}
	config := lego.NewConfig(&myUser)
	config.Certificate.KeyType = certcrypto.RSA2048
	client, err := lego.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Challenge.SetDNS01Provider(chp)
	if err != nil {
		log.Fatal(err)
	}

	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		log.Fatal(err)
	}
	myUser.Registration = reg
	var domains = []string{host.Name}
	if strings.Contains(host.Name, "*.") {
		domains = []string{host.Name[2:], host.Name}
	}
	request := certificate.ObtainRequest{
		Domains: domains,
		Bundle:  true,
	}
	// 申请证书
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%#v\n", certificates)
	var domain = host.Name
	if strings.Contains(host.Name, "*.") {
		domain = host.Name[2:]
	}

	err = os.WriteFile(host.SavePath+domain+"_server.key", certificates.PrivateKey, os.ModePerm)
	if err != nil {
		log.Print(err)
	}
	log.Println("证书保存成功 路径为" + host.SavePath + domain + "_server.key")
	err = os.WriteFile(host.SavePath+domain+"_server.pem", certificates.Certificate, os.ModePerm)
	if err != nil {
		log.Print(err)
	}
	log.Println("证书保存成功 路径为" + host.SavePath + domain + "_server.pem")
	publicKey := string(certificates.Certificate)
	pk := string(certificates.PrivateKey)
	return publicKey, pk, nil
}

type MyUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *MyUser) GetEmail() string {
	return u.Email
}
func (u *MyUser) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *MyUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}
