package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

var status int64
var delay int64

func load(ctx context.Context, url string) {
	startedAt := time.Now()

	stop := false
	done := make(chan bool)

	println(url)
	go func() {
		print("slo")
		for {
			if stop {
				break
			}
			print("o")
			time.Sleep(100 * time.Millisecond)
		}
		done <- true
	}()

	err := chromedp.Run(
		ctx,
		chromeTask(ctx, url),
	)

	if err != nil {
		log.Fatalln(err)
	}

	stop = true
	<-done
	fmt.Printf("w, %.1fs\n", time.Since(startedAt).Seconds())
	time.Sleep(time.Duration(delay) * time.Second)

	if status != 200 {
		log.Fatalln("not 200", status)
	}

}
func main() {
	flag.Int64Var(&delay, "delay", 1, "delay in seconds")
	flag.Parse()
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), append(chromedp.DefaultExecAllocatorOptions[:], chromedp.Flag("headless", false))...)
	defer cancel()
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	response, err := http.Get("https://explorer.raptoreum.com/api/getblockcount")
	if err != nil {
		log.Fatalln("getBlockCount", err)
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatalln("body read", err)
	}

	block, err := strconv.Atoi(string(body))
	if err != nil {
		log.Fatalln("body convert", err)
	}

	err = chromedp.Run(
		ctx,
		chromeTask(ctx, "about:blank"),
	)
	if err != nil {
		log.Fatalln("chrome launch", err)
	}

	for i := block; i >= 0; i-- {
		load(ctx, "https://explorer.raptoreum.com/block-height/"+strconv.Itoa(i))
	}
}

func chromeTask(chromeContext context.Context, url string) chromedp.Tasks {

	chromedp.ListenTarget(chromeContext, func(event interface{}) {
		switch msg := event.(type) {
		case *network.EventRequestWillBeSent:
			request := msg.Request
			if msg.RedirectResponse != nil {
				url = request.URL
				fmt.Printf(" got redirect: %s\n", msg.RedirectResponse.URL)
			}
		case *network.EventResponseReceived:
			if msg.Response.URL == url {
				status = msg.Response.Status
			}
		}

	})

	return chromedp.Tasks{
		network.Enable(),
		chromedp.Navigate(url),
	}
}
