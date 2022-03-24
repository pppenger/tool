package pkg

var (
	tpPackage   = "package %s\n\n"
	tpImport    = "import (\n\t%s\n)\n\n"
	tpVar       = "var (\n\t%s\n)\n"
	tpTestReset = "\n\t\tReset(func() {%s\n\t\t})"
	tpTestFunc  = "func Test%s%s(t *testing.T){%s\n\tConvey(\"%s\", t, func(){\n\t\t%s\tConvey(\"When everything goes positive\", func(){\n\t\t\t%s\n\t\t\t})\n\t\t})%s\n\t})\n}\n\n"
	tpTestMain  = `func TestMain(tm *testing.M) {
    daenerys.Init(daenerys.ConfigPath("%s"))
	conf, err := conf.Init()
	if err != nil {
		panic(err)
	}

	%s

	%s
	tm.Run()
}
`
	tpOldTestMain = `func TestMain(m *testing.M) {
	path := "%s"
	var bc interface{}
	conf, err := rpc.NewConfigToml(path, &bc)
	if err != nil {
		panic(err)
	}

	%s

	%s
	m.Run()
}
`
	tplMock = `	// // import "git.inke.cn/BackendPlatform/golang/redis"
	// // 创建一个内存redis
	// redisClient, closeRedis, _ := redis.NewMockRedis()
	// defer closeRedis()
	// // 注入到框架，redis-name就是配置文件redis的名称
	// daenerys.Default.AddRedisClient("redis-name", redisClient)
	//
	// // import "git.inke.cn/BackendPlatform/golang/sql"
	// // 创建一个mysql实例，sql.sql为初始化sql路径
	// mysqlClient, closeMysql, _ := sql.NewMockSQL("../sql.sql")
	// defer closeMysql()
	// // 注入到框架，mysql-name为配置文件mysql名称
	// daenerys.Default.AddSqlClient("mysql-name", mysqlClient)
	// sql.SQLGroupManager.Add("mysql-name", mysqlClient) // 老框架注入
	//
	// // import "git.inke.cn/BackendPlatform/golang/kafka"
	// // 创建一个同步mockKafkaClient
	// syncClient, mock, _ := kafka.NewMockSyncProducerClient()
	// defer syncClient.Close()
	// mock.ExpectSendMessageAndSucceed() // 期待收到一条消息
	// daenerys.Default.AddSyncKafkaClient("kafka-name", syncClient)
	// 
	// // 创建一个异步mockKafkaClient
	// asyncClient, amock, _ := kafka.NewMockAsyncProducerClient()
	// amock.ExpectInputAndSucceed() // 期待收到一条消息
	// daenerys.Default.AddAsyncKafkaClient("async-producer-name", asyncClient)
	//
	// // import "git.inke.cn/inkelogic/daenerys/http/client"
	// // 创建一个mockHttpClient
	// httpClient := client.Func(func(req *client.Request) (*client.Response, error) {
	// 	resp := &http.Response{
	// 		StatusCode: 200,
	// 		Body:       ioutil.NopCloser(bytes.NewBuffer([]byte("{\"dm_error\":0,\"error_msg\":\"操作成功\"}"))),
	// 	}
	// 	return client.BuildResp(nil, resp)
	// })
	// // 注入到框架，service-name为配置文件下游服务名称
	// daenerys.Default.AddHTTPClient("service-name", httpClient)

	// // 其他函数mock可参考：https://code.inke.cn/BackendPlatform/monkey
`
)
