package trace

import (
	"encoding/json"
	"strconv"
	"strings"
)

type Entry struct {
	ID        int     // Function ID
	Parent    *Entry  // the parent of this Function
	Children  []int   // the IDs of any children the function has
	offset    int     // the byte offset in the file where the function entry details exist
	bytes     int     // # of bytes at offset the entry consists of
	depth     int     // The base indentation level
	memoryIn  int     // memory at function entry
	memoryOut int     // memory at function exit
	timeIn    float64 // time offset at function entry
	timeOut   float64 // time offset at function exit
	trace     *Trace
}

type entryJson struct {
	ID       int `json:"id"`
	ParentID int `json:"parent_id"`
	Location struct {
		File string `json:"file"`
		Line int    `json:"line"`
	} `json:"location"`
	Details struct {
		Func string   `json:"func"`
		Args []string `json:"args"`
	} `json:"details"`
	Mem        int         `json:"memory"`
	Time       float32     `json:"time"`
	Duration   float32     `json:"duration"`
	Children   []entryJson `json:"children"`
	ChildCount int         `json:"child_count"`
	Raw        string      `json:"raw"`
}

func (e *Entry) getEntryJSON() entryJson {
	var rval = entryJson{
		ID:         e.ID,
		ParentID:   e.Parent.ID,
		Mem:        e.memoryOut - e.memoryIn,
		Time:       float32(e.timeIn),
		Duration:   float32(e.timeOut - e.timeIn),
		Children:   []entryJson{},
		ChildCount: len(e.Children),
	}
	if e.ID == 0 {
		return rval
	}
	b := e.Bytes()
	if e == nil {
		return rval
	}
	rval.Raw = string(b)
	parts := strings.Split(strings.TrimSpace(rval.Raw), "\t")
	if parts[6] == "0" {
		rval.Details.Args = []string{strings.Join(parts[11:], "\t")}
		rval.Location.File = parts[8]
	} else {
		rval.Details.Args = []string{parts[8]}
		rval.Location.File = parts[7]
	}
	rval.Details.Func = parts[5]
	rval.Location.Line, _ = strconv.Atoi(parts[9])
	return rval
}

func (e *Entry) GetJSON() []byte {
	rval := e.getEntryJSON()
	if e.Children != nil {
		for _, v := range e.Children {
			rval.Children = append(rval.Children, e.trace.lookupByID[v].getEntryJSON())
			rval.ChildCount++
		}
	}
	rbytes, _ := json.Marshal(rval)
	return rbytes
}

func (e *Entry) Bytes() []byte {
	e.trace.lock.Lock()
	defer e.trace.lock.Unlock()
	var buf = make([]byte, e.bytes)
	e.trace.fp.ReadAt(buf, int64(e.offset))
	return buf
}
