package generator

// Generator generates short URL keys.
type Generator interface {
	ShortURLKey() string
}
