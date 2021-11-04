package fakenews

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/mb-14/gomarkov"
)

func NewGenerator(s Source) *Generator {
	return &Generator{
		source: s,
		chain:  gomarkov.NewChain(1),
	}
}

type Generator struct {
	source Source
	items  []string
	chain  *gomarkov.Chain
}

func (fn *Generator) Init(ctx context.Context) error {
	err := fn.source.Fetch(ctx)
	if err != nil {
		return err
	}
	fn.items = fn.source.Items()
	for _, item := range fn.items {
		fn.chain.Add(strings.Split(item, " "))
	}
	sort.Strings(fn.items)
	return nil
}

func (fn *Generator) isOriginal(item string) bool {
	idx := sort.SearchStrings(fn.items, item)
	if idx < len(fn.items) && fn.items[idx] == item {
		return false
	}
	return true
}

func (fn *Generator) Generate() (string, error) {
	return fn.generate(1)
}

func (fn *Generator) generate(depth int) (string, error) {
	if fn.chain == nil {
		return "", errors.New("generator is not initialized")
	}
	if depth > 100 {
		return "", errors.New("unable to produce original content")
	}
	tokens := []string{gomarkov.StartToken}
	for tokens[len(tokens)-1] != gomarkov.EndToken {
		next, _ := fn.chain.Generate(tokens[(len(tokens) - 1):])
		tokens = append(tokens, next)
	}
	item := strings.Join(tokens[1:len(tokens)-1], " ")
	if fn.isOriginal(item) {
		return item, nil
	}
	return fn.generate(depth + 1)
}
