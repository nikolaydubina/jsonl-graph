package graph

// MapIDStorage maps string tokens to integers and generates new int ids.
type MapIDStorage struct {
	stringToInt map[string]uint64
	intToString map[uint64]string
}

func NewMapIDStorage() MapIDStorage {
	return MapIDStorage{
		stringToInt: map[string]uint64{},
		intToString: map[uint64]string{},
	}
}

// Add generates new id and sets to storage, returns this id.
func (m MapIDStorage) Add(token string) uint64 {
	id := uint64(len(m.intToString)) + 1
	m.intToString[id] = token
	m.stringToInt[token] = id
	return id
}

// Get returns 0 if not found.
func (m MapIDStorage) Get(id string) uint64 {
	return m.stringToInt[id]
}

// GetToken returns empty string if not found.
func (m MapIDStorage) GetToken(id uint64) string {
	return m.intToString[id]
}
