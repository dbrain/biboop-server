package biboop

import (
  "appengine"
  "appengine/user"
  "github.com/dbrain/soggy"
)

type AppEngineMiddleware struct {}
func (middleware *AppEngineMiddleware) Execute(ctx *soggy.Context) {
  aeCtx := appengine.NewContext(ctx.Req.Request)
  ctx.Env["aeCtx"] = aeCtx

  currentUser := user.Current(aeCtx)
  if currentUser != nil {
    ctx.Env["currentUser"] = currentUser
  }

  ctx.Next(nil)
}

func startWebServer() *soggy.Server {
  webServer := soggy.NewServer("/")
  
  webServer.Get("/", WebIndex)
  webServer.Get("/dashboard", WebUserRequired, WebDashboard)
  webServer.Get("/logout", WebLogout)
  
  webServer.All(soggy.ANY_PATH, func (context *soggy.Context) (int, interface{}) {
    return 404, map[string]interface{} { "error": "Path not found" }
  })
  
  webServer.Use(soggy.NewStaticServerMiddleware("/public"), &AppEngineMiddleware{}, webServer.Router)
  return webServer
}

func startApiServer() *soggy.Server {
  apiServer := soggy.NewServer("/api")

  return apiServer
}

func startServer() {
  app := soggy.NewApp()
  app.AddServers(startApiServer())
  app.AddServers(startWebServer())
  app.BindHandlers()
}

func init() {
  startServer()
}
