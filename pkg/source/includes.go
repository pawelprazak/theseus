package source

type Includes map[string]string

func NewIncludes(items ...string) Includes {
	res := make(Includes)

	for _, item := range items {
		res[item] = ""
	}

	return res
}

func (i Includes) ShouldInclude(item string) bool {
	if len(i) == 0 {
		return true
	}

	if _, include := i[item]; include {
		return true
	}

	return false
}
