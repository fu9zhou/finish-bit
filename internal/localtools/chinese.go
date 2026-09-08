package localtools

import (
	"context"
	"embed"
	"strings"
	"sync"

	"github.com/fu9zhou/finish-bit/internal/toolrun"
	"github.com/fu9zhou/finish-bit/pkg/operation"
)

//go:embed data/opencc/*.txt
var chineseData embed.FS

type phraseNode struct {
	children map[rune]*phraseNode
	value    string
}

var dictionaries [2]*phraseNode
var chineseOnce sync.Once

func chineseSpecs() []spec {
	return []spec{{id: "text.chinese", summary: "Convert simplified and traditional Chinese with local phrase dictionaries", alias: "简繁转换", inputs: in("input"), options: []operation.Parameter{str("direction", "s2t or t2s", "s2t")}, run: func(ctx context.Context, v *toolrun.Values, a []string) (map[string]any, error) {
		direction := v.Enum("direction", "s2t", "s2t", "t2s")
		if v.Err != nil {
			return nil, v.Err
		}
		if len([]rune(a[0])) > 200000 {
			return nil, invalid("Chinese conversion supports at most 200000 characters")
		}
		chineseOnce.Do(func() {
			for i, prefix := range []string{"ST", "TS"} {
				root := &phraseNode{children: map[rune]*phraseNode{}}
				for _, name := range []string{"Characters", "Phrases"} {
					b, _ := chineseData.ReadFile("data/opencc/" + prefix + name + ".txt")
					for _, line := range strings.Split(string(b), "\n") {
						parts := strings.SplitN(strings.TrimSpace(line), "\t", 2)
						if len(parts) != 2 {
							continue
						}
						choices := strings.Fields(parts[1])
						if len(choices) == 0 {
							continue
						}
						node := root
						for _, r := range parts[0] {
							if node.children[r] == nil {
								node.children[r] = &phraseNode{children: map[rune]*phraseNode{}}
							}
							node = node.children[r]
						}
						node.value = choices[0]
					}
				}
				dictionaries[i] = root
			}
		})
		index := 0
		if direction == "t2s" {
			index = 1
		}
		root := dictionaries[index]
		runes := []rune(a[0])
		var b strings.Builder
		for i := 0; i < len(runes); {
			if i%1024 == 0 {
				if e := ctx.Err(); e != nil {
					return nil, e
				}
			}
			node := root
			end := i
			replacement := ""
			for j := i; j < len(runes); j++ {
				node = node.children[runes[j]]
				if node == nil {
					break
				}
				if node.value != "" {
					replacement = node.value
					end = j + 1
				}
			}
			if end == i {
				b.WriteRune(runes[i])
				i++
			} else {
				b.WriteString(replacement)
				i = end
			}
		}
		return map[string]any{"text": b.String(), "direction": direction, "dictionary": "OpenCC ver.1.1.7", "algorithm": "longest phrase, first candidate; not OpenCC mmseg or regional vocabulary"}, nil
	}}}
}
