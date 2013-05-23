package biboop

import (
  "appengine"
  "appengine/datastore"
  "appengine/memcache"
  "github.com/dbrain/soggy"
  "time"
  "strconv"
  "errors"
  "log"
)

var DatastoreKindUser = "User"
var DatastoreKindServer = "Server"
var DatastoreKindCommand = "Command"

var ErrUserNotFound = errors.New("User not found")
var ErrServerNotFound = errors.New("Server not found")

type User struct {
  ID int64 `json:"id,omitempty" datastore:"-"`
  Email string `json:"email,omitempty"`
  ServerAPIKey string `json:"serverAPIKey,omitempty"`
}

type Server struct {
  ID int64 `json:"id,omitempty" datastore:"-"`
  UserID int64 `json:"userId,omitempty"`
  ServerID string `json:"serverId,omitempty"`
  Name string `json:"name,omitempty"`
  Description string `json:"description,omitempty"`
  LastPollTime int64 `json:"lastPollTime,omitempty"`
  PendingCommands int `json:"pendingCommands,omitempty"`
  AvailableCommands []*datastore.Key `json:"AvailableCommands,omitempty"`
}

type CommandParam struct {
  Name string `json:"name,omitempty"`
  Description string `json:"description,omitempty"`
  DefaultValue string `json:"defaultValue,omitempty"`
}

type Command struct {
  ID int64 `json:"id,omitempty" datastore:"-"`
  UserID int64 `json:"userId,omitempty"`
  Private bool `json:"private,omitempty"`
  Name string `json:"name,omitempty"`
  Description string `json:"description,omitempty"`
  Command string `json:"command,omitempty"`
  Params []CommandParam `json:"command,omitempty"`
}

func GetOrCreateUser(ctx appengine.Context, email string) (User, error) {
  var user User
  var userKey *datastore.Key

  cacheKey := "User-" + email
  if item, err := memcache.Gob.Get(ctx, cacheKey, &user); err == memcache.ErrCacheMiss {
    if userKey, user, err = getOrCreateUserNoCache(ctx, email); err != nil {
      return user, err
    } else {
      user.ID = userKey.IntID()
      item = &memcache.Item{
        Key: cacheKey,
        Object: user,
      }
      memcache.Gob.Set(ctx, item)
    }
  } else if err != nil {
    return user, err
  }
  return user, nil
}

func getOrCreateUserNoCache(ctx appengine.Context, email string) (*datastore.Key, User, error) {
  var user User
  var userKey *datastore.Key
  var err error

  query := datastore.NewQuery(DatastoreKindUser).
         Filter("Email =", email).
         Limit(1)

  for cursor := query.Run(ctx); ; {
    if cursorKey, err := cursor.Next(&user); err == datastore.Done {
      break
    } else if err != nil {
      return userKey, user, err
    } else {
      userKey = cursorKey
    }
  }

  if userKey == nil {
    userKey = datastore.NewIncompleteKey(ctx, DatastoreKindUser, nil)
    user.Email = email
    user.ServerAPIKey = email + "-" + strconv.FormatInt(time.Now().Unix(), 10) + "-" + soggy.UIDString()
    if userKey, err = datastore.Put(ctx, userKey, &user); err != nil {
      return userKey, user, err
    }
  }

  return userKey, user, err
}

func GetServerForPollRequest(ctx appengine.Context, user User, pollRequest PollRequest) (Server, error) {
  var server Server
  var serverKey *datastore.Key

  cacheKey := "Server-" + strconv.FormatInt(user.ID, 10) + "-" + pollRequest.ServerID
  if item, err := memcache.Gob.Get(ctx, cacheKey, &server); err == memcache.ErrCacheMiss {
    if serverKey, server, err = getServerForPollRequestNoCache(ctx, user, pollRequest); err != nil {
      return server, err
    } else {
      server.ID = serverKey.IntID()
      item = &memcache.Item{
        Key: cacheKey,
        Object: server,
      }
      memcache.Gob.Set(ctx, item)
    }
  } else if err != nil {
    return server, err
  }

  return server, nil
}

func getServerForPollRequestNoCache(ctx appengine.Context, user User, pollRequest PollRequest) (*datastore.Key, Server, error) {
  var server Server
  var serverKey *datastore.Key

  query := datastore.NewQuery(DatastoreKindServer).
           Filter("UserID =", user.ID).
           Filter("ServerID =", pollRequest.ServerID).
           Limit(1)

  for cursor := query.Run(ctx); ; {
    if cursorKey, err := cursor.Next(&server); err == datastore.Done {
      break
    } else if err != nil {
      return serverKey, server, err
    } else {
      serverKey = cursorKey
    }
  }

  if server.ServerID == "" {
    return serverKey, server, ErrServerNotFound
  }

  return serverKey, server, nil
}

func FindUserByServerAPIKey(ctx appengine.Context, serverAPIKey string) (User, error) {
  var user User
  var userKey *datastore.Key

  cacheKey := "User-" + serverAPIKey
  if item, err := memcache.Gob.Get(ctx, cacheKey, &user); err == memcache.ErrCacheMiss {
    if userKey, user, err = findUserByServerAPIKeyNoCache(ctx, serverAPIKey); err != nil {
      return user, err
    } else {
      user.ID = userKey.IntID()
      item = &memcache.Item{
        Key: cacheKey,
        Object: user,
      }
      memcache.Gob.Set(ctx, item)
    }
  } else if err != nil {
    return user, err
  }
  return user, nil
}

func findUserByServerAPIKeyNoCache(ctx appengine.Context, serverAPIKey string) (*datastore.Key, User, error) {
  var user User
  var userKey *datastore.Key

  query := datastore.NewQuery(DatastoreKindUser).
           Filter("ServerAPIKey =", serverAPIKey).
           Limit(1)

  for cursor := query.Run(ctx); ; {
    if cursorKey, err := cursor.Next(&user); err == datastore.Done {
      break
    } else if err != nil {
      return cursorKey, user, err
    } else {
      userKey = cursorKey
    }
  }

  if userKey == nil {
    return userKey, user, ErrUserNotFound
  }

  return userKey, user, nil
}

func UpdateServerForUpdateRequest(ctx appengine.Context, user User, updateRequest UpdateRequest) (Server, error) {
  var server Server
  var serverKey *datastore.Key

  query := datastore.NewQuery(DatastoreKindServer).
           Filter("UserID =", user.ID).
           Filter("ServerID =", updateRequest.ServerID).
           Limit(1)

  for cursor := query.Run(ctx); ; {
    if cursorKey, err := cursor.Next(&server); err == datastore.Done {
      break
    } else if err != nil {
      return server, err
    } else {
      serverKey = cursorKey
    }
  }

  if serverKey == nil {
    log.Println("Creating server")
    serverKey = datastore.NewIncompleteKey(ctx, DatastoreKindServer, nil)
    server.UserID = user.ID
    server.ServerID = updateRequest.ServerID
    server.Name = updateRequest.Name
    server.Description = updateRequest.Description
    server.LastPollTime = time.Now().UTC().Unix()
    server.PendingCommands = 0
  } else {
    server.LastPollTime = time.Now().UTC().Unix()
  }

  if _, err := datastore.Put(ctx, serverKey, &server); err != nil {
    return server, err
  }

  return server, nil
}

func GetServersNoCache(ctx appengine.Context, user User) ([]Server, error) {
  var servers []Server

  query := datastore.NewQuery(DatastoreKindServer).
    Filter("UserID =", user.ID)

  if keys, err := query.GetAll(ctx, &servers); err != nil {
    return servers, err
  } else {
    for i, key := range keys {
      servers[i].ID = key.IntID()
    }
  }

  return servers, nil
}

func CreateCommandNoCache(ctx appengine.Context, user User, commandRequest CreateCommandRequest) (Command, error) {
  var command Command
  commandKey := datastore.NewIncompleteKey(ctx, DatastoreKindCommand, nil)
  command.UserID = user.ID
  command.Name = commandRequest.Name
  command.Description = commandRequest.Description
  command.Command = commandRequest.Command
  command.Params = commandRequest.Params

  if _, err := datastore.Put(ctx, commandKey, &command); err != nil {
    return command, err
  }
  return command, nil
}

func GetCommandsNoCache(ctx appengine.Context, user User) ([]Command, error) {
  var commands []Command

  query := datastore.NewQuery(DatastoreKindCommand).
    Filter("UserID =", user.ID)

  for cursor := query.Run(ctx); ; {
    var command Command
    if _, err := cursor.Next(&command); err == datastore.Done {
      break
    } else if err != nil {
      return commands, err
    } else {
      commands = append(commands, command)
    }
  }

  return commands, nil
}
