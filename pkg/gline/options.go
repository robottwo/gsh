package gline

type Options struct {
	MinHeight int
}

func NewOptions() Options {
	return Options{
		MinHeight: 8,
	}
}
