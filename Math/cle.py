from anchor import Anchor

class Cle():

    def __init__(self, msg, cfg):
        self.tags = []
        self.anchors = []
        self.cfg = cfg
        self.organization = msg["data"]["organization"]
        self.roomid = msg["data"]["roomid"]

        for data in msg["data"]["anchors"]:
            anchor = Anchor(data, cfg, self)
            self.anchors.append(anchor)

        for anchor in self.anchors:
            anchor.relate_to_master(self.anchors, cfg)

