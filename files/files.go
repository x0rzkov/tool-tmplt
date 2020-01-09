/*
This file is based on github.com/helm/helm/pkg/chartutil/files.go
	Copyright The Helm Authors.
	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
Changes:
- file access
- filepath bug fixes
- no 'little' dependencies (protobuf any)

*/
package files

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/golang/glog"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
	"github.com/BurntSushi/toml"
)

// Dir is the path to load relative files from.
type Dir string

// Files is a map of files in a chart that can be accessed from a template.
type Files map[string]string

//TODO type Files map[string][]byte

// Get returns a string representation of the given file.
//
// Fetch the contents of a file as a string. It is designed to be called in a
// template.
//
//	{{.Files.Get "foo"}}
func (dir Dir) Get(name string) string {
	p := filepath.Join(string(dir), name)
	b, err := ioutil.ReadFile(p)
	if err != nil {
		glog.Exitf("Files.Get failed: %v", err)
	}
	return string(b)
}

// Glob takes a glob pattern and returns another files object only containing
// matched  files.
//
// This is designed to be called from a template.
//
// {{ range $name, $content := .Files.Glob("foo/**") }}
// {{ $name }}: |
// {{ .Files.Get($name) | indent 4 }}{{ end }}
func (dir Dir) Glob(pattern string) Files {
	p := filepath.Join(string(dir), pattern)
	fs, err := filepath.Glob(p)
	if err != nil {
		glog.Exitf("Files.Glob %v failed: %v", pattern, err)
	}

	//TODO m := map[string][]byte{}
	m := make(Files, 0)
	for _, f := range fs {
		b, err := ioutil.ReadFile(f)
		if err != nil {
			glog.Exitf("Files.Glob %v failed: %v", pattern, err)
		}
		m[f] = string(b)
	}
	return m
}

// AsConfig turns a Files group and flattens it to a YAML map suitable for
// including in the 'data' section of a Kubernetes ConfigMap definition.
// Duplicate keys will be overwritten, so be aware that your file names
// (regardless of path) should be unique.
//
// This is designed to be called from a template, and will return empty string
// (via ToYaml function) if it cannot be serialized to YAML, or if the Files
// object is nil.
//
// The output will not be indented, so you will want to pipe this to the
// 'indent' template function.
//
//   data:
// {{ (.Files.Glob "config/**").AsConfig() | indent 4 }}
func (f Files) AsConfig() string {
	if f == nil {
		return ""
	}

	m := map[string]string{}

	// Explicitly convert to strings, and file names
	for k, v := range f {
		m[filepath.Base(k)] = string(v)
	}

	return ToYaml(m)
}

// AsSecrets returns the base64-encoded value of a Files object suitable for
// including in the 'data' section of a Kubernetes Secret definition.
// Duplicate keys will be overwritten, so be aware that your file names
// (regardless of path) should be unique.
//
// This is designed to be called from a template, and will return empty string
// (via ToYaml function) if it cannot be serialized to YAML, or if the Files
// object is nil.
//
// The output will not be indented, so you will want to pipe this to the
// 'indent' template function.
//
//   data:
// {{ (.Files.Glob "secrets/*").AsSecrets() }}
func (f Files) AsSecrets() string {
	if f == nil {
		return ""
	}

	m := map[string]string{}

	for k, v := range f {
		m[filepath.Base(k)] = string(base64.StdEncoding.EncodeToString([]byte(v)))
	}

	return ToYaml(m)
}

// Lines returns each line of a named file (split by "\n") as a slice, so it can
// be ranged over in your templates.
//
// This is designed to be called from a template.
//
// {{ range .Files.Lines "foo/bar.html" }}
// {{ . }}{{ end }}
//func (f Files) Lines(path string) []string {
//	if f == nil || f[path] == nil {
//		return []string{}
//	}
//
//	return strings.Split(string(f[path]), "\n")
//}

// ToYaml takes an interface, marshals it to yaml, and returns a string. It will
// always return a string, even on marshal error (empty string).
//
// This is designed to be called from a template.
func ToYaml(v interface{}) string {
	data, err := yaml.Marshal(v)
	if err != nil {
		// Swallow errors inside of a template.
		return ""
	}
	return string(data)
}


// FromYaml converts a YAML document into a map[string]interface{}.
//
// This is not a general-purpose YAML parser, and will not parse all valid
// YAML documents. Additionally, because its intended use is within templates
// it tolerates errors. It will insert the returned error message string into
// m["Error"] in the returned map.
func FromYaml(str string) map[string]interface{} {
	m := map[string]interface{}{}

	if err := yaml.Unmarshal([]byte(str), &m); err != nil {
		m["Error"] = err.Error()
	}
	return m
}

// ToToml takes an interface, marshals it to toml, and returns a string. It will
// always return a string, even on marshal error (empty string).
//
// This is designed to be called from a template.
func ToToml(v interface{}) string {
	b := bytes.NewBuffer(nil)
	e := toml.NewEncoder(b)
	err := e.Encode(v)
	if err != nil {
		return err.Error()
	}
	return b.String()
}

// ToJson takes an interface, marshals it to json, and returns a string. It will
// always return a string, even on marshal error (empty string).
//
// This is designed to be called from a template.
func ToJson(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		// Swallow errors inside of a template.
		return ""
	}
	return string(data)
}

// FromJson converts a JSON document into a map[string]interface{}.
//
// This is not a general-purpose JSON parser, and will not parse all valid
// JSON documents. Additionally, because its intended use is within templates
// it tolerates errors. It will insert the returned error message string into
// m["Error"] in the returned map.
func FromJson(str string) map[string]interface{} {
	m := map[string]interface{}{}

	if err := json.Unmarshal([]byte(str), &m); err != nil {
		m["Error"] = err.Error()
	}
	return m
}