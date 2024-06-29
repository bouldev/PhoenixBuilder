package convertor

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type ConvertRecord struct {
	Name             string `json:"name"`
	SNBTStateOrValue string `json:"states"`
	RTID             uint32 `json:"rtid"`
	isLegacyValue    int
	legacyValue      uint16
}

func (r *ConvertRecord) GetLegacyValue() (val uint16, ok bool) {
	if r.isLegacyValue == 0 {
		// unknown
		legacyBlockValue, err := strconv.Atoi(r.SNBTStateOrValue)
		if err == nil {
			if legacyBlockValue < 0 || legacyBlockValue > 0x8fff {
				panic("should not happen")
			}
			r.isLegacyValue = 1
			r.legacyValue = uint16(legacyBlockValue)
		} else {
			r.isLegacyValue = -1
		}
	}
	if r.isLegacyValue == 1 {
		return r.legacyValue, true
	} else {
		return 0, false
	}
}

func (r *ConvertRecord) String() string {
	return fmt.Sprintf("%v\n%v\n%v\n", r.Name, r.SNBTStateOrValue, r.RTID)
}

type CanReadString interface {
	ReadString(delim byte) (string, error)
}

func ReadRecordFromStream(reader CanReadString) (r *ConvertRecord, err error) {
	blockName, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	blockName = strings.TrimSuffix(blockName, "\n")
	if blockName == "" {
		return nil, fmt.Errorf("no block name")
	}
	snbtState, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	snbtState = strings.TrimSuffix(snbtState, "\n")
	rtidStr, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	rtidStr = strings.TrimSuffix(rtidStr, "\n")
	rtid, err := strconv.Atoi(rtidStr)
	if err != nil {
		return nil, err
	}
	return &ConvertRecord{
		Name:             blockName,
		SNBTStateOrValue: snbtState,
		RTID:             uint32(rtid),
	}, nil
}

func ReadRecordsFromString(dataBytes string) (records []*ConvertRecord, err error) {
	reader := bufio.NewReader(strings.NewReader(dataBytes))
	records = []*ConvertRecord{}
	for {
		r, err := ReadRecordFromStream(reader)
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			return records, nil
		}
		r.GetLegacyValue()
		records = append(records, r)
	}
}
