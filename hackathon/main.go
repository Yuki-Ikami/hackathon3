package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/oklog/ulid"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/oklog/ulid"
)

type User struct {
	Id    string `json:"id"`
	Name  string `json:"id"`
	Email string `json:"id"`
}

type Channel struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"name"`
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

var db *sql.DB

func init() {
	// ①-1
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
	//Get:channel_idからmessageを表示
	//Post:channle_idからmessageを追加
	http.HandleFunc("/channel", messageHandler)
	//messageを編集
	http.HandleFunc("/edit", editMessageHnadler)
	//messageを削除
	http.HandleFunc("/delete", deleteMessageHandler)
	closeDBWithSysCall()

	// 8000番ポートでリクエストを待ち受ける
	log.Println("Listening...")
	if err := http.ListenAndServe(":8000", nil); err != nil {
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
	switch r.Method {
	case http.MethodPost:
		var user User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
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
		_, err := db.Query("INSERT INTO user (id, name, email) VALUES (?, ?, ?);", id.String(), user.Name, user.Email)
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
	switch r.Method {
	case http.MethodGet:
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

		rows2, err := db.Query("SELECT id, channel_id FROM channel_members WHERE user_id = ?", u.Id)
		if err != nil {
			log.Printf("fail: db.Query, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		channels := make([]Channel, 0)
		for rows2.Next() {
			var cu ChannelUser
			if err := rows.Scan(&cu.Id, &cu.ChannelId); err != nil {
				log.Printf("fail: rows.Scan, %v\n", err)

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
			var c Channel
			for rows3.Next() {
				if err := rows.Scan(&c.Id, &c.Name, &c.Description); err != nil {
					log.Printf("fail: rows.Scan, %v\n", err)

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
	}
}

func messageHandler(w http.ResponseWriter, r *http.Request) {
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
		var content string
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
		_, err = db.Query("INSERT INTO message (id, channel_id, user_id, content) VALUES (?, ?, ?, ?)", id.String(), channelId, u.Id, content)
		if err != nil {
			log.Printf("fail: db.Query, %v\n", err)
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
	var m Message
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		fmt.Println(err)
		return
	}
	tx, er := db.Begin()
	if er != nil {
		log.Fatal(er)
	}
	_, err := db.Query("UPDATE message SET content = ? WHERE id = ?", m.Content+"(編集済み)", m.Id)
	if err != nil {
		log.Println("insert error")
		w.WriteHeader(http.StatusInternalServerError)
		tx.Rollback()
		return
	}
	tx.Commit()
}

func deleteMessageHandler(w http.ResponseWriter, r *http.Request) {
	var m Message
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		fmt.Println(err)
		return
	}
	tx, er := db.Begin()
	if er != nil {
		log.Fatal(er)
	}
	_, err := db.Query("DELETE FROM message WHERE id = ?", m.Id)
	if err != nil {
		log.Println("insert error")
		w.WriteHeader(http.StatusInternalServerError)
		tx.Rollback()
		return
	}
	tx.Commit()
}
