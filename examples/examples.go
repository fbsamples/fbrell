// Package examples provides stock and stored examples.
package examples

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/GeertJohan/go.rice"
	"github.com/daaku/go.errcode"
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

type ByteStore interface {
	Store(key string, value []byte) error
	Get(key string) ([]byte, error)
}

type Store struct {
	ByteStore ByteStore
}

type Example struct {
	Name    string `json:"-"`
	Content []byte `json:"-"`
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

var (
	mu = mustLoadBox(rice.MustFindBox("db"))

	// Stock response for the index page.
	emptyExample = &Example{Title: "Welcome", URL: "/", AutoRun: true}
)

func mustLoadBox(box *rice.Box) *DB {
	db, err := loadBox(box)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// Loads a specific examples directory.
func loadBox(box *rice.Box) (*DB, error) {
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

			content, err := box.Bytes(exampleFile)
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
			db.Reverse[ContentID(bytes.TrimSpace(content))] = example
			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// Load an Example for a given version and path.
func (s *Store) Load(path string) (*Example, error) {
	parts := strings.Split(path, "/")
	if len(parts) == 2 && parts[1] == "" {
		return emptyExample, nil
	} else if len(parts) == 4 {
		if parts[1] != "raw" && parts[1] != "simple" {
			return nil, errcode.New(http.StatusNotFound, "Invalid URL: %s", path)
		}
		parts = []string{"", parts[2], parts[3]}
	} else if len(parts) != 3 {
		return nil, errcode.New(http.StatusNotFound, "Invalid URL: %s", path)
	}

	if parts[1] == "saved" {
		content, err := s.ByteStore.Get(makeKey(parts[2]))
		if err != nil {
			return nil, err
		}
		if content == nil {
			return nil, errcode.New(
				http.StatusNotFound, "Example not found: %s", path)
		}
		return &Example{
			Content: content,
			Title:   "Stored Example",
			URL:     path,
		}, nil
	}
	category := GetDB().FindCategory(parts[1])
	if category == nil {
		return nil, errcode.New(http.StatusNotFound, "Could not find category: %s", parts[1])
	}
	example := category.FindExample(parts[2])
	if example == nil {
		return nil, errcode.New(http.StatusNotFound, "Could not find example: %s", parts[2])
	}
	return example, nil
}

// Get the DB for a given SDK Version.
func GetDB() *DB {
	return mu
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
func (s *Store) Save(id string, content []byte) error {
	if len(content) > 10240 {
		return errcode.New(
			http.StatusRequestEntityTooLarge,
			"Maximum allowed size is 10 kilobytes.")
	}
	err := s.ByteStore.Store(makeKey(id), content)
	if err != nil {
		log.Printf("Error in ByteStore.Store: %s", err)
	}
	return err
}

func makeKey(id string) string {
	return "fbrell_examples:" + id
}

func ContentID(content []byte) string {
	h := md5.New()
	_, err := h.Write(content)
	if err != nil {
		log.Fatalf("Error comupting md5 sum: %s", err)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
