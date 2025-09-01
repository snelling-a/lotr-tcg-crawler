package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
)

type CardData struct {
	CardNo    string
	ImageName string
	ImagePath string
	ImageURL  string
	SetNo     string
	Title     string
	Props     map[string]string
}

var spaceRe = regexp.MustCompile(`\s+`)

func normalizeSpaces(s string) string {
	return strings.TrimSpace(spaceRe.ReplaceAllString(s, " "))
}

func normalizeKey(key string) string {
	k := strings.ToLower(key)
	re := regexp.MustCompile(`[^a-z0-9]+`)
	k = re.ReplaceAllString(k, "_")
	return strings.Trim(k, "_")
}

func DownloadImage(url, dir, filename string) (string, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	path := filepath.Join(dir, filename)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return "", err
	}
	return path, nil
}

func ScrapeLOTRCard(r io.Reader, url string) (*CardData, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	h1 := strings.TrimSpace(doc.Find("h1 a").First().Text())
	if h1 == "" {
		return nil, fmt.Errorf("h1 not found")
	}
	title := strings.TrimSpace(strings.Split(h1, "(")[0])

	re := regexp.MustCompile(`(\d+)([A-Z])(\d+)`)
	m := re.FindStringSubmatch(h1)
	if len(m) != 4 {
		return nil, fmt.Errorf("metadata not found in title: %s", h1)
	}
	setNo := fmt.Sprintf("%02s", m[1])
	cardNo := fmt.Sprintf("%03s", m[3])

	props := make(map[string]string)
	doc.Find("table.inline tr").Each(func(i int, s *goquery.Selection) {
		key := strings.TrimSpace(s.Find("td.col0").Text())
		val := strings.TrimSpace(s.Find("td.col1").Text())
		if key == "" || val == "" {
			return
		}
		props[normalizeKey(key)] = normalizeSpaces(val)
	})

	imgSel := doc.Find("p span a img.media").First()
	src, exists := imgSel.Attr("src")
	if !exists {
		return nil, fmt.Errorf("image src not found")
	}
	titleAttr, _ := imgSel.Attr("title")
	fullURL := url + src
	imagePath := sanitizeFilename(title) + filepath.Ext(src)

	return &CardData{
		Title:     title,
		SetNo:     setNo,
		CardNo:    cardNo,
		ImageURL:  fullURL,
		ImagePath: imagePath,
		ImageName: titleAttr,
		Props:     props,
	}, nil
}

func sanitizeFilename(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "'", "")
	re := regexp.MustCompile(`[^a-z0-9]+`)
	s = re.ReplaceAllString(s, "_")
	return strings.Trim(s, "_")
}

func WriteCard(baseDir string, card *CardData) (string, error) {
	dir := filepath.Join(baseDir, card.SetNo)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	filename := fmt.Sprintf("%s%s_%s.md",
		card.SetNo, card.CardNo, sanitizeFilename(card.Title))
	path := filepath.Join(dir, filename)

	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("title: \"%s\"\n", card.Title))
	b.WriteString(fmt.Sprintf("set_no: %s\n", card.SetNo))
	b.WriteString(fmt.Sprintf("card_no: %s\n", card.CardNo))
	b.WriteString(fmt.Sprintf("photo: \"[[./%s|%s]]\"\n", card.ImagePath, card.ImageName))
	for k, v := range card.Props {
		b.WriteString(fmt.Sprintf("%s: \"%s\"\n", k, v))
	}
	b.WriteString("amount:\n")
	b.WriteString("value:\n")
	b.WriteString("sold_price:\n")
	b.WriteString("offer_price:\n")
	b.WriteString("total:\n")
	b.WriteString("currency: €\n")
	b.WriteString("---\n\n")
	b.WriteString("# " + card.Title + "\n")

	if err := os.WriteFile(path, []byte(b.String()), 0644); err != nil {
		return "", err
	}
	return path, nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	baseURL := os.Getenv("BASE_URL")

	tmpDir := "output"

	for _, set := range []int{1, 2, 3} {
		fmt.Printf("Scraping set %d\n", set)

		for i := 1; ; i++ {
			cardNo := fmt.Sprintf("%02d%03d", set, i)
			url := fmt.Sprintf("%s/lotr%s", baseURL, cardNo)
			fmt.Printf("Fetching %s\n", url)

			resp, err := http.Get(url)
			if err != nil {
				fmt.Printf("Error fetching %s: %v\n", url, err)
				break
			}
			if resp.StatusCode != 200 {
				resp.Body.Close()
				break
			}

			card, err := ScrapeLOTRCard(resp.Body, url)
			resp.Body.Close()
			if err != nil {
				fmt.Printf("Error scraping %s: %v\n", url, err)
				break
			}

			card.CardNo = fmt.Sprintf("%03d", i)
			card.SetNo = fmt.Sprintf("%02d", set)

			setDir := filepath.Join(tmpDir, card.SetNo)
			_, err = DownloadImage(card.ImageURL, setDir, card.ImagePath)
			if err != nil {
				fmt.Printf("Failed to download image for %s: %v\n", card.Title, err)
				continue
			}

			mdPath, err := WriteCard(tmpDir, card)
			if err != nil {
				fmt.Printf("Failed to write markdown for %s: %v\n", card.Title, err)
				continue
			}

			fmt.Printf("Card %s saved: %s\n", card.Title, mdPath)
		}
	}
}
