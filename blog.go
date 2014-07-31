package blog

import (
    "appengine"
    "appengine/user"
    "html/template"
    "net/http"
    "time"
)

const htmlHeader =`<html>
 <head>
  <title>{{.Header.Title}}</title>
  <style>
     label, input {
          display: block;
     }
     .Errors {
         color: red;
     }     
  </style>
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

    blogDBList,err := DBgetList(c)
    if err != nil {
          http.Error(w, err.Error(), http.StatusInternalServerError)
          return;
    }

   blogList := make([]BlogListEntry,len(blogDBList))
    for i := range blogDBList {
      blogList[i] = BlogListEntry{ blogDBList[i].Id,
                       blogDBList[i].Title,
                       blogDBList[i].Descr,
                    }  
    }

    data := BlogListPage{pageHeader,blogList}

    err = pageTemplate.Execute(w, data)
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
   Errors []error
   FormBody template.HTML
}

const htmlBlogNewPage = htmlHeader+`
  <ul class="Errors">
      {{range .Errors}}
      <li>{{.}}</li>
      {{end}}
  </ul>
  <form action="/admin/new" method="POST">
     {{.FormBody}}
  </form>
`+htmlFooter;

var pageBlogNewTemplate = template.Must(template.New("blogNewPage").Parse(htmlBlogNewPage))


func handlerAdminNew(w http.ResponseWriter, r *http.Request) {

    c := appengine.NewContext(r)
    pageHeader := createPageHeader("New Blog Entry",w,r)

    items := []FormItem{
         {"hidden","Action","","Store",false},
         {"text","Id","Slug (Id)",r.FormValue("Id"),true},
         {"text","Title","Title",r.FormValue("Title"),true},
         {"textarea","Descr","Description",r.FormValue("Descr"),true},
         {"textarea","Body","Body",r.FormValue("Body"),true},
         {"submit","Submit","","Save",false},
    }

    errors := make([]error,0)

    if r.FormValue("Action") == "Store" {

       errors = FormValidate(r,items)

       if len(errors) == 0 {

	       id := r.FormValue("Id")
	       b := DBBlogEntry{ Id: id,
			  Title: r.FormValue("Title"),
			  Descr: r.FormValue("Descr"),
			  Body:  r.FormValue("Body"),
			  Published: time.Now(),
			  LastModified: time.Now(),
		    }
	      err := DBstoreDBBlogEntry(c,&b)
	      if err != nil {
                   errors = append(errors,err)
	      } else {
		 http.Redirect(w, r, "/", http.StatusFound)
		 return
	      }
      }
    }
    formHtml,err := FormRenderItems( items)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    data := BlogNewPage{
                 Header:pageHeader,
                 Errors:errors,
                 FormBody:formHtml,
    }

    err = pageBlogNewTemplate.Execute(w, data)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}



func init() {
    http.HandleFunc("/", handlerHome)
    http.HandleFunc("/entry/", handlerEntry)
    http.HandleFunc("/admin/new", handlerAdminNew)
}


