package biboop

import (
  "appengine"
  "appengine/datastore"
  "github.com/dbrain/soggy"
  "time"
  "strconv"
)

type User struct {
  Email string `json:"email,omitempty"`
  ServerKey string `json:"serverKey,omitempty"`
}

type Server struct {
  UserKey *datastore.Key `json:"userKey,omitempty"`
  ServerID string `json:"serverId,omitempty"`
}

func GetOrCreateUser(ctx appengine.Context, email string) (User, error) {
  key := datastore.NewKey(ctx, "User", email, 0, nil)
  biboopUser := User{}
  if err := datastore.Get(ctx, key, &biboopUser); err != nil {
    if err == datastore.ErrNoSuchEntity {
      biboopUser.Email = email
      biboopUser.ServerKey = email + "-" + strconv.FormatInt(time.Now().Unix(), 10) + "-" + soggy.UIDString()
      if _, err := datastore.Put(ctx, key, &biboopUser); err != nil {
        return biboopUser, err
      }
    } else {
      return biboopUser, err
    }
  }
  return biboopUser, nil
}

func GetOrCreateServer(ctx appengine.Context, serverKey string, serverId string) (Server, error) {
  server := Server{}
  return server, nil
}