package main

import (
	"NODE/Anchor"
	"NODE/Logger"
	"NODE/ReadAndSetNodeConfig"
	"NODE/ServerForMath"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var server_ip, server_port, login, password, roomid, independent_flag, connect_math_flag, node_server_ip, node_server_port, ref_tag_config = ReadAndSetNodeConfig.ReadAndSetNodeConfig()
var apikey, clientid, name, roomname, organization string
var login_flag, config_flag, start_spam_flag bool = false, false, false
var anchors_array []map[string]interface{}
var rf_config = map[string]interface{}{}
var server_connection *websocket.Conn
var math_connection *websocket.Conn

func CheckAnchors(anchors_mas *[]map[string]interface{}, server_connection *websocket.Conn) {
	for {
		func() {
			defer func() {
				if err := recover(); err != nil {
					Logger.Logger("ERROR : CheckAnchors", err)
				}
			}()
			for i := 0; i < len(*anchors_mas); i++ {
				if (*anchors_mas)[i]["connection"] == nil {
					Anchor.Connect(&(*anchors_mas)[i], server_connection)
					Anchor.SetRfConfig((*anchors_mas)[i], rf_config, server_connection)
				}
			}
		}()
	}
}

func main() {
	if independent_flag == "true" {
		chan_math_connection := make(chan *websocket.Conn)
		go ServerForMath.StartServer(node_server_ip, node_server_port, chan_math_connection, server_connection)
		math_connection = <-chan_math_connection
		close(chan_math_connection)

		var config []map[string]interface{}
		config_file, _ := os.ReadFile("Config.json")
		json.Unmarshal(config_file, &config)

		for i := 0; i < len(config); i++ {
			anchor := map[string]interface{}{}
			anchor["ip"] = config[i]["ip"].(string)
			anchor["number"] = config[i]["number"].(float64)
			anchor["masternumber"] = config[i]["masternumber"].(float64)
			anchor["role"] = config[i]["role"].(string)
			anchor["lag"] = config[i]["lag"].(float64)
			anchor["adrx"] = config[i]["adrx"].(float64)
			anchor["adtx"] = config[i]["adtx"].(float64)
			anchor["x"] = config[i]["x"].(float64)
			anchor["y"] = config[i]["y"].(float64)
			anchor["z"] = config[i]["z"].(float64)
			Anchor.Connect(&anchor, server_connection)
			anchors_array = append(anchors_array, anchor)
		}
		MessageToMath(math_connection, map[string]interface{}{"action": "RoomConfig", "data": map[string]interface{}{"clientid": "clientid", "organization": "clientid", "roomid": "roomid", "roomname": "roomname", "anchors": anchors_array, "ref_tag_config": ref_tag_config}})

		var rf_config []map[string]interface{}
		rf_config_file, _ := os.ReadFile("RfConfig.json")
		json.Unmarshal(rf_config_file, &rf_config)
		for i := 0; i < len(anchors_array); i++ {
			Anchor.SetRfConfig((anchors_array)[i], rf_config[0], server_connection)
		}
		for i := 0; i < len(anchors_array); i++ {
			Anchor.StartSpam((anchors_array)[i], server_connection)
			go Anchor.Handler("apikey", "name", "clientid", "roomid", "organization", &(anchors_array[i]), server_connection, math_connection)
		}

		for {
			var name string
			fmt.Scanf("%s\n", &name)
			if name == "Stop" {
				for i := 0; i < len(anchors_array); i++ {
					Anchor.StopSpam((anchors_array)[i], server_connection)
				}
			}
			if name == "Start" {
				for i := 0; i < len(anchors_array); i++ {
					Anchor.StartSpam((anchors_array)[i], server_connection)
					go Anchor.Handler("apikey", "name", "clientid", "roomid", "organization", &(anchors_array[i]), server_connection, math_connection)
				}
			}
		}
	}

	if independent_flag == "false" {
		for {
			func() {
				defer func() {
					if err := recover(); err != nil {
						Logger.Logger("ERROR : main - server handler", err)
					}
				}()
				URL := url.URL{Scheme: "ws", Host: server_ip + ":" + server_port}
				server_connection, _, err := websocket.DefaultDialer.Dial(URL.String(), nil)
				if err != nil {
					Logger.Logger("ERROR : main - server connection", err)
				} else {
					Logger.Logger("SUCCESS : node connected to the server "+string(server_ip)+":"+string(server_port), nil)

					if connect_math_flag == "true" {
						chan_math_connection := make(chan *websocket.Conn)
						go ServerForMath.StartServer(node_server_ip, node_server_port, chan_math_connection, server_connection)
						math_connection = <-chan_math_connection
						close(chan_math_connection)
					}

					MessageToServer(server_connection, map[string]interface{}{"action": "Login", "login": login, "password": password, "roomid": roomid})

					break_main_receiver_point := false
					for {
						if break_main_receiver_point {
							break
						}
						func() {
							defer func() {
								if err := recover(); err != nil {
									Logger.Logger("ERROR : main receiver", err)
									if err.(string) == "repeated read on failed websocket connection" {
										break_main_receiver_point = true
									}
								}
							}()
							_, message, err := server_connection.ReadMessage()
							if err != nil {
								Logger.Logger("ERROR : message from server", err)
							} else {
								Logger.Logger("SUCCESS : message from server "+string(message), nil)

								var message_map map[string]interface{}
								err := json.Unmarshal(message, &message_map)
								if err != nil {
									Logger.Logger("ERROR : Unmarshal message from server", err)
								} else {
									if message_map["action"] == "Login" && message_map["status"] == "true" {
										Login(message_map, server_connection)
									}
									if message_map["action"] == "SetConfig" && message_map["status"] == "true" {
										SetConfig(message_map, server_connection, math_connection)
									}
									if message_map["action"] == "Start" && message_map["status"] == "true" {
										StartSpam(server_connection, math_connection)
									}
									if message_map["action"] == "Stop" && message_map["status"] == "true" {
										StopSpam(server_connection)
									}
								}
							}
						}()
					}
				}
			}()
		}
	}
}

func Login(message_map map[string]interface{}, server_connection *websocket.Conn) {
	defer func() {
		if err := recover(); err != nil {
			login_flag = false
			Logger.Logger("ERROR : Login", err)
			MessageToServer(server_connection, map[string]interface{}{"action": "Error", "data": "Error: Login"})
		}
	}()
	apikey = string(message_map["data"].(map[string]interface{})["apikey"].(string))
	clientid = string(message_map["data"].(map[string]interface{})["clientid"].(string))
	name = string(message_map["data"].(map[string]interface{})["name"].(string))
	roomname = string(message_map["data"].(map[string]interface{})["roomname"].(string))
	organization = string(message_map["data"].(map[string]interface{})["organization"].(string))

	if len(strings.TrimSpace(apikey)) == 0 ||
		len(strings.TrimSpace(clientid)) == 0 ||
		len(strings.TrimSpace(name)) == 0 ||
		len(strings.TrimSpace(roomname)) == 0 ||
		len(strings.TrimSpace(organization)) == 0 {
		MessageToServer(server_connection, map[string]interface{}{"action": "Error", "data": "Error: Login - empty data"})
		login_flag = false
	} else {
		MessageToServer(server_connection, map[string]interface{}{"action": "Success", "data": "Success: Login"})
		login_flag = true
	}
}

func SetConfig(message_map map[string]interface{}, server_connection *websocket.Conn, math_connection *websocket.Conn) {
	defer func() {
		if err := recover(); err != nil {
			Logger.Logger("ERROR : main SetConfig", err)
			MessageToServer(server_connection, map[string]interface{}{"action": "Error", "data": "Error: main SetConfig"})
		}
	}()
	if !login_flag {
		MessageToServer(server_connection, map[string]interface{}{"action": "Error", "data": "Error: SetConfig - need autorization"})
	} else {
		StopSpam(server_connection)
		for i := 0; i < len(anchors_array); i++ {
			Anchor.DisConnect(&(anchors_array[i]), server_connection)
		}
		anchors_array = []map[string]interface{}{}
		time.Sleep(1 * time.Second)

		for i := 0; i < len(message_map["anchors"].([]interface{})); i++ {
			anchor := map[string]interface{}{}
			anchor["ip"] = message_map["anchors"].([]interface{})[i].(map[string]interface{})["ip"].(string)
			anchor["number"] = message_map["anchors"].([]interface{})[i].(map[string]interface{})["number"].(float64)
			anchor["masternumber"] = message_map["anchors"].([]interface{})[i].(map[string]interface{})["masternumber"].(float64)
			anchor["role"] = message_map["anchors"].([]interface{})[i].(map[string]interface{})["role"].(string)
			anchor["lag"] = message_map["anchors"].([]interface{})[i].(map[string]interface{})["lag"].(float64)
			anchor["adrx"] = message_map["anchors"].([]interface{})[i].(map[string]interface{})["adrx"].(float64)
			anchor["adtx"] = message_map["anchors"].([]interface{})[i].(map[string]interface{})["adtx"].(float64)
			anchor["x"] = message_map["anchors"].([]interface{})[i].(map[string]interface{})["x"].(float64)
			anchor["y"] = message_map["anchors"].([]interface{})[i].(map[string]interface{})["y"].(float64)
			anchor["z"] = message_map["anchors"].([]interface{})[i].(map[string]interface{})["z"].(float64)
			anchor["connection"] = nil
			anchor["id"] = nil
			Anchor.Connect(&anchor, server_connection)
			anchors_array = append(anchors_array, anchor)
		}

		rf_config["chnum"] = message_map["rf_config"].([]interface{})[0].(map[string]interface{})["chnum"].(float64)
		rf_config["prf"] = message_map["rf_config"].([]interface{})[0].(map[string]interface{})["prf"].(float64)
		rf_config["datarate"] = message_map["rf_config"].([]interface{})[0].(map[string]interface{})["datarate"].(float64)
		rf_config["preamblecode"] = message_map["rf_config"].([]interface{})[0].(map[string]interface{})["preamblecode"].(float64)
		rf_config["preamblelen"] = message_map["rf_config"].([]interface{})[0].(map[string]interface{})["preamblelen"].(float64)
		rf_config["pac"] = message_map["rf_config"].([]interface{})[0].(map[string]interface{})["pac"].(float64)
		rf_config["nsfd"] = message_map["rf_config"].([]interface{})[0].(map[string]interface{})["nsfd"].(float64)
		rf_config["diagnostic"] = message_map["rf_config"].([]interface{})[0].(map[string]interface{})["diagnostic"].(float64)
		rf_config["lag"] = message_map["rf_config"].([]interface{})[0].(map[string]interface{})["lag"].(float64)
		for i := 0; i < len(anchors_array); i++ {
			Anchor.SetRfConfig((anchors_array)[i], rf_config, server_connection)
		}

		config_flag = true

		if math_connection != nil {
			MessageToMath(math_connection, map[string]interface{}{"action": "RoomConfig", "data": map[string]interface{}{"clientid": "clientid", "organization": "organization", "roomid": "roomid", "roomname": "roomname", "anchors": anchors_array, "ref_tag_config": ref_tag_config}})
		} else {
			MessageToServer(server_connection, map[string]interface{}{"action": "RoomConfig", "apikey": apikey, "clientid": clientid, "organization": clientid, "roomid": roomid, "roomname": roomname, "anchors": anchors_array, "ref_tag_config": ref_tag_config})
		}
	}
}

func StartSpam(server_connection *websocket.Conn, math_connection *websocket.Conn) {
	defer func() {
		if err := recover(); err != nil {
			Logger.Logger("ERROR : main StartSpam", err)
			MessageToServer(server_connection, map[string]interface{}{"action": "Error", "data": "Error: main StartSpam"})
		}
	}()
	if !config_flag {
		MessageToServer(server_connection, map[string]interface{}{"action": "Error", "data": "Error: StartSpam - need Config"})
	} else {
		if !start_spam_flag {
			for i := 0; i < len(anchors_array); i++ {
				Anchor.StartSpam((anchors_array)[i], server_connection)
				go Anchor.Handler(apikey, name, clientid, roomid, organization, &(anchors_array[i]), server_connection, math_connection)
			}
			start_spam_flag = true
		}
	}
}

func StopSpam(server_connection *websocket.Conn) {
	defer func() {
		if err := recover(); err != nil {
			Logger.Logger("ERROR : main StopSpam", err)
			MessageToServer(server_connection, map[string]interface{}{"action": "Error", "data": "Error: Error main StopSpam"})
		}
	}()
	if start_spam_flag {
		for i := 0; i < len(anchors_array); i++ {
			Anchor.StopSpam(anchors_array[i], server_connection)
		}
		start_spam_flag = false
	}
}

func MessageToServer(server_connection *websocket.Conn, map_message map[string]interface{}) {
	defer func() {
		if err := recover(); err != nil {
			Logger.Logger("ERROR : main - MessageToServer", err)
		}
	}()
	json_message, _ := json.Marshal(map_message)
	server_connection.WriteMessage(websocket.TextMessage, json_message)
	Logger.Logger("SUCCESS : main - Message to server: "+string(json_message), nil)
}

func MessageToMath(math_connection *websocket.Conn, map_message map[string]interface{}) {
	defer func() {
		if err := recover(); err != nil {
			Logger.Logger("ERROR : main - MessageToMath", err)
		}
	}()
	json_message, _ := json.Marshal(map_message)
	math_connection.WriteMessage(websocket.TextMessage, json_message)
	Logger.Logger("SUCCESS : main - Message to math: "+string(json_message), nil)
}
