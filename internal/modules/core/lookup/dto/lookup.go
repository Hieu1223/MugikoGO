package dto

type DictionaryEntry struct {
	Word     string            `json:"word"`
	Readings []DictionaryReading `json:"readings"`
}

type DictionaryReading struct {
	Kana    string `json:"kana"`
	Meaning string `json:"meaning"`
}

type KanjiEntry struct {
	Kanji     string `json:"kanji"`
	Reading   string `json:"reading"`
	Meanings  string `json:"meanings"`
	Strokes   int    `json:"strokes"`
	Radical   string `json:"radical"`
	PenStrokes string `json:"penstrokes"`
	Shape     string `json:"shape"`
	Unicode   string `json:"unicode"`
}

type PrefixSearchResult struct {
	Words []string `json:"words"`
}

type PrefixSearchResponse struct {
	Results []string `json:"results"`
	Limit   int      `json:"limit"`
}

type WordLookupRequest struct {
	Word string `form:"word" binding:"required,min=1,max=100"`
}

type PrefixSearchRequest struct {
	Prefix string `form:"prefix" binding:"required,min=1,max=50"`
	Limit  int    `form:"limit" binding:"omitempty,min=1,max=500"`
}

type KanjiLookupRequest struct {
	Kanji string `form:"kanji" binding:"required,len=1"`
}