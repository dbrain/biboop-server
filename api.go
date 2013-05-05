package biboop

import (
  "github.com/dbrain/soggy"
  "net/http"
)

func ApiMe(ctx* soggy.Context) (int, interface{}) {
  return http.StatusOK, ctx.Env["user"]
}