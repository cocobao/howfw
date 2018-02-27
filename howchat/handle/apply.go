package handle

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"net"

	"github.com/cocobao/log"
)

func ApplyService(addr string) string {
	c, err := net.Dial("tcp", addr)
	if err != nil {
		log.Warn(err)
		return ""
	}

	defer c.Close()
	d, err := json.Marshal(map[string]interface{}{
		"cmd": "apply",
	})

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, int32(len(d)))
	buf.Write(d)
	pkt := buf.Bytes()

	c.Write(pkt)

	lengthBytes := make([]byte, 4)
	_, err = io.ReadFull(c, lengthBytes)
	if err != nil {
		if io.EOF == err {
			log.Warn("conn has been close by peer")
			return ""
		}
		log.Warn("read length bytes fail", err)
		return ""
	}

	if len(lengthBytes) == 0 {
		log.Warn("length bytes is 0")
		return ""
	}

	lengthBuf := bytes.NewReader(lengthBytes)
	var msgLen uint32
	if err = binary.Read(lengthBuf, binary.LittleEndian, &msgLen); err != nil {
		log.Warn("lengthBuf read fail", err)
		return ""
	}

	var rd = make([]byte, msgLen)
	l, err := c.Read(rd)
	if err != nil {
		log.Warn(err)
		return ""
	}
	var m map[string]interface{}
	err = json.Unmarshal(rd, &m)
	if err != nil {
		log.Warn(err)
		return ""
	}
	log.Debug(l, m)

	if v, ok := m["result"].(map[string]interface{}); ok {
		if vv, ok := v["saddr"].(string); ok {
			return vv
		}
	}

	return ""
}
