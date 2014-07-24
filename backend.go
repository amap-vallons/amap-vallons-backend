package main

import "fmt"
import "net/http"
import "github.com/go-martini/martini"
import "github.com/martini-contrib/cors"
import "github.com/codegangsta/martini-contrib/binding"
import "labix.org/v2/mgo"
import "labix.org/v2/mgo/bson"


type User struct {
    ID bson.ObjectId `bson:"_id,omitempty"`
    username string `bson:"username"`
    fullname string `bson:"password"`
    email string
    password string
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
    m.Post("/login", binding.Form(Login{}), func(login Login, res http.ResponseWriter) string {
        result := User{}
        fmt.Println("Before db request " + login.Username)
        err = c.Find(bson.M{}).One(&result)
        fmt.Println("After db request " + result.ID + "***")
        if err != nil {
            panic(err)
        }
        if result.password != login.Password {
            res.WriteHeader(404);
        }
        return "";
    })
    m.Get("permanences", func() string {
        return "[{\"2014\":{\"Juillet\":{\"24\":\"Libre\", \"31\":\"Libre\"}}}]"
    })
    m.Run()
}
