package generics

import "fmt"

/*
将 mapping(map[K0]V0) 转换为 map[K1]V1 。

var_name 指代 mapping 在源代码中的实际名称；
path 指代调用此函数的代码在源代码的路径。
以上两个参数被共同用于追踪引发转换错误的位置
*/
func Map[
	K0 comparable, V0 any,
	K1 comparable, V1 any,
](mapping map[K0]V0, var_name string, path string) (
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
			return nil, fmt.Errorf("Map: Failed to convert key(%T) to type %T; key = %#v, %s = %#v; path = %s", key, k, key, var_name, mapping, path)
		}
		// convert key
		v, success := V.(V1)
		if !success {
			return nil, fmt.Errorf("Map: Failed to convert %s[%#v](%T) to type %T; %s = %#v; path = %s", var_name, k, value, v, var_name, mapping, path)
		}
		// convert value
		result[k] = v
		// sync data
	}
	// convert data and sync
	return
	// return
}

/*
将 slice([]V0) 转换为 []V1 。

var_name 指代 mapping 在源代码中的实际名称；
path 指代调用此函数的代码在源代码的路径。
以上两个参数被共同用于追踪引发转换错误的位置
*/
func Slice[V0 any, V1 any](slice []V0, var_name string, path string) (
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
			return nil, fmt.Errorf("Slice: Failed to convert %s[%d](%T) to type %T; %s = %#v; path = %s", var_name, index, value, v, var_name, slice, path)
		}
		// convert data
		result[index] = v
		// sync data
	}
	// convert data and sync
	return
	// return
}

/*
将 value 转换为类型为 V 的变量。

var_name 指代 mapping 在源代码中的实际名称；
path 指代调用此函数的代码在源代码的路径。
以上两个参数被共同用于追踪引发转换错误的位置
*/
func To[V any](value any, var_name string, path string) (
	result V,
	err error,
) {
	result, success := value.(V)
	if !success {
		err = fmt.Errorf("To: Failed to convert %s(%T) to type %T; %s = %#v; path = %s", var_name, value, result, var_name, value, path)
	}
	return
}
