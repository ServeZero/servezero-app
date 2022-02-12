package config

type CookieConfig struct {
	NAME string
}

func GetCookieConfig() *CookieConfig {
	return &CookieConfig{
		NAME: "zero_session",
	}
}
