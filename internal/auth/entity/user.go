package entity

type User struct {
	ID         uint   `json:"id"`
	Email      string `json:"email"`
	Password   string `json:"-"`
	Name       string `json:"name"`
	Provider   string `json:"provider"`
	ProviderID string `json:"provider_id"`
}
