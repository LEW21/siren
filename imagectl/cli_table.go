package imagectl

import (
	"io"
	"fmt"
	"strconv"

	"github.com/fatih/color"
)

type Column struct {
	Name string
	Get func(interface{}) (string, color.Attribute)
}

func toYesNo(val bool) string {
	if val {
		return "yes"
	} else {
		return "no"
	}
}

func printTable(w io.Writer, columns []Column, data []interface{}) {
	maxlen := make([]int, len(columns))
	for i, col := range columns {
		maxlen[i] = len(col.Name)
	}
	for _, obj := range data {
		for i, col := range columns {
			v, _ := col.Get(obj)
			if l := len(v); l > maxlen[i] {
				maxlen[i] = l
			}
		}
	}

	formats := make([]string, len(columns))

	for i, l := range maxlen {
		formats[i] = "%-" + strconv.Itoa(l) + "v "
	}

	// Header
	for i, col := range columns {
		fmt.Fprintf(w, formats[i], col.Name)
	}
	fmt.Fprintln(w)

	// Data
	for _, obj := range data {
		for i, col := range columns {
			v, attr := col.Get(obj)
			bold := color.Attribute(0)
			if attr != 0 {
				bold = color.Bold
			}
			fmt.Fprint(w, color.New(attr, bold).SprintfFunc()(formats[i], v))
		}
		fmt.Fprintln(w)
	}
}
