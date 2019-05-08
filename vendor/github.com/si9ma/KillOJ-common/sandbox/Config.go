package sandbox

type Config struct {
	ExePath   string `yaml:"exePath"` // sandbox exe path
	EnableLog bool   `yaml:"enableLog"`
	LogPath   string `yaml:"logPath"`
	LogFormat string `yaml:"logFormat"`
}
