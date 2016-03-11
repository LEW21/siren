package main

import (
	"errors"
	"strings"
	"unicode/utf8"
)

func ParseLine(line string) ([]string, error) {
	parts := make([]string, 0, strings.Count(line, " "))

	part := make([]rune, 0, len(line))

	i := 0

	for i < len(line) {
		r, d := utf8.DecodeRuneInString(line[i:])
		switch r {
			case '"':
				i += d

				endFound := false
				for !endFound && i < len(line) {
					r, d = utf8.DecodeRuneInString(line[i:])
					switch r {
						case '"':
							endFound = true
							i += d

						case '\\':
							i += d
							r, d = utf8.DecodeRuneInString(line[i:])
							part = append(part, r)
							i += d

						default:
							part = append(part, r)
							i += d
					}
				}
				if !endFound {
					return nil, errors.New("Invalid line.")
				}

			case '\'':
				i += d

				endFound := false
				for !endFound && i < len(line) {
					r, d = utf8.DecodeRuneInString(line[i:])
					switch r {
						case '\'':
							endFound = true
							i += d

						default:
							part = append(part, r)
							i += d
					}
				}
				if !endFound {
					return nil, errors.New("Invalid line.")
				}

			case ' ':
				parts = append(parts, string(part))
				part = make([]rune, 0, len(line)-i)
				i += d

			case '\\':
				i += d
				r, d = utf8.DecodeRuneInString(line[i:])
				part = append(part, r)
				i += d

			default:
				part = append(part, r)
				i += d
		}
	}

	parts = append(parts, string(part))
	return parts, nil
}

func ParseSirenfile(file string) ([][]string, error) {
	lines := strings.Split(file, "\n")

	commands := make([][]string, 0, len(lines))

	for _, line := range lines {
		if line != "" && line[0] != '#' && line[0] != '.' {
			parts, err := ParseLine(line)
			if err != nil {
				return nil, err
			}
			commands = append(commands, parts)
		}
	}

	return commands, nil
}
