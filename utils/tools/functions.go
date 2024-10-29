package tools

func ContainsAll(a, b []string) bool {
	if len(b) == 0 {
		return true
	}
	if len(a) == 0 {
		return false
	}
	elementMap := make(map[string]struct{})
	for _, v := range a {
		elementMap[v] = struct{}{}
	}

	for _, v := range b {
		if _, found := elementMap[v]; !found {
			return false
		}
	}
	return true
}

func ContainOne(a []string, b [][]string) bool {
	if len(b) == 0 {
		return true
	}
	if len(a) == 0 {
		return false
	}
	elementMap := make(map[string]struct{})
	for _, v := range a {
		elementMap[v] = struct{}{}
	}
	for _, v := range b {
		flag := false
		for _, v1 := range v {
			if _, found := elementMap[v1]; found {
				flag = true
				break
			}
		}
		// 某个多选一未通过
		if !flag {
			return false
		}
	}
	return true
}
