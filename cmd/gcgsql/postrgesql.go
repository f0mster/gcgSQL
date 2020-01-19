package main

var templatesPostrgeSQL = `
// Code generated by gcgSQL. DO NOT EDIT.
package {{.Pkg}}

import ({{if .WithContext}}
	"context"{{end}}
	"database/sql"
	"fmt"
	{{$HaveRepeatableParts := false}}{{range $k, $v :=.QueriesData}}{{if $v.HaveRepeatableParts}}{{$HaveRepeatableParts = true}}{{end}}{{end}}{{if $HaveRepeatableParts}}"strings"
{{end}}
	{{if $HaveRepeatableParts}}"strconv"{{end}}
    {{range $k, $v := .Imports}}{{$v}}
{{end}}
)

{{$havectx:=.WithContext}}{{$Ctx := ""}}{{$Tx := ""}}{{if .WithContext}}{{$Ctx = "ctx context.Context, "}}{{end}}{{if .WithTransaction}}{{$Tx = "conn *sql.Tx"}}{{else}}{{$Tx = "conn *sql.DB"}}{{end}}
{{define "queryAndArgs"}}{{if .HaveRepeatableParts}}query, args ...{{else}}"{{.ReplacePlaceHoldersInQueryIfNoRepeatable '$'| Escape}}"{{.Arguments.PrintNamesAndTypes "," "" false}}{{end}}{{end}}
{{define "returnIfError"}}if err != nil {
		return nil, fmt.Errorf("error while executing function {{.Name}} %w", err)
	}{{end}}
{{define "haveRepeatableArgs"}}{{if .HaveRepeatableParts}}
	var subString strings.Builder
	var first = true
    query := "{{.Query| Escape}}"
	cnt := {{.Arguments.CountArgs}}
	var args = make([]interface{}, cnt)
	i:=0
	subQuery := ""
	{{range $k, $v := .Arguments}}
		{{if $v.Repeatable}}
			first = true
			subString.Reset()
			for _, v:= range {{$v.ArgName}} {
			subQuery = "{{$v.RepeatedQuery | Escape}}"
			{{if ne $v.RepeatedQuery ""}}
				{{range $k2, $v2 := $v.RepeatedArgs}}
					args[i] = v.{{$v2.ArgName}}
					i++
					subQuery = strings.Replace(subQuery, "{{$v2.PlaceHolder}}", "$"+strconv.Itoa(i), -1)
				{{end}}
			{{else}}
				args[i] = v
				i++
			{{end}}
			{{if ne $v.Separator ""}}
				if first {
					first = false
				} else {
					subString.Write([]byte("{{$v.Separator| Escape}}"))
				}
			{{end}}
			subString.Write([]byte({{if eq $v.RepeatedQuery ""}}"$"+strconv.Itoa(i)+subQuery{{else}}subQuery{{end}}))
			}
			query = strings.Replace(query, "{{$v.PlaceHolder|Escape}}", subString.String(), -1)
		{{else}}
			query = strings.Replace(query, "{{$v.PlaceHolder|Escape}}", "$", -1)
			args[i] = $v.ArgName
			i++
		{{end}}
	{{end}}
	{{end}}	
{{end}}

{{$isStructAlreadySet := MakeMapStringBool}}

{{range $k, $v :=.QueriesData}}
{{if $v.HaveRepeatableParts}}
	{{range $k2, $v2 := $v.Arguments}}
		{{if and ($v2.IsGeneratedName) (not (index $isStructAlreadySet $v2.ArgName))}}
			{{$tmp := SetMapStringBoolValue $isStructAlreadySet $v2.ArgName true}}
			type {{$v2.ArgType}} struct {
			{{$v2.RepeatedArgs.PrintNamesAndTypes "	" "\n" true}}
			}
		{{end}}
	{{end}}
{{end}}
{{if gt (len $v.ReturnParams) 0}}
type {{$v.Name}}Struct struct {
{{$v.ReturnParams.PrintNamesAndTypes "	" "\n" true}}
}

func {{.Name}}({{$Ctx}}{{$Tx}}{{.Arguments.PrintNamesAndTypes ", " "" true}}) ([]*{{.Name}}Struct, error) {
	{{template "haveRepeatableArgs" .}}
	res, err := conn.Query{{if $havectx}}Context(ctx, {{else}})({{end}}{{template "queryAndArgs" .}})
	{{template "returnIfError" .}}
	ret{{.Name|Title}}Struct := make([]*{{.Name}}Struct, 0)
	for res.Next() {
		retStructRow := {{.Name}}Struct{}
		err = res.Scan({{.ReturnParams.PrintNamesAndTypes "&retStructRow." "," false}})
		{{template "returnIfError" .}}
		ret{{.Name|Title}}Struct = append(ret{{.Name|Title}}Struct, &retStructRow)
	}
	res.Close()
	return ret{{.Name|Title}}Struct, nil
}
{{else}}
func {{.Name}}({{$Ctx}}{{$Tx}}{{.Arguments.PrintNamesAndTypes ", " "" true}}) (sql.Result, error) {
	{{template "haveRepeatableArgs" .}}	
	res, err := conn.Exec{{if $havectx}}Context(ctx, {{else}})({{end}}{{template "queryAndArgs" .}})
	{{template "returnIfError" .}}
	return res, nil
}
{{end}}{{end}}
`