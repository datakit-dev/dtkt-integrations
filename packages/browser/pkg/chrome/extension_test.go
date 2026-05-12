package chrome

import (
	"bytes"
	"html/template"
	"testing"
)

func TestInstallTemplate(t *testing.T) {
	tmpl, err := template.ParseFS(tmplFS, "template/install.html.tmpl")
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]any{
		"INSTALL_DIR": "foo/bar",
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(buf.String())
}
