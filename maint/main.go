package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jszwec/csvutil"
)

const UserAgent = "Mozilla/5.0 (compatible; rakudo-releases; +https://github.com/skaji/rakudo-releases)"

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
	out := make(Entries, len(es))
	copy(out, es)
	sort.Slice(out, func(i, j int) bool {
		if out[j].SortKey == out[i].SortKey {
			return out[j].URL < out[i].URL
		}
		return out[j].SortKey < out[i].SortKey
	})
	return out
}

func run() error {
	req, err := http.NewRequest(http.MethodGet, "https://rakudo.org/dl/rakudo", nil)
	if err != nil {
		return err
	}
	req.Close = true
	req.Header.Set("User-Agent", UserAgent)
	res, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return err
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return err
	}
	if res.StatusCode/100 != 2 {
		return errors.New(res.Status)
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
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
