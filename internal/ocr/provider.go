// Package ocr wraps a locally managed Tesseract engine. Inputs are never uploaded.
package ocr

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fu9zhou/finish-bit/internal/toolrun"
	"github.com/fu9zhou/finish-bit/pkg/operation"
)

func Register(reg *operation.Registry, resolver toolrun.Resolver) error {
	for _, s := range []struct{ id, summary, alias string }{{"ocr.text", "Recognize printed text in a local image", "图片文字识别 印刷体识别 英文识别"}, {"ocr.words", "Recognize words with bounding boxes and confidence", "OCR文字坐标"}, {"ocr.to-pdf", "Create a searchable PDF from an image using local OCR", "图片转可搜索PDF"}, {"ocr.to-html", "Export local OCR as hOCR with word coordinates", "OCR转HTML"}} {
		options := []operation.Parameter{toolrun.Option("language", operation.TypeString, "eng, chi_sim or chi_sim+eng", "chi_sim+eng"), toolrun.Option("layout", operation.TypeInteger, "Page segmentation mode: 3,4,6,7,8,11,13", 3), toolrun.Option("dpi", operation.TypeInteger, "Assumed input DPI", 300), toolrun.Option("overwrite", operation.TypeBoolean, "Replace an existing file", false)}
		required := s.id == "ocr.to-pdf" || s.id == "ocr.to-html"
		options = append(options, toolrun.Param("output", "Output file; optional for text/words", required))
		def := operation.Definition{ID: s.id, Summary: s.summary, Description: s.summary + ". Tesseract fast Chinese/English models, offline; printed text is the target, not guaranteed handwriting or specialized document field extraction.", Aliases: []string{s.alias}, Tags: []string{"ocr", "image", "offline"}, Inputs: []operation.Parameter{toolrun.Param("input", "Local PNG/JPEG/GIF up to 32 MiB and 25 million pixels", true)}, Options: options, Requirements: []operation.Requirement{{Package: "tesseract"}}, Source: "tesseract"}
		if err := reg.Register(operation.Capability{Definition: def, Runner: operation.Func(func(ctx context.Context, r operation.Request) (operation.Result, error) {
			if err := operation.ValidateRequest(def, r); err != nil {
				return operation.Result{}, err
			}
			return Run(ctx, resolver, s.id, r)
		})}); err != nil {
			return err
		}
	}
	return nil
}
func Run(ctx context.Context, resolver toolrun.Resolver, id string, r operation.Request) (operation.Result, error) {
	if err := operation.RequireInputs(r, 1); err != nil {
		return operation.Result{}, err
	}
	v := &toolrun.Values{Request: r}
	language := v.Enum("language", "chi_sim+eng", "eng", "chi_sim", "chi_sim+eng")
	layout := v.Int("layout", 3, 3, 13)
	v.Check(layout == 3 || layout == 4 || layout == 6 || layout == 7 || layout == 8 || layout == 11 || layout == 13, "layout must be 3,4,6,7,8,11 or 13")
	dpi := v.Int("dpi", 300, 70, 1200)
	out := v.String("output", "")
	overwrite := v.Bool("overwrite", false)
	if v.Err != nil {
		return operation.Result{}, v.Err
	}
	file, e := toolrun.LocalFile(r.Inputs[0])
	if e != nil {
		return operation.Result{}, e
	}
	f, e := os.Open(file)
	if e != nil {
		return operation.Result{}, e
	}
	data, e := io.ReadAll(io.LimitReader(f, (32<<20)+1))
	_ = f.Close()
	if e != nil {
		return operation.Result{}, e
	}
	if len(data) > 32<<20 {
		return operation.Result{}, toolrun.Invalid("OCR input exceeds 32 MiB")
	}
	cfg, format, e := image.DecodeConfig(bytes.NewReader(data))
	if e != nil || cfg.Width < 1 || cfg.Height < 1 || cfg.Width > 32768 || cfg.Height > 32768 || cfg.Width > 25000000/cfg.Height {
		return operation.Result{}, toolrun.Invalid("invalid image or more than 25 million pixels")
	}
	if format != "png" && format != "jpeg" && format != "gif" {
		return operation.Result{}, toolrun.Invalid("OCR input must be PNG/JPEG/GIF")
	}
	binary, e := resolver.Executable("tesseract", "tesseract")
	if e != nil {
		return operation.Result{}, e
	}
	tessdata := filepath.Join(filepath.Dir(binary), "tessdata")
	for _, lang := range strings.Split(language, "+") {
		if _, e := os.Stat(filepath.Join(tessdata, lang+".traineddata")); e != nil {
			return operation.Result{}, &operation.Error{Code: operation.CodeDependencyMissing, Message: "OCR model missing: " + lang, Suggestion: "fnsh pkg repair tesseract"}
		}
	}
	work, e := os.MkdirTemp("", "finishbit-ocr-")
	if e != nil {
		return operation.Result{}, e
	}
	defer os.RemoveAll(work)
	input := "input." + format
	if e = os.WriteFile(filepath.Join(work, input), data, 0600); e != nil {
		return operation.Result{}, e
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	args := []string{input, "stdout", "--tessdata-dir", tessdata, "-l", language, "--oem", "1", "--psm", strconv.Itoa(layout), "--dpi", strconv.Itoa(dpi)}
	switch id {
	case "ocr.words":
		args = append(args, "tsv")
	case "ocr.to-pdf":
		args[1] = "result"
		args = append(args, "pdf")
	case "ocr.to-html":
		args[1] = "result"
		args = append(args, "hocr")
	case "ocr.text":
	default:
		return operation.Result{}, toolrun.Invalid("unknown OCR operation")
	}
	stdout, e := toolrun.RunWithEnv(ctx, resolver, "tesseract", "tesseract", work, append(os.Environ(), "OMP_THREAD_LIMIT=2", "TESSDATA_PREFIX="+tessdata), args...)
	if e != nil {
		return operation.Result{}, e
	}
	result := operation.Result{Operation: id, Data: map[string]any{"language": language, "width": cfg.Width, "height": cfg.Height}}
	if id == "ocr.to-pdf" || id == "ocr.to-html" {
		ext := ".pdf"
		source := "result.pdf"
		if id == "ocr.to-html" {
			ext = ".html"
			source = "result.hocr"
		}
		if strings.ToLower(filepath.Ext(out)) != ext {
			return operation.Result{}, toolrun.Invalid("output extension must be " + ext)
		}
		e = toolrun.Artifact(ctx, out, overwrite, func(p string) error { return toolrun.CopyFile(filepath.Join(work, source), p) })
		if e != nil {
			return operation.Result{}, e
		}
		result.Outputs = []string{out}
		return result, nil
	}
	if id == "ocr.text" {
		result.Data["text"] = string(stdout)
	} else {
		words, e := parseTSV(stdout)
		if e != nil {
			return operation.Result{}, e
		}
		result.Data["words"] = words
	}
	if out != "" {
		payload := stdout
		if id == "ocr.words" {
			payload, e = json.MarshalIndent(result.Data, "", "  ")
			if e != nil {
				return operation.Result{}, e
			}
		}
		e = toolrun.Artifact(ctx, out, overwrite, func(p string) error { return os.WriteFile(p, payload, 0600) })
		if e != nil {
			return operation.Result{}, e
		}
		result.Outputs = []string{out}
		delete(result.Data, "text")
		delete(result.Data, "words")
	}
	return result, nil
}
func parseTSV(data []byte) ([]map[string]any, error) {
	r := csv.NewReader(bytes.NewReader(data))
	r.Comma = '\t'
	r.FieldsPerRecord = -1
	r.LazyQuotes = true
	header, e := r.Read()
	if e != nil || len(header) != 12 || header[0] != "level" {
		return nil, toolrun.Invalid("invalid Tesseract TSV header")
	}
	words := []map[string]any{}
	for {
		row, e := r.Read()
		if e == io.EOF {
			break
		}
		if e != nil || len(row) < 12 {
			return nil, toolrun.Invalid("invalid Tesseract TSV row")
		}
		if row[0] != "5" {
			continue
		}
		item := map[string]any{"text": strings.Join(row[11:], "\t")}
		for i := 1; i < 10; i++ {
			n, e := strconv.Atoi(row[i])
			if e != nil {
				return nil, toolrun.Invalid("invalid OCR coordinate")
			}
			item[header[i]] = n
		}
		confidence, e := strconv.ParseFloat(row[10], 64)
		if e != nil {
			return nil, toolrun.Invalid("invalid OCR confidence")
		}
		item["confidence"] = confidence
		words = append(words, item)
		if len(words) > 100000 {
			return nil, toolrun.Invalid("OCR exceeds 100000 words")
		}
	}
	return words, nil
}
