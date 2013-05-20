package biboop

import (
  "appengine"
  "appengine/datastore"
  "appengine/memcache"
  "github.com/dbrain/soggy"
  "time"
  "strconv"
  "errors"
)

var DatastoreKindUser = "User"
var DatastoreKindServer = "Server"

var ErrUserNotFound = errors.New("User not found")
var ErrServerNotFound = errors.New("Server not found")

type CacheUser struct {
  Key *datastore.Key `json:"key,omitempty"`
  Email string `json:"email,omitempty"`
  ServerAPIKey string `json:"serverAPIKey,omitempty"`
}

type User struct {
  Email string `json:"email,omitempty"`
  ServerAPIKey string `json:"serverAPIKey,omitempty"`
}

type Server struct {
  UserKey *datastore.Key `json:"userKey,omitempty"`
  ServerID string `json:"serverId,omitempty"`
  Name string `json:"name,omitempty"`
  Description string `json:"description,omitempty"`
  LastPollTime int64 `json:"lastPollTime,omitempty"`
  PendingCommands int `json:"pendingCommands,omitempty"`
}

func GetOrCreateUser(ctx appengine.Context, email string) (CacheUser, error) {
  var biboopUser CacheUser
  var userKey *datastore.Key

  cacheKey := "User-" + email
  if item, err := memcache.Gob.Get(ctx, cacheKey, &biboopUser); err == memcache.ErrCacheMiss {
    var dbUser User
    if userKey, dbUser, err = getOrCreateUserWithoutCache(ctx, email); err != nil {
      return biboopUser, err
    } else {
      biboopUser.Key = userKey
      biboopUser.Email = dbUser.Email
      biboopUser.ServerAPIKey = dbUser.ServerAPIKey
      item = &memcache.Item{
        Key: cacheKey,
        Object: biboopUser,
      }
      memcache.Gob.Set(ctx, item)
    }
  } else if err != nil {
    return biboopUser, err
  }
  return biboopUser, nil
}

func getOrCreateUserWithoutCache(ctx appengine.Context, email string) (*datastore.Key, User, error) {
  var biboopUser User
  var userKey *datastore.Key
  var err error

  query := datastore.NewQuery(DatastoreKindUser).
         Filter("Email =", email).
         Limit(1)

  for cursor := query.Run(ctx); ; {
    if cursorKey, err := cursor.Next(&biboopUser); err == datastore.Done {
      break
    } else if err != nil {
      return userKey, biboopUser, err
    } else {
      userKey = cursorKey
    }
  }

  if userKey == nil {
    userKey = datastore.NewIncompleteKey(ctx, DatastoreKindUser, nil)
    biboopUser.Email = email
    biboopUser.ServerAPIKey = email + "-" + strconv.FormatInt(time.Now().Unix(), 10) + "-" + soggy.UIDString()
    if userKey, err = datastore.Put(ctx, userKey, &biboopUser); err != nil {
      return userKey, biboopUser, err
    }
  }

  return userKey, biboopUser, err
}

func GetServerForPollRequest(ctx appengine.Context, user CacheUser, pollRequest PollRequest) (Server, error) {
  var server Server
  cacheKey := "Server-" + user.Key.Encode() + "-" + pollRequest.ServerID
  if item, err := memcache.Gob.Get(ctx, cacheKey, &server); err == memcache.ErrCacheMiss {
    if server, err = getServerForPollRequestWithoutCache(ctx, user, pollRequest); err != nil {
      return server, err
    } else {
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

func getServerForPollRequestWithoutCache(ctx appengine.Context, user CacheUser, pollRequest PollRequest) (Server, error) {
  var server Server

  query := datastore.NewQuery(DatastoreKindServer).
           Filter("UserKey =", user.Key).
           Filter("ServerID =", pollRequest.ServerID).
           Limit(1)

  for cursor := query.Run(ctx); ; {
    if _, err := cursor.Next(&server); err == datastore.Done {
      break
    } else if err != nil {
      return server, err
    }
  }

  if server.ServerID == "" {
    return server, ErrServerNotFound
  }

  return server, nil
}

func FindUserByServerAPIKey(ctx appengine.Context, serverAPIKey string) (CacheUser, error) {
  var biboopUser CacheUser
  var userKey *datastore.Key

  cacheKey := "User-" + serverAPIKey
  if item, err := memcache.Gob.Get(ctx, cacheKey, &biboopUser); err == memcache.ErrCacheMiss {
    var dbUser User
    if userKey, dbUser, err = findUserByServerAPIKeyWithoutCache(ctx, serverAPIKey); err != nil {
      return biboopUser, err
    } else {
      biboopUser.Key = userKey
      biboopUser.Email = dbUser.Email
      biboopUser.ServerAPIKey = dbUser.ServerAPIKey
      item = &memcache.Item{
        Key: cacheKey,
        Object: biboopUser,
      }
      memcache.Gob.Set(ctx, item)
    }
  } else if err != nil {
    return biboopUser, err
  }
  return biboopUser, nil
}

func findUserByServerAPIKeyWithoutCache(ctx appengine.Context, serverAPIKey string) (*datastore.Key, User, error) {
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
