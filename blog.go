package blog

import (
    "appengine"
    "appengine/user"
    "html/template"
    "net/http"
)

const htmlHeader =`<html>
 <head>
  <title>{{.Header.Title}}</title>
 </head>
 <body>
  <div class="Header"><a href="/">Home</a></div>
  <div class="LoginBox">
   {{.Header.LoginMessage}}
   <a href="{{.Header.UserURL}}">{{.Header.UserLabel}}</a>
  </div>
  <h1>{{.Header.Title}}</h1>
`
const htmlFooter = `</body>
</html>
`

type PageHeader struct {
  Title string
  LoginMessage string
  UserLabel string
  UserURL   string 
  IsAdmin   bool
}

const htmlBlogListpage = htmlHeader+`
  <ul>
  {{range .BlogList}}
  <li>
   <h2><a href="/entry/{{.Id}}">{{.Title}}</a></h2>
   <p>{{.Descr}}</p>
  </li>
  {{end}}
  </ul> 
`+htmlFooter;

type BlogListEntry struct {
     Id string
     Title string
     Descr string
}

type BlogListPage struct {
   Header *PageHeader 
   BlogList []BlogListEntry
}


var pageTemplate = template.Must(template.New("blogListPage").Parse(htmlBlogListpage))

func createPageHeader(title string,w http.ResponseWriter,r *http.Request) *PageHeader{
   pageHeader := &PageHeader{ Title: title }
   c := appengine.NewContext(r)
   u := user.Current(c)
   if u == nil {
        url, err := user.LoginURL(c, r.URL.String())
        if err != nil {
            panic("user.LoginURL error: "+err.Error())
        }
        pageHeader.UserURL = url
        pageHeader.UserLabel = "Login"
        pageHeader.IsAdmin = false

   } else {
        url, err := user.LogoutURL(c, r.URL.String())
        if err != nil {
            panic("user.LogoutURL error: "+err.Error())
        }
        pageHeader.UserURL = url
        pageHeader.UserLabel = "Logout"
        pageHeader.LoginMessage = "Hello, "+u.String()+"!"
        pageHeader.IsAdmin = user.IsAdmin(c)
        w.Header().Set("Pragma", "no-cache")
   }

   return pageHeader
}

func handlerHome(w http.ResponseWriter, r *http.Request) {

    pageHeader := createPageHeader("Blog Homepage",w,r)

    blogList := make([]BlogListEntry,1)
    blogList[0] = BlogListEntry{"just-started","title","Descr"}

    data := BlogListPage{pageHeader,blogList}

    err := pageTemplate.Execute(w, data)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

const htmlBlogEntry = htmlHeader+`
   <h1>{{.Title}}</h1>
   {{.Body}}
`+htmlFooter;

type BlogEntryPage struct {
   Header *PageHeader 
   Id string
   Title string
   Body template.HTML
}


var pageEntryTemplate = template.Must(template.New("blogEntryPage").Parse(htmlBlogEntry))

func handlerEntry(w http.ResponseWriter, r *http.Request) {

    pageHeader := createPageHeader("Blog Entry",w,r)


    data := BlogEntryPage{pageHeader,"id","title",template.HTML("<p>descr</p>")}

    err := pageEntryTemplate.Execute(w, data)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}


func init() {
    http.HandleFunc("/", handlerHome)
    http.HandleFunc("/entry/", handlerEntry)
}


