package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"mangacrawler/createepub"
	"mangacrawler/mangacrawler"

	"gopkg.in/yaml.v2"
)

func main() {
	// get infos for the manga we want to download
	var file string
	var forceDl bool
	var forceEpub bool
	var mangas []mangacrawler.MangaYaml
	var skipDl bool

	flag.BoolVar(&forceEpub, "force-epub", false, "Flag for creating an EPUB from the manga")
	flag.BoolVar(&skipDl, "skip-download", false, "Flag for not downloading the manga")
	flag.BoolVar(&forceDl, "force-download", false, "Download already downloaded chapters")
	flag.StringVar(&file, "file", "", "File with manga IDs. If not provided you need to add manga IDs as arguments")
	flag.Parse()

	if len(flag.Args()) == 0 && file == "" {
		fmt.Printf("Usage: %s [options] [--file /path/to/yaml] or [id1 id2 ...]\n\nParameters:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(0)
	}

	if file != "" {
		mangas = parseFile(file)
	} else {
		for _, id := range flag.Args() {
			var manga mangacrawler.MangaYaml

			manga.ID = id
			manga.Chapter = -1
			manga.Completed = false

			mangas = append(mangas, manga)
		}
	}

	homepath, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	for i, manga := range mangas {
		manga.Name, manga.Completed = mangacrawler.GetMangaInfo(manga)
		var newChapter bool

		mangapath := strings.Join([]string{homepath, "mangas/MangaDex", manga.Name}, "/")
		os.MkdirAll(mangapath, 0755)

		if (!manga.Completed && !skipDl) || forceDl {
			manga, newChapter = mangacrawler.GetManga(manga, mangapath, forceDl)
		} else if manga.Completed {
			fmt.Print("  Manga already completed!\n\n")
		}

		if _, err := os.Stat(strings.Join([]string{homepath, "mangas/EPUB", manga.Name + ".epub"}, "/")); err != nil || forceEpub || newChapter {
			epubPath := strings.Join([]string{homepath, "mangas/EPUB"}, "/")
			os.MkdirAll(epubPath, 0755)
			fmt.Println("Generating EPUB")
			createepub.CreateEpub(mangapath, manga.Name, manga.ID)
			fmt.Printf("EPUB created and saved under: %s\n\n", epubPath)
		} else {
			fmt.Print("EPUB exists already!\n\n")
		}

		mangas[i] = manga
	}

	if file != "" {
		writeFile(file, mangas)
	}

	if file == "" {
		yamlPrint, _ := yaml.Marshal(&mangas)
		fmt.Println(string(yamlPrint))
	}
}

func parseFile(file string) []mangacrawler.MangaYaml {
	var fBytes []byte
	var yamlData []mangacrawler.MangaYaml

	if !strings.HasPrefix(file, "/") {
		cwd, _ := os.Getwd()
		if _, err := os.Stat(strings.Join([]string{cwd, file}, "/")); err != nil {
			log.Fatal("File not found: ", file)
		}
		fBytes, _ = os.ReadFile(strings.Join([]string{cwd, file}, "/"))
	} else {
		if _, err := os.Stat(file); err != nil {
			log.Fatal("File not found: ", file)
		}
		fBytes, _ = os.ReadFile(file)
	}

	err := yaml.Unmarshal(fBytes, &yamlData)
	if err != nil {
		panic(err)
	}

	mangas := yamlData

	return mangas
}

func writeFile(file string, mangas []mangacrawler.MangaYaml) {
	var filePath string

	if !strings.HasPrefix(file, "/") {
		cwd, _ := os.Getwd()
		if _, err := os.Stat(strings.Join([]string{cwd, file}, "/")); err != nil {
			log.Fatal("File not found: ", file)
		}
		filePath = strings.Join([]string{cwd, file}, "/")
	} else {
		if _, err := os.Stat(file); err != nil {
			log.Fatal("File not found: ", file)
		}
		filePath = file
	}

	data, err := yaml.Marshal(&mangas)
	if err != nil {
		panic(err)
	}

	_ = os.WriteFile(filePath, data, 0644)
}
