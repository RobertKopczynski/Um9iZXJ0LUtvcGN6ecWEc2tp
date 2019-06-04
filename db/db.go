package db

import (
 "database/sql"
 "fmt"
 _ "github.com/mattn/go-sqlite3"
)
var database,_ = sql.Open("sqlite3", "/tmp/test.db")

type Url struct {
    Id int
    Url string
    Interval int
}

func InitDB() {
    statement, _ := database.Prepare(
        "CREATE TABLE IF NOT EXISTS url" +
        " (url_id INTEGER PRIMARY KEY AUTOINCREMENT, url TEXT, interval INTEGER)")
    statement.Exec()
    statement, _ = database.Prepare(
        "CREATE TABLE IF NOT EXISTS response"+
        " (url_id INTEGER,response TEXT, duration REAL, created_at TEXT)")
    statement.Exec()
}

func InsertUrl(data Url, withoutId bool) (int,error) {
    if withoutId {
        res,_ := database.Exec("INSERT INTO url (url,interval) VALUES (?,?)",data.Url,data.Interval)
        id,err :=res.LastInsertId()
        if err!=nil{
            fmt.Println("Error inserting url withouth id: ",err.Error())
            return -1,err
        } else{
            return int(id),nil
        }
    } else {
        e:= database.QueryRow("SELECT url_id FROM url WHERE url_id = ? ",data.Id)
        if e != nil {
            _,err:=database.Exec("INSERT INTO url (url_id,url,interval) VALUES (?,?,?)",
                                data.Id,data.Url,data.Interval)
            if err != nil{
                fmt.Println("Error inserting url with id: ",err.Error())
                return -1,err
            }
            return data.Id,nil
        }else{
            _,err:=database.Exec("UPDATE url SET url = ?, interval = ? WHERE url_id=?",
                                data.Url,data.Interval,data.Id)
            if err != nil{
                fmt.Println("Error updating id: ",err.Error())
                return -1,err
            }
            return data.Id,nil
        }
    }
}




func SelectAllUrls() []Url{
    var queue []Url
    rows, _ :=database.Query("SELECT * FROM url")
    var id int
    var url string
    var interval int
    for rows.Next(){
        if err:=rows.Scan(&id,&url,&interval); err!=nil{
                fmt.Println("problem fetching Urls", err)
        }
        u:=Url{id,url,interval}
        queue = append(queue,u)
    }
    return queue
}

func DeleteUrl(url_id int) {
    database.Exec("DELETE FROM url WHERE url_id = ?",url_id)
    database.Exec("DELETE FROM response WHERE url_id = ?",url_id)
}

type Response struct {
      Response string
      Duration float64
      Created_at string
}

func SelectHistory(url_id int) []Response{
    var queue []Response
    rows,_:=database.Query("SELECT response, duration, created_at FROM response "+
                           "WHERE url_id = ?",url_id)
    var response string
    var duration float64
    var created_at string
    for rows.Next(){
        if err:=rows.Scan(&response,&duration,&created_at); err!=nil{
            fmt.Println("Problem fetching response for id:",url_id,err)
        }
        r:=Response{response,duration,created_at}
        queue = append(queue,r)
    }
    return queue
}
func InsertResponse(data Response, url_id int){
    database.Exec("INSERT INTO response (url_id,response,duration,created_at)"+
                         "VALUES (?,?,?,CURRENT_TIMESTAMP)",url_id,data.Response,data.Duration)
}
