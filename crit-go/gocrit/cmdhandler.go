package gocrit

import (
	"crit-go/images"
	log "crit-go/logging"
	"encoding/json"
	"fmt"
	"os"
)

func Check(e error) {
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}

func checkfile(e error, f *os.File) {
	if e != nil {
		f.Close()
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}

func Decode(inloc string, outloc string, pretty bool, nopl bool) {
	infile, err := inf(inloc)
	if err != nil {
		log.Fatal("Unable to Open Input File,Exiting ", err)
	}
	img := images.Load(infile, pretty, nopl)
	outfile, err := outf(outloc)
	if err != nil {
		infile.Close()
		log.Fatal("Unable to Open Output File,Exiting ", err)
	}
	if pretty == true {
		encoder := json.NewEncoder(outfile)
		encoder.SetIndent("", "    ")
		err := encoder.Encode(&img)
		if err != nil {
			log.Error("Json Encoder Errored", err)
			outfile.Close()
			os.Exit(1)
		}
		outfile.Close()
	} else {
		encoder := json.NewEncoder(outfile)
		err := encoder.Encode(&img)
		if err != nil {
			log.Error("Json Encoder Errored", err)
			outfile.Close()
			os.Exit(1)
		}
		outfile.Close()
	}
}

func Encode(inloc string, outloc string) {
	infile, err := inf(inloc)
	if err != nil {
		log.Fatal("Unable to Open Input File,Exiting ", err)
	}
	outfile, err := outf(outloc)
	if err != nil {
		infile.Close()
		log.Fatal("Unable to Open Output File,Exiting ", err)
	}
	images.Dump(infile, outfile)
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

func inf(inloc string) (imgfile *os.File, err error) {
	/*
		return a pointer to the file or stdin
	*/
	if inloc == "" {
		imgfile := os.Stdin
		return imgfile, nil
	} else {
		imgfile, err := os.Open(inloc)
		return imgfile, err
	}
}

func outf(outloc string) (outfile *os.File, err error) {
	/*
		return a pointer to the file or stdout
	*/
	if outloc == "" {
		outfile := os.Stdout
		return outfile, nil
	} else {
		outfile, err := os.OpenFile(outloc, os.O_CREATE|os.O_WRONLY, 0644)
		return outfile, err
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
