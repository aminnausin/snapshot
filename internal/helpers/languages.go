package helpers

import (
	"fmt"
	"sort"
	"strings"
)

type LangEntry struct {
	Name string
	Data *LangInfo
}

func SortLanguages(langs map[string]*LangInfo) []LangEntry {
	sorted := make([]LangEntry, 0, len(langs))
	for name, data := range langs {
		sorted = append(sorted, LangEntry{Name: name, Data: data})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Data.Size > sorted[j].Data.Size
	})
	return sorted
}

func BuildProgressHTML(entry LangEntry) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(
		`<span style="background-color: %s; width: %.3f%%;" class="progress-item"></span>`,
		entry.Data.Colour, entry.Data.Prop,
	))
	return b.String()
}

func BuildLangListHTML(entry LangEntry, delay int) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(
		`<li style="animation-delay: %dms;">
				<svg xmlns="http://www.w3.org/2000/svg" class="octicon" style="fill:%s;" viewBox="0 0 16 16" version="1.1" width="16" height="16">
					<path fill-rule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8z"></path>
				</svg>
				<span class="lang">%s</span> <span class="percent">%.2f%%</span>
			</li>`, delay, entry.Data.Colour, entry.Name, entry.Data.Prop,
	))
	return b.String()
}
