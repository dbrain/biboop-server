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
  ServerAPIKey string `json:"serverApiKey,omitempty"`
  ServerID string `json:"serverId,omitempty"`
}

type UpdateRequest struct {
  Name string `json:"name,omitempty"`
  Description string `json:"description,omitempty"`
  MinimumPollTimeSec int `json:"minimumPollTimeSec,omitempty"`
  ServerAPIKey string `json:"serverApiKey,omitempty"`
  ServerID string `json:"serverId,omitempty"`
}

type CreateCommandRequest struct {
  PublicCommand bool `json:"publicCommand,omitempty"`
  Name string `json:"name,omitempty"`
  Description string `json:"description,omitempty"`
  Command string `json:"command,omitempty"`
  Params []CommandParam `json:"params,omitempty"`
  Servers []int64 `json:"servers,omitempty"`
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

  if bodyType, _, err := ctx.Req.GetBody(&pollRequest); err != nil {
    ctx.Next(err)
    return 0, nil
  } else if bodyType != soggy.BodyTypeJson {
    return http.StatusBadRequest, map[string]interface{} { "error": "JSON request expected" }
  } else if pollRequest.ServerID == "" || pollRequest.ServerAPIKey == "" {
    ctx.Next(errors.New("serverId and serverAPIKey are required fields"))
    return 0, nil
  }

  aeCtx := ctx.Env["aeCtx"].(appengine.Context)
  user, err := FindUserByServerAPIKey(aeCtx, pollRequest.ServerAPIKey)
  if err != nil {
    ctx.Next(err)
    return 0, nil
  }

  server, err := GetServerForPollRequest(aeCtx, user, pollRequest)
  if err != nil {
    ctx.Next(err)
    return 0, nil
  }

  return http.StatusOK, map[string]interface{} { "server": server }
}

func ApiServerUpdate(ctx *soggy.Context) (int, interface{}) {
  var updateRequest UpdateRequest

  if bodyType, _, err := ctx.Req.GetBody(&updateRequest); err != nil {
    ctx.Next(err)
    return 0, nil
  } else if bodyType != soggy.BodyTypeJson {
    return http.StatusBadRequest, map[string]interface{} { "error": "JSON request expected" }
  } else if updateRequest.ServerID == "" || updateRequest.ServerAPIKey == "" {
    ctx.Next(errors.New("serverId and serverAPIKey are required fields"))
    return 0, nil
  }

  aeCtx := ctx.Env["aeCtx"].(appengine.Context)
  user, err := FindUserByServerAPIKey(aeCtx, updateRequest.ServerAPIKey)
  if err != nil {
    ctx.Next(err)
    return 0, nil
  }

    server, err := UpdateServerForUpdateRequest(aeCtx, user, updateRequest);
    if err != nil {
      ctx.Next(err)
      return 0, nil
    }

  return http.StatusOK, map[string]interface{} { "server": server }
}

func ApiGetServers(ctx *soggy.Context) (int, interface{}) {
  aeCtx := ctx.Env["aeCtx"].(appengine.Context)
  servers, err := GetServersNoCache(aeCtx, ctx.Env["user"].(User))
  if err != nil {
    ctx.Next(err)
    return 0, nil
  }

  return http.StatusOK, map[string]interface{} { "servers": servers }
}

func ApiCreateCommand(ctx *soggy.Context) (int, interface{}) {
  var createCommandRequest CreateCommandRequest

  if bodyType, _, err := ctx.Req.GetBody(&createCommandRequest); err != nil {
    ctx.Next(err)
    return 0, nil
  } else if bodyType != soggy.BodyTypeJson {
    return http.StatusBadRequest, map[string]interface{} { "error": "JSON request expected" }
  } else if createCommandRequest.Name == "" || createCommandRequest.Command == "" {
    ctx.Next(errors.New("name and command are required fields"))
    return 0, nil
  }

  aeCtx := ctx.Env["aeCtx"].(appengine.Context)
  command, err := CreateCommandNoCache(aeCtx, ctx.Env["user"].(User), createCommandRequest)
  if err != nil {
    ctx.Next(err)
    return 0, nil
  }

  return http.StatusCreated, map[string]interface{} { "command": command }
}

func ApiGetCommands(ctx *soggy.Context) (int, interface{}) {
  aeCtx := ctx.Env["aeCtx"].(appengine.Context)
  commands, err := GetCommandsNoCache(aeCtx, ctx.Env["user"].(User))
  if err != nil {
    ctx.Next(err)
    return 0, nil
  }

  return http.StatusOK, map[string]interface{} { "commands": commands }
}
