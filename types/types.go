package types

type CharSet map[string]float64

type CharClass struct {
	List    []string
	Weights CharSet
}

func (c CharClass) Contains(k string) bool {
	_, ok := c.Weights[k]
	return ok
}
