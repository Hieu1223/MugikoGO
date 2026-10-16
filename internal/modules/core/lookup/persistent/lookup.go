package persistent

import (
	"encoding/binary"
	"fmt"
	"os"
)

func NewBinaryDictionaryProvider(dictPath string) (*BinaryDictionaryProvider, error) {
	data, err := os.ReadFile(dictPath)
	if err != nil {
		return nil, fmt.Errorf("dictionary file not found: %s", dictPath)
	}

	if len(data) < 24 {
		return nil, fmt.Errorf("file too small for header")
	}

	magic := string(data[0:4])
	if magic != MUGIMagic {
		return nil, fmt.Errorf("invalid magic: %s, expected %s", magic, MUGIMagic)
	}

	version := binary.LittleEndian.Uint32(data[4:8])
	if version != MUGIVersion {
		return nil, fmt.Errorf("unsupported version: %d", version)
	}

	provider := &BinaryDictionaryProvider{
		data:       data,
		trieStart:  binary.LittleEndian.Uint32(data[8:12]),
		defStart:   binary.LittleEndian.Uint32(data[12:16]),
		entryCount: binary.LittleEndian.Uint32(data[16:20]),
	}

	return provider, nil
}

func (p *BinaryDictionaryProvider) Close() error {
	p.data = nil
	return nil
}

func (p *BinaryDictionaryProvider) readU32(offset uint32) uint32 {
	return binary.LittleEndian.Uint32(p.data[offset : offset+4])
}

func (p *BinaryDictionaryProvider) readDefinition(offset uint32) []DictionaryReading {
	end := offset
	for end < uint32(len(p.data)) && p.data[end] != 0 {
		end++
	}
	defStr := string(p.data[offset:end])
	parts := splitByPipe(defStr)

	var readings []DictionaryReading
	for i := 0; i < len(parts)-1; i += 2 {
		readings = append(readings, DictionaryReading{
			Kana:    parts[i],
			Meaning: parts[i+1],
		})
	}
	return readings
}

func splitByPipe(s string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '|' {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}

func (p *BinaryDictionaryProvider) Lookup(word string) *DictionaryEntry {
	nodeOffset := p.trieStart

	for _, char := range word {
		tableSize := p.readU32(nodeOffset)
		_ = p.readU32(nodeOffset + 4) // defAddr - unused in loop
		charCode := uint32(char)
		childBase := nodeOffset + 8
		hashVal := charCode % tableSize

		found := false
		for range tableSize {
			childOffset := childBase + hashVal*8
			storedChar := p.readU32(childOffset)
			addr := p.readU32(childOffset + 4)

			if storedChar == 0 && addr == 0 {
				return nil
			}

			if storedChar == charCode {
				nodeOffset = addr
				found = true
				break
			}
			hashVal = (hashVal + 1) % tableSize
		}

		if !found {
			return nil
		}
	}

	defAddr := p.readU32(nodeOffset + 4)
	if defAddr != 0 {
		readings := p.readDefinition(defAddr)
		return &DictionaryEntry{
			Word:     word,
			Readings: readings,
		}
	}
	return nil
}

func (p *BinaryDictionaryProvider) PrefixSearch(prefix string, limit int) []DictionaryEntry {
	nodeOffset := p.trieStart

	for _, char := range prefix {
		tableSize := p.readU32(nodeOffset)
		childBase := nodeOffset + 8
		charCode := uint32(char)
		hashVal := charCode % tableSize

		found := false
		for range tableSize {
			childOffset := childBase + hashVal*8
			storedChar := p.readU32(childOffset)
			addr := p.readU32(childOffset + 4)

			if storedChar == 0 && addr == 0 {
				return nil
			}

			if storedChar == charCode {
				nodeOffset = addr
				found = true
				break
			}
			hashVal = (hashVal + 1) % tableSize
		}

		if !found {
			return nil
		}
	}

	var results []DictionaryEntry
	p.dfsCollect(nodeOffset, prefix, &results, limit)
	return results
}

func (p *BinaryDictionaryProvider) dfsCollect(nodeOffset uint32, currentWord string, results *[]DictionaryEntry, limit int) {
	if len(*results) >= limit {
		return
	}

	tableSize := p.readU32(nodeOffset)
	defAddr := p.readU32(nodeOffset + 4)

	if defAddr != 0 {
		readings := p.readDefinition(defAddr)
		*results = append(*results, DictionaryEntry{
			Word:     currentWord,
			Readings: readings,
		})
		if len(*results) >= limit {
			return
		}
	}

	childBase := nodeOffset + 8
	for i := uint32(0); i < tableSize; i++ {
		childOffset := childBase + i*8
		storedChar := p.readU32(childOffset)
		addr := p.readU32(childOffset + 4)

		if storedChar != 0 && addr != 0 {
			char := rune(storedChar)
			p.dfsCollect(addr, currentWord+string(char), results, limit)
			if len(*results) >= limit {
				return
			}
		}
	}
}

type KanjiDictionaryProvider struct {
	data         []byte
	tableSize    uint32
	entryCount   uint32
	entriesStart uint32
}

const (
	KANJMagic   = "KANJ"
	KANJVersion = 1
)

func NewKanjiDictionaryProvider(dictPath string) (*KanjiDictionaryProvider, error) {
	data, err := os.ReadFile(dictPath)
	if err != nil {
		return nil, fmt.Errorf("kanji dictionary file not found: %s", dictPath)
	}

	if len(data) < 16 {
		return nil, fmt.Errorf("file too small for header")
	}

	magic := string(data[0:4])
	if magic != KANJMagic {
		return nil, fmt.Errorf("invalid magic: %s, expected %s", magic, KANJMagic)
	}

	version := binary.LittleEndian.Uint32(data[4:8])
	if version != KANJVersion {
		return nil, fmt.Errorf("unsupported version: %d", version)
	}

	provider := &KanjiDictionaryProvider{
		data:         data,
		tableSize:    binary.LittleEndian.Uint32(data[8:12]),
		entryCount:   binary.LittleEndian.Uint32(data[12:16]),
		entriesStart: 16,
	}

	return provider, nil
}

func (p *KanjiDictionaryProvider) Close() error {
	p.data = nil
	return nil
}

func (p *KanjiDictionaryProvider) readU32(offset uint32) uint32 {
	return binary.LittleEndian.Uint32(p.data[offset : offset+4])
}

func (p *KanjiDictionaryProvider) readDefinition(offset uint32) *KanjiEntry {
	end := offset
	for end < uint32(len(p.data)) && p.data[end] != 0 {
		end++
	}
	defStr := string(p.data[offset:end])
	parts := splitByPipe(defStr)

	if len(parts) < 8 {
		return nil
	}

	strokes := 0
	if len(parts[3]) > 0 && parts[3][0] >= '0' && parts[3][0] <= '9' {
		fmt.Sscanf(parts[3], "%d", &strokes)
	}

	return &KanjiEntry{
		Kanji:      parts[0],
		Reading:    parts[1],
		Meanings:   parts[2],
		Strokes:    strokes,
		Radical:    parts[4],
		Penstrokes: parts[5],
		Shape:      parts[6],
		Unicode:    parts[7],
	}
}

func (p *KanjiDictionaryProvider) Lookup(kanji string) *KanjiEntry {
	runes := []rune(kanji)
	if len(runes) != 1 {
		return nil
	}

	charCode := uint32(runes[0])
	hashVal := charCode % p.tableSize

	for range p.tableSize {
		entryOffset := p.entriesStart + hashVal*8
		if entryOffset+8 > uint32(len(p.data)) {
			return nil
		}

		storedChar := p.readU32(entryOffset)
		addr := p.readU32(entryOffset + 4)

		if storedChar == 0 && addr == 0 {
			return nil
		}

		if storedChar == charCode {
			return p.readDefinition(addr)
		}

		hashVal = (hashVal + 1) % p.tableSize
	}

	return nil
}

type LookupRepository struct {
	dictProvider  *BinaryDictionaryProvider
	kanjiProvider *KanjiDictionaryProvider
	dictPath      string
	kanjiDictPath string
}

func NewLookupRepository(dictPath, kanjiDictPath string) (*LookupRepository, error) {
	dictProvider, err := NewBinaryDictionaryProvider(dictPath)
	if err != nil {
		return nil, err
	}

	kanjiProvider, err := NewKanjiDictionaryProvider(kanjiDictPath)
	if err != nil {
		dictProvider.Close()
		return nil, err
	}

	return &LookupRepository{
		dictProvider:  dictProvider,
		kanjiProvider: kanjiProvider,
		dictPath:      dictPath,
		kanjiDictPath: kanjiDictPath,
	}, nil
}

func (r *LookupRepository) Close() error {
	var err error
	if r.dictProvider != nil {
		err = r.dictProvider.Close()
	}
	if r.kanjiProvider != nil {
		if e := r.kanjiProvider.Close(); e != nil && err == nil {
			err = e
		}
	}
	return err
}

func (r *LookupRepository) LookupWord(word string) *DictionaryEntry {
	return r.dictProvider.Lookup(word)
}

func (r *LookupRepository) PrefixSearch(prefix string, limit int) []DictionaryEntry {
	return r.dictProvider.PrefixSearch(prefix, limit)
}

func (r *LookupRepository) LookupKanji(kanji string) *KanjiEntry {
	return r.kanjiProvider.Lookup(kanji)
}
