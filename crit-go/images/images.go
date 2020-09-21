package images

import (
	log "crit-go/logging"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"strconv"
)

func closefile(f *os.File) {
	f.Close()
	os.Exit(1)
}

const sizeof_u16 = 2
const sizeof_u32 = 4
const sizeof_u64 = 8

var jsonmap map[string]interface{}
var bynamemap map[string]interface{}
var byvalmap map[string]interface{}

func Load(f *os.File, pretty bool, nopl bool) map[string]interface{} {
	image := make(map[string]interface{})
	/*
	   Convert criu image from binary format to dict(json).
	   Takes a file-like object to read criu image from.
	   Returns criu image in dict(json) format.
	*/
	magicjson()
	m := magichandler(f)
	image["magic"] = m
	switch {
	case m == "GHOST_FILE":
		handler := &ghost_file_handler{}
		handler.m = m
		image["entries"] = handler.Load(f, pretty, nopl)
	case m == "PAGEMAP":
		handler := &pagemap_handler{}
		handler.m = m
		image["entries"] = handler.Load(f, pretty, nopl)
	case m == "PIPES_DATA":
		handler := &pipes_data_extra_handler{}
		handler.m = m
		image["entries"] = handler.Load(f, pretty, nopl)
	case m == "FIFO_DATA":
		handler := &pipes_data_extra_handler{}
		handler.m = m
		image["entries"] = handler.Load(f, pretty, nopl)
	case m == "SK_QUEUES":
		handler := &sk_queues_extra_handler{}
		handler.m = m
		image["entries"] = handler.Load(f, pretty, nopl)
	case m == "IPCNS_SHM":
		handler := &ipc_shm_set_handler{}
		handler.m = m
		image["entries"] = handler.Load(f, pretty, nopl)
	case m == "IPCNS_SEM":
		handler := &ipc_sem_set_handler{}
		handler.m = m
		image["entries"] = handler.Load(f, pretty, nopl)
	case m == "IPCNS_MSG":
		handler := &ipc_msg_queue_handler{}
		handler.m = m
		image["entries"] = handler.Load(f, pretty, nopl)
	case m == "TCP_STREAM":
		handler := &tcp_stream_extra_handler{}
		handler.m = m
		image["entries"] = handler.Load(f, pretty, nopl)
	default:
		handler := &entry_handler{}
		handler.m = m
		image["entries"] = handler.Load(f, pretty, nopl)
	}
	f.Close()
	return image
}

func Dump(ilf *os.File, ouf *os.File) {
	/*
	   Convert criu image from dict(json) format to binary.
	   Takes an image in dict(json) format and file-like
	   object to write to.
	*/
	magicjson()
	var img map[string]interface{}
	file, err := ioutil.ReadAll(ilf)
	if err != nil {
		log.Error("Error reading Input Json file: ", err)
		ilf.Close()
		ouf.Close()
		os.Exit(1)
	}
	if err := json.Unmarshal(file, &img); err != nil {
		log.Error("Unmarshalling Json file: ", err)
		ilf.Close()
		ouf.Close()
		os.Exit(1)
	}
	ilf.Close()
	m := img["magic"].(string)
	magic_val, err := strconv.ParseUint(bynamemap[m].(string), 10, 64)
	if err != nil {
		log.Error("Unable to Parse Magic value: ", err)
		closefile(ouf)
	}
	if m != "INVENTORY" {
		if m == "STATS" || m == "IRMAP_CACHE" {
			bnvmap, err := strconv.ParseUint(bynamemap["IMG_SERVICE"].(string), 10, 64)
			if err != nil {
				log.Error("Magic Error: ", err)
				closefile(ouf)
			}
			bs := make([]byte, 4)
			binary.LittleEndian.PutUint32(bs, uint32(bnvmap))
			_, err = ouf.Write(bs)
			if err != nil {
				log.Error("Error writing to file: ", err)
				closefile(ouf)
			}
		} else {
			bnvmap, err := strconv.ParseUint(bynamemap["IMG_COMMON"].(string), 10, 64)
			if err != nil {
				log.Error("Error in Magic Value Conversion: ", err)
				closefile(ouf)
			}
			bs := make([]byte, 4)
			binary.LittleEndian.PutUint32(bs, uint32(bnvmap))
			_, err = ouf.Write(bs)
			if err != nil {
				log.Error("Error writing to file: ", err)
				closefile(ouf)
			}
		}
	}
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(magic_val))
	_, err = ouf.Write(bs)
	if err != nil {
		log.Error("Error writing to file: ", err)
		closefile(ouf)
	}
	switch {
	case m == "GHOST_FILE":
		handler := &ghost_file_handler{}
		handler.m = m
		handler.Dump(img, ouf)
	case m == "PAGEMAP":
		handler := &pagemap_handler{}
		handler.m = m
		handler.Dump(img, ouf)
	case m == "PIPES_DATA":
		handler := &pipes_data_extra_handler{}
		handler.m = m
		handler.Dump(img, ouf)
	case m == "FIFO_DATA":
		handler := &pipes_data_extra_handler{}
		handler.m = m
		handler.Dump(img, ouf)
	case m == "SK_QUEUES":
		handler := &sk_queues_extra_handler{}
		handler.m = m
		handler.Dump(img, ouf)
	case m == "IPCNS_SHM":
		handler := &ipc_shm_set_handler{}
		handler.m = m
		handler.Dump(img, ouf)
	case m == "IPCNS_SEM":
		handler := &ipc_sem_set_handler{}
		handler.m = m
		handler.Dump(img, ouf)
	case m == "IPCNS_MSG":
		handler := &ipc_msg_queue_handler{}
		handler.m = m
		handler.Dump(img, ouf)
	case m == "TCP_STREAM":
		handler := &tcp_stream_extra_handler{}
		handler.m = m
		handler.Dump(img, ouf)
	default:
		handler := &entry_handler{}
		handler.m = m
		handler.Dump(img, ouf)
	}

}

func magicjson() {
	/*
		Populates the byname and byval magic maps
	*/
	jsondata, err := ioutil.ReadFile("./magic-gen/magic.json")
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	err = json.Unmarshal(jsondata, &jsonmap)
	if err != nil {
		log.Fatal("Error Reading from Magic.Json,Exitng: ", err)
	}
	bynamemap = jsonmap["byname"].(map[string]interface{})
	byvalmap = jsonmap["byval"].(map[string]interface{})
}

func magichandler(f *os.File) string {
	/*
		Fetches the magic of a .img file for the
		load function.
	*/
	bufl := make([]byte, 4)
	_, err := f.Read(bufl)
	if err != nil {
		log.Error("Error Reading from Magic.Json,Exitng: ", err)
		closefile(f)
	}
	img_magic := strconv.FormatUint(uint64(binary.LittleEndian.Uint32(bufl)), 10)
	_, err = f.Seek(4, 0)
	if err != nil {
		log.Error("Error Reading from Magic.Json,Exitng: ", err)
		closefile(f)
	}
	bufl = make([]byte, 4)
	_, err = f.Read(bufl)
	if err != nil {
		log.Error("Error Reading from Magic from Input file,Exitng: ", err)
		closefile(f)
	}
	if img_magic == bynamemap["IMG_COMMON"] || img_magic == bynamemap["IMG_SERVICE"] {
		img_magic = strconv.FormatUint(uint64(binary.LittleEndian.Uint32(bufl)), 10)
	}
	m := byvalmap[img_magic]
	if m == nil {
		log.Error(`Unknown magic,Maybe you are feeding me an image with raw data(i.e. pages.img)?`)
		closefile(f)
	}
	return m.(string)
}

func roundup(size uint32, sizeof uint32) uint32 {
	round := ((size - 1) | (sizeof - 1) + 1)
	return round
}

func parseInt(i interface{}) (int, error) {
	s, ok := i.(string)
	if !ok {
		return 0, errors.New("not string")
	}
	return strconv.Atoi(s)
}
