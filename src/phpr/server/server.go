package server

import (
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"../image"
	"../logger"

	log "github.com/Sirupsen/logrus"
	"github.com/endeveit/go-snippets/cli"
	"github.com/endeveit/go-snippets/config"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"
)

type QueryData struct {
	Width       int
	Height      int
	IsBestfit   bool
	IsWatermark bool
}

var (
	enableAccessLog bool = false
	debugMode       bool = false
	proxyPrefix     string
	proxyTimeout    int
	once            sync.Once
	client          http.Client
	nbRequests      int = 0
	maxRequests     int
)

// Инициализация ключей
func initServer() {
	once.Do(func() {
		var (
			err error
		)

		enableAccessLog, err = config.Instance().Bool("http", "access_log")
		cli.CheckError(err)

		proxyPrefix, err = config.Instance().String("proxy", "url")
		cli.CheckFatalError(err)

		if proxyTimeout, err = config.Instance().Int("proxy", "timeout"); err != nil {
			proxyTimeout = 1000
		}

		if maxRequests, err = config.Instance().Int("proxy", "max_requests"); err != nil {
			maxRequests = 1000
		}

		client = http.Client{
			Timeout: time.Duration(time.Duration(proxyTimeout) * time.Millisecond),
		}
	})
}

// Возвращает объект нового мультиплексора
func NewMux(isDebug bool) *web.Mux {
	initServer()
	debugMode = isDebug

	m := web.New()

	if enableAccessLog {
		m.Use(middleware.RealIP)
		m.Use(mwLogger)
	}

	m.Use(mwRecoverer)

	m.Get(regexp.MustCompile(`^/(.*)$`), handleRequest)

	return m
}

func handleRequest(c web.C, w http.ResponseWriter, r *http.Request) {
	var (
		res *http.Response
		err error
	)

	_ = r.ParseForm()
	query := parseQuery(r.Form)
	url := c.URLParams["$1"]

	if debugMode {
		logger.Instance().WithFields(log.Fields{
			"url": proxyPrefix + url,
		}).Info("Request remote file")
	}

	if res, err = client.Get(proxyPrefix + url); err != nil {
		if debugMode {
			logger.Instance().WithFields(log.Fields{
				"error":         err,
				"response_code": 504,
			}).Error("Error while requesting remote file")
		}

		http.Error(w, "504 Gateway Timeout", 504)
	} else {
		defer res.Body.Close()

		if res.StatusCode == 200 {
			for key, _ := range res.Header {
				if strings.HasPrefix(key, "X-") {
					w.Header().Add(key, res.Header.Get(key))
				}
			}

			img, err := image.FromReader(res.Body)
			if err == nil {
				if query.Width > 0 && query.Height > 0 {
					img = image.Resize(img, query.Width, query.Height, query.IsBestfit)
				}

				if query.IsWatermark {
					img = image.Watermark(img)
				}

				if err := image.ToWriter(img, w); err != nil {
					logger.Instance().WithFields(log.Fields{
						"error": err,
					}).Error("Error encoding image")
				}
			} else {
				if debugMode {
					logger.Instance().WithFields(log.Fields{
						"error":         err,
						"response_code": 502,
					}).Error("Error while reading image from response")
				}

				http.Error(w, "502 Bad Gateway", 502)
			}
		} else {
			if debugMode {
				logger.Instance().WithFields(log.Fields{
					"response_code": res.StatusCode,
				}).Error("Wrong status received")
			}

			http.Error(w, res.Status, res.StatusCode)
		}
	}
}

func parseQuery(query url.Values) QueryData {
	var (
		err    error
		result QueryData
	)

	if valSize, okSize := query["size"]; okSize {
		size := strings.Split(valSize[0], "x")
		if len(size) == 2 {
			if valBestfit, okBestfit := query["bestfit"]; okBestfit {
				result.IsBestfit = valBestfit[0] == "1"
			}

			result.Width, err = strconv.Atoi(size[0])
			result.Height, err = strconv.Atoi(size[1])
		} else {
			if debugMode {
				logger.Instance().WithFields(log.Fields{
					"size": valSize[0],
				}).Error("Wrong size passed")
			}
		}

	}

	if valWm, okWm := query["watermark"]; okWm {
		result.IsWatermark = valWm[0] == "1"
	}

	if err != nil {
		if debugMode {
			logger.Instance().WithFields(log.Fields{
				"error": err,
			}).Error("Error processing query parameters")
		}
	}

	return result
}
