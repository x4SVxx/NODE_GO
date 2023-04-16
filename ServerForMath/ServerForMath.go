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

var connections []*websocket.Conn

func Receiver(connection *websocket.Conn, chan_math_connection chan *websocket.Conn, server_connection *websocket.Conn) {
	math_apikey := ""
	for {
		_, message, err := connection.ReadMessage()
		if err != nil {
			Logger.Logger("ERROR : ReadMessage from node's client", err)
			for i := 0; i < len(connections); i++ {
				if connections[i] == connection {
					connections[i] = nil
				}
			}
			return
		} else {
			var message_map map[string]interface{}
			err := json.Unmarshal(message, &message_map)
			if err != nil {
				Logger.Logger("ERROR : Unmarshal message from node's client", err)
			} else {
				Logger.Logger("SUCCESS : Message from node's client: "+string(message), nil)
				if message_map["action"] == "Login" && message_map["login"] == "mathLogin" && message_map["password"] == "%wPp7VO6k7ump{BP4mu2rm4w?p|J5N%P" {
					chan_math_connection <- connection
					math_apikey = GenerateApikey()
					json_message, _ := json.Marshal(map[string]interface{}{"action": "Login", "apikey": math_apikey})
					connection.WriteMessage(websocket.TextMessage, json_message)
					Logger.Logger("Message to client: "+string(json_message), nil)
					if server_connection != nil {
						json_message_for_server, _ := json.Marshal(map[string]interface{}{"action": "Success", "data": "math connected"})
						server_connection.WriteMessage(websocket.TextMessage, json_message_for_server)
						Logger.Logger("SUCCESS : Message to server: "+string(json_message_for_server), nil)
					}
				} else if message_map["apikey"] == math_apikey {
					if server_connection != nil {
						server_connection.WriteMessage(websocket.TextMessage, message)
						Logger.Logger("SUCCESS : Message from math to server: "+string(message), nil)
					}
					for i := 0; i < len(connections); i++ {
						if connections[i] != connection && connections[i] != nil {
							connections[i].WriteMessage(websocket.TextMessage, message)
							Logger.Logger("SUCCESS : Message from math to client: "+string(message), nil)
						}
					}
				}
			}
		}
	}
}

func StartServer(node_server_ip string, node_server_port string, chan_math_connection chan *websocket.Conn, server_connection *websocket.Conn) {
	var upgrader = websocket.Upgrader{}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		connection, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			Logger.Logger("ERROR : Node's server connection", err)
			if server_connection != nil {
				json_message_for_server, _ := json.Marshal(map[string]interface{}{"action": "Error", "data": "Node's server connection"})
				server_connection.WriteMessage(websocket.TextMessage, json_message_for_server)
				Logger.Logger("SUCCESS: Message to server: "+string(json_message_for_server), nil)
			}
		}
		Logger.Logger("SUCCESS : New websocket connection for node", nil)
		if server_connection != nil {
			json_message_for_server, _ := json.Marshal(map[string]interface{}{"action": "Success", "data": "Node's server connection"})
			server_connection.WriteMessage(websocket.TextMessage, json_message_for_server)
			Logger.Logger("SUCCESS: Message to server: "+string(json_message_for_server), nil)
		}
		connections = append(connections, connection)
		go Receiver(connection, chan_math_connection, server_connection)
	})
	http.ListenAndServe(node_server_ip+":"+node_server_port, nil)
}
