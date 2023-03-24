import clelib as cl
import time

class Tag():
    def __init__(self, msg, cle):
        self.organization = cle.organization
        self.roomid = cle.roomid
        self.cfg = cle.cfg
        self.cle = cle
        self.ID = msg['data']["sender"]
        self.SN = -1
        self.measurements = []
        self.x = 0.0
        self.y = 0.0
        self.h = cle.cfg.hei
        self.DOP = 1.
        self.starttime = time.time()
        self.lasttime = time.time()
        self.lifetime = 0.
        self.state = 0
        self.alpha = 0.5
        self.data2sendflag = 0
        self.anchors_number_to_solve = 0
        self.data2log = None
        self.x_buffer = []
        self.y_buffer = []
        self.buffer_length = cle.cfg.buffer_length
        self.accumulation_mode = cle.cfg.accumulation_mode

    def add_data(self, msg):
        if time.time() - self.lasttime > 10:
            self.x_buffer = []
            self.y_buffer = []
            self.measurements = []
            self.SN = -1
            self.lasttime = time.time()
        if msg['data']["sn"] != self.SN:
            delta = msg['data']["sn"] - self.SN
            if delta < -240:
                delta += 255
            if (delta > 0) or self.SN < 0:
                self.SN = msg['data']["sn"]
            else:
                return
            self.measurements = cl.check_PD(self.measurements, self.cfg)
            flag = cl.coords_calc_2D(self)
            if flag:
                # self.data2send = f"{self.name}\t{round(self.x, 2)}\t{round(self.y, 2)}\t{len(self.measurements)}"
                if self.accumulation_mode:
                    if time.time() - self.lasttime > 1:
                        self.data2sendflag = 1
                        self.x_buffer = []
                        self.y_buffer = []
                    else:
                        self.x_buffer.append(self.x)
                        self.y_buffer.append(self.y)
                        if len(self.x_buffer) == self.buffer_length:
                            sorted_x = sorted(self.x_buffer)
                            sorted_y = sorted(self.y_buffer)
                            self.x = sorted_x[round(self.buffer_length / 2)]
                            self.y = sorted_y[round(self.buffer_length / 2)]
                            self.data2sendflag = 1
                            self.x_buffer = []
                            self.y_buffer = []
                else:
                    self.data2sendflag = 1

                self.lasttime = time.time()
                self.lifetime = self.lasttime - self.starttime

                self.anchors_number_to_solve = len(self.measurements)

            else:
                flag = 0
            if self.cfg.log:
                self.log_cle_tag(flag)
            self.measurements = []
        self.measurements.append(msg)








