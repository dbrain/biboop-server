package biboop

import (
  "github.com/dbrain/soggy"
  "net/http"
)

func ApiMe(ctx* soggy.Context) (int, interface{}) {
  return http.StatusOK, map[string]interface{} { "googleUser": ctx.Env["googleUser"], "user": ctx.Env["user"] }
}