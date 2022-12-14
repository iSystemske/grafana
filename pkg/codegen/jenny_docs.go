package codegen

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"text/template"

	"cuelang.org/go/cue/cuecontext"
	"github.com/grafana/codejen"
	"github.com/grafana/thema/encoding/jsonschema"
)

func DocsJenny(docsPath string) OneToOne {
	return docsJenny{docsPath: docsPath}
}

type docsJenny struct {
	docsPath string
}

func (j docsJenny) JennyName() string {
	return "DocsJenny"
}

func (j docsJenny) Generate(decl *DeclForGen) (*codejen.File, error) {
	// TODO: added it because 1. decl.Lineage() is nil for RawKind; 2. not sure if docs for other kind types are needed
	if !decl.IsCoreStructured() {
		return nil, nil
	}

	f, err := jsonschema.GenerateSchema(decl.Lineage().Latest())
	if err != nil {
		return nil, err
	}
	b, _ := cuecontext.New().BuildFile(f).MarshalJSON()

	// We don't need the entire json obj, only the value for components.schemas.<kindName> path
	var obj struct {
		Components struct {
			Schemas map[string]json.RawMessage
		}
	}
	err = json.Unmarshal(b, &obj)
	if err != nil {
		return nil, err
	}

	kindProps := decl.Properties.Common()
	kindName := strings.ToLower(kindProps.Name)
	kindJSON := obj.Components.Schemas[kindName]

	// Add kind metadata to template and generate template file
	data := templateData{
		KindName:     kindName,
		KindVersion:  decl.Lineage().Latest().Version().String(),
		KindMaturity: string(kindProps.Maturity),
		Markdown:     "{{ .Markdown 1 }}",
	}

	tmpl, err := makeMarkdownTmpl(data)
	if err != nil {
		return nil, err
	}
	_ = tmpl // TODO: Do nothing with template till we figure out how to generate markdown using https://github.com/marcusolsson/json-schema-docs

	jsonBuf := new(bytes.Buffer)
	err = json.Indent(jsonBuf, kindJSON, "", "  ")
	return codejen.NewFile(filepath.Join(j.docsPath, kindName)+".json", jsonBuf.Bytes(), j), nil
}

type templateData struct {
	KindName     string
	KindVersion  string
	KindMaturity string
	Markdown     string
}

func makeMarkdownTmpl(data templateData) ([]byte, error) {
	tmpl, err := template.New("docs").Parse(mdTmpl)
	if err != nil {
		return []byte{}, err
	}
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, data)
	return buf.Bytes(), err
}

const mdTmpl = `
+++
# -------------------------------------------------------------------------
# Do not edit this file. It is automatically generated by json-schema-docs.
# -------------------------------------------------------------------------
title = "{{ .KindName }}.json"
keywords = ["grafana", "schema", "documentation"]
+++

# Kind: {{ .KindName }}

### Maturity: {{ .KindMaturity }}
### Version: {{ .KindVersion }}

{{ .Markdown }}
`
