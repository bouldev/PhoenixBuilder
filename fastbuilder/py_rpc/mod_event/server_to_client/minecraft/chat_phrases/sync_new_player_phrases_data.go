package chat_phrases

import "fmt"

const (
	SourceOfficial = "OFFICIAL_SIGN"
)

type SyncNewPlayerPhrasesData struct {
	Data []PhrasesData `json:"phrasesData"` // ...
	Show bool          `json:"show"`        // e.g. true
}

type PhrasesData struct {
	Source  string `json:"itemId"`  // e.g. "OFFICIAL_SIGN"
	ID      uint64 `json:"id"`      // e.g. uint64(1002)
	Content string `json:"content"` // e.g. "加个好友吧"
}

// Return the event name of s
func (s *SyncNewPlayerPhrasesData) EventName() string {
	return "SyncNewPlayerPhrasesData"
}

// Convert s to go object which only contains go-built-in types
func (s *SyncNewPlayerPhrasesData) MakeGo() (res any) {
	new := map[string]any{}
	for _, value := range s.Data {
		_, ok := new[value.Source]
		if !ok {
			new[value.Source] = make(map[uint64]any)
		}
		new[value.Source].(map[uint64]any)[value.ID] = map[string]any{
			"content": value.Content,
			"id":      value.ID,
			"itemId":  value.Source,
		}
	}
	return map[string]any{
		"phrasesData": new,
		"show":        s.Show,
	}
}

// Sync data to s from obj
func (s *SyncNewPlayerPhrasesData) FromGo(obj any) error {
	object, success := obj.(map[string]any)
	if !success {
		return fmt.Errorf("FromGo: Failed to convert obj to map[string]interface{}; obj = %#v", obj)
	}
	// convert data
	{
		s.Data = make([]PhrasesData, 0)
		// make object
		phrases_data, success := object["phrasesData"].(map[string]any)
		if !success {
			return fmt.Errorf(`FromGo: Failed to convert object["phrasesData"] to map[string]interface{}; object["phrasesData"] = %#v`, object["phrasesData"])
		}
		// get phrases data
		for source_name, multiple_phrases := range phrases_data {
			mp, success := multiple_phrases.(map[uint64]any)
			if !success {
				return fmt.Errorf(`FromGo: Failed to convert phrases_data[%#v] to map[uint64]interface{}; phrases_data[%#v] = %#v`, source_name, source_name, multiple_phrases)
			}
			// convert data
			for id, phrases := range mp {
				p, success := phrases.(map[string]any)
				if !success {
					return fmt.Errorf(`FromGo: Failed to convert mp[%d] to map[string]interface{}; mp[%d] = %#v`, id, id, phrases)
				}
				// convert data
				content, success := p["content"].(string)
				if !success {
					return fmt.Errorf(`FromGo: Failed to convert p["content"] to string; p = %#v`, p)
				}
				id_converted, success := p["id"].(uint64)
				if !success {
					return fmt.Errorf(`FromGo: Failed to convert p["id"] to uint64; p = %#v`, p)
				}
				item_id, success := p["itemId"].(string)
				if !success {
					return fmt.Errorf(`FromGo: Failed to convert p["itemId"] to string; p = %#v`, p)
				}
				// get data
				s.Data = append(s.Data, PhrasesData{
					Source:  item_id,
					ID:      id_converted,
					Content: content,
				})
				// submit sub result
			}
			// for each source, like "OFFICIAL_SIGN"
		}
		// sync phrases data
	}
	// get and sync phrases data
	s.Show, success = object["show"].(bool)
	if !success {
		return fmt.Errorf(`FromGo: Failed to convert object["show"] to bool; object["show"] = %#v`, object["show"])
	}
	// get and sync show data
	return nil
	// return
}
