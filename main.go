package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"unicode"
	"unicode/utf8"
)

func checkErr(e error) {
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}

var runes = make(map[rune]uint)

func updateStat(filename string) {
	buf, err := ioutil.ReadFile(filename)
	checkErr(err)
	for len(buf) > 0 {
		r, n := utf8.DecodeRune(buf)
		runes[unicode.ToUpper(r)]++
		buf = buf[n:]
	}
}

var re *regexp.Regexp

func walk(dirname string) {
	d, err := os.Open(dirname)
	checkErr(err)
	list, err := d.Readdir(-1)
	d.Close()
	for _, fi := range list {
		fname := fi.Name()
		if fname[0] == '.' {
			continue
		}
		path := filepath.Join(dirname, fname)
		switch {
		case fi.IsDir():
			walk(path)
		case fi.Size() > 0 && re.MatchString(fname):
			updateStat(path)
		}
	}
}

type stat struct {
	r rune
	n uint
}

type statSlice []stat

func (p statSlice) Len() int           { return len(p) }
func (p statSlice) Less(i, j int) bool { return p[i].n < p[j].n }
func (p statSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: charstat REGEXP DIR [DIR ...]")
		os.Exit(1)
	}
	regexpStr := os.Args[1]

	for _, dir := range os.Args[2:] {
		fi, err := os.Lstat(dir)
		checkErr(err)
		if !fi.IsDir() {
			fmt.Fprintln(os.Stderr, dir, "should be a directory")
		}
		re, err = regexp.Compile(regexpStr)
		checkErr(err)

		walk(dir)
	}

	var ss statSlice
	var sum uint
	for r, n := range runes {
		sum += n
		ss = append(ss, stat{r, n})
	}
	sort.Sort(ss)
	for _, s := range ss {
		fmt.Printf(
			"%q %.1f %d\n",
			s.r, float64(s.n)*100/float64(sum), s.n,
		)
	}
}
