package ServerForMath

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func GenerateApikey() string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789!№;%:?*()-_+=")
	apikey := make([]rune, 30)
	rand.Seed(time.Now().UnixNano())
	for i := range apikey {
		apikey[i] = letters[rand.Intn(len(letters))]
	}
	return string(apikey)
}

func Receiver(math_connection *websocket.Conn, chan_math_connection chan *websocket.Conn, server_connection *websocket.Conn) {
	apikey := ""
	for {
		_, message_from_math, err := math_connection.ReadMessage()
		if err != nil {
			fmt.Println(err)
		}
		var message_map map[string]interface{}
		json_err := json.Unmarshal(message_from_math, &message_map)
		if json_err != nil {
			fmt.Println(json_err)
		}
		fmt.Println(message_map)
		if message_map["action"] == "Login" && message_map["login"] == "mathLogin" && message_map["password"] == "%wPp7VO6k7ump{BP4mu2rm4w?p|J5N%P" {
			chan_math_connection <- math_connection
			apikey = GenerateApikey()
			json_message, _ := json.Marshal(map[string]interface{}{"action": "Login", "apikey": apikey})
			math_connection.WriteMessage(websocket.TextMessage, json_message)
		} else if message_map["apikey"] == apikey && server_connection != nil {
			json_message, _ := json.Marshal(message_map)
			server_connection.WriteMessage(websocket.TextMessage, json_message)
		}
	}
}

func StartServer(chan_math_connection chan *websocket.Conn, server_connection *websocket.Conn) {
	var upgrader = websocket.Upgrader{}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		math_connection, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("New websocket connection")
		go Receiver(math_connection, chan_math_connection, server_connection)
	})
	http.ListenAndServe("localhost:8000", nil)
}
