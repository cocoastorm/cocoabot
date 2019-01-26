package decide

import (
	"bytes"
	"html/template"
	"math/rand"
	"strings"
	"time"
)

func normalize(input string) []string {
	input = strings.TrimSpace(input)
	items := strings.Split(input, " ")

	for i, item := range items {
		items[i] = strings.ToLower(item)
	}

	return items
}

func readInLine(line string) (string, []string) {
	inputDataForSeed := time.Now().UnixNano()
	choices := normalize(line)

	rand.Seed(inputDataForSeed)

	return choices[rand.Intn(len(choices))], choices
}

func Decide(q string) string {
	const replyTmpl = `
### Decision

Original Query
> {{ .Query }}

Decided!
{{ .Decision }}

Choices
{{ _, choice := range .Choices }}
	- {{ .choice }}
{{ end }}
`

	t := template.Must(template.New("reply").Parse(replyTmpl))
	choice, entrants := readInLine(q)

	reply := struct {
		Query    string
		Decision string
		Choices  []string
	}{
		Query:    q,
		Decision: choice,
		Choices:  entrants,
	}

	buf := bytes.NewBuffer(nil)
	if err := t.Execute(buf, reply); err != nil {
		return ""
	}

	return buf.String()
}
