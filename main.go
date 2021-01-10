package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"testvm/hdrtree"
)

func main() {
	hdrTree := hdrtree.HdrNode{
		Record: make(map[string]string),
		Next:   make(map[string]hdrtree.HdrNode),
	}

	err := hdrTree.LoadFromFile("./conf.json")
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	ipList := []string{}
	fp, err := os.Open("dev_list.txt")
	if err != nil {
		return
	}
	defer fp.Close()
	bufReader := bufio.NewReader(fp)
	for {
		line, _, err := bufReader.ReadLine() // 按行读
		if err != nil {
			break
		} else {
			lineStr := string(line)
			if strings.HasPrefix(lineStr, "#") {
				continue
			}
			ipList = append(ipList, lineStr)
		}
	}

	for _, ip := range ipList {
		output := make(map[string]string)
		err := hdrTree.GetVals(ip, "", "", output)
		if err != nil {
			fmt.Printf("GetVals failed: %v\n", err)
		}
		fmt.Printf("%-15s: %v\n", ip, output)
	}
}
