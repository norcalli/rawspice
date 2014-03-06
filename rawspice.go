// Package rawspice implements the structs and types required to
// interpret a raw spice file output and a method for reading it.
package rawspice

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

type Vector []float64

type PlotType int

const (
	plotUndefined PlotType = iota
)

type VectorType int

const (
	tUndefined VectorType = iota,
		tCurrent,
		tVoltage,
		tTime
)

// SpiceVector is the main type for containing the data.
type SpiceVector struct {
	Name string
	Type string
	Data Vector
}

// NewVector creates a new spice vector.
func NewVector(Name, Type string) *SpiceVector {
	vector := new(SpiceVector)
	vector.Name = Name
	vector.Type = Type
	return vector
}

// Simple method to get the ith element of the data.
func (s *SpiceVector) Get(i int64) float64 {
	return s.Data[i]
}

func (s *SpiceVector) String() string {
	// return fmt.Sprintf("%s of type %s = %v", s.Name, s.Type, s.Data)
	return fmt.Sprintf("%s of type %s = %v...", s.Name, s.Type, s.Data[:minInt(10, len(s.Data))])
}

// SpicePlot is the main data structure for containing the output of
// simulation data.
type SpicePlot struct {
	Title string
	Date  time.Time
	Name  string
	// Type       PlotType
	// Dimensions []string
	// ScaleVector *SpiceVector

	DataVectors []*SpiceVector
	// TimeVector  *SpiceVector

	Real   bool
	Padded bool

	NVariables int64
	NPoints    int64
}

func minInt(a, b int) int {
	if a > b {
		return b
	}
	return a
}

// ReadFile decodes the plots from the raw spice file given a filename.
// - A plot is any simulation command output, .dc, .tf, etc...
// - Currently doesn't support complex numbers.
// - Interpretation is left to the user, as a .dc command will report
// vectors with NPoints = 1.
func ReadFile(filename string) ([]*SpicePlot, error) {
	// file, err := os.OpenFile(filename, os.O_RDONLY, 0666)
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
		// log.Fatal(err)
	}
	reader := bufio.NewReader(file)
	var plots []*SpicePlot
	// plots := make([]*SpicePlot, 1)
	plot := &SpicePlot{}
	var line string
	// for scanner.Scan() {
	for {
		line, err = reader.ReadString('\n')
		switch err {
		case io.EOF:
			return plots, nil
		case nil:
		default:
			return nil, err
		}
		line = strings.TrimSpace(line)
		split := strings.SplitN(line, ":", 2)
		for i, x := range split {
			split[i] = strings.TrimSpace(x)
		}
		switch strings.ToLower(split[0]) {
		case "title":
			plot.Title = split[1]
		case "date":
			plot.Date, err = time.Parse("Mon Jan 2 15:04:05 2006", split[1])
			if err != nil {
				return nil, err
			}
		case "plotname":
			plot.Name = split[1]
		case "flags":
			for _, flag := range strings.Fields(split[1]) {
				flag = strings.ToLower(strings.TrimSpace(flag))
				switch flag {
				case "real":
					plot.Real = true
				case "complex":
					plot.Real = false
				case "padded":
					plot.Padded = true
				case "unpadded":
					plot.Padded = false
				}
			}
		case "no. variables":
			plot.NVariables, err = strconv.ParseInt(split[1], 10, 32)
			if err != nil {
				return nil, err
			}
		case "no. points":
			plot.NPoints, err = strconv.ParseInt(split[1], 10, 32)
			if err != nil {
				return nil, err
			}
		case "variables":
			for i := int64(0); i < plot.NVariables; i++ {
				line, err = reader.ReadString('\n')
				line = strings.TrimSpace(line)
				// scanner.Scan()
				fields := strings.Fields(strings.TrimSpace(line))
				plot.DataVectors = append(plot.DataVectors, NewVector(fields[1], fields[2]))
			}
		case "binary":
			// line, err = reader.ReadString('\n')
			// line = strings.TrimSpace(line)

			// scanner.Scan()
			// bytes_buffer := scanner.Bytes()

			data_size := int64(plot.NVariables * plot.NPoints)
			buffer_size := data_size
			if !plot.Real {
				buffer_size *= 2
			}
			// bytes_buffer := make([]byte, buffer_size*8)

			// io.ReadFull(reader, bytes_buffer)

			// data_buffer := make([]float64, buffer_size)
			// // binary.Read(io.LimitReader(file, buffer_size*8), binary.LittleEndian, &data_buffer)
			// // binary.Read(io.LimitReader(file, buffer_size*8), binary.LittleEndian, &data_buffer)
			// // binary.Read(file, binary.LittleEndian, &buffer)
			// // file.Read(bytes_buffer)
			// binary.Read(bytes.NewBuffer(bytes_buffer), binary.LittleEndian, &data_buffer)
			// fmt.Printf("Plot: %+v\n", plot)
			// fmt.Println(len(bytes_buffer), bytes_buffer)
			// fmt.Println(len(data_buffer), data_buffer)
			// // _ = io.LimitReader
			// plots = append(plots, plot)
			// plot = &SpicePlot{}
			// // buffer := make(Vector, plot.NVariables*plot.NPoints)
			// // binary.Read(bytes.NewBuffer(scanner.Bytes()), binary.LittleEndian, &buffer)

			if plot.Real {
				for _, v := range plot.DataVectors {
					v.Data = make(Vector, plot.NPoints)
				}
				for j := int64(0); j < plot.NPoints; j++ {
					for i := int64(0); i < plot.NVariables; i++ {
						// offset := j*plot.NVariables + i
						binary.Read(reader, binary.LittleEndian, &plot.DataVectors[i].Data[j])
					}
				}
			}
			// fmt.Printf("Plot: %+v\n", plot)
			plots = append(plots, plot)
			plot = &SpicePlot{}
		}
		// fmt.Printf("[%s] = %s\n", , text[i+2:])
	}
	return plots, nil
}
