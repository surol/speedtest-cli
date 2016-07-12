package speedtest

import (
	"time"
	"os"
	"log"
	"io"
	"strings"
	"crypto/rand"
)

const maxUploadDuration = maxDownloadDuration
const uploadStreamLimit = downloadStreamLimit
const uploadRepeats = downloadRepeats

var uploadSizes []int

func init() {

	var uploadSizeSizes = []int{int(1000 * 1000 / 4), int(1000 * 1000 / 2)}

	uploadSizes = make([]int, len(uploadSizeSizes) * 25)
	for _, size := range uploadSizeSizes {
		for i := 0; i < 25; i++ {
			uploadSizes[i] = size
		}
	}
}

const safeChars = "0123456789abcdefghijklmnopqrstuv"

type safeReader struct {
	in io.Reader
}

func (r safeReader) Read(p []byte) (n int, err error) {
	n, err = r.in.Read(p)

	for i := 0; i < n; i++ {
		p[i] = safeChars[p[i] & 31]
	}

	return n, err
}

func (client *Client) uploadFile(url string, start time.Time, size int, ret chan int) {
	totalWrote := 0
	defer func() {
		ret <- totalWrote
	}()

	if (time.Since(start) > maxUploadDuration) {
		return;
	}
	if !client.opts.Quiet {
		os.Stdout.WriteString(".")
		os.Stdout.Sync()
	}

	resp, err := client.Post(
		url,
		"application/x-www-form-urlencoded",
		io.MultiReader(
			strings.NewReader("content1="),
			io.LimitReader(&safeReader{rand.Reader}, int64(size - 9))))
	if err != nil {
		log.Printf("[%s] Upload failed: %v\n", url, err)
		return;
	}

	totalWrote = size

	defer resp.Body.Close()
}

func (server *Server) UploadSpeed() int {
	client := server.client
	if !client.opts.Quiet {
		os.Stdout.WriteString("Testing upload speed: ")
		os.Stdout.Sync()
	}

	starterChan := make(chan int, uploadStreamLimit)
	uploads := uploadRepeats * len(uploadSizes)
	resultChan := make(chan int, uploadStreamLimit)
	start := time.Now()

	go func() {
		for _, size := range uploadSizes {
			for i := 0; i < uploadRepeats; i++ {
				url := server.URL
				starterChan <- 1
				go func() {
					client.uploadFile(url, start, size, resultChan)
					<-starterChan
				}()
			}
		}
		close(starterChan)
	}()

	var totalSize int64 = 0;

	for i := 0; i < uploads; i++ {
		totalSize += int64(<-resultChan)
	}

	if !client.opts.Quiet {
		os.Stdout.WriteString("\n")
		os.Stdout.Sync()
	}

	duration := time.Since(start);

	return int(totalSize * int64(time.Second) / int64(duration))
}
