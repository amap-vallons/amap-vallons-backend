package main

import "fmt"
import "net/http"
import "github.com/go-martini/martini"
import "github.com/martini-contrib/cors"
//import "github.com/martini-contrib/sessions"
import "github.com/codegangsta/martini-contrib/binding"
import "labix.org/v2/mgo"
import "labix.org/v2/mgo/bson"
//import "encoding/json"


type User struct {
    ID bson.ObjectId `bson:"_id,omitempty"`
    Username string
    Password string
    Fullname string
    Email string
}
type Login struct {
    Username    string `form:"username" binding:"required"`
    Password   string `form:"password"`
    unexported  string `form:"-"`
}
func main() {

    /* DB connection */
    session, err := mgo.Dial("mongodb://amap:rochetoirin@ds043168.mongolab.com:43168/amap-vallons")
    if err != nil {
        panic(err)
    }
    defer session.Close()

    c := session.DB("").C("amap.users")

    /* Web Framework */
    m := martini.Classic()
    m.Use(cors.Allow(&cors.Options{
        AllowOrigins: []string{"*"},
    }))
    m.Get("/", func() string {
        return "Hello from heroku"
    })
    m.Put("/user", binding.Form(User{}), func(user User) string {
        err := c.Insert(user)
        if err != nil {
            panic(err)
        }
        return ""
    })
    m.Post("/login", binding.Form(Login{}), func(login Login, res http.ResponseWriter) User {
        user := User{}
        var query = c.Find(bson.M{"username": login.Username})
        count,_ := query.Count()
        if count == 0 {
            fmt.Errorf("Record not found")
            res.WriteHeader(404)
        } else {
            err = query.One(&user)
            if err != nil {
                panic(err)
            }
            if user.Password != login.Password {
                res.WriteHeader(404)
            }
        }
        return user
    })
    m.Get("permanences", func() string {
        return "[{\"2014\":{\"Juillet\":{\"24\":\"Libre\", \"31\":\"Libre\"}}}]"
    })
    m.Run()
}
