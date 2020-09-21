package images

import (
  "bytes"
  log "crit-go/logging"
  "crit-go/protobindings"
  "encoding/base64"
  "encoding/binary"
  "encoding/json"
  "fmt"
  "github.com/golang/protobuf/jsonpb"
  "github.com/golang/protobuf/proto"
  "io"
  "os"
  "strconv"
  "strings"
)

type entry_handler struct {
  m string
}

type ipc_shm_set_handler struct {
  m string
}

type ipc_msg_queue_handler struct {
  m string
}

type ipc_sem_set_handler struct {
  m string
}

type tcp_stream_extra_handler struct {
  m string
}

type sk_queues_extra_handler struct {
  m string
}

type pipes_data_extra_handler struct {
  m    string
  size int
}

type keyvalue map[string]interface{}

type ghost_file_handler struct {
  m string
}

type pagemap_handler struct {
  m string
}

func (m *entry_handler) Load(f *os.File, pretty bool, nopl bool) []keyvalue {
  /*
     Generic CLass to Handle Loading/Dumping criu images
     entries from bin format to dict(json)
  */
  readbuffer := make([]byte, 4)
  var entries []keyvalue
  for {
    var entry map[string]interface{}
    n, err := f.Read(readbuffer)
    if n == 0 && err != nil {
      if err == io.EOF {
        f.Close()
        break
      }
      log.Error("Input file Read Error: ", err)
      closefile(f)
    }
    size := uint64(binary.LittleEndian.Uint32(readbuffer))
    internalrbuffer := make([]byte, size)
    _, err = f.Read(internalrbuffer)
    if err != nil {
      log.Error("Input file Read Error: ", err)
      closefile(f)
    }
    // Do not change this,Below bool used with generic proto parsing struct
    load := true
    entr, err := protohandler(m.m, load, internalrbuffer)
    if err != nil {
      log.Error(err)
      closefile(f)
    }
    if err := json.Unmarshal(entr, &entry); err != nil {
      log.Error("Json Unmarshal Error: ", err)
      closefile(f)
    }
    entries = append(entries, entry)
  }
  return entries
}

func (m *entry_handler) Dump(jsonmap map[string]interface{}, f *os.File) {
  /*
     Generic function to Convert criu image entries from dict(json)
     format to binary.Takes a list of entries and a file-like object
     to write entries in binary format to.
  */
  // Do not change this,Below bool used with generic proto parsing struct
  load := false
  for _, entry := range jsonmap["entries"].([]interface{}) {
    internalbmap, err := json.Marshal(entry)
    if err != nil {
      log.Error("Json Marshalling Error: ", err)
    }
    entr, err := protohandler(m.m, load, internalbmap)
    if err != nil {
      log.Error(err)
      closefile(f)
    }
    bs := make([]byte, 4)
    binary.LittleEndian.PutUint32(bs, uint32(len(entr)))
    _, err = f.Write(bs)
    if err != nil {
      log.Error("File Writing Error : ", err)
      closefile(f)
    }
    _, err = f.Write(entr)
    if err != nil {
      log.Error("File Writing Error : ", err)
      closefile(f)
    }
  }
  f.Close()
}

func (m *ghost_file_handler) Load(f *os.File, pretty bool, nopl bool) []keyvalue {
  /*
     Convert criu image entries from binary format to dict(json).
     Takes a file-like object and returnes a list with entries in
     dict(json) format.
  */
  readbuffer := make([]byte, 4)
  var entries []keyvalue
  var gentry map[string]interface{}
  var entry map[string]interface{}
  _, err := f.Read(readbuffer)
  if err != nil {
    log.Error("Input file Read Error: ", err)
    closefile(f)
  }
  fi, err := f.Stat()
  if err != nil {
    log.Error("Input file Read Error: ", err)
    closefile(f)
  }
  size := uint64(binary.LittleEndian.Uint32(readbuffer))
  internalrbuffer := make([]byte, size)
  _, err = f.Read(internalrbuffer)
  if err != nil {
    log.Error("Input file Read Error: ", err)
    closefile(f)
  }
  protobinding := &protobindings.GhostFileEntry{}
  if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
    log.Warn("Failed to Parse data into protobinding: ", err)
  }
  jsonm := &jsonpb.Marshaler{OrigName: true}
  entr := &bytes.Buffer{}
  err = jsonm.Marshal(entr, protobinding)
  if err != nil {
    log.Warn(err)
  }
  jpb := entr.Bytes()
  if err := json.Unmarshal(jpb, &gentry); err != nil {
    log.Error(err)
    closefile(f)
  }
  gc_prot := &protobindings.GhostChunkEntry{}
  if protobinding.GetChunks() == true {
    entries = append(entries, gentry)
    for {
      internalbuffer := make([]byte, 4)
      bytesread, err := f.Read(internalbuffer)
      if bytesread == 0 && err != nil {
        if err == io.EOF {
          f.Close()
          break
        }
        log.Error("Input file Read Error: ", err)
        closefile(f)
      }
      size := uint64(binary.LittleEndian.Uint32(internalbuffer))
      internalrbuffer := make([]byte, size)
      bytesread, err = f.Read(internalbuffer)
      if bytesread == 0 && err != nil {
        if err == io.EOF {
          f.Close()
          break
        }
        log.Error("Input file Read Error: ", err)
        closefile(f)
      }
      if err = proto.Unmarshal(internalrbuffer, gc_prot); err != nil {
        log.Warn("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, gc_prot)
      if err != nil {
        log.Warn(err)
      }
      jpb = entr.Bytes()
      if err := json.Unmarshal(jpb, &entry); err != nil {
        log.Error(err)
        closefile(f)
      }
      if nopl == true {
        f.Seek(int64(gc_prot.GetLen()), 1)
      } else {
        extradatabuf := make([]byte, gc_prot.GetLen())
        n, err := f.Read(extradatabuf)
        if err != nil {
          if err != io.EOF {
            log.Error(err)
            closefile(f)
          }
        }
        // removing n(bytesread) will add extra zeros/data to extra
        str := base64.StdEncoding.EncodeToString(extradatabuf[:n])
        entry["extra"] = str
      }
      entries = append(entries, entry)
    }
  } else {
    if nopl == true {
      f.Seek(0, 2)
    } else {
      extradatabuf := make([]byte, fi.Size())
      n, err := f.Read(extradatabuf)
      if err != nil {
        if err != io.EOF {
          log.Error(err)
          closefile(f)
        }
      }
      // removing n(bytesread) will add extra zeros/data to extra
      str := base64.StdEncoding.EncodeToString(extradatabuf[:n])
      gentry["extra"] = str
    }
    entries = append(entries, gentry)
  }
  return entries
}

func (m *ghost_file_handler) Dump(jsonmap map[string]interface{}, f *os.File) {
  var jpb []byte
  protobinding := &protobindings.GhostFileEntry{}
  for i, entry := range jsonmap["entries"].([]interface{}) {
    if i == 1 {
      break
    }
    internalbmap, err := json.Marshal(entry)
    if err != nil {
      log.Warn(err)
    }
    jsonm := &jsonpb.Unmarshaler{}
    r := bytes.NewReader(internalbmap)
    err = jsonm.Unmarshal(r, protobinding)
    if err != nil {
      log.Warn(err)
    }
    jpb, err = proto.Marshal(protobinding)
    if err != nil {
      log.Error("Failed to Marshal protobinding,check magic or file a github issue: ", err)
      closefile(f)
    }
    bs := make([]byte, 4)
    binary.LittleEndian.PutUint32(bs, uint32(len(jpb)))
    _, err = f.Write(bs)
    if err != nil {
      log.Error(err)
      closefile(f)
    }
    _, err = f.Write(jpb)
    if err != nil {
      log.Error(err)
      closefile(f)
    }
  }

  if protobinding.GetChunks() == true {
    gc_prot := &protobindings.GhostChunkEntry{}
    for _, entry := range jsonmap["entries"].([]interface{}) {
      enty := entry.(map[string]interface{})
      internalbmap, err := json.Marshal(entry)
      if err != nil {
        log.Warn(err)
      }
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalbmap)
      err = jsonm.Unmarshal(r, gc_prot)
      if err != nil {
        log.Warn(err)
      }
      jpb, err = proto.Marshal(gc_prot)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic: ", err)
        closefile(f)
      }
      bs := make([]byte, len(jpb))
      binary.LittleEndian.PutUint32(bs, uint32(len(jpb)))
      _, err = f.Write(bs)
      if err != nil {
        log.Error(err)
        closefile(f)
      }
      _, err = f.Write(jpb)
      if err != nil {
        log.Error(err)
        closefile(f)
      }
      dcsdata, err := base64.StdEncoding.DecodeString(enty["extra"].(string))
      if err != nil {
        log.Error(err)
        closefile(f)
      }
      _, err = f.Write(dcsdata)
      if err != nil {
        log.Error(err)
        closefile(f)
      }
    }
  } else {
    for i, entry := range jsonmap["entries"].([]interface{}) {
      if i == 1 {
        break
      }
      enty := entry.(map[string]interface{})
      dcsdata, err := base64.StdEncoding.DecodeString(enty["extra"].(string))
      if err != nil {
        log.Error(err)
        closefile(f)
      }
      _, err = f.Write(dcsdata)
      if err != nil {
        log.Error(err)
        closefile(f)
      }
    }
  }
  f.Close()
}

func (m *pagemap_handler) Load(f *os.File, pretty bool, nopl bool) []keyvalue {
  /*
     Convert criu image entries from binary format to dict(json).
     Takes a file-like object and returnes a list with entries in
     dict(json) format.
  */
  readbuffer := make([]byte, 4)
  var entries []keyvalue
  i := 0
  for {
    var entry map[string]interface{}
    n, err := f.Read(readbuffer)
    if n == 0 && err != nil {
      if err == io.EOF {
        f.Close()
        break
      }
      if err != nil {
        log.Error(err)
        closefile(f)
      }
    }
    size := uint64(binary.LittleEndian.Uint32(readbuffer))
    internalrbuffer := make([]byte, size)
    _, err = f.Read(internalrbuffer)
    if err != nil {
      log.Error(err)
      closefile(f)
    }
    if i < 1 {
      protobinding := &protobindings.PagemapHead{}
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Warn("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Warn(err)
      }
      if err := json.Unmarshal(entr.Bytes(), &entry); err != nil {
        log.Error(err)
        closefile(f)
      }
    } else {
      protobinding := &protobindings.PagemapEntry{}
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Warn("Failed to Parse data into protobinding: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Warn(err)
      }
      if err := json.Unmarshal(entr.Bytes(), &entry); err != nil {
        log.Error(err)
        closefile(f)
      }
    }
    entries = append(entries, entry)
    i = i + 1
  }
  return entries
}

func (m *pagemap_handler) Dump(jsonmap map[string]interface{}, f *os.File) {
  /*
     Special dump handler for pagemap.img, which is unique in a way
     that it has a header of pagemap_head type followed by entries
     of pagemap_entry type.
  */
  var jpb []byte
  i := 0
  for _, entry := range jsonmap["entries"].([]interface{}) {
    internalbmap, err := json.Marshal(entry)
    if err != nil {
      log.Warn("Json marshalling error: ", err)
    }
    if i < 1 {
      protobinding := &protobindings.PagemapHead{}
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalbmap)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Warn(err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding: ", err)
        closefile(f)
      }
    } else {
      protobinding := &protobindings.PagemapEntry{}
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalbmap)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Warn(err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding: ", err)
        closefile(f)
      }
    }
    bs := make([]byte, 4)
    binary.LittleEndian.PutUint32(bs, uint32(len(jpb)))
    _, err = f.Write(bs)
    if err != nil {
      log.Error(err)
      closefile(f)
    }
    _, err = f.Write(jpb)
    if err != nil {
      log.Error(err)
      closefile(f)
    }
    i++
  }
  f.Close()
}

func (m *sk_queues_extra_handler) Load(f *os.File, pretty bool, nopl bool) []keyvalue {
  /*
     Convert criu image entries from binary format to dict(json).
     Takes a file-like object and returnes a list with entries in
     dict(json) format.
  */
  readbuffer := make([]byte, 4)
  var entries []keyvalue
  for {
    var entry map[string]interface{}
    n, err := f.Read(readbuffer)
    if n == 0 && err != nil {
      if err == io.EOF {
        f.Close()
        break
      }
      log.Error(err)
      closefile(f)
    }
    size := uint64(binary.LittleEndian.Uint32(readbuffer))
    internalrbuffer := make([]byte, size)
    _, err = f.Read(internalrbuffer)
    if err != nil {
      log.Error("Input file reading error: ", err)
      closefile(f)
    }
    protobinding := &protobindings.SkPacketEntry{}
    if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
      log.Error("Failed to parse data into proto: ", err)
      closefile(f)
    }
    jsonm := &jsonpb.Marshaler{OrigName: true}
    entr := &bytes.Buffer{}
    err = jsonm.Marshal(entr, protobinding)
    if err != nil {
      log.Warn(err)
    }
    jpb := entr.Bytes()
    if err := json.Unmarshal(jpb, &entry); err != nil {
      log.Error("Json unmarshalling error: ", err)
      closefile(f)
    }
    if err := json.Unmarshal(jpb, &entry); err != nil {
      log.Error("Json unmarshalling error: ", err)
      closefile(f)
    }
    if nopl == true {
      harray := [8]string{"", "K", "M", "G", "T", "P", "E", "Z"}
      hreadable := func(num int32) string {
        for _, value := range harray {
          if num < 1024.0 {
            if int32(num) == int32(num) {
              t := fmt.Sprintf("%d%s", num, value)
              return t
            } else {
              t := fmt.Sprintf("%d%s", num, value)
              return t
            }
          }
          num = int32(num) / 1024.0
        }
        t := fmt.Sprintf("%d", num)
        return t
      }
      f.Seek(int64(protobinding.GetLength()), 1)
      humanredaeble := fmt.Sprintf("....< %s >", hreadable(int32(protobinding.GetLength())))
      entry["extra"] = humanredaeble
    } else {
      extradatabuf := make([]byte, protobinding.GetLength())
      n, err := f.Read(extradatabuf)
      if err != nil {
        if err != io.EOF {
          log.Error(err)
          closefile(f)
        }
      }
      /*
         Caution!
         removing n(bytesread) will/can add extra zeros/data to extra
      */
      str := base64.StdEncoding.EncodeToString(extradatabuf[:n])
      entry["extra"] = str
    }
    entries = append(entries, entry)
  }
  return entries
}

func (m *sk_queues_extra_handler) Dump(jsonmap map[string]interface{}, f *os.File) {
  /*
     Convert criu image entries from dict(json) format to binary.
     Takes a list of entries and a file-like object to write entries
     in binary format to.
  */
  protobinding := &protobindings.SkPacketEntry{}
  for _, entry := range jsonmap["entries"].([]interface{}) {
    internalbmap, err := json.Marshal(entry)
    if err != nil {
      log.Error("Json marshalling error: ", err)
    }
    jsonm := &jsonpb.Unmarshaler{}
    r := bytes.NewReader(internalbmap)
    err = jsonm.Unmarshal(r, protobinding)
    if err != nil {
      log.Warn(err)
    }
    entr, err := proto.Marshal(protobinding)
    if err != nil {
      log.Error(err)
      closefile(f)
    }
    bs := make([]byte, len(entr))
    binary.LittleEndian.PutUint32(bs, uint32(len(entr)))
    _, err = f.Write(bs)
    if err != nil {
      log.Error(err)
      closefile(f)
    }
    _, err = f.Write(entr)
    if err != nil {
      log.Error(err)
      closefile(f)
    }
    enty := entry.(map[string]interface{})
    dcsdata, err := base64.StdEncoding.DecodeString(enty["extra"].(string))
    f.Write(dcsdata)
    if err != nil {
      log.Error(err)
      closefile(f)
    }
  }
  f.Close()
}

func (m *ipc_sem_set_handler) Load(f *os.File, pretty bool, nopl bool) []keyvalue {
  /*
     Convert criu image entries from binary format to dict(json).
     Takes a file-like object and returnes a list with entries in
     dict(json) format.
  */
  readbuffer := make([]byte, 4)
  var entries []keyvalue
  var extradata []interface{}
  var entry map[string]interface{}
  for {
    n, err := f.Read(readbuffer)
    if n == 0 && err != nil {
      if err == io.EOF {
        f.Close()
        break
      }
      log.Error("Input file Read Error: ", err)
      closefile(f)
    }
    size := uint64(binary.LittleEndian.Uint32(readbuffer))
    internalrbuffer := make([]byte, size)
    _, err = f.Read(internalrbuffer)
    if err != nil {
      log.Error("Input file Read Error: ", err)
      closefile(f)
    }
    protobinding := &protobindings.IpcSemEntry{}
    if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
      log.Warn("Failed to parse data into proto: ", err)
    }
    jsonm := &jsonpb.Marshaler{OrigName: true}
    entr := &bytes.Buffer{}
    err = jsonm.Marshal(entr, protobinding)
    if err != nil {
      log.Error("Json Marshalling Error", err)
      closefile(f)
    }
    if err := json.Unmarshal(entr.Bytes(), &entry); err != nil {
      log.Error(err)
      closefile(f)
    }
    if nopl == true {
      harray := [8]string{"", "K", "M", "G", "T", "P", "E", "Z"}
      hreadable := func(num int32) string {
        for _, value := range harray {
          if num < 1024.0 {
            if int32(num) == int32(num) {
              t := fmt.Sprintf("%d%s", num, value)
              return t
            } else {
              t := fmt.Sprintf("%d%s", num, value)
              return t
            }
          }
          num = int32(num) / 1024.0
        }
        t := fmt.Sprintf("%d", num)
        return t
      }
      size := sizeof_u16 * uint32(entry["nsems"].(float64))
      rounded := roundup(size, sizeof_u64)
      f.Seek(int64(rounded), 1)
      humanredaeble := fmt.Sprintf("....< %s >", hreadable(int32(size)))
      entry["extra"] = humanredaeble
    } else {
      size := sizeof_u16 * uint32(entry["nsems"].(float64))
      rounded := roundup(size, sizeof_u64)
      var i uint32
      for i = 0; i < size; {
        internalrbuffer := make([]byte, 2)
        _, err := f.Read(internalrbuffer)
        if err != nil {
          log.Error("Input file Read Error: ", err)
          closefile(f)
        }
        extraint := binary.LittleEndian.Uint16(internalrbuffer)
        extradata = append(extradata, extraint)
        i = i + 2
      }
      rsize := int64(rounded - size)
      f.Seek(rsize, 1)
      entry["extra"] = extradata
    }
    entries = append(entries, entry)
  }
  return entries
}

func (m *ipc_sem_set_handler) Dump(jsonmap map[string]interface{}, f *os.File) {
  /*
     Convert criu image entries from dict(json) format to binary.
     Takes a list of entries and a file-like object to write entries
     in binary format to.
  */
  protobinding := &protobindings.IpcSemEntry{}
  for _, entry := range jsonmap["entries"].([]interface{}) {
    internalbmap, err := json.Marshal(entry)
    if err != nil {
      log.Warn(err)
    }
    jsonu := &jsonpb.Unmarshaler{}
    r := bytes.NewReader(internalbmap)
    err = jsonu.Unmarshal(r, protobinding)
    if err != nil {
      log.Warn(err)
    }
    entr, err := proto.Marshal(protobinding)
    if err != nil {
      log.Error("Failed to Marshal protobinding", err)
      closefile(f)
    }
    bs := make([]byte, 4)
    binary.LittleEndian.PutUint32(bs, uint32(len(entr)))
    _, err = f.Write(bs)
    if err != nil {
      log.Error(err)
      closefile(f)
    }
    _, err = f.Write(entr)
    if err != nil {
      log.Error(err)
      closefile(f)
    }
    enty := entry.(map[string]interface{})
    size := enty["nsems"]
    var cleansize uint32
    /*
       Type switch helps in maintaining cross compatibility
       between files generated in python and go version of crit
       python version generates strings,go version doesn not
       for certain data.
    */
    switch v := size.(type) {
    case float64:
      cleansize = sizeof_u16 * uint32(size.(float64))
    case string:
      cs, err := strconv.ParseUint(size.(string), 10, 32)
      if err != nil {
        log.Error(err)
        closefile(f)
      }
      cleansize = sizeof_u16 * uint32(cs)
    default:
      ty := fmt.Sprintf("Unkown type,contact the dev %T!\n", v)
      log.Warn(ty)
    }
    rounded := roundup(cleansize, sizeof_u64)
    var s []uint32
    if len(enty["extra"].([]interface{})) != int(enty["nsems"].(float64)) {
      log.Error("Number of semaphores mismatch")
      closefile(f)
    } else {
      for _, number := range enty["extra"].([]interface{}) {
        s = append(s, uint32(number.(float64)))
        bs = make([]byte, 2)
        binary.LittleEndian.PutUint16(bs, uint16(number.(float64)))
        _, err = f.Write(bs)
        if err != nil {
          log.Error("File Writing Error : ", err)
          closefile(f)
        }
      }
      zerowrite := rounded - cleansize
      for z := 0; z < int(zerowrite); z++ {
        _, err = f.Write([]byte{0})
        if err != nil {
          log.Error("File Writing Error : ", err)
          closefile(f)
        }
      }
    }
  }
  f.Close()
}

func (m *ipc_msg_queue_handler) Load(f *os.File, pretty bool, nopl bool) []keyvalue {
  /*
     Convert criu image entries from binary format to dict(json).
     Takes a file-like object and returnes a list with entries in
     dict(json) format.
  */
  readbuffer := make([]byte, 4)
  var entries []keyvalue
  var extradata []interface{}
  var entry map[string]interface{}
  for {
    n, err := f.Read(readbuffer)
    if n == 0 && err != nil {
      if err == io.EOF {
        f.Close()
        break
      }
      log.Error(err)
      closefile(f)
    }
    size := uint64(binary.LittleEndian.Uint32(readbuffer))
    internalrbuffer := make([]byte, size)
    _, err = f.Read(internalrbuffer)
    if err != nil {
      log.Error("Input file Read Error: ", err)
      closefile(f)
    }
    protobinding := &protobindings.IpcMsgEntry{}
    if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
      log.Warn("Failed to parse data into proto: ", err)
    }
    jsonm := &jsonpb.Marshaler{OrigName: true}
    entr := &bytes.Buffer{}
    err = jsonm.Marshal(entr, protobinding)
    if err != nil {
      log.Error(err)
      closefile(f)
    }
    if err := json.Unmarshal(entr.Bytes(), &entry); err != nil {
      log.Error(err)
      closefile(f)
    }
    if nopl == true {
      harray := [8]string{"", "K", "M", "G", "T", "P", "E", "Z"}
      hreadable := func(num int32) string {
        for _, value := range harray {
          if num < 1024.0 {
            if int32(num) == int32(num) {
              t := fmt.Sprintf("%d%s", num, value)
              return t
            } else {
              t := fmt.Sprintf("%d%s", num, value)
              return t
            }
          }
          num = int32(num) / 1024.0
        }
        t := fmt.Sprintf("%d", num)
        return t
      }
      var plength uint32
      for {
        internalrbuffer := make([]byte, 4)
        n, err = f.Read(internalrbuffer)
        if n == 0 && err != nil {
          if err == io.EOF {
            break
          }
          log.Error(err)
          closefile(f)
        }
        size := uint64(binary.LittleEndian.Uint32(internalrbuffer))
        internalpbuffer := make([]byte, size)
        _, err = f.Read(internalpbuffer)
        if err != nil {
          log.Error("Input File reading error: ", err)
          closefile(f)
        }
        msg_pb := &protobindings.IpcMsg{}
        if err = proto.Unmarshal(internalpbuffer, msg_pb); err != nil {
          log.Error("Failed to parse data into proto: ", err)
          closefile(f)
        }
        rounded := roundup(*msg_pb.Msize, sizeof_u64)
        f.Seek(int64(rounded), 1)
        plength = plength + uint32(size) + *msg_pb.Msize
      }
      humanredaeble := fmt.Sprintf("....< %s >", hreadable(int32(plength)))
      entry["extra"] = humanredaeble
    } else {
      i := 0
      for {
        var extraentry map[string]interface{}
        internalrbuffer := make([]byte, 4)
        n, err = f.Read(internalrbuffer)
        if n == 0 && err != nil {
          if err == io.EOF {
            break
          }
          log.Error(err)
          closefile(f)
        }
        size := uint64(binary.LittleEndian.Uint32(internalrbuffer))
        internalpbuffer := make([]byte, size)
        _, err = f.Read(internalpbuffer)
        if err != nil {
          log.Error("Input file Read Error: ", err)
          closefile(f)
        }
        msg_pb := &protobindings.IpcMsg{}
        if err = proto.Unmarshal(internalpbuffer, msg_pb); err != nil {
          log.Error("Failed to parse data into proto: ", err)
          closefile(f)
        }
        rounded := roundup(*msg_pb.Msize, sizeof_u64)
        extradatabuf := make([]byte, int32(*msg_pb.Msize))
        n, err := f.Read(extradatabuf)
        if err != nil {
          if err != io.EOF {
            log.Error(err)
            closefile(f)
          }
        }
        /*
           Caution!
           removing n(bytesread) will/can add extra zeros/data to extra
        */
        str := base64.StdEncoding.EncodeToString(extradatabuf[:n])
        f.Seek(int64(rounded-*msg_pb.Msize), 1)
        jsonm := &jsonpb.Marshaler{OrigName: true}
        extraentr := &bytes.Buffer{}
        err = jsonm.Marshal(extraentr, msg_pb)
        if err != nil {
          log.Error(err)
          closefile(f)
        }
        if err := json.Unmarshal(extraentr.Bytes(), &extraentry); err != nil {
          log.Error("Failed to Unmarshal extradata to Json", err)
          closefile(f)
        }
        extradata = append(extradata, extraentry)
        extradata = append(extradata, str)
        entry["extra"] = extradata
        i++
      }
    }
    entries = append(entries, entry)
  }
  return entries
}

func (m *ipc_msg_queue_handler) Dump(jsonmap map[string]interface{}, f *os.File) {
  /*
     Convert criu image entries from dict(json) format to binary.
     Takes a list of entries and a file-like object to write entries
     in binary format to.
  */
  protobinding := &protobindings.IpcMsgEntry{}
  for _, entry := range jsonmap["entries"].([]interface{}) {
    internalbmap, err := json.Marshal(entry)
    if err != nil {
      log.Warn("Json unmarshalling error: ", err)
    }
    jsonm := &jsonpb.Unmarshaler{}
    r := bytes.NewReader(internalbmap)
    err = jsonm.Unmarshal(r, protobinding)
    if err != nil {
      log.Warn(err)
    }
    entr, err := proto.Marshal(protobinding)
    if err != nil {
      log.Error("Failed to Marshal protobinding,check magic: ", err)
      closefile(f)
    }
    bs := make([]byte, 4)
    binary.LittleEndian.PutUint32(bs, uint32(len(entr)))
    _, err = f.Write(bs)
    if err != nil {
      log.Error("File writing error", err)
      closefile(f)
    }
    _, err = f.Write(entr)
    if err != nil {
      log.Error("File writing error", err)
      closefile(f)
    }
    enty := entry.(map[string]interface{})
    extradata := enty["extra"].([]interface{})
    msg_pb := &protobindings.IpcMsg{}
    for i := 0; i < len(enty["extra"].([]interface{})); i++ {
      internalbmap, err := json.Marshal(extradata[i])
      if err != nil {
        log.Warn("Json marshalling error: ", err)
      }
      r := bytes.NewReader(internalbmap)
      err = jsonm.Unmarshal(r, msg_pb)
      if err != nil {
        log.Warn("Json unmarshalling error: ", err)
      }
      entr, err := proto.Marshal(msg_pb)
      if err != nil {
        log.Error("Failed to Serialize extra data to protobinding: ", err)
        closefile(f)
      }
      size := len(entr)
      bs := make([]byte, 4)
      binary.LittleEndian.PutUint32(bs, uint32(size))
      _, err = f.Write(bs)
      if err != nil {
        log.Error("Failed to write to data file:", err)
        closefile(f)
      }
      _, err = f.Write(entr)
      if err != nil {
        log.Error("Failed to write to data file:", err)
        closefile(f)
      }
      rounded := roundup(*msg_pb.Msize, sizeof_u64)
      dcsdata, err := base64.StdEncoding.DecodeString(extradata[i+1].(string))
      if err != nil {
        log.Error("Failed to write to data file:", err)
        closefile(f)
      }
      _, err = f.Write(dcsdata[:*msg_pb.Msize])
      if err != nil {
        log.Error("Failed to write to data file:", err)
        closefile(f)
      }
      zerowrite := rounded - *msg_pb.Msize
      for z := 0; z < int(zerowrite); z++ {
        _, err = f.Write([]byte{0})
        if err != nil {
          log.Error("Failed to write to data file:", err)
          closefile(f)
        }
      }
      i = i + 1
    }
  }
  f.Close()
}

func (m *ipc_shm_set_handler) Load(f *os.File, pretty bool, nopl bool) []keyvalue {
  var entries []keyvalue
  var entry map[string]interface{}
  readbuffer := make([]byte, 4)
  i := 0
  for {
    n, err := f.Read(readbuffer)
    if n == 0 && err != nil {
      if err == io.EOF {
        f.Close()
        break
      }
      log.Error("Input file Read Error: ", err)
      closefile(f)
    }
    size := uint64(binary.LittleEndian.Uint32(readbuffer))
    internalrbuffer := make([]byte, size)
    _, err = f.Read(internalrbuffer)
    if err != nil {
      log.Error("Input file Read Error: ", err)
      closefile(f)
    }
    protobinding := &protobindings.IpcShmEntry{}
    if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
      log.Warn("Failed to parse data into proto: ", err)
    }
    jsonm := &jsonpb.Marshaler{OrigName: true}
    entr := &bytes.Buffer{}
    err = jsonm.Marshal(entr, protobinding)
    if err != nil {
      log.Error(err)
      closefile(f)
    }
    if err := json.Unmarshal(entr.Bytes(), &entry); err != nil {
      log.Error(err)
      closefile(f)
    }
    if nopl == true {
      harray := [8]string{"", "K", "M", "G", "T", "P", "E", "Z"}
      hreadable := func(num int32) string {
        for _, value := range harray {
          if num < 1024.0 {
            if int32(num) == int32(num) {
              t := fmt.Sprintf("%d%s", num, value)
              return t
            } else {
              t := fmt.Sprintf("%d%s", num, value)
              return t
            }
          }
          num = int32(num) / 1024.0
        }
        t := fmt.Sprintf("%d", num)
        return t
      }
      size := sizeof_u16 * uint32(entry["size"].(float64))
      rounded := roundup(size, sizeof_u64)
      f.Seek(int64(rounded), 1)
      humanredaeble := fmt.Sprintf("....< %s >", hreadable(int32(size)))
      entry["extra"] = humanredaeble
    } else {
      size, err := parseInt(entry["size"])
      if err != nil {
        log.Error(err)
        closefile(f)
      }
      rounded := roundup(uint32(size), sizeof_u32)
      rsize := int64(rounded - uint32(size))
      extradatabuf := make([]byte, size)
      n, err = f.Read(extradatabuf)
      if err != nil {
        if err != io.EOF {
          log.Error(err)
          closefile(f)
        }
      }
      /*
         Caution!
         removing n(bytesread) will/can add extra zeros/data to extra
      */
      str := base64.StdEncoding.EncodeToString(extradatabuf[:n])
      entry["extra"] = str
      f.Seek(rsize, 1)
    }
    entries = append(entries, entry)
    i++
  }
  return entries
}

func (m *ipc_shm_set_handler) Dump(jsonmap map[string]interface{}, f *os.File) {
  /*
     Convert criu image entries from dict(json) format to binary.
     Takes a list of entries and a file-like object to write entries
     in binary format to.
  */
  protobinding := &protobindings.IpcShmEntry{}
  for _, entry := range jsonmap["entries"].([]interface{}) {
    internalbmap, err := json.Marshal(entry)
    if err != nil {
      log.Warn("Json marshalling error: ", err)
    }
    jsonu := &jsonpb.Unmarshaler{}
    r := bytes.NewReader(internalbmap)
    err = jsonu.Unmarshal(r, protobinding)
    if err != nil {
      log.Warn(err)
    }
    entr, err := proto.Marshal(protobinding)
    if err != nil {
      log.Error("Failed to Marshal protobinding,check magic: ")
      closefile(f)
    }
    bs := make([]byte, 4)
    binary.LittleEndian.PutUint32(bs, uint32(len(entr)))
    _, err = f.Write(bs)
    if err != nil {
      log.Error(err)
      closefile(f)
    }
    _, err = f.Write(entr)
    if err != nil {
      log.Error(err)
      closefile(f)
    }
    enty := entry.(map[string]interface{})
    size := enty["size"]
    var cleansize uint32
    /*
       Type switch helps in maintaining cross compatibility
       between files generated in python and go version of crit
       python version generates strings,go version doesn not
       for certain data.
    */
    switch v := size.(type) {
    case float64:
      cleansize = uint32(size.(float64))
    case string:
      cs, err := strconv.ParseUint(size.(string), 10, 32)
      if err != nil {
        log.Error(err)
        closefile(f)
      }
      cleansize = uint32(cs)
    default:
      ty := fmt.Sprintf("Unkown type,contact the dev %T!\n", v)
      log.Warn(ty)
    }
    rounded := roundup(cleansize, sizeof_u64)
    dcsdata, err := base64.StdEncoding.DecodeString(enty["extra"].(string))
    if err != nil {
      log.Error(err)
      closefile(f)
    }
    if cleansize < uint32(len(dcsdata)) {
      _, err = f.Write(dcsdata[:cleansize])
      if err != nil {
        log.Error("Failed writing to file", err)
        closefile(f)
      }
    } else {
      _, err = f.Write(dcsdata)
      if err != nil {
        log.Error("Failed writing to file", err)
        closefile(f)
      }
    }
    zerowrite := rounded - cleansize
    emptybyte := make([]byte, zerowrite)
    _, err = f.Write(emptybyte)
    if err != nil {
      log.Error("Failed writing to file", err)
      closefile(f)
    }
  }
  f.Close()
}

func (m *pipes_data_extra_handler) Load(f *os.File, pretty bool, nopl bool) []keyvalue {
  readbuffer := make([]byte, 4)
  var entries []keyvalue
  for {
    var entry map[string]interface{}
    n, err := f.Read(readbuffer)
    if n == 0 && err != nil {
      if err == io.EOF {
        f.Close()
        break
      }
      log.Error(err)
      closefile(f)
    }
    size := uint64(binary.LittleEndian.Uint32(readbuffer))
    internalrbuffer := make([]byte, size)
    _, err = f.Read(internalrbuffer)
    if err != nil {
      log.Error("Input file Read Error: ", err)
      closefile(f)
    }
    protobinding := &protobindings.PipeDataEntry{}
    if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
      log.Warn("Failed to parse data into proto: ", err)
    }
    jsonm := &jsonpb.Marshaler{OrigName: true}
    entr := &bytes.Buffer{}
    err = jsonm.Marshal(entr, protobinding)
    if err != nil {
      log.Error("Json Marshalling Error", err)
      closefile(f)
    }
    if err := json.Unmarshal(entr.Bytes(), &entry); err != nil {
      log.Error("Json UnMarshalling Error", err)
      closefile(f)
    }
    if nopl == true {
      harray := [8]string{"", "K", "M", "G", "T", "P", "E", "Z"}
      f.Seek(int64(protobinding.GetBytes()), 1)
      hreadable := func(num int32) string {
        for _, value := range harray {
          if num < 1024.0 {
            if int32(num) == int32(num) {
              t := fmt.Sprintf("%d%s", num, value)
              return t
            } else {
              t := fmt.Sprintf("%d%s", num, value)
              return t
            }
          }
          num = int32(num) / 1024.0
        }
        t := fmt.Sprintf("%d", num)
        return t
      }
      f.Seek(int64(protobinding.GetBytes()), 1)
      humanredaeble := fmt.Sprintf("....< %s >", hreadable(int32(protobinding.GetBytes())))
      entry["extra"] = humanredaeble
    } else {
      extradatabuf := make([]byte, protobinding.GetBytes())
      n, err = f.Read(extradatabuf)
      if err != nil {
        if err != io.EOF {
          log.Error(err)
          closefile(f)
        }
      }
      /*
         Caution!
         removing n(bytesread) will/can add extra zeros/data to extra
      */
      str := base64.StdEncoding.EncodeToString(extradatabuf[:n])
      entry["extra"] = str
    }
    entries = append(entries, entry)
  }
  return entries
}

func (m *pipes_data_extra_handler) Dump(jsonmap map[string]interface{}, f *os.File) {
  /*
     Convert criu image entries from dict(json) format to binary.
     Takes a list of entries and a file-like object to write entries
     in binary format to.
  */
  protobinding := &protobindings.PipeDataEntry{}
  for _, entry := range jsonmap["entries"].([]interface{}) {
    internalbmap, err := json.Marshal(entry)
    if err != nil {
      log.Warn("Json marshalling error: ", err)
    }
    jsonm := &jsonpb.Unmarshaler{}
    r := bytes.NewReader(internalbmap)
    err = jsonm.Unmarshal(r, protobinding)
    if err != nil {
      log.Warn(err)
    }
    entr, err := proto.Marshal(protobinding)
    if err != nil {
      log.Error("Failed to Marshal protobinding,check magic or file a github issue: ", err)
      closefile(f)
    }
    bs := make([]byte, 4)
    binary.LittleEndian.PutUint32(bs, uint32(len(entr)))
    _, err = f.Write(bs)
    if err != nil {
      log.Error(err)
      closefile(f)
    }
    _, err = f.Write(entr)
    if err != nil {
      log.Error(err)
      closefile(f)
    }
    enty := entry.(map[string]interface{})
    dcsdata, err := base64.StdEncoding.DecodeString(enty["extra"].(string))
    if err != nil {
      log.Error(err)
      closefile(f)
    }
    _, err = f.Write(dcsdata)
    if err != nil {
      log.Error(err)
      closefile(f)
    }
  }
  f.Close()
}

func (m *tcp_stream_extra_handler) Load(f *os.File, pretty bool, nopl bool) []keyvalue {
  /*
     Convert criu image entries from binary(.img) format to json.
     Takes a .img file reads in serially,parses it ot the protobuf
     defination and marshalls it ot json and writes it out.
  */
  readbuffer := make([]byte, 4)
  var entries []keyvalue
  for {
    d := make(map[string]interface{})
    var entry map[string]interface{}
    n, err := f.Read(readbuffer)
    if n == 0 && err != nil {
      if err == io.EOF {
        f.Close()
        break
      }
      log.Error("Input file Read Error: ", err)
      closefile(f)
    }
    size := uint64(binary.LittleEndian.Uint32(readbuffer))
    internalrbuffer := make([]byte, size)
    _, err = f.Read(internalrbuffer)
    if err != nil {
      log.Error("Input file Read Error: ", err)
      closefile(f)
    }
    protobinding := &protobindings.TcpStreamEntry{}
    if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
      log.Error("Failed to parse data into proto: ", err)
    }
    jsonm := &jsonpb.Marshaler{OrigName: true}
    entr := &bytes.Buffer{}
    err = jsonm.Marshal(entr, protobinding)
    if err != nil {
      log.Error(err)
      closefile(f)
    }
    if err := json.Unmarshal(entr.Bytes(), &entry); err != nil {
      log.Error("Json UnMarshalling errored: ", err)
    }
    if nopl == true {
      harray := [8]string{"", "K", "M", "G", "T", "P", "E", "Z"}
      hreadable := func(num int32) string {
        for _, value := range harray {
          if num < 1024.0 {
            if int32(num) == int32(num) {
              t := fmt.Sprintf("%d%s", num, value)
              return t
            } else {
              t := fmt.Sprintf("%d%s", num, value)
              return t
            }
          }
          num = int32(num) / 1024.0
        }
        t := fmt.Sprintf("%d", num)
        return t
      }
      f.Seek(0, 2)
      humanredaeble := fmt.Sprintf("....< %s >", hreadable(int32(protobinding.GetInqLen()+protobinding.GetOutqLen())))
      entry["extra"] = humanredaeble
    } else {
      extradatabuf := make([]byte, protobinding.GetInqLen())
      n, err = f.Read(extradatabuf)
      if err != nil {
        if err != io.EOF {
          log.Error("Input file Read Error: ", err)
          closefile(f)
        }
      }
      /*
         Caution!
         removing n(bytesread) will/can add extra zeros/data to extra
      */
      str := base64.StdEncoding.EncodeToString(extradatabuf[:n])
      d["inq"] = str
      extradatabuf = make([]byte, protobinding.GetOutqLen())
      n, err = f.Read(extradatabuf)
      if err != nil {
        if err != io.EOF {
          log.Error("Input file Read Error: ", err)
          closefile(f)
        }
      }
      /*
         Caution!
         removing n(bytesread) will/can add extra zeros/data to extra
      */
      str = base64.StdEncoding.EncodeToString(extradatabuf[:n])
      d["outq"] = str
      entry["extra"] = d
    }
    entries = append(entries, entry)
  }
  return entries
}

func (m *tcp_stream_extra_handler) Dump(jsonmap map[string]interface{}, f *os.File) {
  /*
     Convert criu image entries from dict(json) format to binary.
     Takes a list of entries and a file-like object to write entries
     in binary format to.
  */
  protobinding := &protobindings.TcpStreamEntry{}
  for _, entry := range jsonmap["entries"].([]interface{}) {
    enty := entry.(map[string]interface{})
    /*
       Type switch helps in maintaining cross compatibility
       between files generated in python and go version of crit.
       python version generates strings,go version doesn not
       for certain data.The Below code fixes that.
    */
    switch v := enty["opt_mask"].(type) {
    case string:
      optfix := strings.Trim(enty["opt_mask"].(string), "0x")
      enty["opt_mask"] = optfix
    case float64:
    default:
      ty := fmt.Sprintf("Unkown type,contact the dev. Type : %T!\n", v)
      log.Warn(ty)
    }
    internalbmap, err := json.Marshal(enty)
    if err != nil {
      log.Warn("Json marshalling error: %s", err)
    }
    jsonm := &jsonpb.Unmarshaler{}
    r := bytes.NewReader(internalbmap)
    err = jsonm.Unmarshal(r, protobinding)
    if err != nil {
      log.Warn(err)
    }
    entr, err := proto.Marshal(protobinding)
    if err != nil {
      log.Error("Failed to Marshal protobinding", err)
      closefile(f)
    }
    bs := make([]byte, 4)
    binary.LittleEndian.PutUint32(bs, uint32(len(entr)))
    _, err = f.Write(bs)
    if err != nil {
      log.Error("File Writing Error : ", err)
      closefile(f)
    }
    _, err = f.Write(entr)
    if err != nil {
      log.Error("File Writing Error : ", err)
      closefile(f)
    }
    extra := enty["extra"].(map[string]interface{})
    inq, err := base64.StdEncoding.DecodeString(extra["inq"].(string))
    if err != nil {
      log.Error(err)
      closefile(f)
    }
    outq, err := base64.StdEncoding.DecodeString(extra["outq"].(string))
    if err != nil {
      log.Error(err)
      closefile(f)
    }
    _, err = f.Write(inq)
    if err != nil {
      log.Error("File Writing Error : ", err)
      closefile(f)
    }
    _, err = f.Write(outq)
    if err != nil {
      log.Error("File Writing Error : ", err)
      closefile(f)
    }
  }
  f.Close()
}

func protohandler(m string, l bool, internalrbuffer []byte) (jpb []byte, err error) {
  switch {
  case m == "INVENTORY":
    protobinding := &protobindings.InventoryEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "CORE":
    protobinding := &protobindings.CoreEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "IDS":
    protobinding := &protobindings.TaskKobjIdsEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "CREDS":
    protobinding := &protobindings.CredsEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "UTSNS":
    protobinding := &protobindings.UtsnsEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "TIMENS":
    protobinding := &protobindings.TimensEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "IPC_VAR":
    protobinding := &protobindings.IpcVarEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "FS":
    protobinding := &protobindings.FsEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "MM":
    protobinding := &protobindings.MmEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "CGROUP":
    protobinding := &protobindings.CgroupEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "TCP_STREAM":
    protobinding := &protobindings.TcpStreamEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "STATS":
    protobinding := &protobindings.StatsEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "PSTREE":
    protobinding := &protobindings.PstreeEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "REG_FILES":
    protobinding := &protobindings.RegFileEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "NS_FILES":
    protobinding := &protobindings.NsFileEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "EVENTFD_FILE":
    protobinding := &protobindings.EventfdFileEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "EVENTPOLL_FILE":
    protobinding := &protobindings.EventpollFileEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "EVENTPOLL_TFD":
    protobinding := &protobindings.EventpollTfdEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "SIGNALFD":
    protobinding := &protobindings.SignalfdEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "TIMERFD":
    protobinding := &protobindings.TimerfdEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "INOTIFY_FILE":
    protobinding := &protobindings.InotifyFileEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "INOTIFY_WD":
    protobinding := &protobindings.InotifyWdEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "FANOTIFY_FILE":
    protobinding := &protobindings.FanotifyFileEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "FANOTIFY_MARK":
    protobinding := &protobindings.FanotifyMarkEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "VMAS":
    protobinding := &protobindings.VmaEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "PIPES":
    protobinding := &protobindings.PipeEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "FIFO":
    protobinding := &protobindings.FifoEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "SIGNACT":
    protobinding := &protobindings.SaEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "NETLINK_SK":
    protobinding := &protobindings.NetlinkSkEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "REMAP_FPATH":
    protobinding := &protobindings.RemapFilePathEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "MNTS":
    protobinding := &protobindings.MntEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "TTY_FILES":
    protobinding := &protobindings.TtyFileEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "TTY_INFO":
    protobinding := &protobindings.TtyInfoEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "TTY_DATA":
    protobinding := &protobindings.TtyDataEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "RLIMIT":
    protobinding := &protobindings.RlimitEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "TUNFILE":
    protobinding := &protobindings.TunfileEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "EXT_FILES":
    protobinding := &protobindings.ExtFileEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "IRMAP_CACHE":
    protobinding := &protobindings.IrmapCacheEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "FILE_LOCKS":
    protobinding := &protobindings.FileLockEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "FDINFO":
    protobinding := &protobindings.FdinfoEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "UNIXSK":
    protobinding := &protobindings.UnixSkEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "INETSK":
    protobinding := &protobindings.InetSkEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "PACKETSK":
    protobinding := &protobindings.PacketSockEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "ITIMERS":
    protobinding := &protobindings.ItimerEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "POSIX_TIMERS":
    protobinding := &protobindings.PosixTimerEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "NETDEV":
    protobinding := &protobindings.NetDeviceEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "NETNS":
    protobinding := &protobindings.NetnsEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "USERNS":
    protobinding := &protobindings.UsernsEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "FILES":
    protobinding := &protobindings.FileEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "SECCOMP":
    protobinding := &protobindings.SeccompEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "AUTOFS":
    protobinding := &protobindings.AutofsEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "MEMFD_FILE":
    protobinding := &protobindings.MemfdFileEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  case m == "MEMFD_INODE":
    protobinding := &protobindings.MemfdInodeEntry{}
    if l == true {
      if err = proto.Unmarshal(internalrbuffer, protobinding); err != nil {
        log.Error("Failed to parse data into proto: ", err)
      }
      jsonm := &jsonpb.Marshaler{OrigName: true}
      entr := &bytes.Buffer{}
      err = jsonm.Marshal(entr, protobinding)
      if err != nil {
        log.Error(err)
      }
      jpb = entr.Bytes()
    } else {
      jsonm := &jsonpb.Unmarshaler{}
      r := bytes.NewReader(internalrbuffer)
      err = jsonm.Unmarshal(r, protobinding)
      if err != nil {
        log.Error("Failed to Unmarshal json into protobinding,check json", err)
      }
      jpb, err = proto.Marshal(protobinding)
      if err != nil {
        log.Error("Failed to Marshal protobinding,check magic", err)
      }
    }
  }
  return jpb, err
}
