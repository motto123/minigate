package uuid

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// NewUUID 生成新的uuid
func NewUUID() string {
	u2, err := uuid.NewUUID()
	var no string
	if err != nil {
		log.Println("NewUUID error is:", err.Error())
	} else {
		no = strings.Replace(u2.String(), "-", "", -1)
	}
	return no
}

// NewUUIDWithPre 生成新的带前缀的uuid
func NewUUIDWithPre(pre string) string {
	u2, err := uuid.NewUUID()
	var no string
	if err != nil {
		log.Println("NewUUIDWithPre error is:", err.Error())
	} else {
		no = strings.Replace(u2.String(), "-", "", -1)
		nowStr := time.Now().Unix()
		no = strings.Replace(no, "", pre+strconv.Itoa(int(nowStr)), 1)
	}
	return no[0:32]
}
