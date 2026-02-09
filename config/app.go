package config

type App struct {
	Env                   string   `yaml:"env"`
	Debug                 bool     `yaml:"debug"`
	PublicKey             string   `yaml:"public_key"`
	PrivateKey            string   `yaml:"private_key"`
	AesKey                string   `yaml:"aes_key"`
	AdminEmail            []string `yaml:"admin_email"`
	AllowPhoneRegistration bool    `yaml:"allow_phone_registration"` // 是否允许手机号注册，默认 false
	RequireInviteCode     bool     `yaml:"require_invite_code"`      // 是否需要邀请码注册，默认 false
}
