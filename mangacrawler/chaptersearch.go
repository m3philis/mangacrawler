package mangacrawler

import (
	"encoding/json"
	"strconv"
)

type Chapters struct {
	Data  []ChaptersData `json:"data"`
	Total int            `json:"total"`
}

type ChaptersData struct {
	Id         string             `json:"id"`
	Attributes ChaptersAttributes `json:"attributes"`
	Rels       []ChaptersRels     `json:"relationships"`
}

type ChaptersAttributes struct {
	Volume   string `json:"volume"`
	Chapter  string `json:"chapter"`
	Title    string `json:"title"`
	Language string `json:"translatedLanguage"`
}

type ChaptersRels struct {
	RelsAttr RelsAttributes `json:"attributes"`
}

type RelsAttributes struct {
	Name string `json:"name"`
}

func getChapterInfo(mangaId string) []ChaptersData {
	var tempChapters Chapters
	var chapters Chapters

	url := "https://api.mangadex.org/manga/" + mangaId + "/feed"

	data := GetJson(url)

	if err := json.Unmarshal(data, &tempChapters); err != nil {
		panic(err)
	}

	if tempChapters.Total > 100 {
		var chaptersOffset Chapters
		offset := 1
		maxOffset := tempChapters.Total / 100

		for offset <= maxOffset {
			url = "https://api.mangadex.org/manga/" + mangaId + "/feed?offset=" + strconv.Itoa(offset*100)

			data = GetJson(url)

			if err := json.Unmarshal(data, &chaptersOffset); err != nil {
				panic(err)
			}

			tempChapters.Data = append(tempChapters.Data, chaptersOffset.Data...)

			offset++
		}
	}

	for _, chapter := range tempChapters.Data {
		if chapter.Attributes.Language == "en" {
			chapters.Data = append(chapters.Data, chapter)
		}
	}

	return chapters.Data

}
