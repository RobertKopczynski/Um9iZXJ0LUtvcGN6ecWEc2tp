package main

import (
    "net/http"
    "encoding/json"
    "github.com/go-chi/chi"
    "strconv"
    "io/ioutil"
    "github.com/RobertKopczynski/Um9iZXJ0LUtvcGN6ecWEc2tp/db"
    "time"
)

var MaxFileSize int64 = 1024*1024
var timeout = time.Duration(5 * time.Second)
var smallPool chan func()

func main() {
    db.InitDB()
    smallPool = make(chan func(), 20)
    r := chi.NewRouter()
    r.Get("/", func(w http.ResponseWriter, r *http.Request) {
	    w.Write([]byte("welcome"))
    })
    r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
	    w.Write([]byte("pong"))
    })
    r.Route("/api/fetcher", func(r chi.Router){
        r.Get("/",UrlList)
        r.Post("/",UrlCreate)
        r.Delete("/{id}",UrlDelete)
        r.Get("/{id}/history",UrlHistory)
    })
    http.ListenAndServe(":8080", r)
}

func UrlList(w http.ResponseWriter, r *http.Request){
    data:=db.SelectAllUrls()
    response,_:=json.MarshalIndent(data,"","    ")
    w.Header().Set("Content-Type","application/json")
    w.WriteHeader(200)
    w.Write(response)
}
type returnId struct{
    Id int `json: "id"`
}

func UrlCreate(w http.ResponseWriter, r *http.Request){
    var data db.Url
    if r.ContentLength > MaxFileSize {
        http.Error(w, "Request Entity Too Large", http.StatusExpectationFailed)
        return
    }
    r.Body = http.MaxBytesReader(w, r.Body, MaxFileSize)
    err := json.NewDecoder(r.Body).Decode(&data)
    if err != nil{
        http.Error(w,err.Error(),400)
        return
    }
    var id int
    if data.Id==0 {
        id,_=db.InsertUrl(data,true)
        data.Id=id
    } else {
        id,_=db.InsertUrl(data,false)
    }
    go func(){
                client := http.Client{
                    Timeout: timeout,
                }
                for {
                    start:=time.Now()
                    resp,_:=client.Get(data.Url)
                    body,_:=ioutil.ReadAll(resp.Body)
                    duration:=time.Since(start).Seconds()
                    resp.Body.Close()
                    db.InsertResponse(db.Response{string(body),duration,""},data.Id)
                    time.Sleep(time.Duration(data.Interval)*time.Second)
                }
    }()
    response,_:=json.MarshalIndent(returnId{id},"","    ")
    w.Header().Set("Content-Type","application/json")
    w.WriteHeader(200)
    w.Write(response)
}
func fetchId(w http.ResponseWriter,r *http.Request) (int,error){
     id,err:= strconv.Atoi(chi.URLParam(r, "id"))
     if err!=nil{
        return 0,err
     }else{
        return id,nil
     }
}

func UrlDelete(w http.ResponseWriter, r *http.Request){
    id,err:= fetchId(w,r)
    if err!=nil{
        http.Error(w,"Bad Request",400)
    } else{
        db.DeleteUrl(id)
    }
}
func UrlHistory(w http.ResponseWriter, r *http.Request){
    id,err:= fetchId(w,r)
    if err!=nil{
        http.Error(w,"Bad Request",400)
    }else{
        data:=db.SelectHistory(id)
        if len(data)>0{
            response,_:=json.MarshalIndent(data,"","    ")
            w.Header().Set("Content-Type","application/json")
            w.WriteHeader(200)
            w.Write(response)
        }else{
        http.Error(w,"Not Found",404)
        }
    }
}

