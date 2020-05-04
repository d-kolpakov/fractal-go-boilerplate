package adapters

type InstagramAdapter struct {
}

func (a *InstagramAdapter) GetLink() (string, error) {
	return "", nil
}

func GetInstagramAdapter() AuthAdapter {
	return &InstagramAdapter{}
}
