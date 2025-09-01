package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScrapeAndWriteLOTRCard(t *testing.T) {
	f, err := os.Open("testdata/lotr01001.html")
	if err != nil {
		t.Fatalf("failed to open test HTML: %v", err)
	}
	defer f.Close()

	card, err := ScrapeLOTRCard(f, "https://whatever.com")
	if err != nil {
		t.Fatalf("ScrapeLOTRCard failed: %v", err)
	}

	tmpDir := t.TempDir()
	setDir := filepath.Join(tmpDir, card.SetNo)

	imgPath, err := DownloadImage(card.ImageURL, setDir, card.ImagePath)
	if err != nil {
		t.Fatalf("DownloadImage failed: %v", err)
	}

	// assert image file exists
	if _, err := os.Stat(imgPath); err != nil {
		t.Fatalf("expected image file not created: %v", err)
	}

	mdPath, err := WriteCard(tmpDir, card)
	if err != nil {
		t.Fatalf("WriteCard failed: %v", err)
	}

	// assert markdown exists
	if _, err := os.Stat(mdPath); err != nil {
		t.Fatalf("expected markdown file not created: %v", err)
	}

	data, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	content := string(data)

	// assert frontmatter content
	if !strings.Contains(content, "# The One Ring, Isildur's Bane") {
		t.Errorf("content missing card title:\n%s", content)
	}
	if !strings.Contains(content, "rarity: \"R\"") {
		t.Errorf("content missing rarity:\n%s", content)
	}
	if !strings.Contains(content, "set_no: 01") {
		t.Errorf("content missing set number:\n%s", content)
	}
	if !strings.Contains(content, "card_no: 001") {
		t.Errorf("content missing card number:\n%s", content)
	}
	if !strings.Contains(content, "card_type: \"The One Ring\"") {
		t.Errorf("content missing card_type:\n%s", content)
	}
	if !strings.Contains(content, "strength: \"+1\"") {
		t.Errorf("content missing strength:\n%s", content)
	}
	if !strings.Contains(content, "vitality: \"+1\"") {
		t.Errorf("content missing vitality:\n%s", content)
	}
	if !strings.Contains(content, "game_text:") {
		t.Errorf("content missing game_text:\n%s", content)
	}
	if !strings.Contains(content, "lore:") {
		t.Errorf("content missing lore:\n%s", content)
	}

	// assert frontmatter references the image correctly
	if !strings.Contains(content, card.ImagePath) {
		t.Errorf("frontmatter missing image filename: %s", card.ImagePath)
	}

	wantFilename := filepath.Join(tmpDir, "01", "01001_the_one_ring_isildurs_bane.md")
	if mdPath != wantFilename {
		t.Errorf("filename mismatch: got %q, want %q", mdPath, wantFilename)
	}
}

func TestMain(t *testing.T) {}
