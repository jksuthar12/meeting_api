package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)
type Pre struct {
	  Id string
      Title string
      Participants []Participant
      Starttime string
      Endtime string
      Create_time string
}
type Participant struct{
	Name string
	Email string
	RSPV string
}

func findid() int {
	var id int
	id=0
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	collection:= client.Database("Appointy").Collection("meetings")
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	cursor, err := collection.Find(context.TODO(),bson.D{})
	if err != nil {
		fmt.Println("Finding all documents ERROR:", err)
		defer cursor.Close(ctx)
	} else {
		for cursor.Next(ctx) {
			id++
		}
	}
	return id
}
func overlap(s Pre) bool {
	layout := "2006-01-02T15:04:05.000Z"
	var temp []Participant
	start, err := time.Parse(layout, s.Starttime)
	end, err1 := time.Parse(layout, s.Endtime)
	temp=s.Participants
	if err != nil {
		log.Fatal(err)
	}
	if err1 != nil {
		log.Fatal(err1)
	}
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	//var t []Pre
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	collection := client.Database("Appointy").Collection("meetings")
	cursor, err := collection.Find(context.TODO(), bson.D{})
	if err != nil {
		fmt.Println("Finding all documents ERROR:", err)
		defer cursor.Close(ctx)
	} else {
		for cursor.Next(ctx) {
			var result bson.M
			err := cursor.Decode(&result)
			if err != nil {
				fmt.Println("error", err)
				os.Exit(1)
			} else {
				var s Pre
				bsonBytes, _ := bson.Marshal(result)
				bson.Unmarshal(bsonBytes, &s)
				var pr []Participant
				pr = s.Participants
				layout := "2006-01-02T15:04:05.000Z"
				date1,_ := time.Parse(layout,s.Starttime)
				data2, _ := time.Parse(layout,s.Endtime)
				for i:=0;i<len(pr);i++{
					for j:=0;j<len(temp);j++{
						if pr[i].Email==temp[j].Email && pr[i].RSPV=="YES" && (start.Before(data2)&& date1.Before(end)) {
							return true
						}
					}
				}
			}
		}

	}
	return false
}
func check(s Pre,param1 string,param2 string )bool {
	if s.Starttime=="" || s.Endtime==""{ return false}
	layout := "2006-01-02T15:04:05.000Z"
	t,err := time.Parse(layout,param1)
	t1,err1 := time.Parse(layout,param2)
	t2,err2 :=time.Parse(layout,s.Starttime)
	t3,err3 := time.Parse(layout,s.Endtime)
	if err !=nil{ log.Fatal(err)}
	if err1 != nil {log.Fatal(err1)}
	if err2 != nil{log.Fatal(err2)}
	if err3!=nil {log.Fatal(err3)}
	g1 :=t.Before(t2)
	g2 := t1.After(t3)
	if g1==true && g2==true {
		return  true
	}
	if t==t2 && t1==t3 { return true
	}
	return false
}

func Holdmeeting(w http.ResponseWriter,req *http.Request)  {
	switch req.Method {
	case "POST":{var id int
		id=findid()
		decoder := json.NewDecoder(req.Body)
		var t Pre
		err := decoder.Decode(&t)
		if err!=nil{
			log.Fatal(err)
		}
		clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
		client, err := mongo.Connect(context.TODO(), clientOptions)
		if err != nil {
			log.Fatal(err)
		}
		err = client.Ping(context.TODO(), nil)
		if err != nil {
			log.Fatal(err)
		}
		t.Id=strconv.Itoa(id)
		t.Create_time=time.Now().Format("2006-01-02T15:04:05.000Z")
		if overlap(t) == false {
			collection := client.Database("Appointy").Collection("meetings")
			result, err := collection.InsertOne(context.TODO(), t)
			if err != nil {
				log.Fatal(err)
				log.Fatal(result)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			tt, err := json.Marshal(t)
			if err != nil {
				log.Fatal(err)
			}
			w.Write(tt)
		} else { w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Participant is busy"))}}
	   break
	case "GET":{   param1 := req.URL.Query().Get("start")
		param2 := req.URL.Query().Get("end")
		clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
		client, err := mongo.Connect(context.TODO(), clientOptions)
		if err != nil {
			log.Fatal(err)
		}
		err = client.Ping(context.TODO(), nil)
		if err != nil {
			log.Fatal(err)
		}
		var t []Pre
		ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
		collection:= client.Database("Appointy").Collection("meetings")
		cursor, err := collection.Find(context.TODO(),bson.D{})
		if err != nil {
			fmt.Println("error", err)
			defer cursor.Close(ctx)
		} else {
			for cursor.Next(ctx) {
				var result bson.M
				err := cursor.Decode(&result)
				if err != nil {
					fmt.Println("error", err)
					os.Exit(1)
				} else {
					var s Pre
					bsonBytes, _ := bson.Marshal(result)
					bson.Unmarshal(bsonBytes, &s)
					c :=check(s,param1,param2)
					if c==true{
						t =append(t, s)
					}
				}
			}
         if t!=nil{
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			t, err := json.Marshal(t)
			if err != nil {
				log.Fatal(err)
			}
			w.Write(t)
         } else{w.Write([]byte("no data found "))}
		}}
	 break
	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid request method"))
		break
	}

}
func Showmeeting(w http.ResponseWriter,req *http.Request)  {
	switch req.Method {
	case "GET":{
		var object string
		object =req.URL.Path[len("/meeting/"):]
	//	log.Println(object)
		clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
		client, err := mongo.Connect(context.TODO(), clientOptions)
		if err != nil {
			log.Fatal(err)
		}
		err = client.Ping(context.TODO(), nil)
		if err != nil {
			log.Fatal(err)
		}
		collection:= client.Database("Appointy").Collection("meetings")
		var result Pre
		err1 := collection.FindOne(context.TODO(), bson.M{"id":object}).Decode(&result)
		if err1 != nil {
			fmt.Println("Error calling FindOne():", err)
			os.Exit(1)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			tt, err := json.Marshal(result)
			if err != nil {
				log.Fatal(err)
			}
			w.Write(tt)

		}
	}
	break
	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid method request"))
		break

	}
}

func Allmeeting(w http.ResponseWriter,req *http.Request)  {
	switch req.Method {
	case "GET":{
		email := req.URL.Query().Get("participant")
		clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
		client, err := mongo.Connect(context.TODO(), clientOptions)
		if err != nil {
			log.Fatal(err)
		}
		err = client.Ping(context.TODO(), nil)
		if err != nil {
			log.Fatal(err)
		}
		var t []Pre
		ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
		collection:= client.Database("Appointy").Collection("meetings")
		cursor, err := collection.Find(context.TODO(),bson.D{})
		if err != nil {
			fmt.Println("Finding all documents ERROR:", err)
			defer cursor.Close(ctx)

		} else {

			for cursor.Next(ctx) {
				var result bson.M
				err := cursor.Decode(&result)
				if err != nil {
					fmt.Println("cursor.Next() error:", err)
					os.Exit(1)
				} else {
					var s Pre
					bsonBytes, _ := bson.Marshal(result)
					bson.Unmarshal(bsonBytes, &s)
					var pr []Participant
					pr = s.Participants
					for i:=0;i<len(pr);i++{
						if pr[i].Email==email { t= append(t,s)}
					}

				}
			}
			if t!=nil{
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				t, err := json.Marshal(t)
				if err != nil {
					log.Fatal(err)
				}
				w.Write(t)}else {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("no data found"))
			}
		}
	}
	break
	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid method request"))
		break

	}
}

func main() {
	http.HandleFunc("/meetings", Holdmeeting)
	http.HandleFunc("/meeting/",Showmeeting)
	http.HandleFunc("/articles",Allmeeting)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
