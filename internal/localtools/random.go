package localtools

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/fu9zhou/finish-bit/internal/toolrun"
	"github.com/fu9zhou/finish-bit/pkg/operation"
)

func randomSpecs() []spec {
	rows := []spec{
		{id: "random.integer", summary: "Generate unbiased cryptographic random integers", alias: "随机数生成", inputs: in("min", "max"), options: []operation.Parameter{integer("count", "Number of results", 1), boolean("unique", "Sample without replacement", false)}},
		{id: "random.choose", summary: "Choose or shuffle entries from a user-supplied line list", alias: "随机抽签 今天吃什么", inputs: in("input"), options: []operation.Parameter{integer("count", "Number of items; 0 means shuffle all", 1), boolean("replacement", "Allow drawing the same entry again", false)}},
		{id: "password.generate", summary: "Generate cryptographic passwords from a configurable alphabet", alias: "随机密码生成", options: []operation.Parameter{integer("length", "Characters per password", 20), integer("count", "Number of passwords", 1), str("alphabet", "Distinct characters; blank uses ASCII letters, digits and symbols", "")}},
		{id: "password.inspect", summary: "Inspect password patterns locally without logging or network lookups", alias: "密码安全检测", inputs: in("password")},
		{id: "name.generate", summary: "Combine user-supplied words into random names", alias: "随机网名 项目名生成", inputs: in("prefixes", "suffixes"), options: []operation.Parameter{integer("count", "Number of names", 10), str("separator", "Join text", " ")}},
		{id: "crypto.key", summary: "Generate a 256-bit cryptographic key", alias: "生成加密密钥"},
		{id: "crypto.encrypt", summary: "Encrypt UTF-8 text with authenticated AES-256-GCM", alias: "本地文本加密", inputs: in("input"), options: []operation.Parameter{toolrun.Param("key", "64 hexadecimal key characters (not a password)", true)}},
		{id: "crypto.decrypt", summary: "Decrypt and authenticate a FinishBit AES-256-GCM envelope", alias: "本地文本解密", inputs: in("input"), options: []operation.Parameter{toolrun.Param("key", "64 hexadecimal key characters", true)}},
	}
	for i := range rows {
		id := rows[i].id
		rows[i].run = func(ctx context.Context, v *toolrun.Values, a []string) (map[string]any, error) {
			return runRandom(ctx, id, v, a)
		}
	}
	return rows
}
func randomIndex(n int) (int, error) {
	x, e := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if e != nil {
		return 0, e
	}
	return int(x.Int64()), nil
}
func lineItems(s string) []string {
	out := []string{}
	for _, x := range strings.Split(s, "\n") {
		if x = strings.TrimSpace(x); x != "" {
			out = append(out, x)
		}
	}
	return out
}
func runRandom(ctx context.Context, id string, v *toolrun.Values, a []string) (map[string]any, error) {
	switch id {
	case "random.integer":
		if len(a[0]) > 128 || len(a[1]) > 128 {
			return nil, invalid("bounds exceed 128 digits")
		}
		lo, ok := new(big.Int).SetString(a[0], 10)
		hi, ok2 := new(big.Int).SetString(a[1], 10)
		if !ok || !ok2 || lo.Cmp(hi) > 0 {
			return nil, invalid("invalid integer interval")
		}
		n := v.Int("count", 1, 1, 10000)
		unique := v.Bool("unique", false)
		span := new(big.Int).Add(new(big.Int).Sub(hi, lo), big.NewInt(1))
		if unique && span.Cmp(big.NewInt(int64(n))) < 0 {
			return nil, invalid("unique count exceeds interval size")
		}
		if v.Err != nil {
			return nil, v.Err
		}
		out := []string{}
		swaps := map[string]*big.Int{}
		remaining := new(big.Int).Set(span)
		for i := 0; i < n; i++ {
			if e := ctx.Err(); e != nil {
				return nil, e
			}
			bound := span
			if unique {
				bound = remaining
			}
			r, e := rand.Int(rand.Reader, bound)
			if e != nil {
				return nil, e
			}
			selected := new(big.Int).Set(r)
			if unique {
				if x, ok := swaps[r.String()]; ok {
					selected.Set(x)
				}
				last := new(big.Int).Sub(remaining, big.NewInt(1))
				tail := new(big.Int).Set(last)
				if x, ok := swaps[last.String()]; ok {
					tail.Set(x)
				}
				swaps[r.String()] = tail
				remaining.Sub(remaining, big.NewInt(1))
			}
			out = append(out, new(big.Int).Add(selected, lo).String())
		}
		return map[string]any{"values": out, "source": "crypto/rand"}, nil
	case "random.choose":
		items := lineItems(a[0])
		if len(items) == 0 || len(items) > 100000 {
			return nil, invalid("provide 1 to 100000 nonempty lines")
		}
		for _, item := range items {
			if len(item) > 1024 {
				return nil, invalid("each choice must be at most 1024 bytes")
			}
		}
		n := v.Int("count", 1, 0, 10000)
		replacement := v.Bool("replacement", false)
		if n == 0 {
			n = len(items)
		}
		if n > 10000 {
			return nil, invalid("at most 10000 results")
		}
		if !replacement && n > len(items) {
			return nil, invalid("count exceeds available items")
		}
		if v.Err != nil {
			return nil, v.Err
		}
		out := []string{}
		for i := 0; i < n; i++ {
			if e := ctx.Err(); e != nil {
				return nil, e
			}
			limit := len(items)
			if !replacement {
				limit -= i
			}
			j, e := randomIndex(limit)
			if e != nil {
				return nil, e
			}
			if !replacement {
				j += i
				items[i], items[j] = items[j], items[i]
				out = append(out, items[i])
			} else {
				out = append(out, items[j])
			}
		}
		return map[string]any{"values": out}, nil
	case "password.generate":
		n := v.Int("count", 1, 1, 1000)
		length := v.Int("length", 20, 1, 4096)
		alphabet := v.String("alphabet", "")
		if alphabet == "" {
			alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+"
		}
		runes := []rune(alphabet)
		if len(runes) < 2 || len(runes) > 1024 {
			return nil, invalid("alphabet needs 2 to 1024 distinct characters")
		}
		seen := map[rune]bool{}
		for _, r := range runes {
			if seen[r] || unicode.IsControl(r) {
				return nil, invalid("alphabet must have distinct non-control characters")
			}
			seen[r] = true
		}
		if n*length > 100000 {
			return nil, invalid("total password characters exceed 100000")
		}
		if v.Err != nil {
			return nil, v.Err
		}
		out := []string{}
		for i := 0; i < n; i++ {
			if e := ctx.Err(); e != nil {
				return nil, e
			}
			b := make([]rune, length)
			for j := range b {
				k, e := randomIndex(len(runes))
				if e != nil {
					return nil, e
				}
				b[j] = runes[k]
			}
			out = append(out, string(b))
		}
		return map[string]any{"values": out, "alphabet_size": len(runes), "entropy_bits": float64(length) * math.Log2(float64(len(runes))), "policy": "uniform independent characters; no forced class quota"}, nil
	case "password.inspect":
		x := a[0]
		if len(x) > 4096 {
			return nil, invalid("password exceeds 4096 bytes")
		}
		classes := map[string]bool{"lower": false, "upper": false, "digit": false, "other": false}
		seen := map[rune]bool{}
		for _, r := range x {
			seen[r] = true
			switch {
			case unicode.IsLower(r):
				classes["lower"] = true
			case unicode.IsUpper(r):
				classes["upper"] = true
			case unicode.IsDigit(r):
				classes["digit"] = true
			default:
				classes["other"] = true
			}
		}
		issues := []string{}
		if utf8.RuneCountInString(x) < 12 {
			issues = append(issues, "shorter than 12 characters")
		}
		if len(seen) < 4 {
			issues = append(issues, "very few distinct characters")
		}
		lower := strings.ToLower(x)
		for _, weak := range []string{"password", "123456", "qwerty", "admin", "letmein", "111111", "abcdef"} {
			if strings.Contains(lower, weak) {
				issues = append(issues, "contains a common weak pattern")
				break
			}
		}
		repeat := false
		for size := 1; size <= len(x)/2; size++ {
			if len(x)%size == 0 && strings.Repeat(x[:size], len(x)/size) == x {
				repeat = true
				break
			}
		}
		if repeat {
			issues = append(issues, "repeated substring")
		}
		return map[string]any{"characters": utf8.RuneCountInString(x), "distinct_characters": len(seen), "classes": classes, "issues": issues, "assessment": "heuristic pattern inspection; not a cracking-time estimate or breach check"}, nil
	case "name.generate":
		p, s := lineItems(a[0]), lineItems(a[1])
		n := v.Int("count", 10, 1, 1000)
		sep := v.String("separator", " ")
		if len(p) == 0 || len(s) == 0 || len(p) > 10000 || len(s) > 10000 || len(sep) > 100 {
			return nil, invalid("provide 1 to 10000 prefixes/suffixes and a separator up to 100 bytes")
		}
		for _, group := range [][]string{p, s} {
			for _, item := range group {
				if len(item) > 1024 {
					return nil, invalid("each name component must be at most 1024 bytes")
				}
			}
		}
		if v.Err != nil {
			return nil, v.Err
		}
		out := []string{}
		for i := 0; i < n; i++ {
			j, e := randomIndex(len(p))
			if e != nil {
				return nil, e
			}
			k, e := randomIndex(len(s))
			if e != nil {
				return nil, e
			}
			out = append(out, p[j]+sep+s[k])
		}
		return map[string]any{"values": out, "uniqueness_guaranteed": false}, nil
	case "crypto.key":
		key := make([]byte, 32)
		if _, e := rand.Read(key); e != nil {
			return nil, e
		}
		return map[string]any{"key": hex.EncodeToString(key), "algorithm": "AES-256-GCM"}, nil
	case "crypto.encrypt", "crypto.decrypt":
		key, e := hex.DecodeString(v.String("key", ""))
		if e != nil || len(key) != 32 {
			return nil, invalid("key must be exactly 64 hexadecimal characters")
		}
		block, e := aes.NewCipher(key)
		if e != nil {
			return nil, e
		}
		gcm, e := cipher.NewGCM(block)
		if e != nil {
			return nil, e
		}
		const prefix = "fnsh:aes256gcm:v1:"
		if id == "crypto.encrypt" {
			nonce := make([]byte, gcm.NonceSize())
			if _, e := rand.Read(nonce); e != nil {
				return nil, e
			}
			sealed := gcm.Seal(nonce, nonce, []byte(a[0]), []byte(prefix))
			return value(prefix + base64.RawURLEncoding.EncodeToString(sealed)), nil
		}
		if !strings.HasPrefix(a[0], prefix) {
			return nil, invalid("unknown encrypted envelope format")
		}
		b, e := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(a[0], prefix))
		if e != nil || len(b) < gcm.NonceSize()+gcm.Overhead() {
			return nil, invalid("invalid encrypted envelope")
		}
		plain, e := gcm.Open(nil, b[:gcm.NonceSize()], b[gcm.NonceSize():], []byte(prefix))
		if e != nil {
			return nil, invalid("authentication failed: wrong key or altered ciphertext")
		}
		if !utf8.Valid(plain) {
			return nil, invalid("decrypted payload is not UTF-8")
		}
		return value(string(plain)), nil
	}
	return nil, invalid(fmt.Sprintf("unknown random operation %s", id))
}
