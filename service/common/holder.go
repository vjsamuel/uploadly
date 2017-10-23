package common

type Holder struct {
	File string
	Size int64
	User User
	ContentType string
	Description string
	Object interface{}
}

func (h *Holder) GetProfileID() string {
	return h.User.Profile
}

func (h *Holder) GetProfile() Profile {
	return Profile{
		FirstName: h.User.FirstName,
		LastName:  h.User.LastName,
	}
}
