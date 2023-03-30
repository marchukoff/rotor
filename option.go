package rotor

// Option provide some settings.
type Option interface {
	apply(*options)
}

type options struct {
	delimiter string
	extension string
	postfix   string
	makeDir   bool
}

type delimiter string

func (o delimiter) apply(opts *options) {
	opts.delimiter = string(o)
}

// Delimiter represents separator between name and time for filename.
// Default: "-"
func Delimiter(s string) Option {
	return delimiter(s)
}

type extension string

func (o extension) apply(opts *options) {
	opts.extension = string(o)
}

// Extension represents file name extension.
// Default: "log"
func Extension(ext string) Option {
	return extension(ext)
}

type postfix string

func (o postfix) apply(opts *options) {
	opts.postfix = string(o)
}

// PostfixTime represents format string (layout) for filename.
// Default: "2006-01-02"
func PostfixTime(layout string) Option {
	return postfix(layout)
}

type makeDir bool

func (o makeDir) apply(opts *options) {
	opts.makeDir = bool(o)
}

// MakeDir represents flag to create directory if not exists.
// Default: true
func MakeDir(ok bool) Option {
	return makeDir(ok)
}
