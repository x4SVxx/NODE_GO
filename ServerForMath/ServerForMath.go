package ServerForMath

import (
	"NODE/Logger"
	"encoding/json"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var connections []*websocket.Conn
var math_connection *websocket.Conn
var anchors_array []map[string]interface{}
var ref_tag_config map[string]interface{}

func RoomAndReftagConfig(anchors []map[string]interface{}, ref_tag map[string]interface{}) {
	anchors_array = anchors
	ref_tag_config = ref_tag
	if math_connection != nil {
		MessageToMath(map[string]interface{}{"action": "RoomConfig", "data": map[string]interface{}{"clientid": "clientid", "organization": "clientid", "roomid": "roomid", "roomname": "roomname", "anchors": anchors_array, "ref_tag_config": ref_tag_config}})
	}
}

func GenerateApikey() string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789!№;%:?*()-_+=")
	apikey := make([]rune, 30)
	rand.Seed(time.Now().UnixNano())
	for i := range apikey {
		apikey[i] = letters[rand.Intn(len(letters))]
	}
	return string(apikey)
}

func Receiver(connection *websocket.Conn, server_connection *websocket.Conn) {
	math_apikey := ""
	break_flag := false
	for {
		if break_flag {
			return
		}
		func() {
			defer func() {
				if err := recover(); err != nil {
					Logger.Logger("ERROR : ServerForMath Receiver", err)
					if err.(string) == "repeated read on failed websocket connection" {
						break_flag = true
						if math_connection != nil {
							if math_connection.LocalAddr() == connection.LocalAddr() {
								math_connection.Close()
								math_connection = nil
							}
						}
						for i := 0; i < len(connections); i++ {
							if connections[i] != nil {
								if connections[i].LocalAddr() == connection.LocalAddr() {
									connection.Close()
									connections[i] = nil
								}
							}
						}
					}
				}
			}()

			_, message, _ := connection.ReadMessage()
			var message_map map[string]interface{}
			json.Unmarshal(message, &message_map)
			Logger.Logger("SUCCESS : Message from node's client: "+string(message), nil)
			if message_map["action"] == "Login" && message_map["login"] == "mathLogin" && message_map["password"] == "%wPp7VO6k7ump{BP4mu2rm4w?p|J5N%P" {
				math_connection = connection
				math_apikey = GenerateApikey()
				json_message, _ := json.Marshal(map[string]interface{}{"action": "Login", "apikey": math_apikey})
				connection.WriteMessage(websocket.TextMessage, json_message)
				Logger.Logger("Message to client: "+string(json_message), nil)
				if anchors_array != nil && ref_tag_config != nil {
					MessageToMath(map[string]interface{}{"action": "RoomConfig", "data": map[string]interface{}{"clientid": "clientid", "organization": "clientid", "roomid": "roomid", "roomname": "roomname", "anchors": anchors_array, "ref_tag_config": ref_tag_config}})
				}
				if server_connection != nil {
					json_message_for_server, _ := json.Marshal(map[string]interface{}{"action": "Success", "data": "math connected"})
					server_connection.WriteMessage(websocket.TextMessage, json_message_for_server)
					Logger.Logger("SUCCESS : Message to server: "+string(json_message_for_server), nil)
				}
			} else if message_map["apikey"] == math_apikey {
				if message_map["action"] == "PING" {
					map_answer := map[string]interface{}{
						"action": "PONG",
						"apikey": math_apikey,
					}
					json_answer, _ := json.Marshal(map_answer)
					connection.WriteMessage(websocket.TextMessage, json_answer)
				}
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
		}()
	}
}

func StartServer(node_server_ip string, node_server_port string, server_connection *websocket.Conn) {
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
		go Receiver(connection, server_connection)
	})
	http.ListenAndServe(node_server_ip+":"+node_server_port, nil)
}

func MessageToMath(map_message map[string]interface{}) {
	defer func() {
		if err := recover(); err != nil {
			Logger.Logger("ERROR : MessageToMath", err)
		}
	}()

	if math_connection != nil {
		json_message, _ := json.Marshal(map_message)
		math_connection.WriteMessage(websocket.TextMessage, json_message)
		Logger.Logger("SUCCESS : Message to math: "+string(json_message), nil)
	}
}
