package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"
)

const path = "/dev/sda"
const bootData = "/home/name/osSec"

func main() {
	log.Printf("[+] Reading boot sector of %s\n", path)

	bootSize, bootSector := readBoot(path)

	log.Printf("[+] Reading prev boot sector from log file.\n")

	lastLog := fmt.Sprintf("%s/%s", bootData, openLastBootLog())
	bootLogSize, bootLogSector := readBoot(lastLog)

	if bootSize != bootLogSize {
		log.Fatal("Error size(boot1) != size(boot2)\n")
	}

	result := compare(bootSector, bootLogSector)
	if len(result) == 0{
		fmt.Printf("Boot sector not modified\n")
	} else {
		fmt.Println("Modified bytes")
		for i, v := range result{
			fmt.Print(i, v)
		}
	}

	writeErr := writeCurrentBootSector(bootSector)
	fatalErr(writeErr)
}

func writeCurrentBootSector(data []byte) error {

	name := fmt.Sprintf("%s/%v.txt",bootData, time.Now())
	file, errCreate := os.Create(name)
	if errCreate != nil{
		fmt.Println("no create")
	}
	size, errWrite := file.Write(data)

	if errWrite != nil || size != 512 {
		fmt.Println("Error in writing current boot sector to logList.")
	}
	return errWrite
}

type symbol struct {
	index uint16
	value [2]byte
}

func compare(prevBoot, curBoot []byte) []symbol {
	var s symbol
	modified := make([]symbol, 0, 16)
	for i := 0; i < 512; i++ {
		if prevBoot[i] != curBoot[i] {
			s.index = uint16(i)
			s.value[0] = prevBoot[i]
			s.value[1] = curBoot[i]
			modified = append(modified, s)
		}
	}
	return modified
}

func openLastBootLog() string{
	dir := bootData
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var modTime time.Time
	var names []string
	for _, fi := range files {
		if fi.Mode().IsRegular() {
			if !fi.ModTime().Before(modTime) {
				if fi.ModTime().After(modTime) {
					modTime = fi.ModTime()
					names = names[:0]
				}
				names = append(names, fi.Name())
			}
		}
	}
	if len(names) > 0 {
		fmt.Println(modTime, names)
	}
	len := len(names)
	if len != 0{
		return names[0]
	}
	log.Fatal("mo elems")
	return names[0]
}

func readBoot(filePath string)(int, []byte){
	boot, err := os.Open(filePath)
	fatalErr(err)

	bootSector := make([]byte, 512, 512)
	readBoot, err := io.ReadFull(boot, bootSector)
	fatalErr(err)

	return readBoot, bootSector
}

func firstStart() (n int, boot []byte) {
	dir, errOpen := os.Open(bootData)
	retErr(errOpen, "")
	stat,errStat := dir.Stat()
	retErr(errStat, "")

	if stat.Size() == 0 {
		n, boot = readBoot(path)

		if n != 512 {
			log.Println("[-] Boot sector not equal 512 bytes.")
		} else {
			err := writeCurrentBootSector(boot)
			retErr(err, "")
		}
	} else {
		return 0, []byte{}
	}
	return
}

func retErr(err error, msg string)  {
	if err != nil{
		log.Println(err, msg)
	}
}

func fatalErr(err error)  {
	if err != nil {
		log.Fatal("Error: " + err.Error())
	}
}