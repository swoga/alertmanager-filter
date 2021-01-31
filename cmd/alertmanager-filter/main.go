package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/alertmanager/notify/webhook"
	"github.com/prometheus/alertmanager/template"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/swoga/alertmanager-filter/config"
	"github.com/swoga/alertmanager-filter/version"
	"go.uber.org/zap"
)

var (
	sc  config.SafeConfig
	log *zap.Logger
)

func main() {
	configFlag := flag.String("config.file", "", "")
	debug := flag.Bool("debug", false, "")
	flag.Parse()

	level := zap.InfoLevel
	if *debug {
		level = zap.DebugLevel
	}

	zapConfig := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      false,
		Encoding:         "console",
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	log, _ = zapConfig.Build()
	defer log.Sync()
	log.Info("starting alertmanager-filter", zap.String("version", version.Version), zap.String("revision", version.Revision))

	sc = config.New(*configFlag)
	err := sc.LoadConfig()
	if err != nil {
		log.Fatal("error loading config", zap.Any("err", err))
	}

	// setup config reload
	hup := make(chan os.Signal, 1)
	signal.Notify(hup, syscall.SIGHUP)
	reloadRequest := make(chan chan error)
	go func() {
		for {
			var err error
			select {
			case <-hup:
				log.Debug("config reload triggerd by SIGHUP")
				err = sc.LoadConfig()
			case reloadResult := <-reloadRequest:
				log.Debug("config reload triggerd by API")
				err = sc.LoadConfig()
				reloadResult <- err
			}
			if err != nil {
				log.Error("error reloading config", zap.Any("err", err))
			} else {
				log.Info("reloaded config file")
			}
		}
	}()

	http.HandleFunc("/-/reload", func(w http.ResponseWriter, r *http.Request) {
		reloadResult := make(chan error)
		reloadRequest <- reloadResult
		err := <-reloadResult
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to reload config: %s", err), http.StatusInternalServerError)
		}
	})

	// start http server
	config := sc.Get()
	http.Handle(config.MetricsPath, promhttp.Handler())
	http.HandleFunc(config.AlertsPath, handleRequest)

	log.Info("starting http server", zap.String("metrics_path", config.MetricsPath), zap.String("listen", config.Listen))

	err = http.ListenAndServe(config.Listen, nil)
	if err != nil {
		log.Error("failed to start http server", zap.Any("err", err))
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// create local log variable, so that no one accidentally uses the global one
	log := log

	log.Debug("incoming request", zap.String("from", r.RemoteAddr))

	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(20*time.Second))
	defer cancel()
	r = r.WithContext(ctx)

	c := sc.Get()

	message, err := decodeMessage(r.Body)

	if err == nil {
		log = log.With(zap.String("receiver", message.Receiver))
		log.Debug("received message", zap.Any("message", message))
	}

	var receiver *config.Receiver
	if err == nil {
		receiver, err = getReceiver(c.Receivers, *message)
	}

	if err == nil {
		message, err = filterAlerts(c.TimeIntervals, *receiver, *message)
	}

	if err == nil {
		if message != nil {
			log.Debug("send message", zap.Any("message", message))
			err = sendAlert(ctx, log, w, r, receiver.Target, *message)
		} else {
			log.Debug("no alerts to forward")
		}
	}

	if err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func decodeMessage(reader io.Reader) (*webhook.Message, error) {
	message := &webhook.Message{}
	decoder := json.NewDecoder(reader)
	err := decoder.Decode(message)
	if err != nil {
		return nil, fmt.Errorf("failed to decode message: %s", err)
	}
	return message, nil
}

func getReceiver(receivers map[string]config.Receiver, message webhook.Message) (*config.Receiver, error) {
	receiver, ok := receivers[message.Receiver]
	if !ok {
		return nil, fmt.Errorf("unknown receiver: %s", message.Receiver)
	}

	return &receiver, nil
}

func filterAlerts(timeIntervalsMap config.TimeIntervalsMap, receiver config.Receiver, message webhook.Message) (*webhook.Message, error) {
	now := time.Now()

	matchedAlerts := template.Alerts{}

	for _, alert := range message.Alerts {
		for _, rule := range receiver.Rules {
			if rule.IsMatch(timeIntervalsMap, alert.Labels, now) {
				matchedAlerts = append(matchedAlerts, alert)
				break
			}
		}
	}

	if len(matchedAlerts) == 0 {
		return nil, nil
	}

	message.Alerts = matchedAlerts

	return &message, nil
}

func sendAlert(ctx context.Context, log *zap.Logger, w http.ResponseWriter, r *http.Request, target config.Target, message webhook.Message) error {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	err := encoder.Encode(message)
	if err != nil {
		return err
	}

	r.Body = ioutil.NopCloser(&buf)

	proxy, err := target.CreateProxy(log)
	if err != nil {
		return err
	}
	log.Debug("forward message", zap.Any("url", target.URL.URL))
	proxy.ServeHTTP(w, r)

	return nil
}
