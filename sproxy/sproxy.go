package sproxy

func Start() {
	initConfig()
	initDb()
	startHttp()
}
