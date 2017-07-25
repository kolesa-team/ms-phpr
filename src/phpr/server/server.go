package server

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"../consts"
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
	proxyPrefix     string
	proxyTimeout    int
	once            sync.Once
	lock            sync.RWMutex
	client          http.Client
	nbRequests      int = 0
	maxRequests     int
	statPeriod      int
	statKeys        []string = []string{"nb_requests", "nb_failures", "nb_success"}
	counters        map[string]uint64
	queue           map[string][]uint64
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

		counters = make(map[string]uint64, len(statKeys))
		queue = make(map[string][]uint64, len(statKeys))

		for _, key := range statKeys {
			counters[key] = 0
			queue[key] = make([]uint64, 0)
		}

		if statPeriod, err = config.Instance().Int("stats", "period"); err != nil {
			statPeriod = 300
		}
	})
}

// Возвращает объект нового мультиплексора
func NewMux() *web.Mux {
	initServer()

	m := web.New()

	if enableAccessLog {
		m.Use(middleware.RealIP)
		m.Use(mwLogger)
	}

	m.Use(mwRecoverer)

	m.Get("/_status", handleStatus)
	m.Get("/_version", handleVersion)
	m.Get(regexp.MustCompile(`^/(.*)$`), handleRequest)

	// Считаем статистику за последние 5 минут
	go func() {
		for {
			lock.Lock()
			for _, key := range statKeys {
				queue[key] = append(queue[key], counters[key])

				if len(queue[key]) > statPeriod {
					queue[key] = queue[key][len(queue[key])-statPeriod:]
				}
			}
			lock.Unlock()

			duration, _ := time.ParseDuration("1s")
			time.Sleep(duration)
		}
	}()

	return m
}

func handleVersion(c web.C, w http.ResponseWriter, r *http.Request) {
	http.Error(w, consts.APP_VERSION, http.StatusOK)
}

func handleStatus(c web.C, w http.ResponseWriter, r *http.Request) {
	var (
		head, tail uint64
	)

	w.Header().Set("Content-Type", "text/plain")

	lock.Lock()
	for _, key := range statKeys {
		w.Write([]byte(fmt.Sprintf("%s: %d\n", key, counters[key])))
	}
	lock.Unlock()

	w.Write([]byte("\r\n\r\n"))

	lock.Lock()
	for _, key := range statKeys {
		head = queue[key][0]
		tail = queue[key][len(queue[key])-1]
		w.Write([]byte(fmt.Sprintf("%s_%ds: %d\n", key, statPeriod, tail-head)))
	}
	lock.Unlock()
}

func handleRequest(c web.C, w http.ResponseWriter, r *http.Request) {
	var (
		res *http.Response
		err error
	)

	logRequest("nb_requests")

	_ = r.ParseForm()
	query := parseQuery(r.Form)
	url := c.URLParams["$1"]
	format := strings.ToLower(url[strings.LastIndex(url, ".")+1:])
	res, err = client.Get(proxyPrefix + url)

	if res != nil {
		defer res.Body.Close()
	}

	if err != nil {
		logger.Instance().WithFields(log.Fields{
			"error":         err,
			"response_code": http.StatusGatewayTimeout,
		}).Error("Error while requesting remote file")

		http.Error(w, http.StatusText(http.StatusGatewayTimeout), http.StatusGatewayTimeout)

		logRequest("nb_failures")
	} else {
		if res.StatusCode == 200 {
			for key, _ := range res.Header {
				if strings.HasPrefix(key, "X-") {
					w.Header().Add(key, res.Header.Get(key))
				}
			}

			img, err := image.FromReader(res.Body)

			if err == nil {
				if query.Width > 0 || query.Height > 0 {
					img = image.Resize(img, query.Width, query.Height, query.IsBestfit)
				}

				if query.IsWatermark {
					img = image.Watermark(img)
				}

				if buffer, err := image.ToBuffer(img, format); err == nil {
					w.Header().Add("Content-Length", strconv.Itoa(len(buffer.Bytes())))
					io.Copy(w, buffer)
					logRequest("nb_success")
				} else {
					logger.Instance().WithFields(log.Fields{
						"error":         err,
						"response_code": http.StatusInternalServerError,
					}).Error("Error encoding image")

					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					logRequest("nb_failures")
				}
			} else {
				logger.Instance().WithFields(log.Fields{
					"error":         err,
					"response_code": http.StatusBadGateway,
				}).Error("Error while reading image from response")

				http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
				logRequest("nb_failures")
			}
		} else {
			logger.Instance().WithFields(log.Fields{
				"response_code": res.StatusCode,
			}).Error("Wrong status received")

			http.Error(w, res.Status, res.StatusCode)
			logRequest("nb_failures")
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
			logger.Instance().WithFields(log.Fields{
				"size": valSize[0],
			}).Error("Wrong size passed")
		}

	}

	if valWm, okWm := query["watermark"]; okWm {
		result.IsWatermark = valWm[0] == "1"
	}

	if err != nil {
		logger.Instance().WithFields(log.Fields{
			"error": err,
		}).Error("Error processing query parameters")
	}

	return result
}

func logRequest(key string) {
	lock.Lock()
	if _, ok := counters[key]; ok {
		counters[key]++
	}
	lock.Unlock()
}
