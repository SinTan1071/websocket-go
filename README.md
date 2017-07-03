# TooWhite
一个websocket消息广播系统，使用docker做服务容器

## 系统整体流程图

![alt text](process.png)

## 安装

```ssh
$ https://github.com/RT7oney/websocket-go.git
$ go get github.com/gorilla/websocket
```

## 运行

* 代码部署完成之后进入代码目录，在conf/目录下进行配置具体参数

```ssh
$ docker-compose up
```

## 接口调用

### 服务端:
GET http://xx.xxx.xx.xx:12345/s

####  请求参数
| 字段                     |   必选            |   类型及范围    | 说明                               |
|:-------------------------|:----------------- |:----------------|:-----------------------------------|
|Message|true|string|整体消息体|

### 客户端:
GET ws://xx.xxx.xx.xx:12345/c

#### 请求参数
| 字段                     |   必选            |   类型及范围    | 说明                               |
|:-------------------------|:----------------- |:----------------|:-----------------------------------|
| uid | true | string | 用户的唯一ID |
| token | true | string | 用户的系统访问令牌 |
| app | true | string | 用户的当前访问的平台 ("website", "mobile", "android", "ios")|

## 服务端Message数据字段说明

```json
{
    "Target":["1399338447532dbae5522dc5e8e34adb","569009c82a81047f7f936942c949500c"],
    "Data":"eyJUeXBlIjoxLCJBY3Rpdml0eSI6InRlYW0iLCJUbyI6MiwiQ29udGVudCI6Ijxmb250IHN0eWxlPSdjb2xvcjojNjdiYmZlJz5tZWk8L2ZvbnQ+IFx1NGZlZVx1NjUzOVx1NGU4Nlx1NTZlMlx1OTYxZjxmb250IHN0eWxlPSdjb2xvcjojNjdiYmZlJz5cdTRlZTVcdTUyNGRcdTc2ODRcdTU0MGRcdTViNTc8L2ZvbnQ+XHU3Njg0XHU1NDBkXHU3OWYwXHU0ZTNhPGZvbnQgc3R5bGU9J2NvbG9yOiM2N2JiZmUnPkxPTFx1N2ZhNDwvZm9udD4ifQ=="
}
```

#### Data字段说明：

* Data字段是服务端传递的数据，该数据为经过base64编码后的JSON对象，该数据的格式为

```json
// 动态
{
    "Type" : 1,
    "Activity" : "card", // 卡片动态card 团队动态team 场景动态scene
    "To" : 1, // 卡片或者团队或者场景对应的ID
    "Content" : "xxx修改了xxx"
}
// 消息通知
{
    "Type" : 0,
    "Content" : "欢迎加入团队"
}
```

## 客户端示例

```html
<!DOCTYPE html>
<html>

<head>
    <meta charset="UTF-8">
    <title>
        test websocket
    </title>
</head>
<style type="text/css">
div {
    width: 100%;
    height: 100px;
    text-align: center;
    padding-top: 20px;
    border: 1px solid black;
}
</style>

<body>
    消息：
    <div id="content">
    </div>
</body>
<!-- <script src="base64.js"></script> -->
<script type="text/javascript">
if (typeof(WebSocket) == 'undefined') {
    alert('你的浏览器不支持 WebSocket ，推荐使用Google Chrome 或者 Mozilla Firefox');
}
// var base = new Base64();
var user_token = prompt("用户唯一TOKEN", "");
var user_id = prompt("用户的Member_id", "");
var user_app = prompt("用户的访问平台", "");
if (user_token != "" & user_id != "" & user_app != "") {
    var uri = 'ws://127.0.0.1:12345/c?uid=' + user_id + "&token=" + user_token + "&app=" + user_app;
    // var uri = 'ws://10.65.209.164:12345/c?uid=' + user_token;
    var so;
    createSocket(uri);
}else{
    alert("请刷新页面填写完整数据");
}

/**
 * socket建立
 */
function createSocket(uri) {
    so = new WebSocket(uri);
    console.log(so);

    so.onopen = function() {
        if (so.readyState == 1) {
            console.log('===user register===', "客户端连接成功");
        }else{
            console.log('===user register===', "客户端连接失败");
        }
    }

    so.onclose = function() {
        so = false;
    }

    so.onmessage = function(recv_msg) {
        console.log('===recv_msg===', recv_msg);
        // var decodeStr = Base64.decode(recv_msg.data);
        // console.log('recv_msg.data------' + decodeStr);
    }
}
</script>

</html>

```
