package Logger

import (
	"encoding/json"
	"fmt"
	"os"

	// "runtime"
	"strings"
	"time"
)

var txt_file_for_logger *os.File
var txt_file_count = 0
var time_last_logger_file int64
var log_enable_flag = ReadAndSetConfig()

func ReadAndSetConfig() string {
	var node_config map[string]interface{}
	file, _ := os.ReadFile("NodeConfig.json")
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Error: ", err)
			os.Exit(0)
		}
	}()
	json.Unmarshal(file, &node_config)
	log_enable_flag := node_config["log_enable_flag"].(string)
	if len(strings.TrimSpace(log_enable_flag)) == 0 {
		fmt.Println("Error: Empry data in NodeConfig")
		os.Exit(0)
	}
	return log_enable_flag
}

func Logger(map_message string, err interface{}) {
	if log_enable_flag == "true" {
		time_now := time.Now().Unix()
		if txt_file_count == 0 || time_now-time_last_logger_file > 60*60 {
			time_last_logger_file = time.Now().Unix()
			now_time := strings.Split(time.Now().String(), " ")
			now_time_data := strings.Replace(strings.Split(now_time[0]+" "+now_time[1], ".")[0], ":", "-", 2)
			txt_file_for_logger, _ = os.Create("Logs/" + now_time_data + ".txt")
			txt_file_count += 1
		}
		txt_file_for_logger.WriteString(map_message + "\n")
		fmt.Println(map_message)
		// fmt.Println(runtime.NumGoroutine())
		if err != nil {
			fmt.Println(err)
			txt_file_for_logger.WriteString(err.(error).Error() + "\n")
		}
	}
	if log_enable_flag == "false" {
		fmt.Println(map_message)
		if err != nil {
			fmt.Println(err)
		}
	}
}
