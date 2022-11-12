package beego

import "github.com/beego/beego/v2/server/web"

type UserController struct {
	web.Controller
}

func (c *UserController) GetUser() {
	c.Ctx.WriteString("hello, i am whysk8")
}

func (c *UserController) CreateUser() {
	user := &User{}
	err := c.Ctx.BindJSON(user)
	if err != nil {
		c.Ctx.WriteString(err.Error())
		return
	}

	_ = c.JSONResp(user)
}

type User struct {
	Name string
}
