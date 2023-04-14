package ServerForMath

import (
	"NODE/Logger"
	"encoding/json"
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

func Receiver(client_connections *[]*websocket.Conn, connection *websocket.Conn, chan_math_connection chan *websocket.Conn, server_connection *websocket.Conn) {
	apikey := ""
	for {
		_, message_from_math, err := connection.ReadMessage()
		if err != nil {
			Logger.Logger("Error: ReadMessage from node's client", err)
			var null_connections *websocket.Conn
			for i := 0; i < len(*client_connections); i++ {
				if (*client_connections)[i] == connection {
					(*client_connections)[i] = null_connections
				}
			}
			return
		} else {
			var message_map map[string]interface{}
			json.Unmarshal(message_from_math, &message_map)
			Logger.Logger("Message from node's client: "+string(message_from_math), nil)
			if message_map["action"] == "Login" && message_map["login"] == "mathLogin" && message_map["password"] == "%wPp7VO6k7ump{BP4mu2rm4w?p|J5N%P" {
				chan_math_connection <- connection
				apikey = GenerateApikey()
				json_message, _ := json.Marshal(map[string]interface{}{"action": "Login", "apikey": apikey})
				connection.WriteMessage(websocket.TextMessage, json_message)
				Logger.Logger("Message to client: "+string(json_message), nil)
			} else if message_map["apikey"] == apikey && server_connection != nil {
				json_message, _ := json.Marshal(message_map)
				server_connection.WriteMessage(websocket.TextMessage, json_message)
				Logger.Logger("Message to client: "+string(json_message), nil)
			} else if message_map["action"] == "ECHO" {
				active_client_count := 0
				nil_client_count := 0
				for i := 0; i < len(*client_connections); i++ {
					if (*client_connections)[i] != nil {
						active_client_count += 1
					} else {
						nil_client_count += 1
					}
				}
				json_ws_info_message, _ := json.Marshal(map[string]interface{}{"action": "INFO", "active_ws": active_client_count, "nil_client": nil_client_count})
				for i := 0; i < len(*client_connections); i++ {
					if (*client_connections)[i] != connection && (*client_connections)[i] != nil {
						(*client_connections)[i].WriteMessage(websocket.TextMessage, message_from_math)
						(*client_connections)[i].WriteMessage(websocket.TextMessage, json_ws_info_message)
						Logger.Logger("Message to client: "+string(message_from_math), nil)
						Logger.Logger("Message to client: "+string(json_ws_info_message), nil)
					}
				}
			}
		}
	}
}

func StartServer(node_server_ip string, node_server_port string, chan_math_connection chan *websocket.Conn, server_connection *websocket.Conn) {
	var upgrader = websocket.Upgrader{}
	var connections []*websocket.Conn
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		connection, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			Logger.Logger("Error: Node's server connection", err)
		}
		Logger.Logger("New websocket connection for node", nil)
		Logger.Logger("Node's websockets count: ", nil)
		connections = append(connections, connection)
		go Receiver(&connections, connection, chan_math_connection, server_connection)
	})
	http.ListenAndServe(node_server_ip+":"+node_server_port, nil)
}
