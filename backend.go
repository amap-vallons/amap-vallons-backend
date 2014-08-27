package main

import "fmt"
import "time"
import "strconv"
import "net/http"
import (
    "github.com/go-martini/martini"
    "github.com/martini-contrib/cors"
    "github.com/martini-contrib/sessions"
    "github.com/codegangsta/martini-contrib/binding"
    "github.com/codegangsta/martini-contrib/render"
    )
import (
    "labix.org/v2/mgo"
    "labix.org/v2/mgo/bson"
    )


type User struct {
    ID bson.ObjectId `bson:"_id,omitempty" json:"id"`
    Username string `json:"username"`
    Password string `json:"password"`
    Fullname string `json:"fullname"`
    Email string `json:"email"`
}
type Login struct {
    Username    string `form:"username" binding:"required"`
    Password   string `form:"password"`
    unexported  string `form:"-"`
}
type File struct {
    ID bson.ObjectId `bson:"_id,omitempty"`
    Filename    string
    Size        int
    Content []byte
}
type Date struct {
    ID bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Date time.Time `json:"date"`
	User interface{} `bson:"user,omitempty" json:"user"`
}
type DateUpdate struct {
    Date Date
}
type Dates []Date

func main() {

    var offline = false
    var dbUsers *mgo.Collection = nil
    var dbFiles *mgo.Collection = nil
    var dbDates *mgo.Collection = nil
    /* DB connection */
    DBsession, err := mgo.Dial("mongodb://amap:rochetoirin@ds043168.mongolab.com:43168/amap-vallons")
    if err != nil {
        fmt.Println("offline")
        offline = true
    }
    defer DBsession.Close()

    if ! offline {
        dbUsers = DBsession.DB("").C("amap.users")
        dbFiles = DBsession.DB("").C("amap.files")
        dbDates = DBsession.DB("").C("amap.dates")
    } else {
        dbUsers = nil
        dbFiles = nil
        dbDates = nil
    }

    /* Web Framework */
    m := martini.Classic()
    m.Use(cors.Allow(&cors.Options{
        AllowOrigins: []string{"*"},
        AllowHeaders: []string{"x-request-with", "x-request-by", "Content-Type"},
        AllowMethods: []string{"POST", "GET", "PUT", "OPTIONS", "DELETE"},
        AllowCredentials: true,
        MaxAge: time.Duration(604800) * time.Second,
    }))
    store := sessions.NewCookieStore([]byte("secret123"))
    m.Use(sessions.Sessions("user_session", store))
    m.Use(render.Renderer())
    m.Get("/users/:id", func(session sessions.Session, params martini.Params, r render.Render) {
        user := User{}
        switch params["id"] {
            case "loggedin":
                v := session.Get("user")
                if v == nil {
                    r.Error(404)
                    return
                }

                if ! offline {
                    var query = dbUsers.Find(bson.M{"username": v.(string)})
                    count,_ := query.Count()
                    if count == 0 {
                        fmt.Errorf("Record not found")
                        r.Error(404)
                        return
                    } else {
                        err = query.One(&user)
                        if err != nil {
                            panic(err)
                        }
                    }
                } else {
                    r.Error(404)
                    return
                }
            default:
                v := session.Get("user")
                if v == nil {
                    r.Error(404)
                    return
                }

                if ! offline {
                    var query = dbUsers.Find(bson.M{"_id": params["id"]})
                    count,_ := query.Count()
                    if count == 0 {
                        fmt.Errorf("Record not found")
                        r.Error(404)
                        return
                    } else {
                        err = query.One(&user)
                        if err != nil {
                            panic(err)
                        }
                    }
                } else {
                    r.Error(404)
                    return
                }
        }
        r.JSON(200, map[string]interface{}{"user": user})
    })

    m.Options("/login", func(r render.Render) string {
        fmt.Errorf("Options")
        return "Salut"
    })

    m.Put("/user", binding.Form(User{}), func(user User) string {
        err := dbUsers.Insert(user)
        if err != nil {
            panic(err)
        }
        return ""
    })

    m.Get("/users", func(r render.Render) {
        var users []User
        var query = dbUsers.Find(nil)
        count,_ := query.Count()
        if count == 0 {
            r.JSON(200, map[string]interface{}{ "users" : nil })
        } else {
            err = query.All(&users)
            if err != nil {
                panic(err)
            }
            r.JSON(200, map[string]interface{}{ "users" : users })
        }
    })


    m.Delete("/login", binding.Form(Login{}), func(login Login, session sessions.Session, w http.ResponseWriter) string {
            session.Delete("user")
            w.WriteHeader(200)
            return ""
    })

    /* A new verb GET /logout is created as cross-domain does not work with DELETE */
    m.Get("/logout", binding.Form(Login{}), func(login Login, session sessions.Session, w http.ResponseWriter) string {
            session.Delete("user")
            w.WriteHeader(200)
            return ""
    })

    m.Get("/files/:filename", func(w http.ResponseWriter, params martini.Params, session sessions.Session) []byte {
        v := session.Get("user")
        if v == nil {
            w.WriteHeader(404)
            return []byte("")
        }

        file := File{}
        var query = dbFiles.Find(bson.M{"filename": params["filename"]})
        count,_ := query.Count()
        if count == 0 {
            fmt.Errorf("Record not found")
            w.WriteHeader(http.StatusNotFound)
        } else {
            err = query.One(&file)
            if err != nil {
                panic(err)
            }
            w.Header().Add("Content-Disposition", "attachment; filename=" + params["filename"])
            w.Header().Add("Set-Cookie", "fileDownload=true; path=/")
            return file.Content
        }
        return []byte("")
    })

    m.Post("/files/:filename", func(w http.ResponseWriter, req *http.Request, params martini.Params, session sessions.Session) string {
        v := session.Get("user")
        if v == nil {
            w.WriteHeader(404)
            fmt.Errorf("Not authenticated")
            return ""
        }

        file := File{}
        fileContent, _, err := req.FormFile("file")
        if err != nil {
            fmt.Fprintln(w, err)
            w.WriteHeader(400)
            return ""
        }
        size, err := strconv.Atoi(req.Header["Content-Length"][0])
        if err != nil {
            fmt.Fprintln(w, err)
            w.WriteHeader(400)
            return ""
        }
        fmt.Printf("Size : %d", size)
        var query = dbFiles.Find(bson.M{"filename": params["filename"]})
        count,_ := query.Count()
        if (req.Body != nil) {
            if count == 0 {
                file.Filename = params["filename"]
                file.Content = make([]byte, size)
                file.Size, err = fileContent.Read(file.Content)
                if err != nil {
                    fmt.Fprintln(w, err)
                    w.WriteHeader(400)
                    return ""
                }
                file.Content = file.Content[0:file.Size]
                dbFiles.Insert(file)
            } else {
                err = query.One(&file)
                if err != nil {
                    fmt.Fprintln(w, err)
                    w.WriteHeader(400)
                    return ""
                }
                file.Content = make([]byte, size)
                file.Size, err = fileContent.Read(file.Content)
                if err != nil {
                    fmt.Fprintln(w, err)
                    w.WriteHeader(400)
                    return ""
                }
                file.Content = file.Content[0:file.Size]
                dbFiles.Update(bson.M{"filename": params["filename"]}, file)
                fmt.Printf("Content size : %d\n", len(file.Content))
                w.WriteHeader(200)
            }
        } else {
            w.WriteHeader(406)
        }
        return ""
    })

    m.Post("/login", binding.Form(Login{}), func(login Login, session sessions.Session, r render.Render) {
        user := User{}
        var query = dbUsers.Find(bson.M{"username": login.Username})
        count,_ := query.Count()
        if count == 0 {
            fmt.Errorf("Record not found")
            r.Error(404)
        } else {
            err = query.One(&user)
            if err != nil {
                panic(err)
            }
            if user.Password != login.Password {
                r.Error(404)
            } else {
                session.Set("user", user.Username)
                r.JSON(200, map[string]interface{}{"user": user})
            }
        }
    })

    m.Get("/dates", func(r render.Render, req *http.Request) {
        var dates Dates
        from,err := time.Parse("Mon Jan 2 2006 15:04:05 GMT-0700 (MST)", req.URL.Query().Get("from"))
        if err != nil {
            panic(err)
        }
        var query = dbDates.Find(bson.M{ "date" : bson.M{ "$gt" : from } })
        count,_ := query.Count()
        if count == 0 {
            r.JSON(200, map[string]interface{}{ "dates" : make(Dates, 0, 0) })
        } else {
            limit,_ := strconv.Atoi(req.URL.Query().Get("count"))
            if err != nil {
                panic(err)
            }
            err = query.Limit(limit).All(&dates)
            if err != nil {
                panic(err)
            }
            r.JSON(200, map[string]interface{}{ "dates" : dates })
        }
    })

    m.Put("/dates/:id", binding.Json(DateUpdate{}), func(w http.ResponseWriter, date DateUpdate, params martini.Params, session sessions.Session) {
        err := dbDates.UpdateId(bson.ObjectIdHex(params["id"]), bson.M{"$set": bson.M{ "user": date.Date.User}})
        if err != nil {
            fmt.Printf("%s\n", err.Error())
            w.WriteHeader(404)
            return
        }
        w.WriteHeader(200)
    })

    m.Post("/dates", binding.Json(DateUpdate{}), func(date DateUpdate, session sessions.Session, r render.Render) {
        var update Date
        err := dbDates.Find(bson.M{ "date": date.Date.Date }).One(&update)
        if err != nil {
            err := dbDates.Insert(date.Date)
            if err != nil {
                fmt.Errorf(err.Error())
                r.Error(404)
                return
            }
            r.JSON(200, map[string]interface{}{ "date" : date.Date })
        } else {
            err := dbDates.UpdateId(update.ID, bson.M{"$set": bson.M{ "user": date.Date.User}})
            if err != nil {
                fmt.Printf(err.Error())
                r.Error(404)
                return
            }
        }
    })

    m.Run()
}
