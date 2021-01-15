# Golang

## 管道（Channel）

### 概念

Go在语言层提供的协程间通信的方式

### 初始化

```go
var ch chan int			 			// 声明管道
// 值为 nil

ch1 := make(chan int)			// 无缓冲管道
ch2 := make(chan int, 2)	// 带缓冲管道 

```

### 操作

#### 操作符

操作符 "<-" 表示数据流向，在函数间传递时可以用操作符限制管道的读写

```go
ch <- 1										// 向管道写入数据
<- ch											// 从管道中读取数据

close(ch)									// 关闭管道
// 尝试向关闭的管道写入数据会触发 panic
ch <- 2

// 但关闭的管道仍可读
// 第一个变量表示读出的数据，第二个变量（bool 类型）表示是否成功读取了数据，需要注意的是，第二个变量不用于知识管道的关闭状态
v, ok := <- ch

func ChanParamRW(ch chan int) {
  // 可读可写管道
}

func ChanParamR(ch <-chan int) {
  // 只读管道
}

func chanParamW(ch chan<- int) {
  // 只写管道
}
```

#### 数据读写

管道没有缓冲区，读写数据会阻塞，直到有协程向管道中写读数据。

有缓冲区但没有缓冲数据时读操作也会阻塞协程直到有数据写入才会唤醒该阻塞协程，向管道写入数据，缓冲区满了也会阻塞，直到有协程从缓冲区读取数据。

值为 nil 的管道无论读写都会阻塞，而且是永久阻塞。



使用  select 可以监控多个管道，当其中某一个管道可操作时就触发响应的 case 分支。事实上 select 语句的多个 case 语句的执行顺序是随机的。



#### 其他操作

内置函数 `len()` 和 `cap()` 作用于管道，分别用于查询缓冲区中数据的个数以及缓冲区的大小。

管道实现了一种 FIFO（先入先出）的队列，数据总是按照写入的顺序流出管道。

协程读取管道时阻塞的条件有：

- 管道无缓冲区
- 管道的缓冲区中无数据
- 管道的值为 nil

协程写入管道时阻塞的条件有：

- 管道无缓冲区
- 管道的缓冲区已满
- 管道的值为 nil

### 实现原理

#### 数据结构

`src/runtime/chann.go: hchan`

```go
type hchan struct {
         qcount   uint           // 当前队列中的剩余元素
         dataqsiz uint           // 环形队列长度，即可以存放的元素个数
         buf      unsafe.Pointer // 环形队列指针
         elemsize uint16					// 每个元素的大小
         closed   uint32					// 标识关闭状态
         elemtype *_type // 元素类型
         sendx    uint   // send index 	 队列下标，指示元素写入时存放到队列中的位置
         recvx    uint   // receive index 队列下标，指示下一个被读取的元素在队列中的位置
         recvq    waitq  // list of recv waiters	等待读消息的协程队列
         sendq    waitq  // list of send waiters 等待写消息的协程队列
         lock mutex			// 互斥锁， chan 不允许并发读写
}
```

四个重点：

- 环形队列
- 等待队列（读写各一个）
- 类型消息
- 互斥锁

一般情况下 recvq 和 sendq 至少有一个为空。只有一个例外，那就是同一个协程使用 select 语句向管道一边写入数据一边读取数据，吃屎协程会分别位于两个等待队列中。

##### 向管道写数据

简单过程如下：

- 如果缓冲区中有空余位置，则将数据写入缓冲区，结束发送过程。
- 如果缓冲区中没有空余位置，则将当前协程加入 sendq 队列，并进入睡眠并等待被读协程唤醒

当接收队列 recvq 不为空时，说明缓冲区中没有数据但有协程在等待数据，此时会把数据直接传递给 recvq 队列的第一个协程，而不必再写入缓冲区。

##### 从管道读数据

简单过程如下：

- 如果缓冲区中有数据，则从缓冲区中取出数据，结束读取过程。
- 如果缓冲区没有数据，则将当前协程加入 recvq 队列，进入睡眠并等待被写协程唤醒。

类似的，如果等待发送队列 sendq 不为空，且没有缓冲区，那么直接从 sendq 队列的第一个协程中获取数据

##### 关闭管道

关闭管道会把 recvq 中的协程全部唤醒，这些协程或缺的数据都为 nil。 同时把 sendq 队列中的协程全部唤醒，但这些协程会触发 panic

除此之外，其他会触发 panic 的操作：

- 关闭值为 nil 的管道
- 关闭已经被关闭的管道
- 向已经关闭的管道写入数据



## 切片（slice）

slice 又称动态数组，依托数组实现，可以方便地进行扩容和传递，实际使用时比数组更灵活。

### 初始化

- 声明变量
  - 变量值都为零值，对于切片来讲零值为 nil
- 字面量
  - 空切片是指长度为空，而值不是 nil
  - 声明长度为0的切片时，推荐使用变量声明的方式获得一个 nil 切片，而不是空切片，因为 nil 切片不需要内存分配
- 内置函数make()
  - 推荐指定长度同时指定预估空间，可有效地减少切片扩容时内存分配及拷贝次数
- 切取
  - 切片与原数组或切片共享底层空间

```go
// 1. 声明变量
var s []int  // nil 切片

// 2. 字面量
s1 := []int{} // 空切片
s2 := []int{1, 2, 3} // 长度为3的切片

// 3. 内置函数 make() 
s3 := make([]int, 12) // 指定长度
s4 := make([]int, 12, 100) // 指定长度和空间

// 4. 切取
arr := [5]int{1, 2, 3, 4, 5}
s5 := arr[2:4] // 从数组中切取
s6 ；= s5[1:2] // 从切片中切取

```

### 操作

#### `append()`

当切片空间不足时，`append()` 会先创建新的大容量切片，添加元素后再返回新切片。

扩容规则

- 原 slice 的容量小于 1024，则新 slice 的容量将扩大为原来的 2 倍
- 原 slice 的容量大于或等于 1024，则新 slice 的容量将扩大为原来的 1.25 倍

`append()` 向 slice 添加一个元素的实现步骤如下：

- 加入 slice 的容量够用，则将新元素追加进去，slice.len++，返回 slice
- 原 slice 的容量不够，则将 slice 先扩容，扩容后得到新 slice
- 将新元素追加进新 slice， slice.len++，返回新的 slice

```go
s := make([]int, 0)
s = append(s, 1)  // 添加1个元素
s = append(s, 2, 3, 4 ,5) // 添加多个元素
s = append(s, []int{6, 7}...) //添加一个切片
```

#### `len()` 和 `cap()`

由于切片的本质为结构体，结构体中存储了切片的长度和容量，所以这两个操作的时间复杂度均为 O(1)

#### `copy()`

会将源切片的数据逐个拷贝到目的切片指向的数组中，拷贝数量取两个切片长度的最小值。

例如长度为 10 的切片拷贝到长度为 5 的切片中时，将拷贝 5 个元素

也就是说，拷贝过程中不会发生扩容。

### 实现原理

`src/runtime/slice.go:slice`

```go
type slice struct {
  array unsafe.Pointer
  len		int
  cap		int
}
```

array 指针指向底层数组，len 表示切片长度，cap 表示数组容量

#### 切片表达式

- 简单表达式 s[low : high]
- 扩展表达式 s[low : high : max]

使用简单表达式生成的切片将与原数组或切片共享底层数组。新切片的生成逻辑可以使用一下伪代码表示：

```go
b.array = &a[low]
b.len = high - low
b.cap = len(a) - low

// 扩展表达式中的 max 用于限制新生成切片的容量，新切片的容量为 max - low
b.cap = max - low
```

如果切片表达式发生越界就触发 panic

#### 切取 string

作用于字符串时会则会产生新的字符串，而不是切片

扩展表达式不能用于切取 string



## Map

Go 语言的 map 底层使用 Hash 表实现

### 初始化

```go
// 字面量初始
m := map[string]int {
  "apple": 2,
  "banana": 3, 
}

// 内置函数make()
m := make(map[string]int, 10)
m["apple"] = 2
m["banana"] = 3
```





## 操作 Redis

### 包： [go-redis/redis](https://github.com/go-redis/redis)

```sh
go get -u github.com/go-redis/redis/v8
```

#### 初始化

```go
// Way 1    
rdb := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "", // no password set
        DB:       0,  // use default DB
    })

// Way 2
opt, err := redis.ParseURL("redis://localhost:6379/<db>")
if err != nil {
    panic(err)
}

rdb := redis.NewClient(opt)
```

#### Set & Get & Del

```go
cmd := rdb.Set(ctx, "key", "value", 0)
cmd = rdb.Get(ctx, "key")
cmd = rdb.Del(ctx, "Key")

// Do 用于执行 Redis 命令
cmd = rdb.Do(ctx, "get", "key")
cmd = rdb.Do(ctx, "set", "key", "value", 0)

// 所有返回值为一个 *redis.Cmd 结构体实例，该实例为命令执行的一个状态管理，而 Set(), Get(), Do() 函数相当于只是一个命令的执行开关，而结果是这个命令的返回。

err := cmd.Err()
if err != nil {
  panic(err)
}

val, err := cmd.Result()

// Get 一个没有 Set 的值
val, err := rdb.Get(ctx, "OtherKey").Result()
if err != nil {
    if err == redis.Nil {
        fmt.Println("key does not exists")
        return
    }
    panic(err)
}
fmt.Println(val)

```



## GORM

#### 包：[go-gorm/gorm](https://github.com/go-gorm/gorm)

```sh
go get -u gorm.io/gorm

// 安装驱动
// GORM 官方支持的数据库类型有： MySQL, PostgreSQL, SQlite, SQL Server
go get -u gorm.io/driver/sqlite
go get -u gorm.io/driver/mysql
go get -u gorm.io/driver/postgres
go get -u gorm.io/driver/sqlserver
go get -u gorm.io/driver/clickhouse
```

#### 初始化

以连接到 MySQL 为例

```go
import (
  "gorm.io/driver/mysql"
  "gorm.io/gorm"
)

func main() {
  // 参考 https://github.com/go-sql-driver/mysql#dsn-data-source-name 获取详情

  // 通过 Open 函数获得 *gorm.DB 连接实例，之后的查询都通过这个连接进行
	db, err := gorm.Open(mysql.New(mysql.Config{
  DSN: "gorm:gorm@tcp(localhost:9910)/gorm?charset=utf8&parseTime=True&loc=Local", // data source name, refer https://github.com/go-sql-driver/mysql#dsn-data-source-name
  DefaultStringSize: 256, // add default size for string fields, by default, will use db type `longtext` for fields without size, not a primary key, no index defined and don't have default values
  DisableDatetimePrecision: true, // disable datetime precision support, which not supported before MySQL 5.6
  DontSupportRenameIndex: true, // drop & create index when rename index, rename index not supported before MySQL 5.7, MariaDB
  DontSupportRenameColumn: true, // use change when rename column, rename rename not supported before MySQL 8, MariaDB
  SkipInitializeWithVersion: false, // smart configure based on used version
}), &gorm.Config{})
  
}
```

#### 示例

```go
package main

import "gorm.io/gorm"
import "gorm.io/driver/mysql"

type Product struct {
  ID	  		uint
  Code			string
  Price			uint
  CreatedAt	time.Time
  UpdatedAt	time.Time
}

dsn := "gorm:gorm@tcp(localhost:9910)/gorm?charset=utf8&parseTime=True&loc=Local"
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
  Logger: logger.Default.LogMade(logger.Silent)
})

db.AutoMigrate(&Product{})
db.Migrator().CreateTable(&Product{})
```



#### 模型

##### 定义

以用户模型为例

```go
type User struct {
	gorm.Model
	Username string  `gorm:"uniqueIndex;not null;size:128"` // 用户名
	Password string  `gorm:"not null;size:50"`              // 密码
  Bio			 *string `gorm:"null;text"`											// 个人BB

	// 关联模型
	Articles []Article
}
```

##### 钩子【[更多](https://gorm.io/zh_CN/docs/hooks.html)】

```go
func (u *User) BeforeCreate(db *gorm.DB) error {
	// hex 密码
	u.Password = hex.EncodeToString([]byte(u.Password))

	return nil
}
```

##### 创建【[更多](https://gorm.io/zh_CN/docs/create.html)】

```
user := User{Username: "uptutu", Password: "123456", Bio: nil}

result := db.Create(&user) // 通过数据的指针来创建
// 创建后自增的主键会返回到 user 变量中，所以传值传递的是地址

user.ID             // 返回插入数据的主键
result.Error        // 返回 error
result.RowsAffected // 返回插入记录的条数
```





