package pinba

import (
	"bytes"
	"code.google.com/p/goprotobuf/proto"
	"compress/zlib"
	"sort"
	"strconv"
	"strings"
)

/**
 * ProtoBuf pinba message struct
 */
type Request struct {
	Hostname         *string   `protobuf:"bytes,1,req,name=hostname"`
	ServerName       *string   `protobuf:"bytes,2,req,name=server_name"`
	ScriptName       *string   `protobuf:"bytes,3,req,name=script_name"`
	RequestCount     *uint32   `protobuf:"varint,4,req,name=request_count"`
	DocumentSize     *uint32   `protobuf:"varint,5,req,name=document_size"`
	MemoryPeak       *uint32   `protobuf:"varint,6,req,name=memory_peak"`
	RequestTime      *float32  `protobuf:"fixed32,7,req,name=request_time"`
	RuUtime          *float32  `protobuf:"fixed32,8,req,name=ru_utime"`
	RuStime          *float32  `protobuf:"fixed32,9,req,name=ru_stime"`
	TimerHitCount    []uint32  `protobuf:"varint,10,rep,name=timer_hit_count"`
	TimerValue       []float32 `protobuf:"fixed32,11,rep,name=timer_value"`
	TimerTagCount    []uint32  `protobuf:"varint,12,rep,name=timer_tag_count"`
	TimerTagName     []uint32  `protobuf:"varint,13,rep,name=timer_tag_name"`
	TimerTagValue    []uint32  `protobuf:"varint,14,rep,name=timer_tag_value"`
	Dictionary       []string  `protobuf:"bytes,15,rep,name=dictionary"`
	Status           *uint32   `protobuf:"varint,16,opt,name=status"`
	XXX_unrecognized []byte    ``
    XXX_extensions   []RawMetric ``
}

/**
 * Request Metric
 *
 * unused?
 */
func (m *Request) GetMetric() RawMetric {
	tags := Tags{"host": *m.Hostname, "server_name": *m.ServerName, "script": *m.ScriptName}
	return RawMetric{Count: 1, Value: *m.RequestTime, Tags: tags, Id: tags.String()}
}

/**
 * Request Timers as RawMetrics
 */
func (m *Request) Timers() []RawMetric {
	if len(m.XXX_extensions) == len(m.TimerValue) {
        return m.XXX_extensions
    }
    m.XXX_extensions = make([]RawMetric, len(m.TimerValue))
	var offset uint32 = 0
	for idx, val := range m.TimerValue {
		m.XXX_extensions[idx] = RawMetric{Count: m.TimerHitCount[idx], Value: val}
		m.XXX_extensions[idx].Tags = Tags{"host": *m.Hostname, "server_name": *m.ServerName, "script": *m.ScriptName}
		for k, v := range m.TimerTagName[offset : offset+m.TimerTagCount[idx]] {
			m.XXX_extensions[idx].Tags[m.Dictionary[v]] = m.Dictionary[m.TimerTagValue[int(offset)+k]]
		}
		m.XXX_extensions[idx].Id = m.XXX_extensions[idx].Tags.String()
		offset += m.TimerTagCount[idx]
	}
	return m.XXX_extensions
}

func (m *Request) Tags() Tags {
	return Tags{"host": *m.Hostname, "server_name": *m.ServerName, "script": *m.ScriptName}
}

func (m *Request) Reset() {
	*m = Request{}
}

func (m *Request) String() string {
	return proto.CompactTextString(m)
}

func (*Request) ProtoMessage() {

}

type RawMetric struct {
	Id    string
	Tags  Tags
	Count uint32
	Value float32
}

type Tags map[string]string

/**
 * Tags.String() - from map to k:v string
 */
func (t Tags) String() (result string) {
	keys := make([]string, len(t))
	i := 0
	for k, _ := range t {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	for _, key := range keys {
		result += key + ":" + t[key]
	}
	return result
}

func Decode(content *[]byte) (int, []Request) {
	var out bytes.Buffer
	input := string(*content)
	ts, _ := strconv.Atoi(input[0:10])
	reader, _ := zlib.NewReader(strings.NewReader(input[11:]))
	out.ReadFrom(reader)
	reader.Close()

	// Looks like there is no data
	if len(out.Bytes()) == 0 {
		return ts, make([]Request, 0)
	}

	raw_pb := bytes.Split(out.Bytes(), []byte{0xa, 0x2d, 0x2d, 0xa})
	requests := make([]Request, len(raw_pb))
	for idx, r := range raw_pb {
		tmp := &Request{}
		proto.Unmarshal(r, tmp)
		requests[idx] = *tmp
	}
	return ts, requests
}
