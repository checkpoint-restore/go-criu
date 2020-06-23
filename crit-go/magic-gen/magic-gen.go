package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/thoas/go-funk"
	"os"
	"reflect"
	"strconv"
	"strings"
)

func main() {
	// checks argument correctness
	if len(os.Args) != 3 {
		fmt.Println("Usage: magic-gen.go path/to/magic.h path/to/magic.json")
		os.Exit(1)
	}
	magiccheader := os.Args[1]
	magicjson := os.Args[2]
	// opens magic file
	infile, err := os.Open(magiccheader)
	check(err)
	defer infile.Close()
	outfile, err := os.OpenFile(magicjson, os.O_CREATE|os.O_WRONLY, 0644)
	check(err)
	defer outfile.Close()
	enc := json.NewEncoder(outfile)
	mapper := make(map[string]interface{})
	allmagic := make(map[string]string)
	magic := make(map[string]string)
	byname := make(map[string]string)
	byval := make(map[uint64]interface{})
	scanner := bufio.NewScanner(infile)
	scanner.Split(bufio.ScanLines)
	// scans the file line by line
	for scanner.Scan() {
		split := strings.Fields(scanner.Text())
		if len(split) < 3 {
			continue
		}

		if !itemExists(split, "#define") {
			continue
		}

		key := split[1]
		value := split[2]
		if funk.Contains(allmagic, value) {
			value := allmagic[value]
			allmagic[key] = value
		} else {
			magic[key] = value
			allmagic[key] = value
		}
	}

	check(scanner.Err())

	for k, v := range magic {
		if v == "0x0" || v == "1" || k == "0x0" {
			continue
		}
		if strings.Contains(k, "_MAGIC") == true {
			k = strings.Replace(k, "_MAGIC", "", -1)
		}
		/*
			Although the value of v above is a hex string,
			base 0 has to be used instead of base 16
			to convert without errors.As strconv.ParseUint
			outputs a invalid syntax error for strings that
			have 0x at the beginning when base 16 is given
			example -> https://play.golang.org/p/6dWo6oy9vyo
		*/
		vhexedv, err := strconv.ParseUint(v, 0, 64)
		check(err)
		byname[k] = strconv.FormatUint(vhexedv, 10)
		byval[vhexedv] = k
	}
	// byval and byname append to a master map to be converted to json
	mapper["byname"] = byname
	mapper["byval"] = byval

	err = enc.Encode(&mapper)
	check(err)
}

func check(e error) {
	if e != nil {
		fmt.Println(e)
		os.Exit(1)
	}
}

// function to check if item exists in map
func itemExists(slice interface{}, item interface{}) bool {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		panic("Invalid data-type")
	}
	return s.Index(0).Interface() == item
}
