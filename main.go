package main

import (
	"log"
	"net/http"

	r "gopkg.in/rethinkdb/rethinkdb-go.v6"
)

func main() {
	// establish rethinkdb address
	session, err := r.Connect(r.ConnectOpts{
		Address:  "localhost:28015",
		Database: "chat_database",
	})

	if err != nil {
		log.Panic(err.Error())
	}

	router := NewRouter(session)

	// TODO: handles
	router.Handle("channel add", addChannel)
	router.Handle("channel subscribe", subscribeChannel)
	router.Handle("channel unsubscribe", unsubscribeChannel)

	router.Handle("user edit", editUser)
	router.Handle("user subscribe", subscribeUser)
	router.Handle("user unsubscribe", unsubscribeUser)

	router.Handle("message add", addChannelMessage)
	router.Handle("message subscribe", subscribeChannelMessage)
	router.Handle("message unsubscribe", unsubscribeChannelMessage)

	http.Handle("/", router)
	http.ListenAndServe(":4000", nil)
}
