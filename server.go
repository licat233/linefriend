package main

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

//JSONResult 响应的json数据结构体
type JSONResult struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Res  int    `json:"res"`
}

func main() {
	router := fasthttprouter.New()
	router.GET("/", lineFriend)
	log.Fatal(fasthttp.ListenAndServe(":80", router.Handler))

}

func lineFriend(ctx *fasthttp.RequestCtx) {
	var resjson = &JSONResult{
		Code: 400,
		Msg:  "error",
		Res:  0,
	}
	var fans int
	var err error
	p := ctx.QueryArgs().Peek("lineid")
	if p == nil {
		resjson.Msg = "no lineid"
		goto VIEW
	}
	fans, err = getFans(string(p))
	if err != nil {
		resjson.Msg = err.Error()
		goto VIEW
	}
	resjson.Code = 200
	resjson.Msg = "successfully"
	resjson.Res = fans
VIEW:
	msg, _ := json.Marshal(resjson)
	ctx.Write(msg)
}

//getFans @账号的粉丝数
func getFans(account string) (fans int, err error) {
	if !strings.Contains(account, "@") && !strings.Contains(account, "%40") {
		return
	}
	account = url.QueryEscape(account)
	webURL := "https://social-plugins.line.me/widget/friend?count=true&lineId=" + account
	resp, err := http.Get(webURL)
	if err != nil || resp.StatusCode != 200 {
		return
	}
	defer resp.Body.Close()
	var body string
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, _ := gzip.NewReader(resp.Body)
		for {
			buf := make([]byte, 1024)
			n, err := reader.Read(buf)

			if err != nil && err != io.EOF {
				panic(err)
			}

			if n == 0 {
				break
			}
			body += string(buf)
		}
	default:
		bodyByte, _ := ioutil.ReadAll(resp.Body)
		body = string(bodyByte)
	}

	r := regexp.MustCompile(`friendCount":\d+,`)
	d := r.FindString(body)

	r1 := regexp.MustCompile(`\d+`)
	d1 := r1.FindString(d)
	fans, err = strconv.Atoi(d1)
	return
}
