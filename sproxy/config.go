package sproxy

import (
	"log"
	"path"
	"strconv"

	"github.com/hackallcode/hackonf"
)

type ConfigFile struct {
	Host     string `json:"host"`
	Port     int64  `json:"port"`
	CertDir  string `json:"cert_dir"`
	Protocol string `json:"protocol"`
	DbPath   string `json:"db_path"`
}

var (
	host     string
	port     string
	protocol string
	certPath string
	keyPath  string
	dbPath   string
)

func initConfig() {
	var configFile ConfigFile
	err := hackonf.Load([]string{"config/sproxy.json"}, &configFile)
	if err != nil {
		log.Fatalln(err)
	}

	host = configFile.Host
	port = strconv.FormatInt(configFile.Port, 10)
	protocol = configFile.Protocol
	certPath = path.Join(configFile.CertDir, host+"_"+port+"_cert.pem")
	keyPath = path.Join(configFile.CertDir, host+"_"+port+"_key.pem")
	dbPath = configFile.DbPath
}
