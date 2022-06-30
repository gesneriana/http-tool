package utils

import (
	"encoding/json"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/pkg/errors"
	"http-tool/model"
	"io/ioutil"
	"log"
	"net/http"
)

// GetDnsQuery 查询dns
func GetDnsQuery(client *http.Client, urlStringSet mapset.Set[string]) mapset.Set[*model.DNSQuery] {
	var dataSet = mapset.NewSet[*model.DNSQuery]() // 返回的DNS查询结果列表
	for _, urlString := range urlStringSet.ToSlice() {
		resp, err := client.Get(urlString)
		if err != nil {
			log.Printf("http request err: %s\n", errors.WithStack(err).Error())
			continue
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("http request read body err: %s\n", errors.WithStack(err).Error())
			continue
		}
		_ = resp.Body.Close()
		dns := &model.DNSQuery{}
		err = json.Unmarshal(data, dns)
		if err != nil {
			log.Printf("json Unmarshal body err: %s\n%s", errors.WithStack(err).Error(), string(data))
			continue
		}
		if dns.Answer != nil && len(dns.Answer) > 0 {
			dataSet.Add(dns)
		}
	}

	return dataSet
}
