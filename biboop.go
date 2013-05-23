package biboop

import (
  "appengine"
  "appengine/user"
  "appengine/urlfetch"
  "github.com/dbrain/soggy"
  "strings"
  "net/http"
  "encoding/json"
  "io/ioutil"
  "strconv"
)

type AppEngineWebMiddleware struct {}
func (middleware *AppEngineWebMiddleware) Execute(ctx *soggy.Context) {
  aeCtx := appengine.NewContext(ctx.Req.Request)
  ctx.Env["aeCtx"] = aeCtx

  googleUser := user.Current(aeCtx)
  if googleUser != nil {
    ctx.Env["googleUser"] = googleUser
    biboopUser, err := GetOrCreateUser(aeCtx, googleUser.Email)
    if err != nil {
      ctx.Next(err)
      return
    }
    ctx.Env["user"] = biboopUser
  }

  ctx.Next(nil)
}

type AppEngineApiMiddleware struct{}
func (middleware *AppEngineApiMiddleware) Execute(ctx *soggy.Context) {
  aeCtx := appengine.NewContext(ctx.Req.Request)
  urlfetchClient := urlfetch.Client(aeCtx)
  ctx.Env["aeCtx"] = aeCtx
  ctx.Env["urlfetchClient"] = urlfetchClient

  authHeader := ctx.Req.Request.Header.Get("Authorization");
  if strings.HasPrefix(authHeader, "Bearer ") {
    googleUser := loadUserDetails(ctx, authHeader, urlfetchClient)
    if googleUser != nil {
      ctx.Env["googleUser"] = googleUser
      biboopUser, err := GetOrCreateUser(aeCtx, googleUser["email"].(string))
      if err != nil {
        ctx.Next(err)
        return
      }

      ctx.Env["user"] = biboopUser
      ctx.Next(nil)
    }
  } else {
    ctx.Next(nil)
  }
}

func loadUserDetails(ctx *soggy.Context, authHeader string, urlfetchClient *http.Client) map[string]interface{} {
  req, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v3/userinfo?alt=json", nil)
  req.Header.Add("Authorization", authHeader)
  resp, err := urlfetchClient.Do(req)
  if err != nil || resp.StatusCode != http.StatusOK {
      sendAuthFailure(ctx.Res, http.StatusUnauthorized, "authorization_failed")
  } else {
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
      sendAuthFailure(ctx.Res, http.StatusUnauthorized, "authorization_parse_failed")
    } else {
      var googleUser map[string]interface{}
      err := json.Unmarshal(body, &googleUser)
      if err != nil {
        sendAuthFailure(ctx.Res, http.StatusUnauthorized, "authorization_body_parse_failed")
      } else {
        return googleUser
      }
    }
  }
  return nil
}

func sendAuthFailure(res *soggy.Response, status int, reason string) {
  res.Set("Content-Type", "application/json; charset=utf-8")
  jsonOut, err := json.Marshal(map[string]interface{} { "error": reason })
  if err == nil {
    res.Set("Content-Length", strconv.Itoa(len(jsonOut)))
    res.WriteHeader(status)
    _, err = res.Write(jsonOut)
  }
}

func startWebServer() *soggy.Server {
  webServer := soggy.NewServer("/")

  webServer.Get("/", WebIndex)
  webServer.Get("/dashboard", WebUserRequired, WebDashboard)
  webServer.Get("/me", WebUserRequired, WebMe)
  webServer.Get("/logout", WebLogout)

  webServer.All(soggy.ANY_PATH, func (context *soggy.Context) (int, interface{}) {
    return 404, map[string]interface{} { "error": "Path not found" }
  })

  webServer.Use(&AppEngineWebMiddleware{}, webServer.Router)
  return webServer
}

func startApiServer() *soggy.Server {
  apiServer := soggy.NewServer("/api")
  apiServer.Get("/me", ApiUserRequired, ApiMe)
  apiServer.Post("/server/poll", ApiServerPoll)
  apiServer.Post("/server/update", ApiServerUpdate)
  apiServer.Get("/servers", ApiGetServers)
  apiServer.Post("/commands", ApiCreateCommand)

  apiServer.All(soggy.ANY_PATH, func (context *soggy.Context) (int, interface{}) {
    return 404, map[string]interface{} { "error": "Path not found" }
  })

  apiServer.Use(&AppEngineApiMiddleware{}, apiServer.Router)
  return apiServer
}

func startServer() {
  app := soggy.NewApp()
  app.AddServers(startWebServer())
  app.AddServers(startApiServer())
  app.BindHandlers()
}

func init() {
  startServer()
}
