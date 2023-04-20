from anchor import Anchor

class Cle():

    def __init__(self, msg, cfg):
        self.tags = []
        self.anchors = []
        self.cfg = cfg
        self.organization = msg["data"]["organization"]
        self.roomid = msg["data"]["roomid"]

        # reference tag
        if 'data' in msg:
            if 'ref_tag_config' in msg['data']
                self.ref_tag = msg["data"]["ref_tag_config"]
        if self.ref_tag:
            self.use_ref_tag = True
        else:
            self.use_ref_tag = False

        for data in msg["data"]["anchors"]:
            anchor = Anchor(data, cfg, self)
            self.anchors.append(anchor)

        for anchor in self.anchors:
            anchor.relate_to_master(self.anchors, self.cfg, self.use_ref_tag, self.ref_tag)

