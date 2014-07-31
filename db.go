package blog

import (
    "appengine"
    "appengine/datastore"
    "time"
)

type DBBlogEntry struct {
     Id string
     Title string
     Descr string
     Body  string 
     Published time.Time
     LastModified time.Time
}


func DBstoreDBBlogEntry(c appengine.Context,entity *DBBlogEntry) (err error) {

    entity.LastModified = time.Now()

     key := datastore.NewKey(c,"BlogEntry",entity.Id,0,nil)
     _, err = datastore.Put(c, key, entity)
     if err != nil {
           return err
     }
     return nil
}


func DBgetList (c appengine.Context) (blogDBList []DBBlogEntry, err error) {
    q := datastore.NewQuery("BlogEntry").Order("-Published").Limit(10)
    blogDBList = make([]DBBlogEntry,0,10)
    if _, err := q.GetAll(c, &blogDBList); err != nil {
           return nil,err
    }
    return blogDBList,nil;
}


