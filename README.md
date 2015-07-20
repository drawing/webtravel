webtravel
=========

VPN(虚拟专用网络)利用加密通信，在公共网络上建立专用网络的技术，外部网络通过加密认证，访问内网数据。SSL VPN 是解决远程用户访问敏感公司数据最简单最安全的解决技术。利用浏览器内置 SSL 协议，无需安装客户端，便可从外网访问内网应用。

webtravel 实现通过 SSL 协议访问后端服务的代理技术。其原理如下：

* 客户端浏览器通过访问代理地址 `https://proxy.com/baidu.com`
* 服务器收到 /baidu.com 的请求时，访问 `baidu.com`，把返回的html数据返回给客户端。
* 从baidu.com返回的html数据包含很多图片，css，js信息，如 `http://www.baidu.com/img/bdlogo.gif`，如果把html不加更改的返回客户端，浏览器就会从`baidu.com/img/bdlogo.gif`获取图片，撇开了 proxy.com，所以服务器端必须修改html，把其中的链接改为 `https://proxy.com/baidu.com/img/bdlogo.gif`。

其中除了html中的链接，还包括js，css中的链接，以及客户端js拼接的地址。当然，其中还需要修改http头中的cookie、referer等信息。webtravel 实现了重要逻辑的修改，可访问twitter，webqq等复杂应用。

体验地址：<https://diy-htmlshow6.rhcloud.com/>。输入要访问的URL即可，比如<http://twitter.com>。

