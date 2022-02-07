# jenkins-bridge-client

### 编译

``` shell
go build .
```

### 使用

```
Usage of ./jenkins-bridge-client:
  -downloadArtifacts
    	是否下载产物
  -jobName string
    	要触发的 Jenkins 任务名 (default "github-pipeline")
  -printlog
    	是否打印日志
  -runid int
    	job runid
  -token string
    	bridge server token
  -triggerBuild
    	是否触发编译
```

#### 触发编译, 获取 runid
``` shell
jenkins-bridge-client -token $BRIDGE_TOKEN -triggerBuild # 使用默认jobName
jenkins-bridge-client -token $BRIDGE_TOKEN -triggerBuild -jobName $jobname # 自定义jobname
```


#### 使用runid 打印日志/获取产物

同样可以使用 -jobName 参数来指定Jenkins 任务名

打印日志:
``` shell
jenkins-bridge-client -token "$BRIDGE_TOKEN" -runid "$id" -printlog
```
获取产物
``` shell
jenkins-bridge-client -token "$BRIDGE_TOKEN" -runid "$id" -downloadArtifacts
```
