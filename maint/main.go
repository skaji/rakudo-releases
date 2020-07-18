package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/jszwec/csvutil"
)

type Entry struct {
	Sort     string `json:"sort" csv:"sort"`           //
	Arch     string `json:"arch" csv:"arch"`           // x86_64 / ""
	Backend  string `json:"backend" csv:"backend"`     // moar / null
	BuildRev int    `json:"build_rev" csv:"build_rev"` // 1 / 2 / null
	Format   string `json:"format" csv:"format"`       // asc / tar.gz / zip
	Latest   int    `json:"latest" csv:"latest"`       // 1 / 0
	Name     string `json:"name" csv:"name"`           // rakudo
	Platform string `json:"platform" csv:"platform"`   // linux / macos / win / src
	Type     string `json:"type" csv:"type"`           // archive / sig
	URL      string `json:"url" csv:"url"`             //
	Version  string `json:"ver" csv:"ver"`             //
	Key      string `json:"key" csv:"key"`
}

type Entries []*Entry

func (es Entries) Len() int {
	return len(es)
}

func (es Entries) Less(i, j int) bool {
	if es[i].Platform != es[j].Platform {
		return es[i].Platform > es[j].Platform
	}
	if es[i].Type != es[j].Type {
		return es[i].Type > es[j].Type
	}
	if es[i].Version != es[j].Version {
		return es[i].Version > es[j].Version
	}
	if es[i].BuildRev != es[j].BuildRev {
		return es[i].BuildRev > es[j].BuildRev
	}
	return true
}

func (es Entries) Swap(i, j int) {
	es[i], es[j] = es[j], es[i]
}

func run(typ string) error {
	req, err := http.NewRequest(http.MethodGet, "https://rakudo.org/dl/rakudo", nil)
	if err != nil {
		return err
	}
	req.Close = true
	req.Header.Set("User-Agent", "https://github.com/skaji/rakudo-releases")
	res, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return errors.New(res.Status)
	}
	var allEntries Entries
	if err := json.Unmarshal(body, &allEntries); err != nil {
		return err
	}
	sort.Stable(allEntries)
	var entries Entries
	if typ == "prebuilt" {
		sort := 1
		for _, e := range allEntries {
			if e.Platform != "src" {
				e2 := new(Entry)
				*e2 = *e
				e2.Sort = fmt.Sprintf("%03d", sort)
				e2.Key = fmt.Sprintf("rakudo-%s-%02d", e.Version, e.BuildRev)
				entries = append(entries, e2)
				sort++
			}
		}
	} else {
		sort := 1
		for _, e := range allEntries {
			if e.Platform == "src" {
				e2 := new(Entry)
				*e2 = *e
				e2.Sort = fmt.Sprintf("%03d", sort)
				e2.Key = fmt.Sprintf("rakudo-%s", e.Version)
				entries = append(entries, e2)
				sort++
			}
		}
	}
	b, err := csvutil.Marshal(entries)
	if err != nil {
		return err
	}
	fmt.Print(string(b))
	return nil
}

func main() {
	typ := "prebuilt"
	if len(os.Args) > 1 {
		typ = os.Args[1]
	}
	if err := run(typ); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
