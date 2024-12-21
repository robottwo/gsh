package gline

type Options struct {
	ClearScreen bool
}

func NewOptions() *Options {
	return &Options{
		ClearScreen: false,
	}
}
