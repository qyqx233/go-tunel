package rest

import (
	"reflect"
	"unsafe"

	"github.com/cockroachdb/pebble"
	"github.com/qyqx233/gtool/lib/convert"
	"github.com/rs/zerolog/log"
)

var pdb *pebble.DB

type ByteDecoder[T any] interface {
	decode([]byte)
}

func PebbleGet[T ByteDecoder[T]](key []byte, t T) error {
	value, closer, err := pdb.Get(key)
	if err != nil {
		return err
	}
	defer closer.Close()
	t.decode(value)
	return nil
}

func PebbleGetBytes(key []byte) ([]byte, error) {
	value, closer, err := pdb.Get(key)
	if err != nil {
		return []byte(""), err
	}
	defer closer.Close()
	return value, nil
}

type TransportPdb struct {
	Enable bool
	Export bool
}

func (t *TransportPdb) GetKey(port int) []byte {
	var key = make([]byte, 0, 12)
	key = append(key, "port:"...)
	key = append(key, convert.Uint642Bytes(uint64(port))[0])
	return key
}

type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type basePdb struct {
	size uint8
}

func (t *basePdb) encode() []byte {
	l := int(unsafe.Sizeof(basePdb{}))
	data := reflect.SliceHeader{uintptr(unsafe.Pointer(t)), l, l}
	return *(*[]byte)(unsafe.Pointer(&data))
}

func (t *TransportPdb) Key() string {
	return ""
}

func (t *TransportPdb) decode(data []byte) {
	l := len(data)
	h := reflect.SliceHeader{uintptr(unsafe.Pointer(t)), l, l}
	bs := *(*[]byte)(unsafe.Pointer(&h))
	copy(bs, data)
}

func (t *TransportPdb) encode() []byte {
	l := int(unsafe.Sizeof(TransportPdb{}))
	data := reflect.SliceHeader{uintptr(unsafe.Pointer(t)), l, l}
	return *(*[]byte)(unsafe.Pointer(&data))
}

func InitDB() {
	var err error
	pdb, err = pebble.Open("db", &pebble.Options{})
	if err != nil {
		panic(err)
	}
	// batch := pdb.NewBatch()
	// batch.Set([]byte(""), []byte(""), pebble.NoSync)
	// batch.Commit(&pebble.WriteOptions{Sync: true})
}

func init() {
	log.Info().Msg("init pdb")
	InitDB()
}
