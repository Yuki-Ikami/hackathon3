package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/oklog/ulid"
	_ "github.com/oklog/ulid"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	_ "strings"
	"syscall"
	"time"
)

type User struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type User2 struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Channel struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ChannelUser struct {
	Id        string `json:"id"`
	ChannelId string `json:"channel_id"`
	UserId    string `json:"user_id"`
}

type Message struct {
	Id        string `json:"id"`
	ChannelId string `json:"channel_id"`
	UserId    string `json:"user_id"`
	Content   string `json:"content"`
}

type postMessage struct {
	Content string `json:"content"`
}

type deleteId struct {
	Id string `json:"id"`
}

type Edit struct {
	Id      string `json:"id"`
	Message string `json:"message"`
}

var db *sql.DB

func init() {
	// DB接続のための準備
	mysqlUser := os.Getenv("MYSQL_USER")
	mysqlPwd := os.Getenv("MYSQL_PWD")
	mysqlHost := os.Getenv("MYSQL_HOST")
	mysqlDatabase := os.Getenv("MYSQL_DATABASE")

	connStr := fmt.Sprintf("%s:%s@%s/%s", mysqlUser, mysqlPwd, mysqlHost, mysqlDatabase)
	_db, err := sql.Open("mysql", connStr)
	if err != nil {
		log.Fatalf("fail: sql.Open, %v\n", err)
	}
	// ①-3
	if err := _db.Ping(); err != nil {
		log.Fatalf("fail: _db.Ping, %v\n", err)
	}
	db = _db
}

func main() {
	//新規登録でuserテーブルに追加
	http.HandleFunc("/register", addUserHandler)
	//mypage上でchannelのnameとidを受け取る
	http.HandleFunc("/mypage", getMypageHandler)
	http.HandleFunc("/message", getMessage)
	//Get:channel_idからmessageを表示
	//Post:channle_idからmessageを追加
	http.HandleFunc("/channel", messageHandler)
	//messageを編集
	http.HandleFunc("/edit", editMessageHnadler)
	//messageを削除
	http.HandleFunc("/delete", deleteMessageHandler)
	//channelを作る
	http.HandleFunc("/makeChannel", makeChannelHandler)
	//channelに参加
	http.HandleFunc("/joinChannel", joinChannelHandler)

	closeDBWithSysCall()

	// 8080番ポートでリクエストを待ち受ける
	log.Println("Listening...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func closeDBWithSysCall() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-sig
		log.Printf("received syscall, %v", s)

		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
		log.Printf("success: db.Close()")
		os.Exit(0)
	}()
}

func addUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	switch r.Method {
	case http.MethodPost:
		var user User2
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			fmt.Println(err)
			return
		}
		log.Printf("%v, %v\n", user.Name, user.Email)
		t := time.Now()
		entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
		id := ulid.MustNew(ulid.Timestamp(t), entropy)
		tx, er := db.Begin()
		if er != nil {
			log.Fatal(er)
		}
		_, err := db.Query("INSERT INTO user (id, name, email, password) VALUES (?, ?, ?, ?);", id.String(), user.Name, user.Email, user.Password)
		if err != nil {
			log.Println("insert error")
			w.WriteHeader(http.StatusInternalServerError)
			tx.Rollback()
			return
		}
		tx.Commit()
		log.Printf("id: %v\n", id.String())
		s := User{
			Id: id.String(),
		}
		ans, err := json.Marshal(s)
		if err != nil {
			log.Printf("fail: json.Marshal, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(ans)
		return
	default:
		log.Printf("fail: HTTP Method is %s\n", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func getMypageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	email := r.URL.Query().Get("email")
	log.Println("%v\n", email)
	rows, err := db.Query("SELECT id, name, email FROM user WHERE email = ?", email)
	if err != nil {
		log.Printf("fail: db.Query, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var u User
	for rows.Next() {
		if err := rows.Scan(&u.Id, &u.Name, &u.Email); err != nil {
			log.Printf("fail: rows.Scan, %v\n", err)

			if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
				log.Printf("fail: rows.Close(), %v\n", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		break
	}
	log.Printf("%v", u.Id)
	rows2, err := db.Query("SELECT id, channel_id FROM channnel_user WHERE user_id = ?", u.Id)
	if err != nil {
		log.Printf("fail: db.Query, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	channels := make([]Channel, 0)
	var c Channel
	for rows2.Next() {
		var cu ChannelUser
		if err := rows2.Scan(&cu.Id, &cu.ChannelId); err != nil {
			log.Printf("fail: rows2.Scan, %v\n", err)

			if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
				log.Printf("fail: rows.Close(), %v\n", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		rows3, err := db.Query("SELECT id, name, description FROM channel WHERE id = ?", cu.ChannelId)
		if err != nil {
			log.Printf("fail: db.Query, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		for rows3.Next() {
			if err := rows3.Scan(&c.Id, &c.Name, &c.Description); err != nil {
				log.Printf("fail: rows3.Scan, %v\n", err)
				if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
					log.Printf("fail: rows.Close(), %v\n", err)
				}
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			channels = append(channels, c)
		}
	}

	bytes, err := json.Marshal(channels)
	if err != nil {
		log.Printf("fail: json.Marshal, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
	//log.Printf("end getMypage\n")
	//log.Printf("fail: rows3.Scan, %v, %v, %v\n", c.Id, c.Name, c.Description)

}

func getMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	channel_id := r.URL.Query().Get("channelId")
	rows, err := db.Query("SELECT id, channel_id, user_id, content FROM message WHERE channel_id= ?", channel_id)
	if err != nil {
		log.Printf("fail: db.Query, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var m Message
	messages := make([]Message, 0)
	for rows.Next() {
		if err := rows.Scan(&m.Id, &m.ChannelId, &m.UserId, &m.Content); err != nil {
			log.Printf("fail: rows.Scan, %v\n", err)
			if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
				log.Printf("fail: rows.Close(), %v\n", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		//log.Printf("%v, %v, %v, %v", m.Id, m.ChannelId, m.UserId, m.Content)
		messages = append(messages, m)
	}
	bytes, err := json.Marshal(messages)
	if err != nil {
		log.Printf("fail: json.Marshal, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
	//log.Printf("end /message")
	return
}

func messageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	channelId := r.URL.Query().Get("channelId")
	switch r.Method {
	case http.MethodGet:
		rows, err := db.Query("SELECT id, channel_id, user_id, content FROM message WHERE channel_id = ?", channelId)
		if err != nil {
			log.Printf("fail: db.Query, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		messages := make([]Message, 0)
		for rows.Next() {
			var m Message
			if err := rows.Scan(&m.Id, &m.ChannelId, &m.UserId, &m.Content); err != nil {
				log.Printf("fail: rows.Scan, %v\n", err)

				if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
					log.Printf("fail: rows.Close(), %v\n", err)
				}
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			messages = append(messages, m)
		}
		bytes, err := json.Marshal(messages)
		if err != nil {
			log.Printf("fail: json.Marshal, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	case http.MethodPost:
		email := r.URL.Query().Get("email")
		rows, err := db.Query("SELECT id, name, email FROM user WHERE email = ?", email)
		if err != nil {
			log.Printf("fail: db.Query, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var u User
		for rows.Next() {
			if err := rows.Scan(&u.Id, &u.Name, &u.Email); err != nil {
				log.Printf("fail: rows.Scan, %v\n", err)

				if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
					log.Printf("fail: rows.Close(), %v\n", err)
				}
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			break
		}
		var content postMessage
		if err := json.NewDecoder(r.Body).Decode(&content); err != nil {
			fmt.Println(err)
			return
		}
		t := time.Now()
		entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
		id := ulid.MustNew(ulid.Timestamp(t), entropy)
		tx, er := db.Begin()
		if er != nil {
			log.Fatal(er)
		}
		_, err = db.Query("INSERT INTO message (id, channel_id, user_id, content) VALUES (?, ?, ?, ?)", id.String(), channelId, u.Id, content.Content)
		if err != nil {
			log.Printf("fail:insert db.Query, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		tx.Commit()
	default:
		log.Printf("fail: HTTP Method is %s\n", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func editMessageHnadler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	var e Edit
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		fmt.Println(err)
		return
	}
	tx, er := db.Begin()
	if er != nil {
		log.Fatal(er)
	}
	message := "(編集済み)"
	addMessage := e.Message + message
	log.Printf("%v\n", addMessage)
	_, err := db.Query("UPDATE message SET content = ? WHERE id = ?", addMessage, e.Id)
	if err != nil {
		log.Println("insert error")
		w.WriteHeader(http.StatusInternalServerError)
		tx.Rollback()
		return
	}
	tx.Commit()
}

func deleteMessageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	var d deleteId
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		fmt.Println(err)
		return
	}
	tx, er := db.Begin()
	if er != nil {
		log.Fatal(er)
	}
	_, err := db.Query("DELETE FROM message WHERE id = ?", d.Id)
	if err != nil {
		log.Println("insert error")
		w.WriteHeader(http.StatusInternalServerError)
		tx.Rollback()
		return
	}
	tx.Commit()
}

func makeChannelHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	email := r.URL.Query().Get("email")
	rows, err := db.Query("SELECT id, name, email FROM user WHERE email = ?", email)
	if err != nil {
		log.Printf("fail: db.Query, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var u User
	for rows.Next() {
		if err := rows.Scan(&u.Id, &u.Name, &u.Email); err != nil {
			log.Printf("fail: rows.Scan, %v\n", err)

			if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
				log.Printf("fail: rows.Close(), %v\n", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		break
	}
	var c Channel
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		fmt.Println(err)
		return
	}
	t := time.Now()
	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
	id := ulid.MustNew(ulid.Timestamp(t), entropy)
	tx, er := db.Begin()
	if er != nil {
		log.Fatal(er)
	}
	_, err = db.Query("INSERT INTO channel (id, name, description) VALUES (?, ?, ?)", id.String(), c.Name, c.Description)
	if err != nil {
		log.Printf("fail:insert1 db.Query, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tx.Commit()
	t = time.Now()
	entropy = ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
	id2 := ulid.MustNew(ulid.Timestamp(t), entropy)
	tx, er = db.Begin()
	if er != nil {
		log.Fatal(er)
	}
	_, err = db.Query("INSERT INTO channnel_user (id, channel_id, user_id) VALUES (?, ?, ?)", id2.String(), id.String(), u.Id)
	if err != nil {
		log.Printf("fail:insert2 db.Query, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tx.Commit()
}

func joinChannelHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	email := r.URL.Query().Get("email")
	rows, err := db.Query("SELECT id, name, email FROM user WHERE email = ?", email)
	if err != nil {
		log.Printf("fail: db.Query, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var u User
	for rows.Next() {
		if err := rows.Scan(&u.Id, &u.Name, &u.Email); err != nil {
			log.Printf("fail: rows.Scan, %v\n", err)

			if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
				log.Printf("fail: rows.Close(), %v\n", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		break
	}
	var cu ChannelUser
	if err := json.NewDecoder(r.Body).Decode(&cu); err != nil {
		fmt.Println(err)
		return
	}
	t := time.Now()
	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
	id := ulid.MustNew(ulid.Timestamp(t), entropy)
	tx, er := db.Begin()
	if er != nil {
		log.Fatal(er)
	}
	_, err = db.Query("INSERT INTO channnel_user (id, channel_id, user_id) VALUES (?, ?, ?)", id.String(), cu.ChannelId, u.Id)
	if err != nil {
		log.Printf("fail:insert1 db.Query, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tx.Commit()
}
