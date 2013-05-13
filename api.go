package biboop

import (
  "appengine"
  "github.com/dbrain/soggy"
  "net/http"
  "errors"
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

  var bodyMap = body.(map[string]interface{})

  if bodyMap["serverId"] == nil || bodyMap["serverKey"] == nil {
    ctx.Next(errors.New("serverId and serverKey are required fields"))
    return 0, nil
  }

  server, err := GetOrCreateServerByServerKey(ctx.Env["aeCtx"].(appengine.Context), bodyMap["serverKey"].(string), bodyMap["serverId"].(string))
  if err != nil {
    ctx.Next(err)
    return 0, nil
  }

  return http.StatusOK, map[string]interface{} { "server": server }
}
