package mangacrawler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Manga struct {
	Data MangaData `json:"data"`
}

type MangaData struct {
	Attributes MangaAttributes `json:"attributes"`
}

type MangaAttributes struct {
	Title       Titles   `json:"title"`
	AltTitle    []Titles `json:"altTitles"`
	Status      string   `json:"status"`
	LastChapter string   `json:"lastChapter"`
}

type Titles struct {
	JP string `json:"ja-ro"`
	EN string `json:"en"`
}

func GetMangaInfo(mangaYaml MangaYaml) (string, bool) {
	var manga Manga
	status := false
	homepath, _ := os.UserHomeDir()

	url := "https://api.mangadex.org/manga/" + mangaYaml.ID
	data := GetJson(url)
	if err := json.Unmarshal(data, &manga); err != nil {
		panic(err)
	}

	mangaLastChapter, _ := strconv.ParseFloat(manga.Data.Attributes.LastChapter, 32)
	if manga.Data.Attributes.Status == "completed" && (mangaLastChapter <= mangaYaml.Chapter || manga.Data.Attributes.LastChapter == "") {
		status = true
	}

	// set home directory and create subdir to save manga in
	mangaTitles := []string{manga.Data.Attributes.Title.EN}
	for _, title := range manga.Data.Attributes.AltTitle {
		if title.EN != "" {
			mangaTitles = append(mangaTitles, title.EN)
		} else if title.JP != "" {
			mangaTitles = append(mangaTitles, title.JP)
		}
	}
	for _, title := range mangaTitles {
		if _, err := os.Stat(strings.Join([]string{homepath, "mangas/MangaDex", title}, "/")); err == nil && title != "" {
			fmt.Printf("Title found on system! Using: %s\n", title)
			return title, status
		}
	}

	for i, title := range mangaTitles {
		fmt.Printf("(%d): %s\n", i, title)
	}
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("---\nPlease choose title for the manga: ")
	selection, _ := reader.ReadString('\n')
	selection = strings.TrimSuffix(selection, "\n")
	choice, _ := strconv.Atoi(selection)

	mangaTitle := mangaTitles[choice]

	return mangaTitle, status
}
