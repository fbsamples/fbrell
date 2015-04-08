// Package examples provides stock and stored examples.
package examples

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/daaku/rell/internal/github.com/GeertJohan/go.rice"
	"github.com/daaku/rell/internal/github.com/daaku/go.errcode"
	"github.com/daaku/rell/internal/github.com/facebookgo/parse"
	"github.com/daaku/rell/internal/github.com/golang/groupcache/lru"
)

// Some categories are hidden from the listing.
var hidden = map[string]bool{
	"auth":   true,
	"bugs":   true,
	"fb.api": true,
	"fb.ui":  true,
	"hidden": true,
	"secret": true,
	"tests":  true,
	"xfbml":  true,
	"canvas": true,
}

type Store struct {
	Parse *parse.Client
	DB    *DB
	Cache *lru.Cache
}

type Example struct {
	Name    string `json:"-"`
	Content string `json:"-"`
	AutoRun bool   `json:"autoRun"`
	Title   string `json:"-"`
	URL     string `json:"-"`
}

type Category struct {
	Name    string
	Example []*Example
	Hidden  bool
}

type DB struct {
	Category map[string]*Category
	Reverse  map[string]*Example
}

type parseExample struct {
	Content string `json:"content,omitempty"`
	Hash    string `json:"hash,omitempty"`
}

var (
	// Stock response for the index page.
	emptyExample = &Example{Title: "Welcome", URL: "/", AutoRun: true}
	classExample = &url.URL{Path: "classes/Example"}
)

func MustMakeDB(box *rice.Box) *DB {
	db, err := MakeDB(box)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// Loads a specific examples directory.
func MakeDB(box *rice.Box) (*DB, error) {
	db := &DB{
		Category: make(map[string]*Category),
		Reverse:  make(map[string]*Example),
	}
	db.Reverse[ContentID(emptyExample.Content)] = emptyExample

	err := box.Walk(
		"",
		func(exampleFile string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			categoryName := filepath.Base(filepath.Dir(exampleFile))
			exampleName := filepath.Base(exampleFile)

			category := db.Category[categoryName]
			if category == nil {
				category = &Category{
					Name:   categoryName,
					Hidden: hidden[categoryName],
				}
				db.Category[categoryName] = category
			}

			content, err := box.String(exampleFile)
			if err != nil {
				return fmt.Errorf("Failed to read example %s: %s", exampleFile, err)
			}
			cleanName := exampleName[:len(exampleName)-5] // drop .html
			example := &Example{
				Name:    cleanName,
				Content: content,
				AutoRun: true,
				Title:   categoryName + " · " + cleanName,
				URL:     path.Join("/", categoryName, cleanName),
			}
			category.Example = append(category.Example, example)
			db.Reverse[ContentID(strings.TrimSpace(content))] = example
			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	return db, nil
}

type cacheKey string

func (s *Store) loadFromParse(hash string) (*Example, error) {
	if ex, ok := s.Cache.Get(cacheKey(hash)); ok {
		return ex.(*Example), nil
	}
	j, err := json.Marshal(map[string]string{"hash": hash})
	if err != nil {
		return nil, err
	}
	v := make(url.Values)
	v.Set("where", string(j))
	u := &url.URL{
		Path:     classExample.Path,
		RawQuery: v.Encode(),
	}
	var res struct {
		Results []parseExample `json:"results"`
	}
	if _, err := s.Parse.Get(u, &res); err != nil {
		return nil, err
	}
	if len(res.Results) == 0 {
		return nil, errcode.New(
			http.StatusNotFound, "Example not found: %s", hash)
	}
	ex := &Example{
		Content: res.Results[0].Content,
		Title:   "Stored Example",
		URL:     path.Join("/saved", hash),
	}
	s.Cache.Add(cacheKey(hash), ex)
	return ex, nil
}

// Load an Example for a given version and path.
func (s *Store) Load(path string) (*Example, error) {
	parts := strings.Split(path, "/")
	if len(parts) == 2 && parts[1] == "" {
		return emptyExample, nil
	} else if len(parts) != 3 {
		return nil, errcode.New(http.StatusNotFound, "Invalid URL: %s", path)
	}

	if parts[1] == "saved" {
		return s.loadFromParse(parts[2])
	}
	category := s.DB.FindCategory(parts[1])
	if category == nil {
		return nil, errcode.New(http.StatusNotFound, "Could not find category: %s", parts[1])
	}
	example := category.FindExample(parts[2])
	if example == nil {
		return nil, errcode.New(http.StatusNotFound, "Could not find example: %s", parts[2])
	}
	return example, nil
}

// Find a category by it's name.
func (d *DB) FindCategory(name string) *Category {
	for _, category := range d.Category {
		if category.Name == name {
			return category
		}
	}
	return nil
}

// Find an example by it's name.
func (c *Category) FindExample(name string) *Example {
	for _, example := range c.Example {
		if example.Name == name {
			return example
		}
	}
	return nil
}

// Save an Example.
func (s *Store) Save(hash string, content string) error {
	if len(content) > 10240 {
		return errcode.New(
			http.StatusRequestEntityTooLarge,
			"Maximum allowed size is 10 kilobytes.")
	}

	if _, err := s.loadFromParse(hash); err == nil {
		return nil // it already exists
	}

	pe := &parseExample{
		Content: content,
		Hash:    hash,
	}
	if _, err := s.Parse.Post(classExample, pe, nil); err != nil {
		return err
	}

	ex := &Example{
		Content: content,
		Title:   "Stored Example",
		URL:     path.Join("/saved", hash),
	}
	s.Cache.Add(cacheKey(hash), ex)

	return nil
}

func ContentID(content string) string {
	h := md5.New()
	_, err := fmt.Fprint(h, content)
	if err != nil {
		log.Fatalf("Error comupting md5 sum: %s", err)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
