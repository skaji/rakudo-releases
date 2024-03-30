package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/jszwec/csvutil"
)

const (
	URL       = "https://rakudo.org/dl/rakudo"
	UserAgent = "Mozilla/5.0 (compatible; rakudo-releases; +https://github.com/skaji/rakudo-releases)"
)

type Entry struct {
	SortKey                  string `json:"sort_key" csv:"sort_key"`   //
	Arch                     string `json:"arch" csv:"arch"`           // x86_64 / ""
	Backend                  string `json:"backend" csv:"backend"`     // moar / null
	BuildRevision            int    `json:"build_rev" csv:"build_rev"` // 1 / 2 / null
	Format                   string `json:"format" csv:"format"`       // asc / tar.gz / zip
	Name                     string `json:"name" csv:"name"`           // rakudo
	Platform                 string `json:"platform" csv:"platform"`   // linux / macos / win / src
	Type                     string `json:"type" csv:"type"`           // archive / sig
	URL                      string `json:"url" csv:"url"`             //
	Version                  string `json:"ver" csv:"ver"`             //
	VersionWithBuildRevision string `json:"ver_with_build_rev" csv:"ver_with_build_rev"`
	Padding                  string `json:"padding" csv:"padding"`
}

func (e *Entry) setSortKey() {
	v := e.Version // 2020.08 or 2020.08.1
	if len(v) < len("2020.08.1") {
		v += ".0"
	}
	e.SortKey = strings.Join([]string{
		v + strconv.Itoa(e.BuildRevision),
		e.Platform,
		e.Arch,
	}, "-")
}

func (e *Entry) setVersionWithBuildRevision() {
	if e.Platform == "src" {
		e.VersionWithBuildRevision = fmt.Sprintf("%s", e.Version)
	} else {
		e.VersionWithBuildRevision = fmt.Sprintf("%s-%02d", e.Version, e.BuildRevision)
	}
}

type Entries []*Entry

func (es Entries) Filter() Entries {
	var out Entries
	for _, e := range es {
		if e.Format != "txt" && e.Format != "asc" && e.Platform != "src" && e.Type != "installer" {
			out = append(out, e)
		}
	}
	return out
}

func (es Entries) Sort() Entries {
	out := slices.Clone(es)
	slices.SortFunc(out, func(a, b *Entry) int {
		if a.SortKey != b.SortKey {
			return strings.Compare(b.SortKey, a.SortKey)
		}
		return strings.Compare(b.URL, a.URL)
	})
	return out
}

func httpGet(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Close = true
	req.Header.Set("User-Agent", UserAgent)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, err
	}
	if res.StatusCode/100 != 2 {
		return nil, errors.New(res.Status)
	}
	return body, nil
}

func run(ctx context.Context) error {
	body, err := httpGet(ctx, URL)
	if err != nil {
		return err
	}
	var entries Entries
	if err := json.Unmarshal(body, &entries); err != nil {
		return err
	}
	entries = entries.Filter()
	for _, e := range entries {
		e.setSortKey()
		e.setVersionWithBuildRevision()
	}
	entries = entries.Sort()
	b, err := csvutil.Marshal(entries)
	if err != nil {
		return err
	}
	fmt.Print(string(b))
	return nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
