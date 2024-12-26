package config

type Mysql struct {
	Engine       string `yaml:"engine" default:"InnoDB"`
	Path         string `yaml:"path"`
	Port         string `yaml:"port"`
	Config       string `yaml:"config"`
	DbName       string `yaml:"dbName"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	MaxIdleConns int    `yaml:"maxIdleConns"`
	MaxOpenConns int    `yaml:"maxOpenConns"`
	LogMode      string `yaml:"logMode"`
}

func (m *Mysql) Dsn() string {
	return m.Username + ":" + m.Password + "@tcp(" + m.Path + ":" + m.Port + ")/" + m.DbName + "?" + m.Config
}
