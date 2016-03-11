package main

import (
	"testing"
	"reflect"
)

func TestParseLine(t *testing.T) {
	cases := []struct {
		in string; want []string
	}{
		{"A B C", []string{"A", "B", "C"}},
		{"A 'b c' d", []string{"A", "b c", "d"}},
		{"RUN sed -i 's|#DBPath      = /var/lib/pacman/|DBPath      = /usr/lib/pacman/|' /etc/pacman.conf", []string{"RUN", "sed", "-i", "s|#DBPath      = /var/lib/pacman/|DBPath      = /usr/lib/pacman/|", "/etc/pacman.conf"}},
	}
	for _, c := range cases {
		got, err := ParseLine(c.in)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("ParseLine(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}
