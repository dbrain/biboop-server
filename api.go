package biboop

import (
  "github.com/dbrain/soggy"
  "net/http"
)

func ApiUserRequired(ctx *soggy.Context) (int, interface{}) {
  if ctx.Env["googleUser"] == nil {
    return http.StatusUnauthorized, map[string]interface{} { "error": "This function requires authorization" }
  }
  ctx.Next(nil)
  return 0, nil
}

func ApiMe(ctx* soggy.Context) (int, interface{}) {
  return http.StatusOK, map[string]interface{} { "googleUser": ctx.Env["googleUser"], "user": ctx.Env["user"] }
}

func ApiServerPoll(ctx *soggy.Context) (int, interface{}) {
  bodyType, body, err := ctx.Req.GetBody()
  if err != nil {
    ctx.Next(err)
    return 0, nil
  }

  if bodyType != soggy.BodyTypeJson {
    return http.StatusBadRequest, map[string]interface{} { "error": "JSON request expected" }
  }
  return http.StatusOK, map[string]interface{} { "body": body }
}
