package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/hashicorp/go.net/proxy"
)

func main() {
	// 参数使用方式1: "socks5://127.0.0.1:10808" "https://dns.google/resolve?name=apn-a.point.mysynology.net&type=A,https://dns.google/resolve?name=apn-b.point.mysynology.net&type=A"
	// 参数使用方式2: "socks5://127.0.0.1:10808" "E:\code\go\http-tool\1648344450873.yml"
	if len(os.Args) < 3 {
		log.Printf("参数数量太少, args:%s", strings.Join(os.Args, ","))
		return
	}

	// 使用socks5代理初始化http客户端
	client := &http.Client{}
	tgProxyURL, err := url.Parse(os.Args[1])
	if err != nil {
		log.Printf("Failed to parse proxy URL:%s\n", err.Error())
		return
	}
	tgDialer, err := proxy.FromURL(tgProxyURL, proxy.Direct)
	if err != nil {
		log.Printf("Failed to obtain proxy dialer: %s\n", err.Error())
		return
	}
	var dialContext = func(ctx context.Context, network string, addr string) (net.Conn, error) {
		return tgDialer.Dial(network, addr)
	}
	tgTransport := &http.Transport{
		DialContext: dialContext,
	}
	client.Transport = tgTransport // 使用全局的HttpClient不需要释放连接

	var paramString = os.Args[2] // 需要解析的数据, 有可能是多个域名拼接的字符串, 也有可能是文件路径

	if strings.HasPrefix(paramString, "https") {
		var urlSlice = strings.Split(paramString, ",")
		var dataSlice []*DNSQuery // 返回的DNS查询结果列表

		for _, urlString := range urlSlice {
			resp, err := client.Get(urlString)
			if err != nil {
				log.Printf("http request err: %s\n", err.Error())
				continue
			}

			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Printf("http request read body err: %s\n", err.Error())
				continue
			}
			_ = resp.Body.Close()
			dns := &DNSQuery{}
			err = json.Unmarshal(data, dns)
			if err != nil {
				log.Printf("json Unmarshal body err: %s\n%s", err.Error(), string(data))
				continue
			}
			dataSlice = append(dataSlice, dns)
		}

		data, err := json.Marshal(dataSlice)
		if err != nil {
			log.Printf("json Marshal err: %s\n", err.Error())
			return
		}
		uEnc := base64.URLEncoding.EncodeToString(data)
		fmt.Println(uEnc)
	} else if strings.HasSuffix(paramString, ".yml") {
		readFileData, err := ioutil.ReadFile(paramString)
		if err != nil {
			log.Printf("ioutil ReadFile %s err: %s\n", paramString, err.Error())
		}
		var clashConfig = &ClashConfig{}
		err = yaml.Unmarshal(readFileData, clashConfig)
		if err != nil {
			log.Printf("yaml Unmarshal %s err: %s\n", string(readFileData), err.Error())
			return
		}
		if clashConfig.Proxies == nil || len(clashConfig.Proxies) == 0 {
			log.Println("clashConfig.Proxies no data.")
			return
		}

		var urlMap = make(map[string]string, 0)
		for _, proxies := range clashConfig.Proxies {
			urlMap[proxies.Server] = ""
		}

		for key := range urlMap {
			var urlString = fmt.Sprintf("https://dns.google/resolve?name=%s&type=A", key)
			resp, err := client.Get(urlString)
			if err != nil {
				log.Printf("http request err: %s\n", err.Error())
				continue
			}

			respData, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Printf("http request read body err: %s\n", err.Error())
				continue
			}
			_ = resp.Body.Close()
			dns := &DNSQuery{}
			err = json.Unmarshal(respData, dns)
			if err != nil {
				log.Printf("json Unmarshal body err: %s\n%s", err.Error(), string(respData))
				continue
			}

			if dns.Answer != nil && len(dns.Answer) > 0 {
				urlMap[key] = dns.Answer[0].Data
			}
		}

		for _, proxies := range clashConfig.Proxies {
			if ip, ok := urlMap[proxies.Server]; ok && len(ip) > 0 {
				proxies.Server = ip
			}
		}

		writeFileData, err := yaml.Marshal(clashConfig)
		if err != nil {
			log.Printf("yaml Marshal %s err: %s\n", string(writeFileData), err.Error())
			return
		}
		err = ioutil.WriteFile(paramString, writeFileData, os.ModePerm)
		if err != nil {
			log.Printf("ioutil WriteFile %s err: %s\n", string(writeFileData), err.Error())
			return
		}
	}
}
