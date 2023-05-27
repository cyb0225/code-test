# 串行化方案

每次执行操作时，对整个 S 数组加锁。

在 select 和 update 时进行了一定时间的睡眠，简单模拟了查询和修改时的性能消耗。

> 实验环境:
> ubuntu20.04（本机）  16GB 内存  16核 CPU

```shell
# 压测数据
# 没有开启事务，线程并发运行，数据有并发风险。
cyb@cyb-ThinkBook:~/project/codetest/serializable$ go run .
time usage: 226.156886ms

# 串行化执行时间消耗
cyb@cyb-ThinkBook:~/project/codetest/serializable$ go run .
time usage: 22.598271495s
```


