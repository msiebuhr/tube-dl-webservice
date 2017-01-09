package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"
)

func main() {
	backgroundCtx := context.Background()

	http.HandleFunc("/dl/", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(backgroundCtx, 5*time.Minute)

		// Flush data every N seconds by default
		if flusher, ok := w.(http.Flusher); ok {
			go func() {
				ticker := time.NewTicker(1 * time.Second)
				for {
					select {
					case <-ticker.C:
						flusher.Flush()
					case <-ctx.Done():
						ticker.Stop()
						return
					}
				}
			}()
		}

		// TODO: Figure out /dl/{video,audio}/... + possibly some flags for kids-stuff separately
		urlCopy := url.URL(*r.URL)
		parts := strings.SplitN(urlCopy.Path, "/", 3)
		fmt.Printf("%+v\n", parts)
		if len(parts) < 2 {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		fmt.Println(parts)
		urlCopy.Path = strings.TrimPrefix(urlCopy.Path, "/dl/")

		if strings.HasPrefix(urlCopy.Path, "https:") {
			urlCopy.Scheme = "https"
			urlCopy.Path = strings.TrimPrefix(urlCopy.Path, "https:/")
		} else if strings.HasPrefix(urlCopy.Path, "http:") {
			urlCopy.Scheme = "http"
			urlCopy.Path = strings.TrimPrefix(urlCopy.Path, "http:/")
		} else {
			urlCopy.Scheme = "http"
		}

		cmd := exec.CommandContext(ctx, "echo", "youtube-dl", urlCopy.String())

		cmd.Stdout = w
		cmd.Stderr = w

		err := cmd.Run()
		if err != nil {
			fmt.Fprint(w, "ERRORED OUT", err)
		}

		// Remember to stop all the other things we've started
		cancel()
	})

	panic(http.ListenAndServe(":8080", nil))
}
