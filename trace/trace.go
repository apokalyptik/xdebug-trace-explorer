package trace

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Trace struct {
	fp       *os.File
	FileInfo os.FileInfo

	lock *sync.Mutex

	prefix string

	rootNode     *Entry
	lookupByName map[string][]int
	lookupByID   map[int]*Entry
}

func New(filename string) (*Trace, error) {
	fi, err := os.Lstat(filename)
	if err != nil {
		return nil, err
	}
	fp, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	var rval = &Trace{
		lookupByName: map[string][]int{},
		lookupByID:   map[int]*Entry{},
		fp:           fp,
		FileInfo:     fi,
		lock:         &sync.Mutex{},
	}

	if err := rval.index(); err != nil {
		return nil, err
	}

	return rval, nil
}

func (t *Trace) index() error {
	var err error
	t.lock.Lock()
	defer t.lock.Unlock()

	var offset = 0

	t.fp.Seek(0, 0)
	reader := bufio.NewReader(t.fp)

	for {
		line, _ := reader.ReadString('\n')
		offset += len(line)
		if strings.HasPrefix(line, "TRACE START ") {
			break
		}
	}
	line, _ := reader.ReadString('\n')
	offset += len(line)
	parts := strings.Split(line, "\t")
	t.rootNode = &Entry{}
	t.rootNode.Parent = t.rootNode
	t.rootNode.offset, _ = strconv.Atoi(parts[0])
	t.rootNode.trace = t
	t.lookupByID[0] = t.rootNode
	parent := t.rootNode
	for {
		parent, offset, err = t.indexFunctionEntry(reader, parent, offset)
		if err != nil {
			break
		}
	}
	return nil
}

func (t *Trace) indexFunctionEntry(r *bufio.Reader, parent *Entry, offset int) (newParent *Entry, newOffset int, err error) {
	var line string
	line, err = r.ReadString('\n')
	newOffset = offset + len(line)
	if err != nil {
		return
	}
	parts := strings.Split(strings.TrimSpace(line), "\t")
	if len(parts) == 2 {
		parts = append([]string{fmt.Sprintf("%d", t.rootNode.depth), "0", "1"}, parts...)
		log.Printf("%#v", parts)
	}
	if len(parts) == 1 {
		return
	}
	if parts[2] == "0" {
		entry := t.enterFunction(parts)
		entry.offset = offset
		entry.bytes = len(line) - 1
		entry.Parent = parent
		entry.Parent.Children = append(entry.Parent.Children, entry.ID)
		entry.trace = t
		t.lookupByID[entry.ID] = entry
		t.lookupByName[parts[5]] = append(t.lookupByName[parts[5]], entry.ID)
		newParent = entry
	} else {
		entry := t.exitfunction(parts)
		newParent = entry.Parent
	}
	return
}

func (t *Trace) enterFunction(parts []string) *Entry {
	var entry = &Entry{}
	entry.ID, _ = strconv.Atoi(parts[1])
	entry.depth, _ = strconv.Atoi(parts[0])
	entry.timeIn, _ = strconv.ParseFloat(parts[3], 64)
	entry.memoryIn, _ = strconv.Atoi(parts[4])
	return entry
}

func (t *Trace) exitfunction(parts []string) *Entry {
	ID, _ := strconv.Atoi(parts[1])
	entry, ok := t.lookupByID[ID]
	if !ok {
		log.Fatal("failed to look up entry", ID)
	}
	entry.timeOut, _ = strconv.ParseFloat(parts[3], 64)
	entry.memoryOut, _ = strconv.Atoi(parts[4])
	return entry
}

func (t *Trace) ByID(ID int) *Entry {
	rval, _ := t.lookupByID[ID]
	return rval
}
