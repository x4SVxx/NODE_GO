package main

import (
	"NODE/Anchor"
	"NODE/ServerForMath"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

func ReadAndSetConfig() (string, string, string, string, string, string, string) {
	var node_config map[string]interface{}
	file, _ := os.ReadFile("NodeConfig.json")
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Error: ReadAndSetConfig -", err)
			os.Exit(0)
		}
	}()
	json.Unmarshal(file, &node_config)
	server_ip := node_config["server_ip"].(string)
	server_port := node_config["server_port"].(string)
	login := node_config["login"].(string)
	password := node_config["password"].(string)
	roomid := node_config["roomid"].(string)
	independent_flag := node_config["independent_flag"].(string)
	connect_math_flag := node_config["connect_math_flag"].(string)

	if len(strings.TrimSpace(server_ip)) == 0 || len(strings.TrimSpace(server_port)) == 0 || len(strings.TrimSpace(login)) == 0 || len(strings.TrimSpace(password)) == 0 || len(strings.TrimSpace(roomid)) == 0 || len(strings.TrimSpace(independent_flag)) == 0 || len(strings.TrimSpace(connect_math_flag)) == 0 {
		fmt.Println("Error: ReadAndSetConfig - empty data in node config")
		os.Exit(0)
	} else {
		fmt.Println("Success: ReadAndSetConfig \n"+"server_ip:", server_ip, "\n"+"server_port:", server_port, "\n"+"login:", login, "\n"+"password:", password, "\n"+"roomid:", roomid, "\n"+"independent_flag:", independent_flag)
	}
	return server_ip, server_port, login, password, roomid, independent_flag, connect_math_flag
}

func main() {
	var server_ip, server_port, login, password, roomid, independent_flag, connect_math_flag string
	server_ip, server_port, login, password, roomid, independent_flag, connect_math_flag = ReadAndSetConfig()
	var apikey, clientid, name, roomname string
	var login_flag, config_flag, rf_config_flag, start_spam_flag, stop_spam_flag bool = false, false, false, false, false
	var anchors_array []map[string]interface{}
	var rf_config = map[string]interface{}{}

	if independent_flag == "true" {
		var server_connection *websocket.Conn
		chan_math_connection := make(chan *websocket.Conn)
		go ServerForMath.StartServer(chan_math_connection, server_connection)
		math_connection := <-chan_math_connection
		close(chan_math_connection)
		print(math_connection)

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
			Anchor.Connect(&anchor, &anchors_array)
		}
		MessageToMath(math_connection, map[string]interface{}{"action": "RoomConfig", "data": map[string]interface{}{"clientid": "clientid", "organization": "clientid", "roomid": "roomid", "roomname": "roomname", "anchors": anchors_array}})

		var rf_config []map[string]interface{}
		rf_config_file, _ := os.ReadFile("RfConfig.json")
		json.Unmarshal(rf_config_file, &rf_config)
		for i := 0; i < len(anchors_array); i++ {
			Anchor.SetRfConfig((anchors_array)[i], rf_config[0])
		}
		for i := 0; i < len(anchors_array); i++ {
			Anchor.StartSpam((anchors_array)[i])
			go Anchor.Handler("apikey", "name", "clientid", "roomid", (anchors_array)[i], server_connection, math_connection)
		}

		for {
			var name string
			fmt.Scanf("%s\n", &name)
			if name == "Stop" {
				for i := 0; i < len(anchors_array); i++ {
					Anchor.StopSpam((anchors_array)[i])
				}
			}
			if name == "Start" {
				for i := 0; i < len(anchors_array); i++ {
					Anchor.StartSpam((anchors_array)[i])
					go Anchor.Handler("apikey", "name", "clientid", "roomid", (anchors_array)[i], server_connection, math_connection)
				}
			}
		}

	}

	if independent_flag == "false" {
		for {
			func() {
				defer func() {
					if err := recover(); err != nil {
						fmt.Println("Error: server connection - ", err)
					}
				}()
				URL := url.URL{Scheme: "ws", Host: server_ip + ":" + server_port}
				server_connection, _, err := websocket.DefaultDialer.Dial(URL.String(), nil)
				if err != nil {
					fmt.Println("Error: server connection -", err)
				} else {
					fmt.Println("Success: connect to the server " + server_ip + ":" + server_port)

					var math_connection *websocket.Conn
					if connect_math_flag == "true" {
						chan_math_connection := make(chan *websocket.Conn)
						go ServerForMath.StartServer(chan_math_connection, server_connection)
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
									fmt.Println("Error: main Receiver -", err)
									if err.(string) == "repeated read on failed websocket connection" {
										break_main_receiver_point = true
									}
								}
							}()
							_, message, err := server_connection.ReadMessage()
							if err != nil {
								fmt.Println("Error: message from server -", err)
							} else {
								fmt.Println("Success: message from server", string(message))
								var message_map map[string]interface{}
								err := json.Unmarshal(message, &message_map)
								if err != nil {
									fmt.Println(err)
								}
								fmt.Println(message_map["action"])
								if message_map["action"] == "Login" && message_map["status"] == "true" {
									Login(message_map, &apikey, &clientid, &name, &roomname, server_connection, &login_flag)
								}
								if message_map["action"] == "SetConfig" && message_map["status"] == "true" {
									SetConfig(message_map, &anchors_array, &rf_config, &apikey, &clientid, &roomid, &roomname, &login_flag, &config_flag, &rf_config_flag, server_connection, math_connection)
								}
								if message_map["action"] == "SetRfConfig" && message_map["status"] == "true" {
									SetRfConfig(message_map, &rf_config, &anchors_array, server_connection, &login_flag, &rf_config_flag, &config_flag)
								}
								if message_map["action"] == "Start" && message_map["status"] == "true" {
									StartSpam(&apikey, &clientid, &roomid, &name, &roomname, &anchors_array, server_connection, math_connection, &start_spam_flag, &stop_spam_flag, &config_flag, &rf_config_flag)
								}
								if message_map["action"] == "Stop" && message_map["status"] == "true" {
									StopSpam(&anchors_array, server_connection, &stop_spam_flag, &start_spam_flag)
								}
							}
						}()
					}
				}
			}()
		}
	}
}

func Login(message_map map[string]interface{}, apikey *string, clientid *string, name *string, roomname *string, server_connection *websocket.Conn, login_flag *bool) {
	defer func() {
		if err := recover(); err != nil {
			*login_flag = false
			fmt.Println("Error: Login -", err)
			MessageToServer(server_connection, map[string]interface{}{"action": "Error", "data": "Error: Login"})
		}
	}()
	*apikey = string(message_map["data"].(map[string]interface{})["apikey"].(string))
	*clientid = string(message_map["data"].(map[string]interface{})["clientid"].(string))
	*name = string(message_map["data"].(map[string]interface{})["name"].(string))
	*roomname = string(message_map["data"].(map[string]interface{})["roomname"].(string))

	if len(strings.TrimSpace(*apikey)) == 0 || len(strings.TrimSpace(*clientid)) == 0 || len(strings.TrimSpace(*name)) == 0 || len(strings.TrimSpace(*roomname)) == 0 {
		MessageToServer(server_connection, map[string]interface{}{"action": "Error", "data": "Error: Login - empty data"})
		*login_flag = false
	} else {
		MessageToServer(server_connection, map[string]interface{}{"action": "Success", "data": "Success: Login"})
		*login_flag = true
	}
}

func SetConfig(message_map map[string]interface{}, anchors_array *[]map[string]interface{}, rf_config *map[string]interface{}, apikey *string, clientid *string, roomid *string, roomname *string, login_flag *bool, config_flag *bool, rf_config_flag *bool, server_connection *websocket.Conn, math_connection *websocket.Conn) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Error: main SetConfig -", err)
			MessageToServer(server_connection, map[string]interface{}{"action": "Error", "data": "Error: SetConfig"})
		}
	}()
	if *login_flag {
		Anchor.DisConnect(anchors_array, server_connection)
		time.Sleep(1 * time.Second)
		for i := 0; i < len(message_map["data"].([]interface{})); i++ {
			anchor := map[string]interface{}{}
			anchor["ip"] = message_map["data"].([]interface{})[i].(map[string]interface{})["ip"].(string)
			anchor["number"] = message_map["data"].([]interface{})[i].(map[string]interface{})["number"].(float64)
			anchor["masternumber"] = message_map["data"].([]interface{})[i].(map[string]interface{})["masternumber"].(float64)
			anchor["role"] = message_map["data"].([]interface{})[i].(map[string]interface{})["role"].(string)
			anchor["lag"] = message_map["data"].([]interface{})[i].(map[string]interface{})["lag"].(float64)
			anchor["adrx"] = message_map["data"].([]interface{})[i].(map[string]interface{})["adrx"].(float64)
			anchor["adtx"] = message_map["data"].([]interface{})[i].(map[string]interface{})["adtx"].(float64)
			anchor["x"] = message_map["data"].([]interface{})[i].(map[string]interface{})["x"].(float64)
			anchor["y"] = message_map["data"].([]interface{})[i].(map[string]interface{})["y"].(float64)
			anchor["z"] = message_map["data"].([]interface{})[i].(map[string]interface{})["z"].(float64)
			if len(strings.TrimSpace(anchor["ip"].(string))) == 0 || len(strings.TrimSpace(anchor["role"].(string))) == 0 {
				MessageToServer(server_connection, map[string]interface{}{"action": "Error", "data": "Error: SetConfig - empty data in config"})
			} else {
				Anchor.Connect(&anchor, anchors_array, server_connection)
			}
		}
		if math_connection != nil {
			MessageToMath(math_connection, map[string]interface{}{"action": "RoomConfig", "data": map[string]interface{}{"clientid": *clientid, "organization": *clientid, "roomid": *roomid, "roomname": *roomname, "anchors": *anchors_array}})
		} else {
			MessageToServer(server_connection, map[string]interface{}{"action": "RoomConfig", "apikey": *apikey, "clientid": *clientid, "organization": *clientid, "roomid": *roomid, "roomname": *roomname, "data": *anchors_array})
		}

		*config_flag = true
		if *rf_config_flag {
			for i := 0; i < len(*anchors_array); i++ {
				Anchor.SetRfConfig((*anchors_array)[i], *rf_config, server_connection)
			}
		}
	} else {
		MessageToServer(server_connection, map[string]interface{}{"action": "Error", "data": "Error: SetConfig - need autorization"})
	}
}

func SetRfConfig(message_map map[string]interface{}, rf_config *map[string]interface{}, anchors_array *[]map[string]interface{}, server_connection *websocket.Conn, login_flag *bool, rf_config_flag *bool, config_flag *bool) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Error: main SetRfConfig -", err)
			MessageToServer(server_connection, map[string]interface{}{"action": "Error", "data": "Error: SetRfConfig"})
		}
	}()
	if *login_flag {
		(*rf_config)["chnum"] = message_map["data"].([]interface{})[0].(map[string]interface{})["chnum"].(float64)
		(*rf_config)["prf"] = message_map["data"].([]interface{})[0].(map[string]interface{})["prf"].(float64)
		(*rf_config)["datarate"] = message_map["data"].([]interface{})[0].(map[string]interface{})["datarate"].(float64)
		(*rf_config)["preamblecode"] = message_map["data"].([]interface{})[0].(map[string]interface{})["preamblecode"].(float64)
		(*rf_config)["preamblelen"] = message_map["data"].([]interface{})[0].(map[string]interface{})["preamblelen"].(float64)
		(*rf_config)["pac"] = message_map["data"].([]interface{})[0].(map[string]interface{})["pac"].(float64)
		(*rf_config)["nsfd"] = message_map["data"].([]interface{})[0].(map[string]interface{})["nsfd"].(float64)
		(*rf_config)["diagnostic"] = message_map["data"].([]interface{})[0].(map[string]interface{})["diagnostic"].(float64)
		(*rf_config)["lag"] = message_map["data"].([]interface{})[0].(map[string]interface{})["lag"].(float64)
		*rf_config_flag = true
		if *config_flag {
			for i := 0; i < len(*anchors_array); i++ {
				Anchor.SetRfConfig((*anchors_array)[i], *rf_config, server_connection)
			}
		}
	} else {
		MessageToServer(server_connection, map[string]interface{}{"action": "Error", "data": "Error: SetRfConfig - need autorization"})
	}
}

func StartSpam(apikey *string, clientid *string, roomid *string, name *string, roomname *string, anchors_array *[]map[string]interface{}, server_connection *websocket.Conn, math_connection *websocket.Conn, start_spam_flag *bool, stop_spam_flag *bool, config_flag *bool, rf_config_flag *bool) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Error: main StartSpam - ", err)
			MessageToServer(server_connection, map[string]interface{}{"action": "Error", "data": "Error: StartSpam"})
		}
	}()
	if *config_flag && *rf_config_flag {
		for i := 0; i < len(*anchors_array); i++ {
			Anchor.StartSpam((*anchors_array)[i], server_connection)
			go Anchor.Handler(*apikey, *name, *clientid, *roomid, (*anchors_array)[i], server_connection, math_connection)
		}
		*start_spam_flag = true
		*stop_spam_flag = false
	} else {
		if !*config_flag {
			MessageToServer(server_connection, map[string]interface{}{"action": "Error", "data": "Error: StartSpam - need Config"})
		}
		if !*rf_config_flag {
			MessageToServer(server_connection, map[string]interface{}{"action": "Error", "data": "Error: StartSpam - need RfConfig"})
		}
	}
}

func StopSpam(anchors_array *[]map[string]interface{}, server_connection *websocket.Conn, stop_spam_flag *bool, start_spam_flag *bool) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Error main StopSpam - ", err)
			MessageToServer(server_connection, map[string]interface{}{"action": "Error", "data": "Error: StopSpam"})
		}
	}()
	if *start_spam_flag {
		for i := 0; i < len(*anchors_array); i++ {
			Anchor.StopSpam((*anchors_array)[i], server_connection)
		}
	}
	*stop_spam_flag = true
	*start_spam_flag = false
}

func MessageToServer(server_connection *websocket.Conn, map_message map[string]interface{}) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Error: MessageToServer - ", err)
		}
	}()
	json_message, _ := json.Marshal(map_message)
	server_connection.WriteMessage(websocket.TextMessage, json_message)
	fmt.Println("Message to server: " + string(json_message))
}

func MessageToMath(math_connection *websocket.Conn, map_message map[string]interface{}) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Error: MessageToServer - ", err)
		}
	}()
	json_message, _ := json.Marshal(map_message)
	math_connection.WriteMessage(websocket.TextMessage, json_message)
	fmt.Println("Message to math: " + string(json_message))
}
