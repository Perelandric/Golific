package main

func (self *FileData) GatherStructImports() {
	if len(self.Structs) == 0 {
		return
	}
	self.Imports["encoding/json"] = true
}

// If any EnumRepr is `bitflag`, `strings` is needed
func (self *FileData) DoStructSummary() bool {
	return false
}

var struct_tmpl = `
{{- define "generate_struct"}}
{{- range $struct := .}}
{{- $privateType := $struct.GetPrivateTypeName}}
{{- $jsonType := $struct.GetJSONTypeName}}

/*****************************

{{$struct.Name}} struct

******************************/

type {{$privateType}} struct {
  {{- range $f := $struct.Fields}}
  {{- if $f.IsPrivate -}}
  {{$f.Name}} {{$f.Type}} ` + "`" + `{{$f.Tag}}` + "`" + `
  {{end -}}
  {{end -}}
}

type {{$jsonType}} struct {
  *{{- $privateType}}
  {{- range $f := $struct.Fields}}
  {{- if $f.IsPublic}}
  {{$f.Name}} {{$f.Type}} ` + "`" + `{{$f.Tag}}` + "`" + `
  {{- end -}}
  {{end -}}
}

type {{$struct.Name}} struct {
  private {{$privateType}}
  {{- range $f := $struct.Fields}}
  {{- if $f.IsPublic}}
  {{$f.Name}} {{$f.Type}} ` + "`" + `{{$f.Tag}}` + "`" + `
  {{- end -}}
  {{end -}}
}

{{- range $f := $struct.Fields}}
{{- if $f.DoRead}}
func (self *{{$struct.Name}}) {{$f.Read}} () {{$f.Type}} {
  {{- if $f.IsPrivate}}
  return self.private.{{$f.Name}}
  {{- else -}}
  return self.{{$f.Name}}
  {{end -}}
}
{{end -}}
{{if $f.DoWrite}}
func (self *{{$struct.Name}}) {{$f.Write}} ( v {{$f.Type}} ) {
  {{- if $f.IsPrivate}}
  self.private.{{$f.Name}} = v
  {{- else -}}
  self.{{$f.Name}} = v
  {{end -}}
}
{{end -}}
{{end -}}

{{- if $struct.DoJson}}
func (self *{{$struct.Name}}) MarshalJSON() ([]byte, error) {
  return json.Marshal({{$jsonType}} {
    &self.private,
    {{- range $f := $struct.Fields}}
    {{- if $f.IsPublic -}}
    self.{{$f.Name}},
    {{end -}}
    {{end -}}
  })
}

func (self *{{$struct.Name}}) UnmarshalJSON(j []byte) error {
  var temp {{$jsonType}}
  if err := json.Unmarshal(j, &temp); err != nil {
    return err
  }
  self.private = *temp.{{$privateType}}
  {{range $f := $struct.Fields -}}
  {{if $f.IsPublic -}}
  self.{{$f.Name}} = temp.{{$f.Name}}
  {{end -}}
  {{end -}}
  return nil
}
{{end -}}

{{end -}}
{{end -}}
`
