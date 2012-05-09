// Package examples implements the DB and S3 backed examples backend for Rell.
package examples

import (
	"crypto/md5"
	"errors"
	"flag"
	"fmt"
	"github.com/nshah/go.flag.pkgpath"
	"io/ioutil"
	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/s3"
	"log"
	"path/filepath"
	"strings"
)

// Some categories are hidden from the listing.
var hidden = map[string]bool{
	"bugs":   true,
	"fbml":   true,
	"hidden": true,
	"secret": true,
	"tests":  true,
}

type Example struct {
	Name    string `json:"-"`
	Content []byte `json:"-"`
	AutoRun bool   `json:"autoRun"`
	Title   string `json:"-"`
}

type Category struct {
	Name    string
	Example []*Example
	Hidden  bool
}

type DB struct {
	Category []*Category
}

var (
	// Directory for disk backed DBs.
	oldExamplesDir string
	newExamplesDir string

	// We have two disk backed DBs.
	old *DB
	mu  *DB

	// For stored examples.
	bucketMemo *s3.Bucket
	bucketName string
	auth       aws.Auth

	// Stock response for the index page.
	emptyExample = &Example{Title: "Welcome"}
)

// We load up the examples into memory on server start.
func init() {
	flag.StringVar(
		&bucketName,
		"rell.amazon.bucket",
		"fbrell_examples",
		"The Amazon bucket to store examples.")
	flag.StringVar(
		&auth.AccessKey,
		"rell.amazon.key",
		"",
		"The Amazon API key to access the bucket.")
	flag.StringVar(
		&auth.SecretKey,
		"rell.amazon.secret",
		"",
		"The Amazon API secret to access the bucket.")
	pkgpath.DirVar(
		&oldExamplesDir,
		"rell.examples.old",
		"github.com/nshah/rell/examples/db/old",
		"The directory containing examples for the old SDK.")
	pkgpath.DirVar(
		&newExamplesDir,
		"rell.examples.new",
		"github.com/nshah/rell/examples/db/mu",
		"The directory containing examples for the new SDK.")
}

// Get's the shared S3 bucket instance.
func bucket() *s3.Bucket {
	if bucketMemo == nil {
		bucketMemo = s3.New(auth, aws.USEast).Bucket(bucketName)
	}
	return bucketMemo
}

// Loads a specific examples directory.
func loadDir(path string) (*DB, error) {
	categories, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("Failed to read directory %s: %s", path, err)
	}
	db := &DB{Category: make([]*Category, 0, len(categories))}
	for _, categoryFileInfo := range categories {
		categoryName := categoryFileInfo.Name()
		if !categoryFileInfo.IsDir() {
			log.Printf(
				"Got unexpected file instead of directory for category: %s",
				categoryName)
			continue
		}
		category := &Category{
			Name:   categoryName,
			Hidden: hidden[categoryName],
		}
		categoryDir := filepath.Join(path, categoryName)
		examples, err := ioutil.ReadDir(categoryDir)
		if err != nil {
			return nil, fmt.Errorf("Failed to read category %s: %s", categoryDir, err)
		}
		category.Example = make([]*Example, 0, len(examples))
		for _, example := range examples {
			exampleName := example.Name()
			exampleFile := filepath.Join(categoryDir, exampleName)
			content, err := ioutil.ReadFile(exampleFile)
			if err != nil {
				return nil, fmt.Errorf(
					"Failed to read example %s: %s", exampleFile, err)
			}
			cleanName := exampleName[:len(exampleName)-5]
			category.Example = append(category.Example, &Example{
				Name:    cleanName,
				Content: content,
				AutoRun: true,
				Title:   categoryName + " · " + cleanName,
			})
		}
		db.Category = append(db.Category, category)
	}
	return db, nil
}

// Load an Example for a given version and path.
func Load(version, path string) (*Example, error) {
	parts := strings.Split(path, "/")
	if len(parts) == 2 && parts[1] == "" {
		return emptyExample, nil
	} else if len(parts) == 4 {
		if parts[1] != "raw" && parts[1] != "simple" {
			return nil, fmt.Errorf("Invalid URL: %s", path)
		}
		parts = []string{"", parts[2], parts[3]}
	} else if len(parts) != 3 {
		return nil, fmt.Errorf("Invalid URL: %s", path)
	}

	if parts[1] == "saved" {
		content, err := bucket().Get("/" + parts[2])
		if err != nil {
			s3Err, ok := err.(*s3.Error)
			if ok && s3Err.StatusCode == 404 {
				return nil, errors.New("Could not find saved example.")
			}
			log.Printf("Unknown S3 error: %s", err)
			return nil, err
		}
		return &Example{Content: content, Title: "Stored Example"}, nil
	}
	category := GetDB(version).FindCategory(parts[1])
	if category == nil {
		return nil, fmt.Errorf("Could not find category: %s", parts[1])
	}
	example := category.FindExample(parts[2])
	if example == nil {
		return nil, fmt.Errorf("Could not find example: %s", parts[2])
	}
	return example, nil
}

// Get the DB for a given SDK Version.
func GetDB(version string) *DB {
	var err error
	if version == "mu" {
		if mu == nil {
			mu, err = loadDir(newExamplesDir)
			if err != nil {
				log.Fatal(err)
			}
		}
		return mu
	}
	if old == nil {
		old, err = loadDir(oldExamplesDir)
		if err != nil {
			log.Fatal(err)
		}
	}
	return old
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
func Save(content []byte) (string, error) {
	if len(content) > 10240 {
		return "", errors.New("Maximum allowed size is 10 kilobytes.")
	}
	h := md5.New()
	_, err := h.Write(content)
	if err != nil {
		return "", fmt.Errorf("Error comupting md5 sum: %s", err)
	}
	hash := fmt.Sprintf("%x", h.Sum(nil))
	err = bucket().Put(hash, content, "text/plain", s3.Private)
	return hash, err
}
