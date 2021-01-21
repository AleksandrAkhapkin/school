package config

type Config struct {
	PostgresDsn      string              `yaml:"postgres_dsn"`
	ServerPort       string              `yaml:"server_port"`
	SecretKeyJWT     string              `yaml:"secret_key_jwt"`
	HtmlRecoveryPath string              `yaml:"html_recovery_path"`
	HtmlNewPassPath  string              `yaml:"html_new_pass_path"`
	Email            *ConfigForSendEmail `yaml:"server_email"`
	Soc              *SocAuth            `yaml:"soc_auth"`
	Telegram         *Telegram           `yaml:"telegram"`
	VideoDir         string              `yaml:"video_directory_path"`
}

type ConfigForSendEmail struct {
	EmailHost  string `yaml:"host"`
	EmailPort  string `yaml:"port"`
	EmailLogin string `yaml:"login"`
	EmailPass  string `yaml:"pass"`
}

type SocAuth struct {
	VKAppID       string `yaml:"vk_app_id"`
	VKCallbackURL string `yaml:"vk_callback"`
	VKSecretKey   string `yaml:"vk_secret_key"`
}

type Telegram struct {
	TelegramToken string `yaml:"telegram_token"`
	ChatID        string `yaml:"chat_id"`
	ChannelID     string `yaml:"channel_id"`
}
