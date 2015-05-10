package search

type Options struct {
}

type Result struct {
	Name      string
	MagnetURL string
	Seeders   uint
	Size      uint64
}

type Searchable interface {
	Search(query string, options Options) ([]*Result, error)
}
