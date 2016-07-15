package speedtest

import (
	"time"
	"log"
	"io"
	"fmt"
	"os"
)

const downloadStreamLimit = 6
const maxDownloadDuration = 10 * time.Second
const downloadBufferSize = 4096
const downloadRepeats = 5

var downloadImageSizes = []int{350, 500, 750, 1000, 1500, 2000, 2500, 3000, 3500, 4000}

func (client *Client) downloadFile(url string, start time.Time, ret chan int) {
	totalRead := 0
	defer func() {
		ret <- totalRead
	}()

	if (time.Since(start) > maxDownloadDuration) {
		return;
	}
	if !client.opts.Quiet {
		os.Stdout.WriteString(".")
		os.Stdout.Sync()
	}

	resp, err := client.Get(url)
	if err != nil {
		log.Printf("[%s] Download failed: %v\n", url, err)
		return;
	}

	defer resp.Body.Close()

	buf := make([]byte, downloadBufferSize)
	for time.Since(start) <= maxDownloadDuration {
		read, err := resp.Body.Read(buf)
		totalRead += read
		if err != nil {
			if err != io.EOF {
				log.Printf("[%s] Download error: %v\n", url, err)
			}
			break
		}
	}
}

func (server *Server) DownloadSpeed() int {
	client := server.client
	if !client.opts.Quiet {
		os.Stdout.WriteString("Testing download speed: ")
		os.Stdout.Sync()
	}

	starterChan := make(chan int, downloadStreamLimit)
	downloads := downloadRepeats * len(downloadImageSizes)
	resultChan := make(chan int, downloadStreamLimit)
	start := time.Now()

	go func() {
		for _, size := range downloadImageSizes {
			for i := 0; i < downloadRepeats; i++ {
				url := server.RelativeURL(fmt.Sprintf("random%dx%d.jpg", size, size))
				starterChan <- 1
				go func() {
					client.downloadFile(url, start, resultChan)
					<-starterChan
				}()
			}
		}
		close(starterChan)
	}()

	var totalSize int64 = 0;

	for i := 0; i < downloads; i++ {
		totalSize += int64(<-resultChan)
	}

	if !client.opts.Quiet {
		os.Stdout.WriteString("\n")
		os.Stdout.Sync()
	}

	duration := time.Since(start);

	return int(totalSize * int64(time.Second) / int64(duration))
}
