package mangacrawler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type MangaYaml struct {
	Name      string
	ID        string
	Chapter   float64
	Completed bool
}

func GetManga(manga MangaYaml, filepath string, forceDl bool) (MangaYaml, bool) {
	chaptersData := getChapterInfo(manga.ID)
	newChapter := false
	latestChapter := manga.Chapter

	// set subdirs for chapters in style volume-chapter-name
	for _, chapter := range chaptersData {
		// chapterVolume := chapter.Attributes.Volume
		chapterIndex, _ := strconv.ParseFloat(chapter.Attributes.Chapter, 32)
		if chapterIndex > manga.Chapter || forceDl {
			newChapter = true
			chapterChapter := fmt.Sprintf("%03s", chapter.Attributes.Chapter)
			extraChapter := strings.Split(chapterChapter, ".")
			if len(extraChapter) > 1 {
				chapterChapter = strings.Join([]string{fmt.Sprintf("%03s", extraChapter[0]), "z" + extraChapter[1]}, "-")
			}
			chapterTitle := chapter.Attributes.Title

			fmt.Printf("Working on Chapter: %s %s\n", chapterChapter, chapterTitle)
			chapterpath := strings.Join([]string{filepath, "chapter" + chapterChapter}, "/")
			if len(chapterTitle) > 0 {
				chapterpath = strings.Join([]string{filepath, "chapter" + chapterChapter + "-" + chapterTitle}, "/")
			}
			os.MkdirAll(chapterpath, 0755)
			chapterDownload(chapter.Id, chapterpath, chapterChapter)
			fmt.Println()
			time.Sleep(1 * time.Second)
		}
		if chapterIndex > latestChapter {
			latestChapter = chapterIndex
		}
	}
	if !newChapter {
		fmt.Print("  No new chapter released yet!\n\n")
	}

	manga.Chapter = latestChapter

	return manga, newChapter
}

func GetJson(url string) []byte {
	response, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	return data
}
