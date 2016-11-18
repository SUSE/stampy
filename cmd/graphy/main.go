package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"
)

var usage = `Consumes a csv file in the format: YYYY-MM-DD HH:MM:SS,file,timeseries,(start|done)
and outputs a new csv to stdout that aggregates the time differences between start/end

Usage:
  graphy [flags] <csvfilename>

Flags:
      --orientation string   Display orientation: horizontal or vertical (default "horizontal")`

var (
	orientation = flag.String("orientation", "horizontal", "Control orientation of the data")
)

func main() {
	for _, arg := range os.Args {
		switch arg {
		case "-h", "--help", "-help":
			help()
			return
		}
	}

	flag.Parse()

	if err := graphyCmd(flag.Args()); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func help() {
	fmt.Fprintln(os.Stderr, usage)
}

func graphyCmd(args []string) error {
	if len(args) < 1 {
		help()
		return errors.New("need more arguments")
	}

	filename := args[0]
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	events := make(map[string][]uint64)
	starts := make(map[string]time.Time)

	reader := csv.NewReader(f)
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("failed to read row: %v", err)
		}

		date, err := time.Parse("2006-01-02 15:04:05", row[0])
		if err != nil {
			return fmt.Errorf("failed to parse time value: %v", err)
		}
		series := row[2]
		event := row[3]
		if event == "start" {
			starts[series] = date
		} else if event == "done" {
			startDate, ok := starts[series]
			if !ok {
				return fmt.Errorf("found done event with no start event: %s", series)
			}
			events[series] = append(events[series], uint64(date.Sub(startDate)))
			delete(starts, series)
		}
	}

	switch *orientation {
	case "vertical":
		return writeHeadersTop(events)
	default:
		return writeHeadersLeft(events)
	}
}

func writeHeadersTop(events map[string][]uint64) error {
	writer := csv.NewWriter(os.Stdout)
	var headers []string
	for series := range events {
		headers = append(headers, series)
	}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write row header: %v", err)
	}

	times := make([]string, len(headers))
	for i := 0; ; i++ {
		foundTiming := false
		for j := 0; j < len(headers); j++ {
			timings := events[headers[j]]
			if i < len(timings) {
				times[j] = strconv.Itoa(int(timings[i]))
				foundTiming = true
			} else {
				times[j] = "0"
			}
		}

		if !foundTiming {
			break
		}

		if err := writer.Write(times); err != nil {
			return fmt.Errorf("failed to write row: %v", err)
		}
	}
	writer.Flush()

	return nil
}

func writeHeadersLeft(events map[string][]uint64) error {
	writer := csv.NewWriter(os.Stdout)
	var row []string
	for ev, timings := range events {
		row = append(row, ev)
		for _, t := range timings {
			row = append(row, strconv.Itoa(int(t)))
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %v", err)
		}
		row = row[:0]
	}
	writer.Flush()

	return nil
}
