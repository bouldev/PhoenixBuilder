package chat_phrases

import "fmt"

const (
	SourceOfficial = "OFFICIAL_SIGN"
)

type SyncNewPlayerPhrasesData []PhrasesData
type PhrasesData struct {
	Source  string `json:"itemId"`  // e.g. "OFFICIAL_SIGN"
	ID      int64  `json:"id"`      // e.g. int64(1002)
	Content string `json:"content"` // e.g. "加个好友吧"
}

// Return the event name of s
func (s *SyncNewPlayerPhrasesData) EventName() string {
	return "SyncNewPlayerPhrasesData"
}

// Convert s to go object which only contains go-built-in types
func (s *SyncNewPlayerPhrasesData) MakeGo() (res any) {
	new := map[string]any{}
	for _, value := range *s {
		_, ok := new[value.Source]
		if !ok {
			new[value.Source] = make(map[string]any)
		}
		new[value.Source].(map[string]any)[fmt.Sprintf("%d", value.ID)] = map[string]any{
			"content": value.Content,
			"id":      value.ID,
			"itemId":  value.Source,
		}
	}
	return map[string]any{"phrasesData": new}
}

// Sync data to s from obj
func (s *SyncNewPlayerPhrasesData) FromGo(obj any) error {
	object, success := obj.(map[string]any)
	if !success {
		return fmt.Errorf("FromGo: Failed to convert obj to map[string]interface{}; obj = %#v", obj)
	}
	// convert data
	phrases_data, success := object["phrasesData"].(map[string]any)
	if !success {
		return fmt.Errorf(`FromGo: Failed to convert object["phrasesData"] to map[string]interface{}; object["phrasesData"] = %#v`, object["phrasesData"])
	}
	// get phrases data
	*s = make(SyncNewPlayerPhrasesData, 0)
	// make object
	for source_name, multiple_phrases := range phrases_data {
		mp, success := multiple_phrases.(map[string]any)
		if !success {
			return fmt.Errorf(`FromGo: Failed to convert phrases_data[%#v] to map[string]interface{}; phrases_data[%#v] = %#v`, source_name, source_name, multiple_phrases)
		}
		// convert data
		for id, phrases := range mp {
			p, success := phrases.(map[string]any)
			if !success {
				return fmt.Errorf(`FromGo: Failed to convert mp[%#v] to map[string]interface{}; mp[%#v] = %#v`, id, id, phrases)
			}
			// convert data
			content, success := p["content"].(string)
			if !success {
				return fmt.Errorf(`FromGo: Failed to convert p["content"] to map[string]interface{}; p = %#v`, p)
			}
			id_converted, success := p["id"].(int64)
			if !success {
				return fmt.Errorf(`FromGo: Failed to convert p["id"] to map[string]interface{}; p = %#v`, p)
			}
			item_id, success := p["itemId"].(string)
			if !success {
				return fmt.Errorf(`FromGo: Failed to convert p["itemId"] to map[string]interface{}; p = %#v`, p)
			}
			// get data
			*s = append(*s, PhrasesData{
				Source:  item_id,
				ID:      id_converted,
				Content: content,
			})
			// submit sub result
		}
		// for each source, like "OFFICIAL_SIGN"
	}
	// sync data
	return nil
	// return
}
