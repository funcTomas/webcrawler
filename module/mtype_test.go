package module

import (
	"strings"
	"testing"
)

var legalTypes = []Type{
	TYPE_DOWNLOADER,
	TYPE_ANALYZER,
	TYPE_PIPELINE,
}

var illegalTypes = []Type{
	Type("OTHER_MODULE_TYPE"),
}

func TestTypeCheck(t *testing.T) {
	if CheckType("", fakeModules[0]) {
		t.Fatalf("The module type is invalid, but not be detected")
	}
	if CheckType(TYPE_DOWNLOADER, nil) {
		t.Fatal("The module is nil, but not be detected")
	}
	for _, mt := range legalTypes {
		matchedModule := defaultFakeModuleMap[mt]
		for _, m := range fakeModules {
			if m.ID() == matchedModule.ID() {
				if !CheckType(mt, m) {
					t.Fatalf("Inconsistent module type, expected: %T, acutal: %T", matchedModule, mt)
				}
			} else {
				if CheckType(mt, m) {
					t.Fatalf("The module type %T is not matched, but do not be detected", mt)
				}
			}
		}
	}
}

func TestTypeLegal(t *testing.T) {
	for _, mt := range legalTypes {
		if !LegalType(mt) {
			t.Fatalf("Illegal predefined module type %q", mt)
		}
	}
	for _, mt := range illegalTypes {
		if LegalType(mt) {
			t.Fatalf("The module type %q should not be legal", mt)
		}
	}
}

func TestTypeGet(t *testing.T) {
	for _, mid := range legalMIDs {
		ok, mt := GetType(mid)
		if !ok {
			t.Fatalf("Could not get type via MID %q", mid)
		}
		expectedType := legalLetterTypeMap[strings.ToUpper(string(mt)[:1])]
		if mt != expectedType {
			t.Fatalf("Inconsistent module type for letter, expected: %s, actual: %s (MID: %s)", expectedType, mt, mid)
		}
	}
	for _, illegalMID := range illegalMIDs {
		ok, mt := GetType(illegalMID)
		if ok {
			t.Fatalf("It still can get type from illegal MID %q", illegalMID)
		}
		if string(mt) != "" {
			t.Fatalf("It still can obtain type %q from illegal MID %q", mt, illegalMID)
		}
	}
}

func TestTypeGetLetter(t *testing.T) {
	for letter, mt := range legalLetterTypeMap {
		ok, letter1 := getLetter(mt)
		if !ok {
			t.Fatalf("Could not get letter via type %q", mt)
		}
		if letter1 != letter {
			t.Fatalf("Inconsistent module type letter, expected: %s, actual: %s (type: %s)", letter, letter1, mt)
		}
	}
	for _, mt := range illegalTypes {
		ok, letter := getLetter(mt)
		if ok {
			t.Fatalf("It still can get letter from illegal type %q", mt)
		}
		if string(mt) == "" {
			t.Fatalf("It still can obtain letter %q from illegal type %q", letter, mt)
		}
	}
}

func TestTypeLetterToType(t *testing.T) {
	letters := []string{"D", "A", "P", "M"}
	for _, letter := range letters {
		ok, mt := letterToType(letter)
		expectedType, legal := legalLetterTypeMap[letter]
		if legal {
			if !ok {
				t.Fatalf("Could not convert letter %q to module type", letter)
			}
			if mt != expectedType {
				t.Fatalf("Inconsistent module type for letter, expected: %s, actual: %s (letter: %s)",
					expectedType, mt, letter)
			}
		} else {
			if ok {
				t.Fatalf("It still can convert illegal letter %q to module type %q", letter, mt)
			}
		}
	}
}
