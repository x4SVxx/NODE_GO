package Anchor

import (
	"NODE/ReportsAndMessages"
	"encoding/json"
	"fmt"
	"net"

	"github.com/gorilla/websocket"
)

func SendToMath(message map[string]interface{}, apikey string, name string, clientid string, roomid string, math_connection *websocket.Conn, server_connection *websocket.Conn) {
	if message["type"] != "Unknow" {
		if math_connection != nil {
			math_map_message := map[string]interface{}{
				"action": message["type"],
				"data": map[string]interface{}{
					"apikey":       apikey,
					"orgname":      name,
					"organization": clientid,
					"roomid":       roomid,
					"type":         message["type"],
					"timestamp":    message["timestamp"],
					"receiver":     message["receiver"],
					"sender":       message["sender"],
				},
			}
			if message["type"] == "CS_RX" || message["type"] == "CS_TX" {
				math_map_message["data"].(map[string]interface{})["seq"] = message["seq"]
			}
			if message["type"] == "BLINK" {
				math_map_message["data"].(map[string]interface{})["sn"] = message["sn"]
				math_map_message["data"].(map[string]interface{})["state"] = message["state"]

			}
			json_math_message, _ := json.Marshal(math_map_message)
			fmt.Println(string(json_math_message))
			math_connection.WriteMessage(websocket.TextMessage, json_math_message)
		}
		if server_connection != nil {
			math_map_message := map[string]interface{}{
				"action":       "SendToMath",
				"apikey":       apikey,
				"orgname":      name,
				"organization": clientid,
				"roomid":       roomid,
				"type":         message["type"],
				"timestamp":    message["timestamp"],
				"receiver":     message["receiver"],
				"sender":       message["sender"],
			}
			if message["type"] == "CS_RX" || message["type"] == "CS_TX" {
				math_map_message["seq"] = message["seq"]
			}
			if message["type"] == "BLINK" {
				math_map_message["sn"] = message["sn"]
			}
			json_math_message, _ := json.Marshal(math_map_message)
			fmt.Println(string(json_math_message))
			server_connection.WriteMessage(websocket.TextMessage, json_math_message)
		}

	}
}

func Handler(apikey string, name string, clientid string, roomid string, anchor map[string]interface{}, server_connection *websocket.Conn, math_connection *websocket.Conn) {
	break_point := false
	for {
		if break_point {
			fmt.Println("BREAK HANDLER")
			return
		}
		func() {
			defer func() {
				if err := recover(); err != nil {
					fmt.Println("Error: AnchorHandler -", err)
					MessageToServer(map[string]interface{}{"action": "Error", "data": "Error: AnchorHandler"}, server_connection)
					break_point = true
				}
			}()
			buffer_header := make([]byte, 3)
			anchor["connection"].(net.Conn).Read(buffer_header)
			number_of_bytes := buffer_header[1]
			buffer_anchor_message := make([]byte, number_of_bytes)
			anchor["connection"].(net.Conn).Read(buffer_anchor_message)
			buffer_ending := make([]byte, 3)
			anchor["connection"].(net.Conn).Read(buffer_ending)
			message := ReportsAndMessages.DecodeAnchorMessage(buffer_anchor_message)
			message["receiver"] = anchor["id"].(string)
			if message["type"] == "CS_TX" {
				message["sender"] = message["receiver"]
			}
			SendToMath(message, apikey, name, clientid, roomid, math_connection, server_connection)
		}()
	}
}

func Connect(anchor *map[string]interface{}, anchors_array *[]map[string]interface{}, server_conn ...*websocket.Conn) {
	var server_connection *websocket.Conn
	if len(server_conn) > 0 {
		server_connection = server_conn[0]
	}
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Error: AnchorConnect -", err)
			MessageToServer(map[string]interface{}{"action": "Error", "data": "Error: AnchorConnect"}, server_connection)
		}
	}()
	anchor_connection, err := net.Dial("tcp", (*anchor)["ip"].(string)+":"+"3000")
	if err != nil {
		fmt.Println("Error:", err)
		MessageToServer(map[string]interface{}{"action": "Error", "data": "Error: AnchorConnect " + (*anchor)["ip"].(string)}, server_connection)
	} else {
		buffer_skip := make([]byte, 3)
		anchor_connection.Read(buffer_skip)
		buffer_anchor_connect := make([]byte, 500)
		anchor_connection.Read(buffer_anchor_connect)
		(*anchor)["connection"] = anchor_connection
		(*anchor)["id"] = ReportsAndMessages.DecodeAnchorMessage(buffer_anchor_connect)["receiver"].(string)
		*anchors_array = append(*anchors_array, (*anchor))
		fmt.Println("Success: AnchorConnect " + (*anchor)["ip"].(string))
		fmt.Println(ReportsAndMessages.DecodeAnchorMessage(buffer_anchor_connect))
		MessageToServer(map[string]interface{}{"action": "Success", "data": "Sucess: AnchorConnect " + (*anchor)["ip"].(string), "more": ReportsAndMessages.DecodeAnchorMessage(buffer_anchor_connect)}, server_connection)
	}
}

func DisConnect(anchors_array *[]map[string]interface{}, server_conn ...*websocket.Conn) {
	var server_connection *websocket.Conn
	if len(server_conn) > 0 {
		server_connection = server_conn[0]
	}
	for i := 0; i < len(*anchors_array); i++ {
		(*anchors_array)[i]["connection"].(net.Conn).Close()
		fmt.Println("Anchor disconnect")
		MessageToServer(map[string]interface{}{"action": "Success", "data": "Sucess: AnchorDisConnect " + (*anchors_array)[i]["ip"].(string)}, server_connection)
	}
	*anchors_array = []map[string]interface{}{}
}

func SetRfConfig(anchor map[string]interface{}, rf_config map[string]interface{}, server_conn ...*websocket.Conn) {
	var server_connection *websocket.Conn
	if len(server_conn) > 0 {
		server_connection = server_conn[0]
	}
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Error: AnchorSetRfConfig -", err)
			MessageToServer(map[string]interface{}{"action": "Error", "data": "Error: AnchorSetRfConfig"}, server_connection)
		}
	}()
	var PRF map[int]int = map[int]int{
		16: 1,
		64: 2,
	}
	var DATARATE map[float64]int = map[float64]int{
		110: 0,
		850: 1,
		6.8: 2,
	}
	var PREAMBLE_LEN map[int]int = map[int]int{
		64:   int(0x04),
		128:  int(0x14),
		256:  int(0x24),
		512:  int(0x34),
		1024: int(0x08),
		1536: int(0x18),
		2048: int(0x28),
		4096: int(0x0C),
	}
	var PAC map[int]int = map[int]int{
		8:  0,
		16: 1,
		32: 2,
		64: 3,
	}
	var anchor_role int
	if anchor["role"].(string) == "Master" {
		anchor_role = 1
	} else if anchor["role"].(string) == "Slave" {
		anchor_role = 0
	}
	RTLS_CMD_SET_CFG_CCP := ReportsAndMessages.Build_RTLS_CMD_SET_CFG_CCP(
		int(anchor_role),
		int(rf_config["chnum"].(float64)),
		int(PRF[int(rf_config["prf"].(float64))]),
		int(DATARATE[rf_config["datarate"].(float64)]),
		int(rf_config["preamblecode"].(float64)),
		int(PREAMBLE_LEN[int(rf_config["preamblelen"].(float64))]),
		int(PAC[int(rf_config["pac"].(float64))]),
		int(rf_config["nsfd"].(float64)),
		int(anchor["adrx"].(float64)),
		int(anchor["adtx"].(float64)),
		int(rf_config["diagnostic"].(float64)),
		int(rf_config["lag"].(float64)))

	anchor["connection"].(net.Conn).Write(RTLS_CMD_SET_CFG_CCP)
	fmt.Println("SetRfConfig on the anchor " + anchor["ip"].(string))
	MessageToServer(map[string]interface{}{"action": "Success", "data": "Success: SetRfConfig on the anchor " + anchor["ip"].(string)}, server_connection)
}

func StartSpam(anchor map[string]interface{}, server_conn ...*websocket.Conn) {
	var server_connection *websocket.Conn
	if len(server_conn) > 0 {
		server_connection = server_conn[0]
	}
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Error: AnchorStartSpam -", err)
			MessageToServer(map[string]interface{}{"action": "Error", "data": "Error: AnchorStartSpam"}, server_connection)
		}
	}()
	anchor["connection"].(net.Conn).Write(ReportsAndMessages.Build_RTLS_START_REQ(1))
	fmt.Println("AnchorStartSpam " + anchor["ip"].(string))
	MessageToServer(map[string]interface{}{"action": "Success", "data": "Sucess: AnchorStartSpam " + anchor["ip"].(string)}, server_connection)
}

func StopSpam(anchor map[string]interface{}, server_conn ...*websocket.Conn) {
	var server_connection *websocket.Conn
	if len(server_conn) > 0 {
		server_connection = server_conn[0]
	}
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Error: AnchorStopSpam -", err)
			MessageToServer(map[string]interface{}{"action": "Error", "data": "Error: AnchorStopSpam"}, server_connection)
		}
	}()
	anchor["connection"].(net.Conn).Write(ReportsAndMessages.Build_RTLS_START_REQ(0))
	fmt.Println("AnchorStopSpam " + anchor["ip"].(string))
	MessageToServer(map[string]interface{}{"action": "Success", "data": "Sucess: AnchorStopSpam " + anchor["ip"].(string)}, server_connection)
}

func MessageToServer(map_message map[string]interface{}, server_conn ...*websocket.Conn) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Error: AnchorMessageToServer: ", err)
		}
	}()

	var server_connection *websocket.Conn
	if len(server_conn) > 0 {
		server_connection = server_conn[0]
		if server_connection != nil {
			json_message, _ := json.Marshal(map_message)
			server_connection.WriteMessage(websocket.TextMessage, json_message)
		}
	}
}
