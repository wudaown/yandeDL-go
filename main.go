package main

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/wudaown/yandeDL-go/crawler"
)

func main() {
	tagFlag := flag.String("t", "", "`tag` to search for")
	concurrentFlag := flag.Int("c", 1, "`concurrent` task number.")
	dirFlag := flag.String("d", "", "`directory` to keep the images. Defaults to `pwd`/`tag`")
	flag.Parse()
	if *tagFlag == "" {
		// fmt.Println(flag.Args())
		flag.Usage()
		os.Exit(2) // the same exit code flag.Parse uses
	}
	if *dirFlag == "" {
		*dirFlag = *tagFlag
	}
	start := time.Now()
	wg := new(sync.WaitGroup)
	ch := make(chan bool, *concurrentFlag)
	crawler.CreateDir(*dirFlag)
	url := crawler.AskTag(*tagFlag)
	html, err := crawler.GetSource(url)
	if err != nil {
		return
	}
	mapping := crawler.GetImgLink(html)
	for filename, url := range mapping {
		wg.Add(1)
		go crawler.DownloadFile(wg, ch, url, filename)
	}

	wg.Wait()
	elapsed := time.Since(start)
	fmt.Printf("Took %s", elapsed)
}
