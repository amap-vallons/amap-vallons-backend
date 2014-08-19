package main

import "fmt"
import "time"
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
func main() {

    var offline = false
    var c *mgo.Collection = nil
    /* DB connection */
    DBsession, err := mgo.Dial("mongodb://amap:rochetoirin@ds043168.mongolab.com:43168/amap-vallons")
    if err != nil {
        fmt.Println("offline")
        offline = true
    }
    defer DBsession.Close()

    if ! offline {
        c = DBsession.DB("").C("amap.users")
    } else {
        c = nil
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
                    var query = c.Find(bson.M{"username": v.(string)})
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
                r.Error(404)
                return
        }
        r.JSON(200, map[string]interface{}{"user": user})
    })

    m.Options("/login", func(r render.Render) string {
        fmt.Errorf("Options")
        return "Salut"
    })

    m.Put("/user", binding.Form(User{}), func(user User) string {
        err := c.Insert(user)
        if err != nil {
            panic(err)
        }
        return ""
    })

    m.Delete("/login", binding.Form(Login{}), func(login Login, session sessions.Session, r render.Render) {
            session.Set("user", "{ \"user\": {}}")
    })

    /* A new verb GET /logout is created as cross-domain does not work with DELETE */
    m.Get("/logout", binding.Form(Login{}), func(login Login, session sessions.Session, r render.Render) {
            session.Set("user", "{ \"user\": {}}")
    })

    m.Post("/login", binding.Form(Login{}), func(login Login, session sessions.Session, r render.Render) {
        user := User{}
        var query = c.Find(bson.M{"username": login.Username})
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

    m.Get("/permanences", func() string {
        return "[{\"2014\":{\"Juillet\":{\"24\":\"Libre\", \"31\":\"Libre\"}}}]"
    })
    m.Run()
}
