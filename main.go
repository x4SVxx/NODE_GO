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

func main() {
	if connect_math_flag == "true" || independent_flag == "true" {
		go ServerForMath.StartServer(node_server_ip, node_server_port, server_connection)
	}

	if independent_flag == "true" {
		var anchors_reader []interface{}
		var rf_config_reader []interface{}
		config_file, _ := os.ReadFile("Config.json")
		rf_config_file, _ := os.ReadFile("RfConfig.json")
		json.Unmarshal(config_file, &anchors_reader)
		json.Unmarshal(rf_config_file, &rf_config_reader)

		login_flag = true
		SetConfig(map[string]interface{}{"anchors": anchors_reader, "rf_config": rf_config_reader})
		ServerForMath.RoomAndReftagConfig(anchors_array, ref_tag_config)
		StartSpam()

		go CheckAnchors(server_connection)

		for {
			var name string
			fmt.Scanf("%s\n", &name)
			if name == "Stop" {
				StopSpam()
			}
			if name == "Start" {
				StartSpam()
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
				var error_server_connection error
				server_connection, _, error_server_connection = websocket.DefaultDialer.Dial(URL.String(), nil)
				if error_server_connection != nil {
					Logger.Logger("ERROR : main - server connection", error_server_connection)
				} else {
					Logger.Logger("SUCCESS : node connected to the server "+string(server_ip)+":"+string(server_port), nil)
					MessageToServer(map[string]interface{}{"action": "Login", "login": login, "password": password, "roomid": roomid})

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
							_, message, error_read_message_from_server := server_connection.ReadMessage()
							if error_read_message_from_server != nil {
								Logger.Logger("ERROR : message from server", error_read_message_from_server)
							} else {
								Logger.Logger("SUCCESS : message from server "+string(message), nil)
								var message_map map[string]interface{}
								error_unmarshal_json := json.Unmarshal(message, &message_map)
								if error_unmarshal_json != nil {
									Logger.Logger("ERROR : Unmarshal message from server", error_unmarshal_json)
								} else {
									if message_map["action"] == "Login" && message_map["status"] == "true" {
										Login(message_map)
									}
									if message_map["action"] == "SetConfig" && message_map["status"] == "true" {
										SetConfig(message_map)
									}
									if message_map["action"] == "Start" && message_map["status"] == "true" {
										StartSpam()
									}
									if message_map["action"] == "Stop" && message_map["status"] == "true" {
										StopSpam()
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

func Login(message_map map[string]interface{}) {
	defer func() {
		if err := recover(); err != nil {
			login_flag = false
			Logger.Logger("ERROR : Login", err)
			MessageToServer(map[string]interface{}{"action": "Error", "data": "Error: Login"})
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
		MessageToServer(map[string]interface{}{"action": "Error", "data": "Error: Login - empty data"})
		login_flag = false
	} else {
		MessageToServer(map[string]interface{}{"action": "Success", "data": "Success: Login"})
		login_flag = true
	}
}

func SetConfig(message_map map[string]interface{}) {
	defer func() {
		if err := recover(); err != nil {
			Logger.Logger("ERROR : main SetConfig", err)
			MessageToServer(map[string]interface{}{"action": "Error", "data": "Error: main SetConfig"})
		}
	}()
	if !login_flag {
		MessageToServer(map[string]interface{}{"action": "Error", "data": "Error: SetConfig - need autorization"})
	} else {
		if start_spam_flag {
			StopSpam()
		}
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
		go CheckAnchors(server_connection)

		if independent_flag == "true" || connect_math_flag == "true" {
			ServerForMath.RoomAndReftagConfig(anchors_array, ref_tag_config)
		} else {
			MessageToServer(map[string]interface{}{"action": "RoomConfig", "apikey": apikey, "clientid": clientid, "organization": organization, "roomid": roomid, "roomname": roomname, "anchors": anchors_array, "ref_tag_config": ref_tag_config})
		}
	}
}

func StartSpam() {
	defer func() {
		if err := recover(); err != nil {
			Logger.Logger("ERROR : main StartSpam", err)
			MessageToServer(map[string]interface{}{"action": "Error", "data": "Error: main StartSpam"})
		}
	}()
	if !config_flag {
		MessageToServer(map[string]interface{}{"action": "Error", "data": "Error: StartSpam - need config"})
	} else {
		if !start_spam_flag {
			start_spam_flag = true
			for i := 0; i < len(anchors_array); i++ {
				Anchor.StartSpam((anchors_array)[i], server_connection)
				go Anchor.Handler(apikey, name, clientid, roomid, organization, independent_flag, connect_math_flag, &(anchors_array[i]), server_connection)
			}
		}
	}
}

func StopSpam() {
	defer func() {
		if err := recover(); err != nil {
			Logger.Logger("ERROR : main StopSpam", err)
			MessageToServer(map[string]interface{}{"action": "Error", "data": "Error: Error main StopSpam"})
		}
	}()
	if start_spam_flag {
		start_spam_flag = false
		for i := 0; i < len(anchors_array); i++ {
			Anchor.StopSpam(anchors_array[i], server_connection)
		}
	}
}

func CheckAnchors(server_connection *websocket.Conn) {
	for {
		func() {
			defer func() {
				if err := recover(); err != nil {
					Logger.Logger("ERROR : CheckAnchors", err)
				}
			}()
			for i := 0; i < len(anchors_array); i++ {
				if anchors_array[i]["connection"] == nil {
					Anchor.Connect(&anchors_array[i], server_connection)
					Anchor.SetRfConfig(anchors_array[i], rf_config, server_connection)
					if start_spam_flag {
						Anchor.StartSpam(anchors_array[i], server_connection)
						if independent_flag == "false" {
							go Anchor.Handler(apikey, name, clientid, roomid, organization, independent_flag, connect_math_flag, &(anchors_array[i]), server_connection)
						} else {
							go Anchor.Handler("apikey", "name", "clientid", "roomid", "organization", independent_flag, connect_math_flag, &(anchors_array[i]), server_connection)
						}
					}
				}
			}
		}()
		time.Sleep(5 * time.Second)
	}
}

func MessageToServer(map_message map[string]interface{}) {
	defer func() {
		if err := recover(); err != nil {
			Logger.Logger("ERROR : main - MessageToServer", err)
		}
	}()
	if server_connection != nil {
		json_message, _ := json.Marshal(map_message)
		server_connection.WriteMessage(websocket.TextMessage, json_message)
		Logger.Logger("SUCCESS : main - Message to server: "+string(json_message), nil)
	}
}
