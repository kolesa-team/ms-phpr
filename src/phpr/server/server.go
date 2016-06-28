package server

import (
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"../image"
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
	once            sync.Once
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

	m.Get(regexp.MustCompile(`^/(.*)$`), handleRequest)

	return m
}

func handleRequest(c web.C, w http.ResponseWriter, r *http.Request) {
	var (
		res *http.Response
	)

	_ = r.ParseForm()
	query := parseQuery(r.Form)
	url := c.URLParams["$1"]

	res, _ = http.Get(proxyPrefix + url)

	defer res.Body.Close()

	if res.StatusCode == 200 {
		img := image.FromReader(res.Body)

		if query.Width > 0 && query.Height > 0 {
			img = image.Resize(img, query.Width, query.Height, query.IsBestfit)
		}

		if query.IsWatermark {
			img = image.Watermark(img)
		}

		image.ToWriter(img, w)
	} else {
		http.Error(w, res.Status, res.StatusCode)
	}
}

func parseQuery(query url.Values) QueryData {
	var (
		err    error
		result QueryData
	)

	if valSize, okSize := query["size"]; okSize {
		size := strings.Split(valSize[0], "x")
		if len(size) != 2 {
			panic("Invalid size")
		}

		if valBestfit, okBestfit := query["bestfit"]; okBestfit {
			result.IsBestfit = valBestfit[0] == "1"
		}

		result.Width, err = strconv.Atoi(size[0])
		result.Height, err = strconv.Atoi(size[1])
	}

	if valWm, okWm := query["watermark"]; okWm {
		result.IsWatermark = valWm[0] == "1"
	}

	if err != nil {
		panic(err)
	}

	return result
}
