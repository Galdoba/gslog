package logger

import (
	"reflect"
	"sync"
)

type typeCache struct {
	fieldNames []string
	fieldInfos []fieldInfo
}

var cache sync.Map // map[reflect.Type]*typeCache

type fieldInfo struct {
	fullName string
	index    []int
}

// collectFields recursively collects all field names (dotted) of a struct type.
func collectFields(t reflect.Type, prefix string) []string {
	var result []string
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}
		name := prefix + f.Name
		if f.Type.Kind() == reflect.Struct {
			nested := collectFields(f.Type, name+".")
			result = append(result, nested...)
		} else {
			result = append(result, name)
		}
	}
	return result
}

// getFieldInfos returns cached fieldInfo for the given struct type.
// If t is not a struct, it returns an empty slice.
func getFieldInfos(t reflect.Type) []fieldInfo {
	if t.Kind() != reflect.Struct {
		return nil
	}
	if cached, ok := cache.Load(t); ok {
		return cached.(*typeCache).fieldInfos
	}
	fieldNames := collectFields(t, "")
	infos := buildFieldInfos(t, nil, "")
	cache.Store(t, &typeCache{
		fieldNames: fieldNames,
		fieldInfos: infos,
	})
	return infos
}

// buildFieldInfos recursively builds fieldInfo for each leaf field of a struct.
func buildFieldInfos(t reflect.Type, parentIndex []int, prefix string) []fieldInfo {
	var result []fieldInfo
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}
		idx := make([]int, len(parentIndex)+1)
		copy(idx, parentIndex)
		idx[len(parentIndex)] = i
		fullName := prefix + f.Name
		if f.Type.Kind() == reflect.Struct {
			nested := buildFieldInfos(f.Type, idx, fullName+".")
			result = append(result, nested...)
		} else {
			result = append(result, fieldInfo{
				fullName: fullName,
				index:    idx,
			})
		}
	}
	return result
}

// isZeroValue reports whether v is the zero value for its type or if v is invalid (e.g., nil).
func isZeroValue(v reflect.Value) bool {
	// If v is the zero reflect.Value (invalid), treat it as zero.
	if !v.IsValid() {
		return true
	}
	return v.IsZero()
}
