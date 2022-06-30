package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"http-tool/model"
	"http-tool/utils"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func main() {
	// 参数使用方式1: "socks5://127.0.0.1:10808" "https://dns.google/resolve?name=apn-a.point.mysynology.net&type=A,https://dns.google/resolve?name=apn-b.point.mysynology.net&type=A"
	// 参数使用方式2: "socks5://127.0.0.1:10808" "E:\code\go\http-tool\1648344450873.yml" "wait=10"
	if len(os.Args) < 3 {
		log.Printf("参数数量太少, args:%s", strings.Join(os.Args, ","))
		return
	}
	var client = utils.GetHttpClient()
	var paramString = os.Args[2] // 需要解析的数据, 有可能是多个域名拼接的字符串, 也有可能是文件路径
	urlStringSet := mapset.NewSet[string]()

	if strings.HasPrefix(paramString, "https") {
		var urlSlice = strings.Split(paramString, ",")
		for _, urlString := range urlSlice {
			urlStringSet.Add(urlString)
		}

		var dnsSet = utils.GetDnsQuery(client, urlStringSet) // 返回的DNS查询结果列表
		data, err := json.Marshal(dnsSet.ToSlice())
		if err != nil {
			log.Printf("json Marshal err: %s\n", errors.WithStack(err).Error())
			return
		}
		uEnc := base64.URLEncoding.EncodeToString(data)
		fmt.Println(uEnc)
	} else if strings.HasSuffix(paramString, ".yml") {
		utils.ParseExtCommand()
		readFileData, err := ioutil.ReadFile(paramString)
		if err != nil {
			log.Printf("ioutil ReadFile %s err: %s\n", paramString, errors.WithStack(err).Error())
		}
		var clashConfig = &model.ClashConfig{}
		err = yaml.Unmarshal(readFileData, clashConfig)
		if err != nil {
			log.Printf("yaml Unmarshal %s err: %s\n", string(readFileData), errors.WithStack(err).Error())
			return
		}
		if clashConfig.Proxies == nil || len(clashConfig.Proxies) == 0 {
			log.Println("clashConfig.Proxies no data.")
			return
		}

		// 插入clash规则
		utils.InsertClashRules(paramString, clashConfig)

		var urlMap = make(map[string]string, 0)
		for _, proxies := range clashConfig.Proxies {
			address := net.ParseIP(proxies.Server)
			if address == nil && len(proxies.Server) > 0 {
				urlMap[proxies.Server] = "" // 当 proxies.Server 是一个域名的时候才会尝试去解析
			}
		}

		if len(urlMap) == 0 {
			log.Println("urlMap no data.")
			return
		}

		for key := range urlMap {
			var urlString = fmt.Sprintf("https://dns.google/resolve?name=%s&type=A", key)
			urlStringSet.Add(urlString)
		}

		var dnsSet = utils.GetDnsQuery(client, urlStringSet)
		for _, dnsData := range dnsSet.ToSlice() {
			urlMap[strings.Trim(dnsData.Answer[0].Name, ".")] = dnsData.Answer[0].Data
		}

		for _, proxies := range clashConfig.Proxies {
			if ip, ok := urlMap[proxies.Server]; ok && len(ip) > 0 {
				proxies.Server = ip
			}
		}

		utils.SaveClashConfig(paramString, clashConfig)
	}
}
