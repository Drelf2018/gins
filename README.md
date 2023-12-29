# gins

gin 的超集。

厌倦了写函数再手动绑定到 `RouterGroup` 的过程？没问题，你只需要编写结构体的方法，剩下的交给 `gins` 帮你完成。

### 注意

使用此方法绑定的方法会有少许性能损耗，来源为官方反射函数的执行: `reflect.Value.Call`

### 使用

```go
package gins_test

import (
	"net/http"
	"testing"

	"github.com/Drelf2018/gins"
	"github.com/gin-gonic/gin"
)

type Auth struct{}

func (Auth) Use(c *gin.Context) {
	uid := c.Query("uid")
	if uid == "" {
		uid = "visitor"
	}
	c.Set("uid", uid)
}

func (Auth) GetPing(c *gin.Context) {
	c.JSON(200, gin.H{"msg": "hello " + c.GetString("uid")})
}

type Admin struct{}

func (*Admin) Use(c *gin.Context) {
	if c.GetString("uid") != "admin" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failure."})
		c.Abort()
	}
}

func (*Admin) GetData(c *gin.Context) {
	c.JSON(200, gin.H{"data": "some important data."})
}

func (*Admin) StaticFileCode() (string, string) {
	return "/code", "./gins.go"
}

type Main struct {
	Auth
	*Admin `router:"admin"`
}

func (Main) StaticFileIcon() (string, string) {
	return "favicon.ico", "./favicon.ico"
}

func TestGin(t *testing.T) {
	r, err := gins.Default(Main{Admin: &Admin{}})
	if err != nil {
		t.Fatal(err)
	}
	r.Run()
}
```

这是一段简单的使用说明代码，你只需要创建一个结构体，并为他编写对应请求方式开头的方法即可。方法名前缀一览：

```go
const (
	MethodGet     = "Get"
	MethodHead    = "Head"
	MethodPost    = "Post"
	MethodPut     = "Put"
	MethodPatch   = "Patch" // RFC 5789
	MethodDelete  = "Delete"
	MethodConnect = "Connect"
	MethodOptions = "Options"
	MethodTrace   = "Trace"

	StaticFileFS = "StaticFileFS"
	StaticFile   = "StaticFile"
	StaticFS     = "StaticFS"
	Static       = "Static"
)
```

方法按照原先 `gin.HandlerFunc` 格式编写即可，只不过要多加上接收器。

```go
func (Auth) GetPing(c *gin.Context) {
	c.JSON(200, gin.H{"msg": "hello " + c.GetString("uid")})
}
```

### 测试

```
[GIN-debug] GET    /favicon.ico              --> ... (3 handlers)
[GIN-debug] HEAD   /favicon.ico              --> ... (3 handlers)
[GIN-debug] GET    /ping                     --> reflect.methodValueCall (4 handlers)
[GIN-debug] GET    /admin/code               --> ... (5 handlers)
[GIN-debug] HEAD   /admin/code               --> ... (5 handlers)
[GIN-debug] GET    /admin/data               --> reflect.methodValueCall (5 handlers)
```

以上是运行测试代码后的方法绑定情况，可以看到 `Main` 结构体的方法 `StaticFileIcon` 先被绑定了。在绑定完当前结构体的方法后，会顺序遍历并绑定他的子字段。

于是 `Auth` 的 `GetPing` 方法被绑定到了 `/ping` 下，同时后面的 `4 handlers` 也增加了一个，这是因为 `Auth` 有 `Use` 方法，它会在绑定其他方法前被优先绑定。

```go
func (s *Scanner) String() string {
	s.len = len(s.s)
	s.index = -1
	buf := &bytes.Buffer{}
	for s.Next() {
		b := s.Read()
		if 'A' <= b && b <= 'Z' {
			buf.Write([]byte{'/', b + 32})
		} else if b == '_' && s.Next() {
			b = s.Read()
			switch b {
			case '8':
				buf.WriteString("/*")
			case '1':
				buf.WriteString("/:")
			default:
				buf.WriteByte(b)
			}
		} else {
			buf.WriteByte(b)
		}
	}
	return buf.String()
}
```

注意到库中有一个扫描器，他会分析方法的名字（前缀会自动去除），对于前面提到的 `GetPing` 方法名，在经过分析后会变成 `/ping`，这也就解释了前面为何会绑定在这个地址。

接着遇到字段 `*Admin` ，这是一个指针结构体，同样会分析其方法。注意到他有标签 `router:"admin"` ，这代表要在当前 `RouterGroup` 上新建子 `Group` ，若使用 `router:"-"` 则代表忽略该字段。同时当目前字段名与所用结构体名不同时也会新建子 `Group` 。

```go
// bind fields' methods
for i, field := range Reflect.FieldOf(elem.Type()) {
    next := r
    if path, ok := field.Tag.Lookup("router"); ok {
        if path == "-" {
            continue
        }
        next = r.Group(path)
    } else if field.Name != field.Type.Name() {
        next = r.Group(ParseName(field.Name))
    }

    err = UnsafeBind(next, elem.Field(i))
    if err != nil {
        return
    }
}
```

也就是说 `Admin` 的方法会被绑定到 `/admin` 路径下，事实也确实如此，`Admin` 先通过 `Use` 方法鉴权，再决定是否返回数据，测试如下：

```py
# http://localhost:8080/admin/data
{
    "error": "Authentication failure."
}

# http://localhost:8080/admin/data?uid=admin
{
    "data": "some important data."
}
```