package biboop

import (
  "appengine"
  "appengine/datastore"
  "github.com/dbrain/soggy"
  "time"
  "strconv"
  "errors"
)

var DatastoreKindUser = "User"
var DatastoreKindServer = "Server"

var ErrUserNotFound = errors.New("User not found")

type User struct {
  Email string `json:"email,omitempty"`
  ServerKey string `json:"serverKey,omitempty"`
}

type Server struct {
  ServerAPIKey string `json:"serverAPIKey,omitempty"`
  ServerID string `json:"serverId,omitempty"`
  Name string `json:"name,omitempty"`
  Description string `json:"description,omitempty"`
}

func GetOrCreateUser(ctx appengine.Context, email string) (User, error) {
  key := datastore.NewKey(ctx, DatastoreKindUser, email, 0, nil)
  var biboopUser User
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

func GetOrCreateServerForPollRequest(ctx appengine.Context, pollRequest PollRequest) (Server, error) {
  return GetOrCreateServer(ctx, pollRequest.ServerAPIKey, pollRequest.ServerID, pollRequest.Name, pollRequest.Description)
}

func GetOrCreateServer(ctx appengine.Context, serverApiKey string, serverId string, name string, description string) (Server, error) {
  serverKey := datastore.NewKey(ctx, DatastoreKindServer, serverApiKey + "-" + serverId, 0, nil)
  var server Server
  if err := datastore.Get(ctx, serverKey, &server); err != nil {
    if err == datastore.ErrNoSuchEntity {
      server.ServerAPIKey = serverApiKey
      server.ServerID = serverId
      server.Name = name
      server.Description = description
      if _, err := datastore.Put(ctx, serverKey, &server); err != nil {
        return server, err
      }
    } else {
      return server, err
    }
  }
  return server, nil

}

func FindUserByServerKey(ctx appengine.Context, serverKey string) (*datastore.Key, User, error) {
  var user User
  var userKey *datastore.Key

  query := datastore.NewQuery(DatastoreKindUser).
           Filter("ServerKey =", serverKey).
           Limit(1)

  for cursor := query.Run(ctx); ; {
    var err error
    userKey, err = cursor.Next(&user)
    if err == datastore.Done {
      break
    }
    if err != nil {
      return userKey, user, err
    }
  }

  if user.Email == "" {
    return userKey, user, ErrUserNotFound
  }

  return userKey, user, nil
}
