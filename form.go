package blog

import (
   "bytes"
   "errors"
   "html/template"
   "net/http"
   "strings"
)

type FormItem struct {
     ItemType string
     Name string
     Label string
     Value string
     Required bool 
}

const formHtml = `
   {{range .}}
      {{if ne .ItemType "hidden"}}
        {{if ne .ItemType "submit"}}
         <label for="{{.Name}}">{{.Label}}:{{if .Required}}*{{end}}</label>
        {{end}}
      {{end}}
      {{if ne .ItemType "textarea"}}
         <input type="{{.ItemType}}" name="{{.Name}}" value="{{.Value}}" />
      {{else}}
         <textarea name="{{.Name}}">{{.Value}}</textarea>
      {{end}}
   {{end}}
`
var formTemplate = template.Must(template.New("formTemplate").Parse(formHtml))

func FormRenderItems(items []FormItem) (html template.HTML,err error){
    var doc bytes.Buffer 
    err = formTemplate.Execute(&doc, items)
    if err != nil {
          html = template.HTML("ERROR")
    } else {
          html = template.HTML(doc.String())
    }
    return html,err
}

func FormValidate(r *http.Request,items []FormItem) ( err []error) {
     err = make([]error,0)
     for i:= range items {
        if  items[i].Required {
             if strings.TrimSpace(r.FormValue( items[i].Name )) == "" {
                 err =  append(err,errors.New("Field "+items[i].Label+" is required"))
             }
        }
    }
    return err
}

