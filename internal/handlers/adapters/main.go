package adapters

type AuthAdapter interface {
	GetLink() (string, error)
}

var list map[string]AuthAdapter

func GetAdaptersList() map[string]AuthAdapter {
	if list == nil || len(list) <= 0 {
		list = make(map[string]AuthAdapter, 5)

		list["instagram"] = GetInstagramAdapter()
	}

	return list
}
