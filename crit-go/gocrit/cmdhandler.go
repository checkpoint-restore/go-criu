package gocrit

import (
	"crit-go/images"
	"encoding/json"
	"fmt"
	"os"
)

func Check(e error) {
	if e != nil {
		fmt.Println(e)
		os.Exit(1)
	}
}

func checkfile(e error, f *os.File) {
	if e != nil {
		f.Close()
		fmt.Println(e)
		os.Exit(1)
	}
}

func Decode(inloc string, outloc string, pretty bool, nopl bool) {
	img := images.Load(inf(inloc), pretty, nopl)
	ouf := outf(outloc)
	if pretty == true {
		encoder := json.NewEncoder(ouf)
		encoder.SetIndent("", "    ")
		err := encoder.Encode(&img)
		checkfile(err, ouf)
		ouf.Close()
	} else {
		encoder := json.NewEncoder(ouf)
		err := encoder.Encode(&img)
		checkfile(err, ouf)
		ouf.Close()
	}
}

func Encode(inloc string, outloc string) {
	images.Dump(inf(inloc), outf(outloc))
}

func Info(inloc string) {
	fmt.Println("test placeholder - Info called ")
}

func Explore(dir string, what string) {
	switch what {
	case "ps":
		explore_ps(dir)
	case "fds":
		explore_fds(dir)
	case "mss":
		explore_mems(dir)
	case "rss":
		explore_rss(dir)
	}
}

func Show(inloc string) {
	fmt.Println("test placeholder - show called")

}

func inf(inloc string) *os.File {
	/*
		return a pointer to the file or stdin
	*/

	if inloc == "" {
		imgfile := os.Stdin
		return imgfile
	} else {
		imgfile, err := os.Open(inloc)
		if err != nil {
			fmt.Println("Failed to open input file: %s", err)
		}
		return imgfile
	}
}

func outf(outloc string) *os.File {
	/*
		return a pointer to the file or stdout
	*/
	if outloc == "" {
		outfile := os.Stdout
		return outfile
	} else {
		outfile, err := os.OpenFile(outloc, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("Failed to open output file: %s", err)
		}
		return outfile
	}
}

func explore_ps(dir string) {
	fmt.Println("placeholder")
}

func explore_fds(dir string) {
	fmt.Println("placeholder")
}

func explore_mems(dir string) {
	fmt.Println("placeholder")
}

func explore_rss(dir string) {
	fmt.Println("placeholder")
}
