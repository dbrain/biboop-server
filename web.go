package biboop

import (
  "appengine"
  "appengine/user"
  "github.com/dbrain/soggy"
  "net/http"
)

func WebUserRequired(ctx *soggy.Context) {
  if ctx.Env["googleUser"] == nil {
    aeCtx := ctx.Env["aeCtx"].(appengine.Context)
    url, _ := user.LoginURL(aeCtx, ctx.Req.URL.Path)
    http.Redirect(ctx.Res, ctx.Req.Request, url, 302)
    return
  }
  ctx.Next(nil)
}

func WebIndex() (string, interface{}) {
  return "index.html", map[string]interface{} {}
}

func WebDashboard(ctx *soggy.Context) (string, interface{}) {
  return "dashboard.html", map[string]interface{} {}
}

func WebMe(ctx *soggy.Context) (int, interface{}) {
  return http.StatusOK, map[string]interface{} { "googleUser": ctx.Env["googleUser"], "user": ctx.Env["user"] }
}

func WebLogout(ctx *soggy.Context) {
  url, _ := user.LogoutURL(ctx.Env["aeCtx"].(appengine.Context), "/")
  http.Redirect(ctx.Res, ctx.Req.Request, url, 302)
}
