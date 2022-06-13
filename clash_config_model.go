package main

type ClashConfig struct {
	Port               int    `yaml:"port"`
	SocksPort          int    `yaml:"socks-port"`
	RedirPort          int    `yaml:"redir-port"`
	AllowLan           bool   `yaml:"allow-lan"`
	Mode               string `yaml:"mode"`
	LogLevel           string `yaml:"log-level"`
	ExternalController string `yaml:"external-controller"`
	Secret             string `yaml:"secret"`
	DNS                struct {
		Enable         bool     `yaml:"enable"`
		Ipv6           bool     `yaml:"ipv6"`
		Listen         string   `yaml:"listen"`
		EnhancedMode   string   `yaml:"enhanced-mode"`
		FakeIPRange    string   `yaml:"fake-ip-range"`
		Nameserver     []string `yaml:"nameserver"`
		Fallback       []string `yaml:"fallback"`
		FallbackFilter struct {
			Geoip  bool     `yaml:"geoip"`
			Ipcidr []string `yaml:"ipcidr"`
		} `yaml:"fallback-filter"`
	} `yaml:"dns"`
	Proxies []*struct {
		Name       string `yaml:"name"`
		Type       string `yaml:"type"`
		Server     string `yaml:"server"`
		Port       int    `yaml:"port"`
		UUID       string `yaml:"uuid"`
		AlterID    int    `yaml:"alterId"`
		Cipher     string `yaml:"cipher"`
		UDP        bool   `yaml:"udp"`
		Servername string `yaml:"servername"`
		Network    string `yaml:"network"`
		WsOpts     struct {
			Path    string `yaml:"path"`
			Headers struct {
				Host string `yaml:"Host"`
			} `yaml:"headers"`
		} `yaml:"ws-opts"`
		WsPath    string `yaml:"ws-path"`
		WsHeaders struct {
			Host string `yaml:"Host"`
		} `yaml:"ws-headers"`
		TLS            bool `yaml:"tls"`
		SkipCertVerify bool `yaml:"skip-cert-verify"`
	} `yaml:"proxies"`
	ProxyGroups []struct {
		Name    string   `yaml:"name"`
		Type    string   `yaml:"type"`
		Proxies []string `yaml:"proxies"`
	} `yaml:"proxy-groups"`
	Rules []string `yaml:"rules"`
}
