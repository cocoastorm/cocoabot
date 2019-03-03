package main

import (
	"bytes"
	"html/template"
	"log"
	"math/rand"
	"strings"
	"time"
)

func normalize(input string) []string {
	input = strings.TrimSpace(input)
	items := strings.Split(input, ",")

	for i, item := range items {
		word := strings.TrimSpace(item)
		items[i] = strings.ToLower(word)
	}

	return items
}

func readDecide(line string) (string, []string) {
	today := time.Now()
	choices := normalize(line)

	rand.Seed(today.UnixNano())

	return choices[rand.Intn(len(choices))], choices
}

func Decide(q string) string {
	const replyTmpl = `
Decision

**{{ .Answer }}**

Choices Given:
{{ range .Choices }}
	* {{ . }}
{{end}}
`

	t := template.Must(template.New("decide").Parse(replyTmpl))
	choice, list := readDecide(q)

	reply := struct {
		Answer  string
		Choices []string
	}{
		Answer:  choice,
		Choices: list,
	}

	buf := bytes.NewBuffer(nil)
	if err := t.Execute(buf, reply); err != nil {
		log.Println("OH NOES!")
		return "error, please try again later"
	}

	return buf.String()
}
