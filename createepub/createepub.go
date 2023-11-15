package createepub

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mangacrawler/mangacrawler"
	"net/http"
	"os"
  "os/exec"
	"os/user"
	"regexp"
	"strings"

	"github.com/go-shiori/go-epub"
)

type MangaPlus struct {
	Data MangaRelationships `json:"data"`
}

type MangaRelationships struct {
	Rels []MangaRels     `json:"relationships"`
	Attr MangaAttributes `json:"attributes"`
}

type MangaRels struct {
	Attributes MangaAuthor `json:"attributes"`
	Type       string      `json:"type"`
}

type MangaAuthor struct {
	Name string `json:"name"`
	File string `json:"fileName"`
}

type MangaAttributes struct {
	Desc MangaDesc `json:"description"`
}

type MangaDesc struct {
	En string `json:"en"`
}

func CreateEpub(mangaPath string, mangaTitle string, mangaId string) {
	var author MangaPlus

	url := "https://api.mangadex.org/manga/" + mangaId + "?includes[]=author&includes[]=cover_art"
	data := mangacrawler.GetJson(url)
	homepath, err := user.Current()
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(data, &author); err != nil {
		panic(err)
	}

	fmt.Println("Downloading and adding cover page for EPUB")
	var coverPath string
	var coverFile string
	for _, rels := range author.Data.Rels {
		if rels.Type == "cover_art" {
			coverPath, coverFile = getCoverPage(mangaId, rels.Attributes.File, mangaPath)
		}
	}

	book := epub.NewEpub(mangaTitle)
	book.SetAuthor(author.Data.Rels[0].Attributes.Name)
	book.SetDescription(author.Data.Attr.Desc.En)
	bookCss, _ := book.AddCSS(strings.Join([]string{homepath.HomeDir, "mangas/EPUB", "epub.css"}, "/"), "")
	bookCover, _ := book.AddImage(coverPath, coverFile)

	book.SetCover(bookCover, "")
	fmt.Println("Cover page added")

	fmt.Println("Adding pages to EPUB. Each chapter is a section\nIf chapter title is available that will be used for section title")
	addPages(book, mangaPath, bookCss)

	fmt.Println("Writing EPUB to disk...")
	err = book.Write(strings.Join([]string{homepath.HomeDir, "mangas/EPUB", mangaTitle + ".epub"}, "/"))
	if err != nil {
		log.Fatal(err)
	}

  fmt.Println("Adding EPUB to calibre DB")
  cmd := exec.Command("calibredb", "add", "--automerge", "overwrite", strings.Join([]string{homepath.HomeDir, "mangas/EPUB", mangaTitle + ".epub"}, "/"))
  var out strings.Builder
  cmd.Stdout = &out
  fmt.Println(cmd)
  err = cmd.Run()
  if err != nil {
    log.Fatal(err)
  }
  fmt.Printf(out.String())

}

func addPages(book *epub.Epub, mangaPath string, bookCss string) *epub.Epub {
	chapters, err := os.ReadDir(mangaPath)
	if err != nil {
		panic(err)
	}

	titleCompile, _ := regexp.Compile(`^[A-Za-z][^\d]`)
	bonusChapterCompile, _ := regexp.Compile(`^(z\d+)`)
	chapterIndexCompile, _ := regexp.Compile(`chapter0*(\d+)`)

	for _, chapter := range chapters {
		var section string

		if !strings.HasPrefix(chapter.Name(), "chapter") {
			continue
		}

		chapterNo := chapterIndexCompile.FindStringSubmatch(chapter.Name())[1]
		// fmt.Println(chapterNo)
		pages, _ := os.ReadDir(strings.Join([]string{mangaPath, chapter.Name()}, "/"))

		for i, page := range pages {
			var sectionBody string
			var subSectionBody string

			bookPage, _ := book.AddImage(strings.Join([]string{mangaPath, chapter.Name(), page.Name()}, "/"), page.Name())

			if i == 0 {
				sectionBody = fmt.Sprintf("<img src=\"%s\" class=\"chapter\">\n", bookPage)

				if len(chapter.Name()) > 10 {
					titleMatch := titleCompile.MatchString(chapter.Name()[11:])
					bonusChapterMatch := bonusChapterCompile.MatchString(chapter.Name()[11:])

					if bonusChapterMatch && len(chapter.Name()) > 13 {
						bonusChapterNo := bonusChapterCompile.FindStringSubmatch(chapter.Name()[11:])
						section, err = book.AddSection(sectionBody, "Chapter "+chapterNo+strings.Replace(bonusChapterNo[1], "z", ".", 1)+": "+chapter.Name()[14:], "", bookCss)
						if err != nil {
							panic(err)
						}
					} else if bonusChapterMatch {
						bonusChapterNo := bonusChapterCompile.FindStringSubmatch(chapter.Name()[11:])
						section, err = book.AddSection(sectionBody, "Chapter "+chapterNo+strings.Replace(bonusChapterNo[1], "z", ".", 1), "", bookCss)
						if err != nil {
							panic(err)
						}
					} else if titleMatch {
						section, err = book.AddSection(sectionBody, "Chapter "+chapterNo+": "+chapter.Name()[11:], "", bookCss)
						if err != nil {
							panic(err)
						}

					}
				} else {
					section, err = book.AddSection(sectionBody, "Chapter "+chapterNo, "", bookCss)
					if err != nil {
						panic(err)
					}
				}
			} else {
				subSectionBody = fmt.Sprintf("<img src=\"%s\" class=\"page\">\n", bookPage)
				_, _ = book.AddSubSection(section, subSectionBody, "", "", bookCss)
			}
		}
	}
	return book
}

func getCoverPage(mangaId string, coverFile string, mangaPath string) (string, string) {
	_, err := os.Stat(strings.Join([]string{mangaPath, coverFile}, "/"))
	if err == nil {
		return strings.Join([]string{mangaPath, coverFile}, "/"), coverFile
	}
	url := strings.Join([]string{"https://uploads.mangadex.org/covers", mangaId, coverFile}, "/")
	result, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer result.Body.Close()

	file, err := os.Create(strings.Join([]string{mangaPath, coverFile}, "/"))
	if err != nil {
		panic(err)
	}

	_, err = io.Copy(file, result.Body)
	if err != nil {
		panic(err)
	}

	file.Close()

	return strings.Join([]string{mangaPath, coverFile}, "/"), coverFile
}
