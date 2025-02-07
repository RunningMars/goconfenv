env环境变量管理Client(golang依赖的版本：go1.11.2)

Step0: 构建变量管理系统,并开放变量查询接口 (localhost:8002/api/variableList) 返回数据结构如下:
{
    "code": 0,
    "data": {
        "PUSHER_APP_ID": "appId",
        "QY_WEIXIN_KEY": "dsjlxshadf",
        "SLB_A": "127.0.0.1"
    },
    "msg": "success"
}

Step1: 切换到项目根路径下，执行 go build，得到一个可执行文件 goconfenvcli

Step2: 在项目跟路径下新增cem.conf文件（参考本仓库的cem.conf文件），并配置对应的.env.template文件（参考本仓库.env.template文件）。

Step3: 执行goconfenvcli -c cem.conf -e qa -h 127.0.0.1:8002/api/variableList -p i-admin-manage-h5-interface  (-c 配置文件地址，-e 环境变量，-h envmanager service端host，-p 项目名称)
