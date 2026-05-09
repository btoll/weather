package main

import (
	"bufio"
	"flag"
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

var (
	defaultLocations = locations{
		{"Valley Springs", "California"},
		{"Phoenix", "Arizona"},
		{"Princeton", "Massachusetts"},
		{"Placitas", "New Mexico"},
		{"Mechanicsburg", "Pennsylvania"},
	}
	numWorkers int
)

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

func pipe1(l locations) <-chan *GeoResults {
	out := make(chan *GeoResults)
	go func() {
		defer close(out)
		for _, location := range l {
			city, state := location[0], location[1]
			out <- GetLocation(city, state)
		}
	}()
	return out
}

func pipe2(in <-chan *GeoResults) <-chan *Forecast {
	out := make(chan *Forecast)
	go func() {
		defer close(out)
		var wg sync.WaitGroup
		for range numWorkers {
			wg.Go(func() {
				for location := range in {
					if location.Err != nil {
						log.Println(location.Err)
					}
					var g GeoResult
					for _, loc := range location.Results {
						if loc.State == location.State {
							g = *loc
							break
						}
					}
					if !g.IsZero() {
						out <- GetForecast(g)
					}
				}
			})
		}
		wg.Wait()
	}()
	return out
}

func pipeline(l locations) <-chan *Forecast {
	return pipe2(pipe1(l))
}

func main() {
	flag.IntVar(&numWorkers, "workers", 3, "The number of concurrent workers.")
	flag.Parse()

	l := locations{}
	if len(os.Args) > 1 {
		reader, err := getFile(flag.Args()[0])
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
	for forecast := range pipeline(l) {
		if forecast.Err != nil {
			log.Println(forecast.Err)
		}
		fmt.Printf("\n%+v\n", forecast)
	}
}
