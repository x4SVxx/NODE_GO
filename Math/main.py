import threading
from ctypes import wstring_at
import json
import time
import websocket
import time
from cle import Cle
from config import Config
from tag import Tag
import socket
import numpy as np


def process_config(msg, CLE, cfg):
    cle = get_current_cle(msg, CLE)
    if cle:
        CLE.remove(cle)
    cle = Cle(msg, cfg)
    print("New CLE")    
    CLE.append(cle)

def process_TX(msg, cle, ws):
    for anchor in cle.anchors:
        if msg["data"]["receiver"] == anchor.master_ID:
            anchor.add_tx(msg)
            if anchor.data2sendflag:
                data2send = get_anchor_info(anchor)
                print('Process_TX data2send ', data2send)
                ws.send(json.dumps({'action':'ECHO','data':data2send,'apikey':apikey}))
                anchor.data2sendflag = 0

def process_RX(msg, cle, ws):
        for anchor in cle.anchors:
            if msg["data"]["receiver"] == anchor.ID and msg["data"]["sender"] == anchor.master_ID:
                anchor.add_rx(msg)
                if anchor.data2sendflag:
                    data2send = get_anchor_info(anchor)
                    print('Process_RX data2send ', data2send)
                    ws.send(json.dumps({'action':'ECHO','data':data2send,'apikey':apikey}))
                    anchor.data2sendflag = 0
                break

def process_BLINK(msg, cle, ws):
    
    
    for anchor in cle.anchors:
        if msg["data"]["receiver"] == anchor.ID:
            break

    if cle.use_ref_tag and cle.ref_tag:
        ID = msg["data"]["sender"]
        if ID == cle.ref_tag["name"]:
            for anchor in cle.anchors:
                if msg["data"]["receiver"] == anchor.master_ID:
                    anchor.add_master_ref_tag(msg)
                    if anchor.data2sendflag:
                        data2send = get_anchor_info(anchor)
                        ws.send(json.dumps({'action':'ECHO','data':data2send,'apikey':apikey}))
                        anchor.data2sendflag = 0
            for anchor in cle.anchors:
                if msg["data"]["receiver"] == anchor.ID:
                    anchor.add_slave_ref_tag(msg)
                    if anchor.data2sendflag:
                        data2send = get_anchor_info(anchor)
                        ws.send(json.dumps({'action':'ECHO','data':data2send,'apikey':apikey}))
                        anchor.data2sendflag = 0
                    break
            return

    if anchor.sync_flag:
        msg["data"]["anchor_number"] = anchor.number
        msg["data"]["corrected_timestamp"] = anchor.correct_timestamp(msg["data"]["timestamp"])
        if cle.use_ref_tag and cle.ref_tag:
            msg["data"]["corrected_timestamp"] += anchor.ref_tag_correction
        match_flag = 0
        for tag in cle.tags:
            if msg["data"]["sender"] == tag.ID:
                match_flag = 1
                break

        if match_flag == 0:
            tag = Tag(msg, cle)
            print(f"New tag {tag.ID}")
            cle.tags.append(tag)
        tag.add_data(msg)
        if tag.data2sendflag:
            data2send = get_tag_info(tag, cle)
            print('Process_BLINK data2send ', data2send)
            ws.send(json.dumps({'action':'ECHO','data':data2send,'apikey':apikey}))
            tag.data2sendflag = 0

def get_current_cle(msg, CLE):
    cle = []
    for cle in CLE:
        if cle.organization == msg["data"]["organization"] and cle.roomid == msg["data"]["roomid"]:
            break
    return cle

def get_tag_info(tag, cle):
    tag_info = {}
    tag_info["type"] = "tag"
    tag_info["ID"] = tag.ID

    tag.x_f = tag.x_f_past * cle.cfg.alpha + tag.x * (1 - cle.cfg.alpha)
    tag.y_f = tag.y_f_past * cle.cfg.alpha + tag.y * (1 - cle.cfg.alpha)
    tag.x_f_past = tag.x_f
    tag.y_f_past = tag.y_f
    tag_info["x"] = tag.x_f
    tag_info["y"] = tag.y_f
    tag_info["z"] = tag.h
    tag_info["dop"] = tag.DOP
    tag_info["lifetime"] = tag.lifetime
    tag_info["time"] = time.strftime("%d.%m.%Y %H:%M:%S", time.localtime(tag.lasttime))
    tag_info["anchors"] = tag.anchors_number_to_solve
    tag_info["static"] = tag.static
    tag_info["SOS"] = tag.SOS
    tag_info["organization"] = tag.organization
    tag_info["roomid"] = tag.roomid
    return tag_info

def get_anchor_info(anchor):
    anchor_info = {}
    anchor_info["type"] = "anchor"
    anchor_info["number"] = anchor.number
    anchor_info["ID"] = anchor.ID
    anchor_info["master_ID"] = anchor.master_ID
    anchor_info["role"] = anchor.Role
    anchor_info["x"] = anchor.x
    anchor_info["y"] = anchor.y
    anchor_info["z"] = anchor.z
    anchor_info["delta"] = float(anchor.X[0])
    anchor_info["delta_t"] = float(anchor.X[1])
    anchor_info["sqrt(Dx[1,1])"] = np.sqrt(float(anchor.Dx[0][0]))
    anchor_info["sqrt(Dx[2,2])"] = np.sqrt(float(anchor.Dx[1][1]))
    anchor_info["current_tx"] = float(anchor.current_tx)
    anchor_info["current_rx"] = float(anchor.current_rx)
    anchor_info["k_skip"] = int(anchor.k_skip)
    anchor_info["sync_flag"] = anchor.sync_flag
    anchor_info["ref_tag_correction"] = anchor.ref_tag_correction
    anchor_info["organization"] = anchor.organization
    anchor_info["roomid"] = anchor.roomid
    return anchor_info

def on_message(ws, json_message):
    print(json_message)
    message = json.loads(json_message)
    if message["action"] == "RoomConfig":
        process_config(message, CLE, cfg)
    elif message["action"] == "CS_TX":
        cle = get_current_cle(message, CLE)
        if cle:
            process_TX(message, cle, ws)
    elif message["action"] == "CS_RX":
        cle = get_current_cle(message, CLE)
        if cle:
            process_RX(message, cle, ws)
    elif message["action"] == "BLINK":
        cle = get_current_cle(message, CLE)
        if cle:
            process_BLINK(message, cle, ws)
    elif message["action"] == "Login":
        global apikey
        apikey = message["apikey"]
        print("Login succsess")


def on_error(ws, error):
    print(error)

def on_close(ws):
    ws.close()

def on_open(ws):
    ws.send("{\"action\":\"Login\",\"login\":\"mathLogin\",\"password\":\"%wPp7VO6k7ump{BP4mu2rm4w?p|J5N%P\",\"roomid\":\"1\"}")


if __name__ == "__main__":
    cfg = Config()
    CLE = []
    apikey = ""
    websocket.enableTrace(False)
    input = websocket.WebSocketApp("ws://127.0.0.1:8000",
                                on_message = on_message,
                                on_error = on_error,
                                on_close = on_close)
    input.on_open = on_open
    input.run_forever()