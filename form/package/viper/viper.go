package viperData


type Config struct {
	Port        int    `mapstructure:"port"`
	Version     string `mapstructure:"version"`
	MysqlConfig `mapstructure:"mysql"`
}

type MysqlConfig struct {
	Host   string `mapstructure:"host"`
	Port   int    `mapstructure:"port"`
	DbName string `mapstructure:"dbname"`
}
