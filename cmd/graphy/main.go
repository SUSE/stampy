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
and outputs a new csv to stdout that aggregates the time differences between start/end. The output
format looks like: event,seconds,seconds on horizontal mode. Vertical output transposes this.

Usage:
  graphy [flags] <csvfilename | '-' for stdin>

Example:
  graphy filename.csv

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

	var ioReader io.Reader
	filename := args[0]
	if filename == "-" {
		ioReader = os.Stdin
	} else {
		f, err := os.Open(filename)
		if err != nil {
			return err
		}
		ioReader = f
		defer f.Close()
	}

	events := make(map[string][]uint64)
	starts := make(map[string]time.Time)

	reader := csv.NewReader(ioReader)
	for line := 1; ; line++ {
		row, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("(line %d) failed to read row: %v", line, err)
		}

		date, err := time.Parse("2006-01-02 15:04:05", row[0])
		if err != nil {
			return fmt.Errorf("(line %d) failed to parse time value: %v", line, err)
		}
		series := row[2]
		event := row[3]
		if event == "start" {
			starts[series] = date
		} else if event == "done" {
			startDate, ok := starts[series]
			if !ok {
				return fmt.Errorf("(line %d) found done event with no start event: %s", line, series)
			}
			events[series] = append(events[series], uint64(date.Sub(startDate)/time.Second))
			delete(starts, series)
		} // else invalid event - ignore
	}

	switch *orientation {
	case "vertical":
		return writeVerticalCSV(events)
	default:
		return writeHorizontalCSV(events)
	}
}

func writeVerticalCSV(events map[string][]uint64) error {
	writer := csv.NewWriter(os.Stdout)
	var headers []string
	for series := range events {
		headers = append(headers, series)
	}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write row header: %v", err)
	}

	// Write timings[i] for each event until timings[i] doesn't exist
	// for any event.
	times := make([]string, len(headers))
	for i := 0; ; i++ {
		foundTiming := false
		for j := 0; j < len(headers); j++ {
			timings := events[headers[j]]
			if i < len(timings) {
				times[j] = strconv.Itoa(int(timings[i]))
				foundTiming = true
			} else {
				times[j] = ""
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
	return writer.Error()
}

func writeHorizontalCSV(events map[string][]uint64) error {
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
	return writer.Error()
}
