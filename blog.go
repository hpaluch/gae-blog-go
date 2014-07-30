package blog

import (
    "appengine"
    "appengine/datastore"
    "appengine/user"
    "html/template"
    "net/http"
    "time"
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
  {{if .Header.IsAdmin}}
  <div class="AdminMenu">
     <a href="/admin/new">New Blog Entry</a>
  </div>
  {{end}}
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

    c := appengine.NewContext(r)
    pageHeader := createPageHeader("Blog Homepage",w,r)

    q := datastore.NewQuery("BlogEntry").Order("-Published").Limit(10)
    blogDBList := make([]DBBlogEntry,0,10)
    if _, err := q.GetAll(c, &blogDBList); err != nil {
           http.Error(w, err.Error(), http.StatusInternalServerError)
           return
    }
    blogList := make([]BlogListEntry,len(blogDBList))
    for i := range blogDBList {
      blogList[i] = BlogListEntry{ blogDBList[i].Id,
                       blogDBList[i].Title,
                       blogDBList[i].Descr,
                    }  
    }

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


//////// ADMIN 

type BlogNewPage struct {
   Header *PageHeader 
   Error string
   Id string
   Title string
   Descr string
   Body string
}

type DBBlogEntry struct {
     Id string
     Title string
     Descr string
     Body  string 
     Published time.Time
     LastModified time.Time
}

const htmlBlogNewPage = htmlHeader+`
  <h2>{{.Error}}</h2>
  <form action="/admin/new" method="POST">
     <input type="hidden" name="Action" value="Store" />
     <input type="text" name="Id" value="{{.Id}}" />
     <input type="text" name="Title" value="{{.Title}}" />
     <textarea name="Descr">
         {{.Descr}}
     </textarea>
     <textarea name="Body">
         {{.Body}}
     </textarea>
     <input type="submit" name="Save"   value="Save2"/>
  </form>
`+htmlFooter;

var pageBlogNewTemplate = template.Must(template.New("blogNewPage").Parse(htmlBlogNewPage))


func handlerAdminNew(w http.ResponseWriter, r *http.Request) {

    c := appengine.NewContext(r)
    pageHeader := createPageHeader("New Blog Entry",w,r)

    errorStr := "NoError"
    if r.FormValue("Action") == "Store" {
       errorStr = "Saving data" 
       id := r.FormValue("Id")
       b := DBBlogEntry{ Id: id,
                  Title: r.FormValue("Title"),
                  Descr: r.FormValue("Descr"),
                  Body:  r.FormValue("Body"),
                  Published: time.Now(),
                  LastModified: time.Now(),
            }
      key := datastore.NewKey(c,"BlogEntry",id,0,nil)
      _, err := datastore.Put(c, key, &b)
      if err != nil {
           errorStr = err.Error()
      } else {
         http.Redirect(w, r, "/", http.StatusFound)
         return
      }

    } else {
       errorStr = "Clean form" 
    }

    data := BlogNewPage{Header:pageHeader,Error:errorStr}

    err := pageBlogNewTemplate.Execute(w, data)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}



func init() {
    http.HandleFunc("/", handlerHome)
    http.HandleFunc("/entry/", handlerEntry)
    http.HandleFunc("/admin/new", handlerAdminNew)
}


