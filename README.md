2019.7.4

通过了规模为1000的压力测试。有52个dns请求在1s内未回应。

2019.7.9

参数为200个udp协程

通过了规模为2000的压力测试。有0~40个dns请求在2s内未回应。

高压力下，平均每秒完成950个请求，丢失20个请求。

参数为1000个udp协程

通过了规模为2000的压力测试。有0~40个dns请求在2s内未回应。

高压力下，平均每秒完成950个请求，丢失20个请求。

更高压力下，平均每秒完成1500个请求，丢失2500个请求。

可能是受远程dns局限

发送了10万条消息，udp没有发生错误。