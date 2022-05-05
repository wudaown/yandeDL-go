package crawler

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/schollz/progressbar/v3"
)

func CreateDir(dirname string) {
	path, err := filepath.Abs(dirname)
	if err != nil {
		log.Println(err)
	}

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}

	os.Chdir(path)
}

func AskTag(tag string) (full_url string) {
	// reader := bufio.NewReader(os.Stdin)
	url := "https://yande.re/post/?tags="
	// fmt.Printf("Input tag to search for: ")
	// tag, _ := reader.ReadString('\n')
	// tag = strings.Replace(tag, "\n", "", -1)
	full_url = fmt.Sprintf("%s%s", url, tag)
	return full_url
}

func GetSource(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	html, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	// convert []byte to string
	// fmt.Printf("%s\n", string(html))
	html_string := string(html)
	re := regexp.MustCompile(`Nobody`)
	tag_not_exists := re.MatchString(html_string)
	if tag_not_exists {
		return "", errors.New("tag does not exists. please try again")
	}
	return html_string, nil
}

func GetImgLink(html string) map[string]string {
	linkFileName := make(map[string]string)
	re := regexp.MustCompile(`directlink \w{5}img"(.+?.jpg)`)
	unprocess_link := re.FindAllString(html, -1)
	for _, v := range unprocess_link {
		link := v[27:]
		path, err := url.PathUnescape(link)
		if err != nil {
			log.Fatal(err)
		}
		fileName := strings.SplitN(path, "yande.re", 3)[2]
		linkFileName[fileName] = link
	}
	return linkFileName
}

func DownloadFile(wg *sync.WaitGroup, ch chan bool, URL, fileName string) error {
	defer wg.Done()
	ch <- true
	log.Println("Downloading: ", fileName)
	// Get the response bytes from the url
	response, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.New("Received non 200 response code")
	}
	//Create a empty file
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	bar := progressbar.DefaultBytes(
		response.ContentLength,
		"downloading",
		// fmt.Sprintf("Downloading %s", fileName[:11]),
	)

	//Write the bytes to the fiel
	_, err = io.Copy(io.MultiWriter(file, bar), response.Body)
	if err != nil {
		return err
	}

	<-ch
	return nil
}
