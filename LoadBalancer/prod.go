package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/valyala/fasthttp"
)

type Config struct {
	Tiempo time.Duration `json:"Tiempo"`
}
type MyHandler struct {
	Conf         Config            `json:"Conf"`
	Servers      []Server          `json:"Servers"`
	Count        int               `json:"Count"`
	TotalRequest int               `json:"TotalRequest"`
	LimitRequest int               `json:"LimitRequest"`
	Request      *fasthttp.Request `json:"Request"`
}
type Server struct {
	addr string
	port int
}

func main() {

	var port string
	if runtime.GOOS == "windows" {
		port = ":81"
	} else {
		port = ":81"
	}

	pass := &MyHandler{Request: fasthttp.AcquireRequest(), Count: 0, Servers: []Server{Server{addr: "10.128.0.10", port: 80}, Server{addr: "10.128.0.11", port: 80}}}
	pass.Request.Header.SetMethod("GET")
	pass.Request.Header.SetContentType("application/json")

	con := context.Background()
	con, cancel := context.WithCancel(con)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGHUP)
	defer func() {
		signal.Stop(signalChan)
		cancel()
	}()
	go func() {
		for {
			select {
			case s := <-signalChan:
				switch s {
				case syscall.SIGHUP:
					pass.Conf.init()
				case os.Interrupt:
					cancel()
					os.Exit(1)
				}
			case <-con.Done():
				log.Printf("Done.")
				os.Exit(1)
			}
		}
	}()
	go func() {
		fasthttp.ListenAndServe(port, pass.HandleFastHTTP)
	}()
	if err := run(con, pass, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func (h *MyHandler) HandleFastHTTP(ctx *fasthttp.RequestCtx) {

	//var path []byte = ctx.Path()
	//var method []byte = ctx.Method()

	//fmt.Println("IP: ", ctx.RemoteIP())
	//fmt.Println("HOST: ", string(ctx.Request.Host()), ctx.Request.Host())
	//fmt.Println("MATHOD: ", string(ctx.Method()), ctx.Method())
	//fmt.Println("PATH: ", string(ctx.Path()), ctx.Path())

	if string(ctx.Method()) == "GET" {
		switch string(ctx.Path()) {
		case "/":
			// INDEX HTML
			ctx.SetBody(h.Send2("", []byte{}))
		case "/count":
			fmt.Println(h.Count)
			ctx.SetBody(h.Send("", []byte{}))
		case "/favicon.ico":
			ctx.SetBody(h.Send("", []byte{65, 66}))
		default:
			ctx.Error("Not Found", fasthttp.StatusNotFound)
		}
	}
}

func (h *MyHandler) Send(ip string, data []byte) []byte {

	num := h.Count % len(h.Servers)
	h.Count++
	uri := fmt.Sprintf("http://%v:%v", h.Servers[num].addr, h.Servers[num].port)
	req := fasthttp.AcquireRequest()
	req.SetBody(data)
	req.Header.SetMethod("GET")
	req.Header.SetContentType("application/json")
	req.SetRequestURI(uri)
	res := fasthttp.AcquireResponse()
	if err := fasthttp.Do(req, res); err != nil {

	}
	fasthttp.ReleaseRequest(req)
	body := res.Body()
	fasthttp.ReleaseResponse(res)

	return body
}

func (h *MyHandler) Send2(ip string, data []byte) []byte {

	num := h.Count % len(h.Servers)
	h.Count++
	uri := fmt.Sprintf("http://%v:%v", h.Servers[num].addr, h.Servers[num].port)
	var body []byte
	statusCode, body, err := fasthttp.Get(body, uri)
	if err != nil {
		fmt.Printf("Client get failed: %s\n", err)
		return nil
	}
	if statusCode != fasthttp.StatusOK {
		fmt.Printf("Expected status code %d but got %d\n", fasthttp.StatusOK, statusCode)
		return nil
	}
	return body
}

// DAEMON //
func (h *MyHandler) StartDaemon() {
	h.Conf.Tiempo = 10 * time.Second
	fmt.Println("Count: ", h.Count)
}
func (c *Config) init() {
	var tick = flag.Duration("tick", 1*time.Second, "Ticking interval")
	c.Tiempo = *tick
}
func run(con context.Context, c *MyHandler, stdout io.Writer) error {
	c.Conf.init()
	log.SetOutput(os.Stdout)
	for {
		select {
		case <-con.Done():
			fmt.Println("ETAPA 1")
			return nil
		case <-time.Tick(c.Conf.Tiempo):
			fmt.Println("ETAPA 2")
			c.StartDaemon()
		}
	}
}
