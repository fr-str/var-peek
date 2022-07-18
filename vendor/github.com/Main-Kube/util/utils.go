package util

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"math"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/Main-Kube/util/slice"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/google/uuid"
	"github.com/wI2L/jsondiff"
	"syslabit.com/git/syslabit/log"
)

const chars = "abcdefghijklmnopqrstuvwxyz0123456789"

var mdFive = md5.New()

// IsNumeric quickly checks if string is a number. 12x faster then strconv.ParseFloat(s, 64)
func IsNumeric(s string) bool {
	if s == "" {
		return false
	}

	dotFound := false
	for _, v := range s {
		if v == '.' {
			if dotFound {
				return false
			}
			dotFound = true
		} else if v < '0' || v > '9' {
			return false
		}
	}
	return true
}

// RoundTo rounds `n` float to `decimals` number after comma
//   RoundTo(1.123, 1) = 1.1
//   RoundTo(1.655, 2) = 1.66
func RoundTo(n float64, decimals uint8) float64 {
	return math.Round(n*math.Pow(10, float64(decimals))) / math.Pow(10, float64(decimals))
}

func CleanString(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsGraphic(r) || r == '\n' {
			return r
		}
		return -1
	}, str)
}

func GenerateID() string {
	return strings.ReplaceAll(
		uuid.New().String()+
			uuid.New().String()+
			uuid.New().String()[:16], // cut to < 90 len, becouse of backup filename length limit
		"-", "",
	)
}

// Create patch from two structs
func JsonPatch(old, new any) (patch []byte, err error) {

	jpatch, err := jsondiff.Compare(old, new)
	if log.Error(err) != nil {
		return nil, err
	}
	if jpatch.String() != "" {
		patch = []byte("[" + strings.ReplaceAll(jpatch.String(), "\n", ",") + "]")
	}

	return
}

// Create merge patch from two structs
func MergePatch(old, new any) (patch []byte, err error) {
	jsonOld, err := json.Marshal(old)
	if log.Error(err) != nil {
		return nil, err
	}

	jsonNew, err := json.Marshal(new)
	if log.Error(err) != nil {
		return nil, err
	}

	patch, err = jsonpatch.CreateMergePatch(jsonOld, jsonNew)
	if log.Error(err) != nil {
		return nil, err
	}

	return
}

func Must[T any](out T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return out
}

// RandomString returns random seeded string with provied length.
func RandomString(strlen int, seed ...string) string {

	s := strings.Join(seed, "")
	if s == "" {
		s = time.Now().String()
	}

	io.WriteString(mdFive, s)

	r := rand.New(rand.NewSource(int64(binary.BigEndian.Uint64(mdFive.Sum(nil)))))

	result := make([]byte, strlen)
	for i := range result {
		result[i] = chars[r.Intn(len(chars))]
	}
	return string(result)
}

func CompareMaps[K, T comparable](a, b map[K]T) bool {
	if len(a) != len(b) {
		return true
	}

	for k, v := range a {
		if b[k] != v {
			return true
		}
	}

	return false
}

func ReadAndReplaceInFile(filepath string, replacer *strings.Replacer) (string, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	return replacer.Replace(string(data)), nil
}

func DeepCopy[T any](src T) *T {
	data, err := json.Marshal(src)
	if err != nil {
		return nil
	}

	dst := new(T)
	err = json.Unmarshal(data, dst)
	if err != nil {
		return nil
	}
	return dst
}

// WaitWithTimeout waits for the waitgroup for the specified max timeout.
// Returns error if waiting timed out.
func WaitWithTimeout(wg *sync.WaitGroup, timeout time.Duration) error {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return nil
	case <-time.After(timeout):
		return errors.New("timeout")
	}
}

// Check if all fields in struct are nil except 'fields'
// returns flase if field is not in []fields and value is not nil
// returns false if obj is not a struct type
func IsNilExcept(obj interface{}, fields []string) bool {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return false
	}
	for i := 0; i < v.NumField(); i++ {
		f := v.Type().Field(i)
		if !f.Anonymous && !slice.Contains(fields, f.Name) {
			if !v.Field(i).IsNil() {
				return false
			}
		}
	}
	return true
}
