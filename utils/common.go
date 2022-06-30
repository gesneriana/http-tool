package utils

import (
	"context"
	"http-tool/model"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go.net/proxy"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// PathExists 判断所给路径文件/文件夹是否存在
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		return false
	}
	return false
}

var proxyClient = &http.Client{}

// GetHttpClient 获取全局的http代理客户端
func GetHttpClient() *http.Client {
	// 使用socks5代理初始化http客户端
	tgProxyURL, err := url.Parse(os.Args[1])
	if err != nil {
		log.Printf("Failed to parse proxy URL:%s\n", errors.WithStack(err).Error())
		return proxyClient
	}
	proxyDialer, err := proxy.FromURL(tgProxyURL, proxy.Direct)
	if err != nil {
		log.Printf("Failed to obtain proxy dialer: %s\n", errors.WithStack(err).Error())
		return proxyClient
	}
	var dialContext = func(ctx context.Context, network string, addr string) (net.Conn, error) {
		return proxyDialer.Dial(network, addr)
	}
	tgTransport := &http.Transport{
		DialContext: dialContext,
	}
	proxyClient.Transport = tgTransport // 使用全局的HttpClient不需要释放连接
	return proxyClient
}

// SaveClashConfig 保存clash配置文件到磁盘
func SaveClashConfig(path string, config *model.ClashConfig) {
	writeFileData, err := yaml.Marshal(config)
	if err != nil {
		log.Printf("yaml Marshal %s err: %s\n", string(writeFileData), errors.WithStack(err).Error())
		return
	}
	err = ioutil.WriteFile(path, writeFileData, os.ModePerm)
	if err != nil {
		log.Printf("ioutil WriteFile %s err: %s\n", string(writeFileData), errors.WithStack(err).Error())
		return
	}
}

// InsertClashRules 插入clash规则
func InsertClashRules(path string, config *model.ClashConfig) {
	rulesFilePath := "./clash-rules.txt"
	if PathExists(rulesFilePath) {
		rulesFileData, err := ioutil.ReadFile(rulesFilePath)
		if err != nil {
			log.Printf("ioutil ReadFile %s err: %s\n", rulesFilePath, errors.WithStack(err).Error())
			return
		}
		if len(rulesFileData) == 0 {
			return
		}
		var ruleText = string(rulesFileData)
		var ruleSlice = strings.Split(ruleText, "\n")
		var insertRuleMap = make(map[string]struct{}, 0)
		for _, rule := range ruleSlice {
			insertRuleMap[strings.Trim(rule, "\r")] = struct{}{}
		}
		for _, rule := range config.Rules {
			delete(insertRuleMap, rule) // 如果已经存在了，去掉
		}
		for rule, _ := range insertRuleMap {
			config.Rules = append([]string{rule}, config.Rules...) // 追加到前面
		}
		SaveClashConfig(path, config)
	}
}

// ParseExtCommand 执行扩展的命令参数
func ParseExtCommand() {
	if len(os.Args) > 3 {
		var cmdArg = make(map[string]string, 0)
		var extArgs = os.Args[3:]
		for _, arg := range extArgs {
			if strings.Count(arg, "=") == 1 {
				argSlice := strings.Split(arg, "=")
				cmdArg[argSlice[0]] = argSlice[1]
			}
		}

		for k, v := range cmdArg {
			if k == "wait" {
				sec, err := strconv.Atoi(v)
				if err != nil {
					log.Printf("wait command err: %s\n", errors.WithStack(err).Error())
					continue
				}
				if sec < 10 {
					sec = 10
				}
				time.Sleep(time.Second * time.Duration(sec)) // 添加更多用法，等待一段时间，等clash启动成功并且开启代理再去更新配置文件
			}
		}
	}
}
