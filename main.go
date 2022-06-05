package main

import (
	"context"
	"fmt"
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
	client.Transport = tgTransport

	resp, err := client.Get(os.Args[2])
	if err != nil {
		log.Printf("http request err: %s\n", err.Error())
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("http request read body err: %s\n", err.Error())
		return
	}
	fmt.Println(string(data))
}
