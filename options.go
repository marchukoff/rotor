package rotor

import (
	"path/filepath"
	"strings"
)

type Option func(*Rotor)

func WithFileName(name string) Option {
	return func(rot *Rotor) {
		rot.name = strings.Map(func(r rune) rune {
			switch r {
			case ':', '\\', '/', '<', '>', '?', '*', '|':
				return '_'
			default:
				return r
			}
		}, name)

		if rot.name == "" {
			rot.name = defaultSuffix
		}
	}
}

func WithFilePath(path string) Option {
	return func(r *Rotor) {
		p, err := filepath.Abs(path)
		if err != nil {
			p, _ = filepath.Abs(defaultPath)
			r.path = p
		}
		r.path = p
	}
}

func WithKeepFiles(num uint8) Option {
	return func(r *Rotor) {
		r.total = int(num)
		if r.total == 0 {
			r.total += 1
		}
	}
}
