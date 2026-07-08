package vo

type UserVO struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
	Phone    string `json:"phone,omitempty"`
	Status   string `json:"status"`
}

type LoginVO struct {
	Token     string `json:"token"`
	ExpiredAt  string `json:"expiredAt"`
	User      UserVO `json:"user"`
}
