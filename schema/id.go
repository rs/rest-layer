package schema

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"time"
)

var (
	// NewID is a field hook handler that generates a new unique id if none exist,
	// to be used in schema with OnInit.
	//
	// The generated ID is a Mongo like base64 object id (mgo/bson code has been embedded
	// into this function to prevent dep)
	NewID = func(value interface{}) interface{} {
		if value == nil {
			value = newID()
		}
		return value
	}

	// IDField is a common schema field configuration that generate an UUID for new item id.
	IDField = Field{
		Required:   true,
		ReadOnly:   true,
		OnInit:     &NewID,
		Filterable: true,
		Sortable:   true,
		Validator: &String{
			Regexp: "^[0-9a-zA-Z_-]{16}$",
		},
	}

	// machineID stores machine id generated once and used in subsequent calls
	// to NewObjectId function.
	machineID = readMachineID()

	// objectIDCounter is atomically incremented when generating a new ObjectId
	// using NewObjectId() function. It's used as a counter part of an id.
	objectIDCounter uint32
)

// ID validates and serialize unique id
type ID struct{}

// Validate implements FieldValidator interface
func (v ID) Validate(value interface{}) (interface{}, error) {
	s, ok := value.(string)
	if !ok {
		return nil, errors.New("invalid id")
	}
	if len(s) != 16 {
		return nil, errors.New("invalid id length")
	}
	_, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %s", err)
	}
	return s, nil
}

// readMachineID generates machine id and puts it into the machineId global
// variable. If this function fails to get the hostname, it will cause
// a runtime error.
func readMachineID() []byte {
	var sum [3]byte
	id := sum[:]
	hostname, err1 := os.Hostname()
	if err1 != nil {
		_, err2 := io.ReadFull(rand.Reader, id)
		if err2 != nil {
			panic(fmt.Errorf("cannot get hostname: %v; %v", err1, err2))
		}
		return id
	}
	hw := md5.New()
	hw.Write([]byte(hostname))
	copy(id, hw.Sum(nil))
	return id
}

// newID returns a new globally unique id using a copy of the mgo/bson algorithm.
func newID() string {
	var b [12]byte
	// Timestamp, 4 bytes, big endian
	binary.BigEndian.PutUint32(b[:], uint32(time.Now().Unix()))
	// Machine, first 3 bytes of md5(hostname)
	b[4] = machineID[0]
	b[5] = machineID[1]
	b[6] = machineID[2]
	// Pid, 2 bytes, specs don't specify endianness, but we use big endian.
	pid := os.Getpid()
	b[7] = byte(pid >> 8)
	b[8] = byte(pid)
	// Increment, 3 bytes, big endian
	i := atomic.AddUint32(&objectIDCounter, 1)
	b[9] = byte(i >> 16)
	b[10] = byte(i >> 8)
	b[11] = byte(i)
	return base64.URLEncoding.EncodeToString(b[:])
}
