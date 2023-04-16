package ReadAndSetNodeConfig

import (
	"NODE/Logger"
	"encoding/json"
	"os"
	"strings"
)

func ReadAndSetNodeConfig() (string, string, string, string, string, string, string, string, string, map[string]interface{}) {
	defer func() {
		if err := recover(); err != nil {
			Logger.Logger("ERROR : ReadAndSetConfig", err)
			os.Exit(0)
		}
	}()
	var node_config map[string]interface{}
	file, _ := os.ReadFile("NodeConfig.json")
	json.Unmarshal(file, &node_config)

	server_ip := node_config["server_ip"].(string)
	server_port := node_config["server_port"].(string)
	login := node_config["login"].(string)
	password := node_config["password"].(string)
	roomid := node_config["roomid"].(string)
	independent_flag := node_config["independent_flag"].(string)
	connect_math_flag := node_config["connect_math_flag"].(string)
	node_server_ip := node_config["node_server_ip"].(string)
	node_server_port := node_config["node_server_port"].(string)
	log_enable_flag := node_config["log_enable_flag"].(string)
	ref_tag_config := node_config["ref_tag_config"].(map[string]interface{})
	ref_tag_config_json, _ := json.Marshal(ref_tag_config)

	if len(strings.TrimSpace(server_ip)) == 0 ||
		len(strings.TrimSpace(server_port)) == 0 ||
		len(strings.TrimSpace(login)) == 0 ||
		len(strings.TrimSpace(password)) == 0 ||
		len(strings.TrimSpace(roomid)) == 0 ||
		len(strings.TrimSpace(independent_flag)) == 0 ||
		len(strings.TrimSpace(connect_math_flag)) == 0 ||
		len(strings.TrimSpace(node_server_ip)) == 0 ||
		len(strings.TrimSpace(node_server_port)) == 0 ||
		len(strings.TrimSpace(log_enable_flag)) == 0 {
		Logger.Logger("ERROR : ReadAndSetNodeConfig - empty data in NodeConfig", nil)
		os.Exit(0)
	} else {
		Logger.Logger("SUCCESS : ReadAndSetNodeConfig \n"+
			"--- server_ip: "+string(server_ip)+" ---"+"\n"+
			"--- server_port: "+string(server_port)+" ---"+"\n"+
			"--- login: "+string(login)+" ---"+"\n"+
			"--- password: "+string(password)+" ---"+"\n"+
			"--- roomid: "+string(roomid)+" ---"+"\n"+
			"--- connect_math_flag: "+string(connect_math_flag)+" ---"+"\n"+
			"--- independent_flag: "+string(independent_flag)+" ---"+"\n"+
			"--- node_server_ip: "+string(node_server_ip)+" ---"+"\n"+
			"--- node_server_port: "+string(node_server_port)+" ---"+"\n"+
			"--- log_enable_flag: "+string(log_enable_flag)+" ---"+"\n"+
			"--- ref_tag_config: "+string(ref_tag_config_json)+" ---"+"\n", nil)
	}
	return server_ip, server_port, login, password, roomid, independent_flag, connect_math_flag, node_server_ip, node_server_port, ref_tag_config
}
