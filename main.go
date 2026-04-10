package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

// https://open-meteo.com/en/docs

type locations [][2]string

var defaultLocations = locations{
	{"Valley Springs", "California"},
	{"Phoenix", "Arizona"},
	{"Princeton", "Massachusetts"},
	{"Placitas", "New Mexico"},
	{"Mechanicsburg", "Pennsylvania"},
}

func getFile(arg string) (io.ReadCloser, error) {
	after, found := strings.CutPrefix(arg, "/dev/fd/")
	if found {
		if fd, err := strconv.Atoi(after); err == nil {
			file := os.NewFile(uintptr(fd), "")
			if _, err = file.Stat(); err == nil {
				return file, nil
			}
		}
	}
	if arg == "-" {
		return io.NopCloser(os.Stdin), nil
	}
	return os.Open(arg)
}

func main() {
	l := locations{}
	if len(os.Args) > 1 {
		reader, err := getFile(os.Args[len(os.Args)-1])
		if err != nil {
			log.Fatal(err)
		}
		defer reader.Close()
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			a := strings.Split(scanner.Text(), ",")
			l = append(l, [2]string{
				strings.Trim(a[0], " "),
				strings.Trim(a[1], " "),
			})
		}
	} else {
		l = defaultLocations
	}

	var wg sync.WaitGroup
	for _, location := range l {
		city, state := location[0], location[1]
		wg.Go(func() {
			gresults, err := GetLocation(city)
			if err != nil {
				log.Fatal(err)
			}

			var g GeoResult
			for _, location := range gresults.Results {
				if location.State == state {
					g = *location
					break
				}
			}

			forecast, err := GetForecast(g)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("\n%s, %s\n%+v\n", city, state, forecast)
		})
	}
	wg.Wait()
}
