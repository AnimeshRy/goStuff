package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/thedevsaddam/renderer"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var rnd *renderer.Render
var db *mgo.Database

const (
	hostname		string = "localhost:27017"
	dbName			string = "demo_todo"
	collectionName	string =  "todo"
	port			string = ":9000"
)

type (
	todoModel struct {
		ID        bson.ObjectId `bson:"_id,omitempty"`
		Title     string        `bson:"title"`
		Completed bool          `bson:"completed"`
		CreatedAt time.Time     `bson:"createAt"`
	}

	todo struct {
		ID        string    `json:"id"`
		Title     string    `json:"title"`
		Completed bool      `json:"completed"`
		CreatedAt time.Time `json:"created_at"`
	}
)

func init(){
	rnd = renderer.New()
	sess, err := mgo.Dial(hostname)
	checkErr(err)
	sess.SetMode(mgo.Monotonic, true) // Changes the consistently mode for the session (strong -> more gaurantee less load distribution, eventual -> few guarantee more distribution, monotonic -> the in between one)
	db = sess.DB(dbName)
}
