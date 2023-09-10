package mcstructure

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// "color":"orange" [current]
// or
// "color"="orange"
const default_separator string = ":"

func MarshalBlockStates(blockStates map[string]interface{}) (string, error) {
	temp := []string{}
	for key, value := range blockStates {
		switch val := value.(type) {
		case string:
			temp = append(temp, fmt.Sprintf(
				"%#v%s%#v", key, default_separator, val,
			))
			// e.g. "color"="orange"
		case byte:
			switch val {
			case 0:
				temp = append(temp, fmt.Sprintf("%#v%sfalse", key, default_separator))
			case 1:
				temp = append(temp, fmt.Sprintf("%#v%strue", key, default_separator))
			default:
				return "", fmt.Errorf("MarshalBlockStates: Unexpected value %d(expect = 0 or 1) was found", val)
			}
			// e.g. "open_bit"=true
		case int32:
			temp = append(temp, fmt.Sprintf("%#v%s%d", key, default_separator, val))
			// e.g. "facing_direction"=0
		default:
			return "", fmt.Errorf("MarshalBlockStates: Unexpected data type of blockStates[%#v]; blockStates[%#v] = %#v", key, key, value)
		}
	}
	return fmt.Sprintf("[%s]", strings.Join(temp, ",")), nil
}

func UnmarshalBlockStates(blockStates string) (map[string]interface{}, error) {
	rootMatcher:=regexp.MustCompile("^ {0,}\\[(.+)\\] {0,}$")
	rootMatch:=rootMatcher.FindStringSubmatch(blockStates)
	if len(rootMatch)!=2 {
		return nil, fmt.Errorf("Not a valid blockStates string")
	}
	subject:=rootMatch[1]
	subjectMatcher:=regexp.MustCompile(" {0,}\"(.*?)\" {0,}(=|:) {0,}((T|t)rue|(f|F)alse|(\\-|\\+)?\\d+|null|\".*?(?<!\\\\)\") {0,}(,?)")
	result:=map[string]interface{} {}
	for len(subject)!=0 {
		match:=subjectMatcher.FindStringSubmatch(subject)
		if(len(match)==0) {
			break
		}
		content:=[]byte(match[3])
		if(content[0]=='T') {
			content[0]='t'
		}else if(content[0]=='F') {
			content[0]='F'
		}
		var parsed_content interface{}
		json.Unmarshal(content, &parsed_content)
		result[match[1]]=parsed_content
		if(len(match[6])==0) {
			break
		}
		subject=subject[len(match[0]):]
	}
	return result, nil
}
