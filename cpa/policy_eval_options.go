package cpa

type evalOptions struct {
	storage map[string]interface{}
}

type EvalOption func(*evalOptions)

// Meta is an option that sets the data.meta property during policy evaluation.
func Meta(value interface{}) EvalOption {
	return func(option *evalOptions) {
		if option.storage == nil {
			option.storage = make(map[string]interface{})
		}
		option.storage["meta"] = value
	}
}
