package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"os"
	"strings"
	"time"
)

func unsetEnv(key string) {
	os.Unsetenv(key)
	os.Unsetenv(strings.ToLower(key))
	os.Unsetenv(strings.ToUpper(key))
}

func main() {
	unsetEnv("http_proxy")
	unsetEnv("https_proxy")
	url := "https://localhost"
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			ForceAttemptHTTP2: true,
		},
	}

	for i := 0; i < 20; i++ {
		go func() {
			for {
				trace := &httptrace.ClientTrace{
					GotConn: func(connInfo httptrace.GotConnInfo) {
					},
				}
				ctx := httptrace.WithClientTrace(context.Background(), trace)
				req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &myReader{})
				if err != nil {
					panic(err)
				}
				resp, err := client.Do(req)
				if err != nil {
					fmt.Println(err)
					continue
				}
				fmt.Print(resp.Header)
				resp.Body.Close()
			}
		}()
	}
	time.Sleep(1 * time.Hour)
}

type myReader struct {
	count int
}

func (r *myReader) Read(p []byte) (n int, err error) {
	if r.count >= 20 {
		return 0, io.EOF
	}
	r.count += 1
	p[0] = 'a'
	return 1, nil
}
