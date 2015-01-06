package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

var prefix = ""
var lookup = map[string][]*entry{}

type entry struct {
	ID       int
	offset   int
	bytes    int
	depth    int
	memIn    int
	memOut   int
	timeIn   float64
	timeOut  float64
	parent   *entry
	children map[int]*entry
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

func (e *entry) getJSON() entryJson {
	rval := e.get()
	rval.Children = []entryJson{}
	for _, v := range e.children {
		rval.Children = append(rval.Children, v.get())
	}
	return rval
}

func (e *entry) get() entryJson {
	rval := entryJson{
		ID:       e.ID,
		ParentID: e.parent.ID,
		Mem:      e.memOut - e.memIn,
		Time:     float32(e.timeIn),
		Duration: float32(e.timeOut - e.timeIn),
	}
	if e.children != nil {
		rval.ChildCount = len(e.children)
	}
	if e.bytes == 0 {
		return rval
	}
	buf := make([]byte, e.bytes)
	t.fp.ReadAt(buf, int64(e.offset))
	rval.Raw = string(buf)
	parts := strings.Split(strings.TrimSpace(string(buf)), "\t")
	rval.Details.Func = parts[5]
	if i := strings.Index(rval.Details.Func, ":{/"); i > 0 {
		rval.Details.Func = rval.Details.Func[:i]
	}
	rval.Location.File = parts[8][len(prefix):]
	rval.Location.Line, _ = strconv.Atoi(parts[9])
	if parts[6] == "1" {
		rval.Details.Args = []string{parts[7]}
	} else {
		rval.Details.Args = []string{strings.Join(parts[11:], "\t")}
		for k, v := range rval.Details.Args {
			if len(v) > 1 && v[0:0] == "'" {
				rval.Details.Args[k] = v[1 : len(v)-1]
			}
		}
	}
	return rval
}

func getEntry(ID int) (*entry, bool) {
	if e, ok := functions[ID]; ok {
		return e, false
	}
	e := &entry{ID: ID, children: map[int]*entry{}}
	functions[ID] = e
	return e, true
}

var root = &entry{}

var baseline bool
var baselevel int
var basefunction int
var baseoffset int

var functions = map[int]*entry{
	0: root,
}

type trace struct {
	fp *os.File
	fi os.FileInfo
	l  sync.Mutex
}

func (t *trace) getFunc(fnum int) []byte {
	var data = struct {
		Success bool      `json:"success"`
		Data    entryJson `json:"data"`
	}{}
	t.l.Lock()
	defer t.l.Unlock()
	if fun, ok := functions[fnum]; ok {
		data.Success = true
		data.Data = fun.getJSON()
	}
	out, err := json.Marshal(data)
	if err != nil {
		log.Printf(err.Error())
	}
	return out
}

func (t *trace) index() {
	t.l.Lock()
	defer t.l.Unlock()

	d := byte('\n')
	tab := []byte("\t")
	offset := 0

	t.fp.Seek(0, 0)
	reader := bufio.NewReader(t.fp)

	line, _ := reader.ReadBytes(d) // Version: 2.2.6
	offset += len(line)
	line, _ = reader.ReadBytes(d) // File format: 2
	offset += len(line)
	line, _ = reader.ReadBytes(d) // TRACE START [2015-01-02 00:26:56]
	offset += len(line)
	line, _ = reader.ReadBytes(d) // The beginning of the trace (exit code of the call to start tracing)
	offset += len(line)

	baseoffset = offset
	parts := bytes.Split(line, tab)

	baselevel, _ = strconv.Atoi(string(parts[0]))
	basefunction, _ = strconv.Atoi(string(parts[1]))

	root.parent = root
	functions[basefunction] = root
	root.ID = basefunction
	root.children = map[int]*entry{}
	current := root

	parentNodes := map[int]*entry{
		baselevel: root,
	}

	currentLevel := baselevel

	i := 0
	for {
		if line, err := reader.ReadBytes(d); err != nil {
			log.Fatal(err)
		} else {
			i++
			parts := bytes.Split(bytes.TrimSpace(line), tab)
			if len(parts[0]) < 1 {
				break
			}
			if len(parts) == 2 {
				if endTime, err := strconv.ParseFloat(string(parts[0]), 64); err == nil {
					if endMem, err := strconv.Atoi(string(parts[1])); err == nil {
						root.timeOut = endTime
						root.memOut = endMem
						continue
					}
				}
			}
			if len(parts) < 3 {
				continue
			}
			newLevel, _ := strconv.Atoi(string(parts[0]))
			kind, _ := strconv.Atoi(string(parts[2]))
			fnum, _ := strconv.Atoi(string(parts[1]))
			e, n := getEntry(fnum)
			if newLevel != currentLevel {
				if newLevel > currentLevel {
					parentNodes[newLevel] = e
					//log.Println(strings.Repeat(">", newLevel), fnum)
				} else {
					delete(parentNodes, currentLevel)
					//log.Println(strings.Repeat("<", newLevel), fnum)
				}
				// log.Printf("%#v", parentNodes)
				currentLevel = newLevel
			} else {
				//log.Println(strings.Repeat(" ", currentLevel), fnum)
			}
			if i == 1 {
				root.memIn, _ = strconv.Atoi(string(parts[4]))
				root.timeIn, _ = strconv.ParseFloat(string(parts[3]), 64)
			}
			if n {
				e.parent = current
				e.offset = offset
				e.bytes = len(line) - 1
				e.parent.children[fnum] = e
				e.memIn, _ = strconv.Atoi(string(parts[4]))
				e.timeIn, _ = strconv.ParseFloat(string(parts[3]), 64)
				e.children = map[int]*entry{}
				current = e
				if v, ok := lookup[string(parts[5])]; ok {
					lookup[string(parts[5])] = append(v, e)
				} else {
					lookup[string(parts[5])] = []*entry{e}
				}
				if prefix == "" {
					prefix = string(parts[8])
				} else {
					prefix = LCS(string(parts[8]), prefix)
				}
			}
			e.timeOut, _ = strconv.ParseFloat(string(parts[3]), 64)
			e.memOut, _ = strconv.Atoi(string(parts[4]))
			if kind == 1 {
				current = e.parent
			}
			offset += len(line)
		}
	}
}

func LCS(s1 string, s2 string) string {
	var m = make([][]int, 1+len(s1))
	for i := 0; i < len(m); i++ {
		m[i] = make([]int, 1+len(s2))
	}
	longest := 0
	x_longest := 0
	for x := 1; x < 1+len(s1); x++ {
		for y := 1; y < 1+len(s2); y++ {
			if s1[x-1] == s2[y-1] {
				m[x][y] = m[x-1][y-1] + 1
				if m[x][y] > longest {
					longest = m[x][y]
					x_longest = x
				}
			} else {
				m[x][y] = 0
			}
		}
	}
	return s1[x_longest-longest : x_longest]
}
