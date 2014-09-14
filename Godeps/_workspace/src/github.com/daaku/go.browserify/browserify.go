// Package browserify provides a wrapper around the CLI.
package browserify

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const Path = "/browserify/"
const defaultBinary = "browserify"

var (
	browserifyPathOverride = flag.String(
		"browserify.path", "", "The path to the browserify command.")
	cacheMaxAge = flag.Duration(
		"browserify.max-age",
		time.Hour*87658,
		"Max age to use in the cache headers.")
	// Internal content cache.
	cache = make(map[string][]byte)
)

// Alias module names.
type Alias map[string]string

// Register plugins and associated configuration (will be JSON encoded).
type Plugin map[string]interface{}

// Define a script just as you would with the browserify CLI.
type Script struct {
	Dir         string // the working directory
	Require     string
	Entry       string
	Exports     string
	Ignore      string
	Alias       Alias
	Debug       bool
	Plugin      Plugin
	OmitPrelude bool
	Override    string
	url         string
}

// Command line arguments for the configured Alias.
func (a Alias) args() ([]string, error) {
	args := make([]string, 0)
	for key, val := range a {
		args = append(args, "--alias", key+":"+val)
	}
	return args, nil
}

// Command line arguments for the configured Plugin.
func (p Plugin) args() ([]string, error) {
	args := make([]string, 0)
	for key, i := range p {
		val, err := json.Marshal(i)
		if err != nil {
			return nil, fmt.Errorf(
				"Failed json.Marshal for argument %v for plugin %s with eror %s.",
				key, val, err)
		}
		args = append(args, "--alias", key+":"+string(val))
	}
	return args, nil
}

// Try harder to look for browserify.
func (s *Script) browserifyPath() (string, error) {
	if *browserifyPathOverride != "" {
		return *browserifyPathOverride, nil
	}

	// prefer the one in `npm bin` if one exists
	npmPath, err := exec.LookPath("npm")
	if err == nil && npmPath != "" {
		cmd := &exec.Cmd{
			Path: npmPath,
			Args: []string{"npm", "bin"},
			Dir:  s.Dir,
		}
		npmBin, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("Failed to run npm bin: %s", err)
		}
		localPath := filepath.Join(strings.TrimSpace(string(npmBin)), defaultBinary)
		st, err := os.Stat(localPath)
		if !os.IsNotExist(err) && st != nil {
			return localPath, nil
		}
	}

	// fallback to the global one
	path, err := exec.LookPath(defaultBinary)
	if err == nil && path != "" {
		return path, nil
	}

	return "", errors.New("Could not find browserify or npm.")
}

// Command line arguments for the browserify command to generate script content.
func (s *Script) args() ([]string, error) {
	args := make([]string, 0)
	if s.Require != "" {
		args = append(args, "--require", s.Require)
	}
	if s.Entry != "" {
		args = append(args, "--entry", s.Entry)
	}
	if s.Exports != "" {
		args = append(args, "--exports", s.Exports)
	}
	if s.Ignore != "" {
		args = append(args, "--ignore", s.Ignore)
	}
	aliasArgs, err := s.Alias.args()
	if err != nil {
		return nil, err
	}
	args = append(args, aliasArgs...)
	pluginArgs, err := s.Plugin.args()
	if err != nil {
		return nil, err
	}
	args = append(args, pluginArgs...)
	if s.OmitPrelude {
		args = append(args, "--prelude", "false")
	}
	if s.Debug {
		args = append(args, "--debug")
	}
	return args, nil
}

// Get the contents of this script.
func (s *Script) Content() ([]byte, error) {
	if s.Override == "" {
		browserify, err := s.browserifyPath()
		if err != nil {
			return nil, err
		}
		args, err := s.args()
		if err != nil {
			return nil, err
		}
		cmd := &exec.Cmd{
			Path: browserify,
			Args: args,
			Dir:  s.Dir,
		}
		out, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf(
				"Failed to execute command %v with error %s and output %s",
				args, err, string(out))
		}
		return out, nil
	}
	return ioutil.ReadFile(s.Override)
}

// Get the URL suffix for easier debuggibility.
func (s *Script) suffix() string {
	if s.Entry != "" {
		return s.Entry
	}
	return s.Require
}

// Get a a content-addressable URL for this script.
func (s *Script) URL() (string, error) {
	if s.url == "" {
		h := md5.New()
		content, err := s.Content()
		if err != nil {
			return "", err
		}
		_, err = h.Write(content)
		if err != nil {
			return "", err
		}
		hex := fmt.Sprintf("%x", h.Sum(nil))
		s.url = path.Join(Path, hex[:10], s.suffix())
		cache[s.url] = content
	}
	return s.url, nil
}

// Serves the static scripts.
func Handle(w http.ResponseWriter, r *http.Request) {
	header := w.Header()
	header.Set("Content-Type", "application/x-javascript; charset=utf-8")
	header.Set("X-Content-Type-Options", "nosniff")
	content, ok := cache[r.URL.Path]
	if !ok {
		header.Set("Cache-Control", "no-cache")
		header.Set("Pragma", "no-cache")
		w.WriteHeader(404)
		w.Write([]byte("alert('browserify resource not found!')"))
		return
	}
	header.Set(
		"Cache-Control",
		fmt.Sprintf("public, max-age=%d", int(cacheMaxAge.Seconds())))
	w.Write(content)
}
