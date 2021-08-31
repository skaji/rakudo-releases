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
	Latest                   int    `json:"latest" csv:"latest"`       // 1 / 0
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
	arch := "_"
	if e.Arch != "" {
		arch = string(e.Arch[0])
	}
	e.SortKey = strings.Join([]string{
		string(e.Platform[0]),
		string(arch),
		string(e.Type[0]),
		v,
		strconv.Itoa(e.BuildRevision),
	}, "")
}

func (e *Entry) setVersionWithBuildRevision() {
	if e.Platform == "src" {
		e.VersionWithBuildRevision = fmt.Sprintf("%s", e.Version)
	} else {
		e.VersionWithBuildRevision = fmt.Sprintf("%s-%02d", e.Version, e.BuildRevision)
	}
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
	var entries []*Entry
	if err := json.Unmarshal(body, &entries); err != nil {
		return err
	}
	for _, e := range entries {
		e.setSortKey()
		e.setVersionWithBuildRevision()
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[j].SortKey < entries[i].SortKey
	})
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
