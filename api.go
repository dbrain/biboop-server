package biboop

import (
  "appengine"
  "github.com/dbrain/soggy"
  "net/http"
  "errors"
)

type PollRequest struct {
  Name string `json:"name,omitempty"`
  Description string `json:"description,omitempty"`
  MinimumPollTimeSec int `json:"minimumPollTimeSec,omitempty"`
  ServerKey string `json:"serverKey,omitempty"`
  ServerID string `json:"serverId,omitempty"`
}

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
  var pollRequest PollRequest
  bodyType, _, err := ctx.Req.GetBody(&pollRequest)
  if err != nil {
    ctx.Next(err)
    return 0, nil
  }

  if bodyType != soggy.BodyTypeJson {
    return http.StatusBadRequest, map[string]interface{} { "error": "JSON request expected" }
  }

  if pollRequest.ServerID == "" || pollRequest.ServerKey == "" {
    ctx.Next(errors.New("serverId and serverKey are required fields"))
    return 0, nil
  }

  aeCtx := ctx.Env["aeCtx"].(appengine.Context)
  server, err := GetOrCreateServerByServerKey(aeCtx, pollRequest.ServerKey, pollRequest.ServerID, pollRequest.Name, pollRequest.Description)
  if err != nil {
    ctx.Next(err)
    return 0, nil
  }

  return http.StatusOK, map[string]interface{} { "server": server }
}
