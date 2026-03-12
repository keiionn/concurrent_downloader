
# 一个简单的多线程并发下载器

# 功能
- 支持多线程并发下载
- 支持自定义下载目录
- 支持自定义线程数
- 支持显示下载进度

# 使用
- 在与main.go 同级的文件下创建 tasks.json文件配置下载任务
	- 每个任务包含URL、文件名、线程数、下载目录
	- 示例：
	```json
	{
        "tasks":[
		    "url":"https://something/example.mp4",
            "filepath":"downloads/example.mp4",
            "filename":"example.mp4",
            "concurrency":4
        ]
	}
	```

- 运行下载器
	- 运行main.go即可
	- 下载器会读取tasks.json中的任务配置，并发下载文件
	- 每个任务完成后，会在下载目录中保存文件
	- 下载进度会在控制台中显示

- 日志文件会产生在logger/目录下
	- 日志文件包含下载器的运行信息、错误信息等

# 主要是自己学习，不建议在生产环境中使用，大佬不喜勿喷喵
