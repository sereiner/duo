package conf

import (
	"github.com/spf13/viper"
	"strings"
)

type ISystemConf interface {
	AppName() string
	IsDebug() bool
	GetAddr() string
}

type DBConf interface {
}

type ZipKinConf interface {
	ZipKinEnable() bool
	GetZipKinReportURL() string
}

type IServerConf interface {
	ISystemConf
	ZipKinConf
}

type ServerConf struct {
}

func NewServerConf(configPath string) *ServerConf {

	arr := strings.Split(configPath, "/")
	Init(strings.Join(arr[:len(arr)-1], "/"), arr[len(arr)-1])

	return &ServerConf{}
}

func Init(configDir, configName string) {

	viper.SetConfigName(configName)

	viper.AddConfigPath(configDir)

	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func (c *ServerConf) AppName() string {
	return viper.GetString("app.name")
}

func (c *ServerConf) IsDebug() bool {
	return viper.GetBool("is_debug")
}

func (c *ServerConf) GetAddr() string {

	return viper.GetString("app.port")

}

func (c *ServerConf) ZipKinEnable() bool {
	return viper.GetBool("zipkin.enable")
}
func (c *ServerConf) GetZipKinReportURL() string {
	return viper.GetString("zipkin.url")
}
