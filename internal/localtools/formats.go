package localtools

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/fu9zhou/finish-bit/internal/toolrun"
	"github.com/fu9zhou/finish-bit/pkg/operation"
	pinyin "github.com/mozillazg/go-pinyin"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	yaml "go.yaml.in/yaml/v3"
)

func formatSpecs() []spec {
	rows := []spec{
		{id: "yaml.to-json", summary: "Convert a single YAML document into lossless JSON-compatible values", alias: "YAML转JSON", inputs: in("input")},
		{id: "json.to-yaml", summary: "Convert JSON into YAML with explicit scalar types", alias: "JSON转YAML", inputs: in("input")},
		{id: "yaml.format", summary: "Format a bounded YAML document", alias: "YAML格式化", inputs: in("input")},
		{id: "yaml.validate", summary: "Validate a single YAML document and its JSON-compatible mapping", alias: "YAML校验", inputs: in("input")},
		{id: "xml.format", summary: "Format XML without fetching external entities", alias: "XML格式化", inputs: in("input")},
		{id: "xml.validate", summary: "Validate XML well-formedness without external entities or DTDs", alias: "XML校验", inputs: in("input")},
		{id: "markdown.to-html", summary: "Render Markdown and GFM to HTML locally with raw HTML disabled", alias: "Markdown渲染", inputs: in("input"), options: []operation.Parameter{boolean("standalone", "Write a standalone HTML page", true), str("title", "Standalone page title", "Document")}},
		{id: "text.pinyin", summary: "Convert Chinese characters into pinyin or initials", alias: "文字转拼音", inputs: in("input"), options: []operation.Parameter{str("style", "plain, tone, number or initials", "plain"), boolean("heteronym", "Include alternate pronunciations", false), str("separator", "Syllable separator", " ")}},
	}
	for i := range rows {
		id := rows[i].id
		rows[i].run = func(ctx context.Context, v *toolrun.Values, a []string) (map[string]any, error) {
			return runFormat(ctx, id, v, a)
		}
	}
	return rows
}
func runFormat(ctx context.Context, id string, v *toolrun.Values, a []string) (map[string]any, error) {
	x := a[0]
	switch id {
	case "yaml.to-json", "yaml.format", "yaml.validate":
		var root yaml.Node
		d := yaml.NewDecoder(strings.NewReader(x))
		if e := d.Decode(&root); e != nil {
			return nil, invalid("invalid YAML: " + e.Error())
		}
		var extra yaml.Node
		if e := d.Decode(&extra); e != io.EOF {
			return nil, invalid("expected one YAML document")
		}
		budget := 100000
		data, e := yamlValue(&root, 0, &budget)
		if e != nil {
			return nil, e
		}
		if id == "yaml.validate" {
			return map[string]any{"valid": true, "profile": "single document; no aliases; unique string mapping keys"}, nil
		}
		if id == "yaml.format" {
			var b bytes.Buffer
			enc := yaml.NewEncoder(&b)
			enc.SetIndent(2)
			if e := enc.Encode(&root); e != nil {
				return nil, e
			}
			if e := enc.Close(); e != nil {
				return nil, e
			}
			return value(b.String()), nil
		}
		b, e := json.MarshalIndent(data, "", "  ")
		return value(string(b) + "\n"), e
	case "json.to-yaml":
		data, e := parseJSON(x)
		if e != nil {
			return nil, e
		}
		budget := 100000
		node, e := jsonYAML(data, 0, &budget)
		if e != nil {
			return nil, e
		}
		var b bytes.Buffer
		enc := yaml.NewEncoder(&b)
		enc.SetIndent(2)
		if e := enc.Encode(node); e != nil {
			return nil, e
		}
		if e := enc.Close(); e != nil {
			return nil, e
		}
		return value(b.String()), nil
	case "xml.format", "xml.validate":
		dec := xml.NewDecoder(strings.NewReader(x))
		var b bytes.Buffer
		enc := xml.NewEncoder(&b)
		enc.Indent("", "  ")
		depth, roots, tokens := 0, 0, 0
		for {
			if e := ctx.Err(); e != nil {
				return nil, e
			}
			t, e := dec.Token()
			if e == io.EOF {
				break
			}
			if e != nil {
				return nil, invalid("invalid XML: " + e.Error())
			}
			tokens++
			if tokens > 100000 {
				return nil, invalid("XML exceeds 100000 tokens")
			}
			switch z := t.(type) {
			case xml.StartElement:
				if depth == 0 {
					roots++
				}
				depth++
				if depth > 128 {
					return nil, invalid("XML exceeds depth 128")
				}
			case xml.EndElement:
				depth--
			case xml.Directive:
				return nil, invalid("XML directives/DTDs are not accepted")
			case xml.CharData:
				if depth == 0 && strings.TrimSpace(string(z)) != "" {
					return nil, invalid("text outside XML root")
				}
			}
			if id == "xml.format" {
				if e := enc.EncodeToken(t); e != nil {
					return nil, e
				}
			}
		}
		if roots != 1 || depth != 0 {
			return nil, invalid("XML needs exactly one complete root element")
		}
		if id == "xml.validate" {
			return map[string]any{"valid": true, "validation": "well-formedness; no XSD or DTD validation"}, nil
		}
		if e := enc.Flush(); e != nil {
			return nil, e
		}
		return value(b.String() + "\n"), nil
	case "markdown.to-html":
		if len(x) > 1<<20 {
			return nil, invalid("Markdown exceeds 1 MiB")
		}
		var b bytes.Buffer
		md := goldmark.New(goldmark.WithExtensions(extension.GFM))
		if e := md.Convert([]byte(x), &b); e != nil {
			return nil, e
		}
		out := b.String()
		if v.Bool("standalone", true) {
			title := v.String("title", "Document")
			out = "<!doctype html>\n<html><head><meta charset=\"utf-8\"><meta name=\"viewport\" content=\"width=device-width\"><meta http-equiv=\"Content-Security-Policy\" content=\"default-src 'none'; img-src data:; style-src 'unsafe-inline'\"><title>" + escapeHTML(title) + "</title><style>body{max-width:70ch;margin:3em auto;padding:0 1em;font:18px/1.6 system-ui}pre{overflow:auto}img{max-width:100%}table{border-collapse:collapse}td,th{padding:.4em;border:1px solid #aaa}</style></head><body>" + out + "</body></html>"
		}
		return value(out), nil
	case "text.pinyin":
		if len([]rune(x)) > 100000 {
			return nil, invalid("pinyin input exceeds 100000 characters")
		}
		args := pinyin.NewArgs()
		switch v.Enum("style", "plain", "plain", "tone", "number", "initials") {
		case "tone":
			args.Style = pinyin.Tone
		case "number":
			args.Style = pinyin.Tone3
		case "initials":
			args.Style = pinyin.FirstLetter
		}
		args.Heteronym = v.Bool("heteronym", false)
		args.Fallback = func(r rune, a pinyin.Args) []string { return []string{string(r)} }
		parts := pinyin.Pinyin(x, args)
		words := make([]string, len(parts))
		for i, part := range parts {
			words[i] = strings.Join(part, "/")
		}
		sep := v.String("separator", " ")
		if len(sep) > 100 {
			return nil, invalid("separator exceeds 100 bytes")
		}
		return map[string]any{"text": strings.Join(words, sep), "syllables": parts, "disambiguation": "dictionary candidates; no contextual disambiguation"}, nil
	}
	return nil, invalid("unknown format operation")
}
func yamlValue(n *yaml.Node, depth int, budget *int) (any, error) {
	*budget--
	if depth > 128 || *budget < 0 {
		return nil, invalid("YAML exceeds depth 128 or 100000 nodes")
	}
	switch n.Kind {
	case yaml.DocumentNode:
		if len(n.Content) != 1 {
			return nil, invalid("expected one YAML root")
		}
		return yamlValue(n.Content[0], depth+1, budget)
	case yaml.AliasNode:
		return nil, invalid("YAML aliases are not accepted; expand them explicitly")
	case yaml.MappingNode:
		m := map[string]any{}
		for i := 0; i < len(n.Content); i += 2 {
			k := n.Content[i]
			if k.Kind != yaml.ScalarNode || k.Tag != "!!str" {
				return nil, invalid("YAML mapping keys must be strings")
			}
			if _, ok := m[k.Value]; ok {
				return nil, invalid("duplicate YAML key: " + k.Value)
			}
			v, e := yamlValue(n.Content[i+1], depth+1, budget)
			if e != nil {
				return nil, e
			}
			m[k.Value] = v
		}
		return m, nil
	case yaml.SequenceNode:
		a := []any{}
		for _, c := range n.Content {
			v, e := yamlValue(c, depth+1, budget)
			if e != nil {
				return nil, e
			}
			a = append(a, v)
		}
		return a, nil
	case yaml.ScalarNode:
		switch n.Tag {
		case "!!str", "!!timestamp":
			return n.Value, nil
		case "!!null":
			return nil, nil
		case "!!bool":
			return strings.EqualFold(n.Value, "true"), nil
		case "!!int", "!!float":
			s := strings.ReplaceAll(n.Value, "_", "")
			if _, e := decimal(s); e != nil {
				return nil, invalid("YAML numeric scalars must be finite decimal values; convert hex/octal explicitly")
			}
			if strings.HasPrefix(s, "+") {
				s = s[1:]
			}
			if !json.Valid([]byte(s)) {
				r, _ := decimal(s)
				if r.IsInt() {
					s = r.Num().String()
				} else {
					return nil, invalid("YAML number is not a JSON number")
				}
			}
			return json.Number(s), nil
		default:
			return nil, invalid("unsupported YAML scalar tag: " + n.Tag)
		}
	}
	return nil, invalid("unsupported YAML node")
}
func jsonYAML(x any, depth int, budget *int) (*yaml.Node, error) {
	*budget--
	if depth > 128 || *budget < 0 {
		return nil, invalid("JSON exceeds depth 128 or 100000 nodes")
	}
	n := &yaml.Node{}
	switch v := x.(type) {
	case map[string]any:
		n.Kind = yaml.MappingNode
		n.Tag = "!!map"
		keys := []string{}
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			c, e := jsonYAML(v[k], depth+1, budget)
			if e != nil {
				return nil, e
			}
			n.Content = append(n.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: k}, c)
		}
	case []any:
		n.Kind = yaml.SequenceNode
		n.Tag = "!!seq"
		for _, i := range v {
			c, e := jsonYAML(i, depth+1, budget)
			if e != nil {
				return nil, e
			}
			n.Content = append(n.Content, c)
		}
	case json.Number:
		n.Kind = yaml.ScalarNode
		n.Tag = "!!int"
		if strings.ContainsAny(string(v), ".eE") {
			n.Tag = "!!float"
		}
		n.Value = string(v)
	case string:
		n.Kind = yaml.ScalarNode
		n.Tag = "!!str"
		n.Value = v
	case bool:
		n.Kind = yaml.ScalarNode
		n.Tag = "!!bool"
		n.Value = fmt.Sprint(v)
	case nil:
		n.Kind = yaml.ScalarNode
		n.Tag = "!!null"
		n.Value = "null"
	}
	return n, nil
}
