/**
 * Copyright (c) 2014-present, Facebook, Inc. All rights reserved.
 *
 * You are hereby granted a non-exclusive, worldwide, royalty-free license to use,
 * copy, modify, and distribute this software in source code or binary form for use
 * in connection with the web services and APIs provided by Facebook.
 *
 * As with any software that integrates with the Facebook platform, your use of
 * this software is subject to the Facebook Developer Principles and Policies
 * [http://developers.facebook.com/policy/]. This copyright notice shall be
 * included in all copies or substantial portions of the software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
 * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

// Package examples provides stock and stored examples.
package examples

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

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
	"saved":  true,
}

type Store struct {
	DB *DB
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

var (
	// Stock response for the index page.
	emptyExample = &Example{Title: "Welcome", URL: "/", AutoRun: true}
	classExample = &url.URL{Path: "classes/Example"}
)

func MustMakeDB(dir string) *DB {
	db, err := MakeDB(dir)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// Loads a specific examples directory.
func MakeDB(dir string) (*DB, error) {
	db := &DB{
		Category: make(map[string]*Category),
		Reverse:  make(map[string]*Example),
	}
	db.Reverse[ContentID(emptyExample.Content)] = emptyExample

	err := filepath.Walk(
		dir,
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

			contentBytes, err := ioutil.ReadFile(exampleFile)
			if err != nil {
				return fmt.Errorf("Failed to read example %s: %s", exampleFile, err)
			}

			autoRun := true
			if categoryName == "saved" {
				autoRun = false
			}

			content := string(contentBytes)
			cleanName := exampleName[:len(exampleName)-5] // drop .html
			example := &Example{
				Name:    cleanName,
				Content: content,
				AutoRun: autoRun,
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

// Load an Example for a given version and path.
func (s *Store) Load(path string) (*Example, error) {
	parts := strings.Split(path, "/")
	if len(parts) == 2 && parts[1] == "" {
		return emptyExample, nil
	} else if len(parts) != 3 {
		return nil, errcode.New(http.StatusNotFound, "Invalid URL: %s", path)
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

func ContentID(content string) string {
	h := md5.New()
	_, err := fmt.Fprint(h, content)
	if err != nil {
		log.Fatalf("Error comupting md5 sum: %s", err)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
