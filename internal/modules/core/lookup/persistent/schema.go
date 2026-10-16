package persistent

type DictionaryEntry struct {
	Word     string
	Readings []DictionaryReading
}

type DictionaryReading struct {
	Kana    string
	Meaning string
}

type KanjiEntry struct {
	Kanji      string
	Reading    string
	Meanings   string
	Strokes    int
	Radical    string
	Penstrokes string
	Shape      string
	Unicode    string
}

type BinaryDictionaryProvider struct {
	data       []byte
	trieStart  uint32
	defStart   uint32
	entryCount uint32
}

const (
	MUGIMagic   = "MUGI"
	MUGIVersion = 1
)
