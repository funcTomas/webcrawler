package module

type Type string

const (
	TYPE_DOWNLOADER Type = "downloader"
	TYPE_ANALYZER   Type = "analyzer"
	TYPE_PIPELINE   Type = "pipeline"
)

var legalTypeLetterMap = map[Type]string{
	TYPE_DOWNLOADER: "D",
	TYPE_ANALYZER:   "A",
	TYPE_PIPELINE:   "P",
}

var legalLetterTypeMap = map[string]Type{
	"D": TYPE_DOWNLOADER,
	"A": TYPE_ANALYZER,
	"P": TYPE_PIPELINE,
}

func CheckType(moduleType Type, module Module) bool {
	if moduleType == "" || module == nil {
		return false
	}
	switch moduleType {
	case TYPE_DOWNLOADER:
		if _, ok := module.(Downloader); ok {
			return true
		}
	case TYPE_ANALYZER:
		if _, ok := module.(Analyzer); ok {
			return true
		}
	case TYPE_PIPELINE:
		if _, ok := module.(Pipleline); ok {
			return true
		}
	}
	return false
}

func LegalType(moduleType Type) bool {
	if _, ok := legalTypeLetterMap[moduleType]; ok {
		return true
	}
	return false
}
func GetType(mid MID) (bool, Type) {
	parts, err := SplitMID(mid)
	if err != nil {
		return false, ""
	}
	mt, ok := legalLetterTypeMap[parts[0]]
	return ok, mt
}

func getLetter(moduleType Type) (bool, string) {
	var letter string
	var found bool
	for l, t := range legalLetterTypeMap {
		if t == moduleType {
			letter = l
			found = true
			break
		}
	}
	return found, letter
}

func typeToLetter(moduleType Type) (bool, string) {
	switch moduleType {
	case TYPE_ANALYZER:
		return true, "A"
	case TYPE_DOWNLOADER:
		return true, "D"
	case TYPE_PIPELINE:
		return true, "P"
	}
	return false, ""
}

func letterToType(letter string) (bool, Type) {
	switch letter {
	case "D":
		return true, TYPE_DOWNLOADER
	case "A":
		return true, TYPE_ANALYZER
	case "P":
		return true, TYPE_DOWNLOADER
	default:
		return false, ""
	}
}
