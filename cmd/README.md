# [花瓣网图片批量上传工具/图床/typora插件](https://github.com/pzx521521/typora-plugin-huaban)
## 参考
[python版本的](https://github.com/Pingze-github/HuabanBatchUpload)这个太老了 已经不能用了,现在全部是V3接口
[github](https://github.com/pzx521521/typora-plugin-huaban)
[huabanv3 api](https://github.com/pzx521521/huabanv3)

## 功能说明  
这是一个用于向花瓣网批量上传/下载 图片的程序。需要用户名密码  
+ 多线程并发，可以快速上传大量文件
+ 自动传输错误重试

## 比较  
|          | 跨域 | 网页直接使用 | 不压缩          |
|----------|---|--------|--------------|
| bilibili | ✅ | ❌(需要no-referrer)  | ❌(部分压缩机制不知道) |
| huaban   | ❌ | ✅      | ✅            |

## 获取可执行版方法  
[github-release](https://github.com/pzx521521/typora-plugin-huaban/releases)
[Windows](http://app.parap.us.kg/huaban/huaban.exe)
[Mac](http://app.parap.us.kg/huaban/huaban)
[Linux](http://app.parap.us.kg/huaban/huaban_linux)
## `config.json`参数解释  
```json
{
  "name": "your_account",//用户名
  "pass": "your_password",//密码
  "board": "your_board",//画板名字 可以不填,或者不配置,此时不会添加到画板中
  "dir": ["/Users/parapeng/Downloads", "./upload.gif"],//上传的文件或者文件夹,在传参的时候,参数会替换改配置
  "debug": true //是否输出日志,建议开启不影响typora使用
}
```
## 使用方法 1
`config.json`配置 **用户名 密码 画板名**,

```json
{
  "name": "your_account",
  "pass": "your_password",
  "board": "your_board"
}
```
把文件/文件夹 拖上去就好, 或者在命令行中执行：`huaban /path/to/images`
## 使用方法 2
`config.json`配置好 **用户名 密码 画板名** 和 文件/文件夹
```json
{
  "name": "your_account",
  "pass": "your_password",
  "board": "your_board",
  "dir": ["/Users/parapeng/Downloads", "./upload.gif"],
  "debug": true
}
```
运行就好

## 使用方法 3
在命令行中执行：`huaban -name=your_account -pass=your_password -board=your_board -dir=/path/to/images -debug=true`

## typora使用
### `config.json`配置好 **用户名 密码 画板名** 
```json
{
  "name": "your_account",
  "pass": "your_password",
  "board": "your_board",
  "debug": false
}
```
### 配置插件

![QQ_1730350431222](https://gd-hbimg.huaban.com/21626e75dd404d7c5de6003edee2022f07d24abf23968-e0dCE7)

+ 点击测试, 如果失败 请设置debug为true,查看报错信息
  + 如果出现了`操作过快,怀疑你是机器人`,手动去网页登录一下,会有一个右滑验证,验证完可以正常使用

![QQ_1730350472124](https://gd-hbimg.huaban.com/b5c7a212001b0ad6b4e15e70d9d4b5384f911e4f2cd95-lCcIAs)

### 可以直接复制粘贴/拖上来

![output](https://gd-hbimg.huaban.com/1c528e76c82479b818d39069980c7a5d1743e96f5797f-8iQEPg)

### 可以输入链接后上传

![upload](https://gd-hbimg.huaban.com/e5c68d00cad1add003e94fbc42426100403f3e6f1187a7-dJC6l9)

### typora免费版下载
- [Windows](https://typora.io/windows/dev_release.html)
- [windows x64 国内OSS镜像下载](https://jiali0126.oss-cn-shenzhen.aliyuncs.com/typora/typora-update-x64-1117.exe)
- [Mac](https://typora.io/dev_release.html)

### MacOS 注意事项
Macos 平台的都是需要授权该可执行文件的  
M1芯片的Mac，需要执行以下命令  
chmod a+x ./ 文件名  
非M1芯片的，设置打开方式为终端打开，尝试打开时会提示无权限，然后去系统偏好设置->通用，点击允许   
