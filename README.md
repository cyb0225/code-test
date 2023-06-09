# code-test

MonographData 公司笔试题

## 题目要求

给定长度数组 S，有 M 个 worker 并发访问并更新 S, 每个 worker 重复如下操作：

随机生成 i, j （在 S 范围内），使 S[j] = S[i] + S[i + 1] + S[i + 2], 其中使用 i + 2 mod len(S) 避免数组越界。

> 题目提供的思路

- 注意读写锁的区别
- j 可能落在 S[i, i+2] 区间
- 如果出现死锁如何避免

## 初步思考

以数据库的视角来考虑这道题的话， S 就是一张表，M 是多个客户端来访问这张表。

使用 select 查询出 S[i], S[i + 1], S[i + 2] 的数据，并使用 update 更新到 S[j] 可以看成一个完整的事务。

而我们要处理的问题是**事务并发执行的隔离性问题**。

## 分析可能产生的问题

要求保证一个事务里不会出现锁征用死锁的问题，如 j 和 i 相同导致死锁。

首先在获取 select 数据时，我们需要获取到的数据是开启事务时的数据，如果在获取过程中发生提交事务，那么就无法保证结果的正确性。
但是，按照题目的意思，应该是直接使用读写锁，无需考虑其他事务中途提交的数据，且这个事务是一定会执行成功的，所以我只需要保证并发正确性即可，给每条记录加读写锁。

## 锁的粒度

如果直接对在整个 S 加锁，类似串行化的隔离级别，那么处理起来会很方便，但是执行起来效率会很低，具体可以查看 [串行化](pkg/serializable.go) 下的代码和实验结果。

设计行锁，由于题目提供的并发执行的 worker 数量及每次操作的纪录数是远小于总数据量的，所以很少会造成记录上的冲突，所以可以使用行锁来提升性能。

行锁使用读写锁，可以在两个协程并发读取数据时获得更好的性能。


## tow-phase locking 二阶段锁

加锁是每执行一条语句就往上加锁，但是锁的释放是在事务执行完毕后释放的。

二阶段锁带来的优势是保证操作存在相同数据区间的事务是串行化处理的，保证了事务的隔离性，满足并发保护及原子性操作。

> 原理：在一个事务执行过程中，他会持有所有锁，这样在读写操作或写写操作冲突时，只能有一个事务执行成功。

## 死锁问题

### 多个协程互相占有对方的锁

#### 问题描述

worker1 获取到的 j=4，i=1

worker2 获取到的 j=1，i=4

假设我们在操作时，依次获取 S[i], S[i+1], S[i+2], S[j] 的锁。

那么在这种情况下， worker1 先获取了 S[1], S[2], S[3] 的锁，worker2 先获取了 S[4],S[5],S[6] 的锁，没有发生锁冲突。
但是在 worker1 获取 S[j] 即 S[4] 时由于 worker2 已经持有锁了，所以开始等待锁释放。 
worker 在获取 S[1] 的时候也因为锁被 worker1 持有了，发生等待，所以产生死锁。

#### 解决方案

首先需要**检测出死锁**，然后针对死锁采取不同的策略。

1. 比如设置超时时间，检测慢事务，将其所在的事务强制关闭，或关闭协程。由于我们这一步操作执行是比较快的，我们可以针对慢事务设置一定的时间阈值判断。
2. 主动退避，当前事务获取锁失败后立马释放自己的其他锁并关闭事务，避免由于无法获取锁并占有其他锁导致死锁。

综合考虑，由于第一种方案阈值难以设计，设置短了容易误关闭事务，设置长了影响性能，且本身这个监控在 go 语言中实现比较困难，所以选择第二种方式。
第二种方式可以使用 go sync.Mutex 的 `tryLock` 方法保证，每次获取锁都使用 `tryLock`，遇到锁被占用的情况就及时主动退避。

但是，第二种方案其实仍然有弊端，会产生**活锁**。

但是这个发生的概率十分的小，但是确确实实会发生，目前的解决方法是第一个被迫退出的事务在下一次开启事务前先进行短暂的睡眠等待产生锁征用事务执行完毕后，
再次进行获取锁操作。


### 同一个协程多次获取同一个记录

#### 问题描述

worker 随机出来的 j 与 i、i + 1、i + 2 可能是相同的，那么如果第二次不加以判断，直接在获取锁，大概率是会锁死的。

#### 解决方案

在查询数据前，先判断一下 j 是否与 i、i + 1、i + 2 相等，如果存在相等的字段，则直接对当前字段加写锁即可，且使用 `x += y` 的模式更新字段。

## 其他改进方向

> 我认为还可以做性能优化的地方，但是目前还没有真正实现，目前做到的版本还是使用读写锁的方式。

### MVCC 

该案例只对 S[j] 进行 update 操作， 对于其他的 S[i], S[i+1], S[i+2] 都是 select 操作，这在 MySQL 里可以称之为快照读，针对这部分读操作，
或许我们并不需要加锁，可以设计 MVCC 视图来获取数据，在性能上取得进一步提升（性能提升的程度主要看锁冲突的程度），也可以满足**更高级别的隔离性要求**。

其实，在一开始，我们其实分析出来了这个模型是没有幻读和不可重复读问题的，他没有范围读取，也不会对同一个数据读取两次。所以其实可以按照 MySQL 读提交和可重复读的方式进行设计。
那么这两种方式的实现差异其实就是 MVCC 创建 read view 的时机。

使用 MVCC 就无需对某个记录加锁，那么只需要对 update 的数据加锁即可，不会产生死锁的情况。

### 并发原语 Atomic

引用上面得出来的结论，我们在每个事务其实只需要加一个锁给更新的记录，那么针对改记录的修改可以使用 cas 乐观锁模式。

go 语言的 goroutine 是在用户代码层调用的，使用正常的互斥锁无法达到用户代码层面的调度。所以 go sync 里面的锁也都是使用 atomic 原子操作实现的（即
cpu 提供的并发原语）。所以，可以使用 atomic 来减少锁调用的开销。


## 实验数据

> 实验环境:
> ubuntu20.04（本机）  16GB 内存  16核 CPU

### serializable 加表锁

```shell
cyb@cyb-ThinkBook:~/project/codetest$ go build -o main .
cyb@cyb-ThinkBook:~/project/codetest$ ./main -mode table
time usage: 25.00223121s
```

### record_lock 加记录锁

可以看到会遇到死锁的情况，但是相对所有操作来说不多，时间上性能比加表锁快上不少。

```shell
cyb@cyb-ThinkBook:~/project/codetest$ go build -o main .
cyb@cyb-ThinkBook:~/project/codetest$ ./main -mode record
dead lock
retry
dead lock
retry
dead lock
retry
dead lock
retry
dead lock
retry
dead lock
retry
dead lock
retry
dead lock
retry
dead lock
retry
dead lock
retry
dead lock
retry
dead lock
retry
dead lock
retry
dead lock
retry
dead lock
retry
dead lock
retry
dead lock
retry
dead lock
retry
dead lock
retry
dead lock
retry
dead lock
retry
dead lock
retry
dead lock
retry
dead lock
retry
dead lock
retry
time usage: 203.387967ms
```
