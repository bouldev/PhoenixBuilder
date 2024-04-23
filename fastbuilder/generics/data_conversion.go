package generics

import "fmt"

// Convert map[K0]V0 to map[K1]V1
func Map[
	K0 comparable, V0 any,
	K1 comparable, V1 any,
](mapping map[K0]V0) (
	result map[K1]V1,
	err error,
) {
	result = make(map[K1]V1)
	// prepare
	for key, value := range mapping {
		var K any = key
		var V any = value
		// prepare
		k, success := K.(K1)
		if !success {
			return nil, fmt.Errorf("Map: Failed to convert key(%T) to type %T; key = %#v, mapping = %#v", key, k, key, mapping)
		}
		// convert key
		v, success := V.(V1)
		if !success {
			return nil, fmt.Errorf("Map: Failed to convert mapping[%#v](%T) to type %T; mapping = %#v", k, value, v, mapping)
		}
		// convert value
		result[k] = v
		// sync data
	}
	// convert data and sync
	return
	// return
}

// Convert []V0 to []V1
func Slice[V0 any, V1 any](slice []V0) (
	result []V1,
	err error,
) {
	result = make([]V1, len(slice))
	// prepare
	for index, value := range slice {
		var val any = value
		// prepare
		v, success := val.(V1)
		if !success {
			return nil, fmt.Errorf("Slice: Failed to convert slice[%d](%T) to type %T; slice = %#v", index, value, v, slice)
		}
		// convert data
		result[index] = v
		// sync data
	}
	// convert data and sync
	return
	// return
}

// Convert value to type V
func To[V any](value any) (result V, err error) {
	result, success := value.(V)
	if !success {
		err = fmt.Errorf("To: Failed to convert value(%T) to type %T; value = %#v", value, result, value)
	}
	return
}
