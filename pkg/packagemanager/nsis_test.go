package packagemanager

import (
	"strings"
	"testing"
)

func TestNSISListingValidation(t *testing.T) {
	good := "Path = tessdata\\eng.traineddata\nSize = 100\nAttributes = \n\nPath = tesseract.exe\nSize = 200\n"
	rows, e := nsisEntries(good, t.TempDir())
	if e != nil || rows["tessdata/eng.traineddata"] != 100 {
		t.Fatalf("%v %v", rows, e)
	}
	for _, bad := range []string{strings.Replace(good, "tesseract.exe", "../escape", 1), good + "\nPath = TESSERACT.EXE\nSize = 2\n", "Path = x\nSize = 1073741825\n", "Path = x\nSize = 1\nSymbolic Link = elsewhere\n", "Path = x\nSize = -1\n"} {
		if _, e := nsisEntries(bad, t.TempDir()); e == nil {
			t.Fatalf("accepted %q", bad)
		}
	}
}
func TestSupplementaryArtifactValidation(t *testing.T) {
	r, e := BuiltinRegistry()
	if e != nil {
		t.Fatal(e)
	}
	pkg, ok := r.Find("tesseract")
	if !ok {
		t.Fatal("missing OCR package")
	}
	a := pkg.Artifacts["win32-x64"]
	if e := validateArtifact(a); e != nil {
		t.Fatal(e)
	}
	a.Downloads = append([]ResourceDownload{}, a.Downloads...)
	a.Downloads[0].Path = "../outside"
	if e := validateArtifact(a); e == nil {
		t.Fatal("accepted resource traversal")
	}
	a = pkg.Artifacts["win32-x64"]
	a.Extractor = "tesseract"
	if e := validateArtifact(a); e == nil {
		t.Fatal("accepted unknown extractor")
	}
}
