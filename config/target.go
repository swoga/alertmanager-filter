package config

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httputil"

	am_config "github.com/prometheus/alertmanager/config"
	commoncfg "github.com/prometheus/common/config"
	"go.uber.org/zap"
)

type Target struct {
	URL        am_config.URL              `yaml:"url"`
	HTTPConfig commoncfg.HTTPClientConfig `yaml:"http_config"`
	HTTPClient *http.Client
}

// UnmarshalYAML implements the Unmarshaller interface for Target
func (t *Target) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Target
	if err := unmarshal((*plain)(t)); err != nil {
		return err
	}

	client, err := commoncfg.NewClientFromConfig(t.HTTPConfig, "target", false)
	if err != nil {
		return err
	}
	t.HTTPClient = client

	return nil
}

func (t *Target) CreateProxy(log *zap.Logger) (*httputil.ReverseProxy, error) {
	director := func(req *http.Request) {
		req.URL = t.URL.URL
	}
	proxy := &httputil.ReverseProxy{Director: director}
	proxy.Transport = t.HTTPClient.Transport
	stdLog, err := zap.NewStdLogAt(log, zap.ErrorLevel)
	if err != nil {
		return nil, err
	}
	proxy.ErrorLog = stdLog

	if log.Core().Enabled(zap.DebugLevel) {
		proxy.ModifyResponse = func(res *http.Response) error {
			defer res.Body.Close()
			data, err := ioutil.ReadAll(res.Body)
			if err != nil {
				return err
			}
			log.Debug("target response", zap.String("data", string(data)))

			buf := bytes.NewReader(data)
			reader := ioutil.NopCloser(buf)

			res.Body = reader

			return nil
		}
	}

	return proxy, nil
}
