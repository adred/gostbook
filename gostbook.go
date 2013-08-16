package main

import (
  "html/template"
  "net/http"
  "labix.org/v2/mgo/bson"
  "time"
)

type Entry struct {
  ID        bson.ObjectId `bson:"_id,omitempty"`
  Timestamp time.Time
  Name      string
  Message   string
}

func NewEntry() *Entry {
  return &Entry{
      Timestamp: time.Now(),
  }
}

func hello(w http.ResponseWriter, req *http.Request) {
  //grab a clone of the session and close it when the
  //function returns
  s := session.Clone()
  defer s.Close()

  //set up the collection and query
  coll := s.DB("gostbook").C("entries")
  query := coll.Find(nil).Sort("-timestamp")

  //execute the query
  //TODO: add pagination :)
  var entries []Entry
  if err := query.All(&entries); err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
  }

  //execute the template
  if err := index.Execute(w, entries); err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
  }
}

func sign(w http.ResponseWriter, req *http.Request) {
  //make sure we got post
  if req.Method != "POST" {
      http.NotFound(w, req)
      return
  }

  entry := NewEntry()
  entry.Name = req.FormValue("name")
  entry.Message = req.FormValue("message")

  if entry.Name == "" {
      entry.Name = "Some dummy who forgot a name"
  }
  if entry.Message == "" {
      entry.Message = "Some dummy who forgot a message."
  }

  s := session.Clone()
  defer s.Close()

  coll := s.DB("gostbook").C("entries")
  if err := coll.Insert(entry); err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
  }

  http.Redirect(w, req, "/", http.StatusTemporaryRedirect)
}

var index = template.Must(template.ParseFiles(
  "templates/_base.html",
  "templates/index.html",
))

var session *mgo.Session

func main() {
  var err error
  session, err = mgo.Dial("localhost")
  if err != nil {
      panic(err)
  }

  http.HandleFunc("/", hello)
  http.HandleFunc("/sign", sign)

  if err = http.ListenAndServe(":8080", nil); err != nil {
      panic(err)
  }
}