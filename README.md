# LOTR TCG Crawler

This is a Go-based scraper that extracts Lord of the Rings TCG card data, downloads their images, and generates Markdown files with frontmatter metadata for each card.

The script currently scrapes sets 1, 2, and 3 by default.

## Features

• Fetches card data (title, set number, card number, type, stats, lore, etc.).
• Downloads card images into per-set folders.
• Generates Markdown files with YAML frontmatter for each card:
• Title, set number, card number.
• Scraped stats (strength, vitality, culture, rarity, etc.).
• Placeholders for collection info (amount, value, sold_price, etc.).
• Link to the downloaded image.
• Configurable via .env file (base URL for scraping).

⸻

Requirements
• Go 1.21+
• Dependencies (installed via go mod tidy):
• PuerkitoBio/goquery — HTML scraping.
• joho/godotenv — .env file support.

⸻

Setup 1. Clone the repo:

git clone <https://github.com/snelling-a/lotr-tcg-crawler.git>
cd lotr-tcg-crawler

    2. Create a .env file in the project root:

```sh
export BASE_URL=https://link/to/card/data
```

    3. Build the project:

go build -o crawler

## Usage

Run the crawler:

./crawler

It will:

- Loop through sets 1–3.
- Download images into `output/<setNo>/`.
- Write card Markdown files into `output/<setNo>/`.

### Example output

```
output/
01/
    01001_the_one_ring_isildurs_bane.md
    the_one_ring_isildurs_bane.jpg
02/
...
```

### Example Markdown

```markdown
---
amount:
card_no: 001
card_type: "The One Ring"
currency: €
offer_price:
photo: "[[./the_one_ring_isildurs_bane.jpg|The One Ring, Isildur's Bane]]"
rarity: "R"
set_no: 01
sold_price:
strength: "7"
title: "The One Ring, Isildur's Bane"
total:
value:
vitality: "3"
---

# The One Ring, Isildur's Bane
```

## Notes

- Stops scraping a set once it hits a missing card (non-200 response).
- You can change the sets to scrape by editing the `for \_, set := range []int{1, 2, 3}` loop in main.go.
- Markdown is ready for integration into Obsidian or static site generators.
