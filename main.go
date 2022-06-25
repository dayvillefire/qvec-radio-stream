package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/schollz/progressbar/v3"
)

var (
	duration = flag.Int("duration", 5, "Duration of recording segments in hours")
	timeout  = flag.Int("timeout", 5, "Timeout in seconds")
	url      = flag.String("url", "https://broadcastify.cdnstream1.com/31315", "Streaming URL")

	lastStarted time.Time
)

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())

	log.Printf("Starting up")
	client := http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   time.Duration(*timeout*3) * time.Second,
				KeepAlive: time.Duration(*timeout*3) * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   time.Duration(*timeout) * time.Second,
			ResponseHeaderTimeout: time.Duration(*timeout) * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	//client := http.DefaultClient

	go func() {
		for {
			if lastStarted.IsZero() {
				goto ENDLOOP
			}

			if time.Now().Sub(lastStarted) >= time.Hour*time.Duration(*duration) {
				cancel()
			}

		ENDLOOP:
			time.Sleep(time.Second)
		}
	}()

	for {
		lastStarted = time.Now()

		fn := fmt.Sprintf("QVEC-%s.mp3", lastStarted.Format("2006-01-02-15-04"))

		fp, err := os.Create(fn)
		if err != nil {
			log.Printf("ERR: unable to create file: %v\n", err)
			os.Exit(1)
		}
		defer fp.Close()

		req, err := http.NewRequest("GET", *url, nil)
		if err != nil {
			log.Printf("ERR: creating request: %v\n", err)
			os.Exit(1)
		}
		resp, err := client.Do(req.WithContext(ctx))
		if err != nil {
			log.Printf("ERR: %s", err.Error())
			os.Exit(1)
		}
		bar := progressbar.DefaultBytes(
			-1,
			"downloading",
		)
		log.Printf("Writing to %s", fn)
		io.Copy(io.MultiWriter(fp, bar), resp.Body)

		fmt.Println("") // make more readable

		log.Printf("Cancel detected, moving to next loop")
		ctx, cancel = context.WithCancel(context.Background())
		time.Sleep(100 * time.Millisecond) // keep from piling on
	}
}
