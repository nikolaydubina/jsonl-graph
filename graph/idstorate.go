package graph

// MapIDStorage maps string ids to integers and generates new int ids.
type MapIDStorage struct {
	StringToInt map[string]uint64
	IntToString map[uint64]string
}

func NewMapIDStorage() MapIDStorage {
	return MapIDStorage{
		StringToInt: map[string]uint64{},
		IntToString: map[uint64]string{},
	}
}

// Add generates new id and sets to storage, returns this id.
func (m MapIDStorage) Add(id string) uint64 {
	intid := uint64(len(m.IntToString)) + 1
	m.IntToString[intid] = id
	m.StringToInt[id] = intid
	return intid
}

// Get returns 0 if not found.
func (m MapIDStorage) Get(id string) uint64 {
	return m.StringToInt[id]
}

// GetToken returns empty string if not found.
func (m MapIDStorage) GetToken(id uint64) string {
	return m.IntToString[id]
}
