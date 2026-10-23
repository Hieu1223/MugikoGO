package exception

const (
	MsgWordParameterRequired       = "Word parameter is required"
	MsgWordNotFound                = "Word not found"
	MsgPrefixParameterRequired     = "Prefix parameter is required"
	MsgKanjiParameterRequired      = "Kanji parameter is required"
	MsgKanjiMustBeSingleCharacter  = "Kanji must be a single character"
	MsgKanjiNotFound               = "Kanji not found"
	MsgDictionaryFileNotFound      = "dictionary file not found: %s"
	MsgKanjiDictionaryFileNotFound = "kanji dictionary file not found: %s"
	MsgFileTooSmallForHeader       = "file too small for header"
	MsgInvalidMagic                = "invalid magic: %s, expected %s"
	MsgUnsupportedVersion          = "unsupported version: %d"
)
