package internal

func Merge(a, b any) any {
	switch va := a.(type) {
	case map[string]any:
		vb, ok := b.(map[string]any)
		if !ok {
			return b
		}
		return MergeMaps(va, vb)
	case []any:
		vb, ok := b.([]any)
		if !ok {
			return b
		}
		return mergeSlice(va, vb)
	default:
		return b
	}
}

func MergeMaps(a, b map[string]any) map[string]any {
	result := make(map[string]any)
	for k, va := range a {
		if vb, ok := b[k]; ok {
			if vb == nil {
				continue
			}
			result[k] = Merge(va, vb)
		} else {
			result[k] = va
		}
	}

	for k, vb := range b {
		if _, ok := result[k]; ok {
			continue
		}
		if vb == nil {
			continue
		}
		result[k] = vb
	}

	return result
}

func mergeSlice(a, b []any) []any {
	result := make([]any, 0, max(len(a), len(b)))

	for i := 0; i < cap(result); i++ {
		switch {
		case len(a) > i && len(b) > i:
			if b[i] == nil {
				continue
			}
			result = append(result, Merge(a[i], b[i]))
		case len(a) > i:
			result = append(result, a[i])
		case len(b) > i:
			result = append(result, b[i])
		}
	}

	return result
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
