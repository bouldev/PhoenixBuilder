package protocol

import (
	"fmt"
	"reflect"

	// A Python library which named "msgpack"
	"github.com/ugorji/go/codec"
)

// The following fields describe the
// type of map in the PyRpc packet
const (
	MapKeyTypeUnknown   = uint8(iota) // ...
	MapKeyTypeString                  // map[string]any
	MapKeyTypeUint64                  // map[uint64]any
	MapKeyTypeInt64                   // map[int64]any
	MapKeyTypeInterface               // map[any]any (NOT SUPPORTED/BLOCKED)
)

// Marshal obj to binary msg pack
func MarshalMsgpack(obj any) (msg_pack []byte, err error) {
	var msg_pack_handler codec.MsgpackHandle
	err = codec.NewEncoderBytes(&msg_pack, &msg_pack_handler).Encode(obj)
	if err != nil {
		err = fmt.Errorf("MarshalMsgpack: %v", err)
	}
	return
}

// Unmarshal msg_pack to go values which
// only contains go-built-in types
func UnmarshalMsgpack(msg_pack []byte) (result any, err error) {
	var msg_pack_handler codec.MsgpackHandle
	msg_pack_handler.RawToString = true
	err = codec.NewDecoderBytes(msg_pack, &msg_pack_handler).Decode(&result)
	if err != nil {
		err = fmt.Errorf("UnmarshalMsgpack: %v", err)
	}
	return
}

/*
Format mapping to make all the map included
are "map[string]any" or "map[uint64]any" or
"map[int64]any".

Formatting may cause changes to the original
value, which is uncertain.

Maps with the key type interface{} are not
allowed because they cannot be Jsonized,
which is fatal for packet research and Bug
fixing.
*/
func FormatMapInMsgpack(mapping map[any]any) (result any, err error) {
	map_type := MapKeyTypeUnknown
	if len(mapping) == 0 {
		map_type = MapKeyTypeString
	}
	// prepare
	for key := range mapping {
		switch key.(type) {
		case string:
			if map_type != MapKeyTypeUnknown && map_type != MapKeyTypeString {
				return nil, fmt.Errorf("FormatMapInMsgpack: We expect the map key is string, but it is %T; mapping = %#v", key, mapping)
			}
			map_type = MapKeyTypeString
		case uint64:
			if map_type != MapKeyTypeUnknown && map_type != MapKeyTypeUint64 {
				return nil, fmt.Errorf("FormatMapInMsgpack: We expect the map key is uint64, but it is %T; mapping = %#v", key, mapping)
			}
			map_type = MapKeyTypeUint64
		case int64:
			if map_type != MapKeyTypeUnknown && map_type != MapKeyTypeInt64 {
				return nil, fmt.Errorf("FormatMapInMsgpack: We expect the map key is int64, but it is %T; mapping = %#v", key, mapping)
			}
			map_type = MapKeyTypeInt64
		default:
			if map_type != MapKeyTypeUnknown && map_type != MapKeyTypeInterface {
				return nil, fmt.Errorf("FormatMapInMsgpack: We expect the map key is interface{}, but it is %T; mapping = %#v", key, mapping)
			}
			map_type = MapKeyTypeInterface
		}
	}
	// determine the key type of the map.
	// actually speaking, this is just the assumption we make
	// that the types of keys in the map are consistent
	for key, value := range mapping {
		if value == nil {
			continue
		}
		switch reflect.TypeOf(value).Kind() {
		case reflect.Map:
			val, success := value.(map[any]any)
			if !success {
				return nil, fmt.Errorf("FormatMapInMsgpack: Unsupported map type %T; value = %#v", value, value)
			}
			// convert data
			mapping[key], err = FormatMapInMsgpack(val)
			if err != nil {
				return nil, fmt.Errorf("FormatMapInMsgpack: %v", err)
			}
			// format sub map
		case reflect.Slice:
			val, success := value.([]any)
			if !success {
				return nil, fmt.Errorf("FormatMapInMsgpack: Unsupported slice type %T; value = %#v", value, value)
			}
			// convert data
			mapping[key], err = FormatSliceInMsgpack(val)
			if err != nil {
				return nil, fmt.Errorf("FormatMapInMsgpack: %v", err)
			}
			// format sub slice
		}
	}
	// format specific element in the map
	switch map_type {
	case MapKeyTypeString:
		new_map := map[string]any{}
		for k, val := range mapping {
			new_map[k.(string)] = val
		}
		return new_map, nil
	case MapKeyTypeUint64:
		new_map := map[uint64]any{}
		for k, val := range mapping {
			new_map[k.(uint64)] = val
		}
		return new_map, nil
	case MapKeyTypeInt64:
		new_map := map[int64]any{}
		for k, val := range mapping {
			new_map[k.(int64)] = val
		}
		return new_map, nil
	default:
		return nil, fmt.Errorf("FormatMapInMsgpack: Unsupported map key type interface{}; mapping = %#v", mapping)
	}
	// sync data to the new map and return
}

/*
Format slice to make all the map included
are "map[string]any" or "map[uint64]any"
or "map[int64]any".

Formatting may cause changes to the original
value, which is uncertain.
*/
func FormatSliceInMsgpack(slice []any) (result any, err error) {
	new := make([]any, len(slice))
	// prepare
	for index, value := range slice {
		if value == nil {
			new[index] = value
			continue
		}
		// check data
		switch reflect.TypeOf(value).Kind() {
		case reflect.Map:
			val, success := value.(map[any]any)
			if !success {
				return nil, fmt.Errorf("FormatSliceInMsgpack: Unsupported map type %T; value = %#v", value, value)
			}
			// convert data
			new[index], err = FormatMapInMsgpack(val)
			if err != nil {
				return nil, fmt.Errorf("FormatSliceInMsgpack: %v", err)
			}
			// format sub map
		case reflect.Slice:
			val, success := value.([]any)
			if !success {
				return nil, fmt.Errorf("FormatSliceInMsgpack: Unsupported slice type %T; value = %#v", value, value)
			}
			// convert data
			new[index], err = FormatSliceInMsgpack(val)
			if err != nil {
				return nil, fmt.Errorf("FormatSliceInMsgpack: %v", err)
			}
			// format sub slice
		default:
			new[index] = value
			// just sync data if the type of
			// value is not map or slice
		}
	}
	// format each element in the slice
	return new, nil
	// return
}
