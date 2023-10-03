package mangacrawler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type Chapter struct {
	Url  string      `json:"baseUrl"`
	Data ChapterData `json:"chapter"`
}

type ChapterData struct {
	Pages []string `json:"data"`
	Hash  string   `json:"hash"`
}

func chapterDownload(chapterId string, chapterPath string, chapterNo string) {
	var pages Chapter

	url := "https://api.mangadex.org/at-home/server/" + chapterId
	data := GetJson(url)

	if err := json.Unmarshal(data, &pages); err != nil {
		log.Fatal(err)
	}

	for _, page := range pages.Data.Pages {
		url = strings.Join([]string{pages.Url, "data", pages.Data.Hash, page}, "/")
		pageDownload(url, chapterPath, page, chapterNo)
	}
}

func pageDownload(url string, path string, page string, chapterNo string) {
	filepage := page
	regMatch, _ := regexp.MatchString(`^\D`, filepage)
	if regMatch {
		filepage = filepage[1:]
	}
	fileSplit := strings.Split(filepage, ".")
	filepage = strings.Join([]string{fmt.Sprintf("%067s", fileSplit[0]), fileSplit[1]}, ".")

	if _, err := os.Stat(path + "/chapter" + chapterNo + "_" + filepage); err == nil {
		return
	}

	result, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer result.Body.Close()

	if result.StatusCode != 200 {
		pageDownload(url, path, page, chapterNo)
		return
	}

	file, err := os.Create(path + "/chapter" + chapterNo + "_" + filepage)
	if err != nil {
		panic(err)
	}

	_, err = io.Copy(file, result.Body)
	if err != nil {
		panic(err)
	}

	file.Close()

	fmt.Printf("Downloading: %s\n", filepage)
}
