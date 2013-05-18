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

type User struct {
  Email string `json:"email,omitempty"`
  ServerKey string `json:"serverKey,omitempty"`
}

type Server struct {
  ServerAPIKey string `json:"serverAPIKey,omitempty"`
  ServerID string `json:"serverId,omitempty"`
  Name string `json:"name,omitempty"`
  Description string `json:"description,omitempty"`
  LastPollTime int64 `json:"lastPollTime,omitempty"`
  PendingCommands int `json:"pendingCommands,omitempty"`
}

func GetOrCreateUser(ctx appengine.Context, email string) (User, error) {
  var biboopUser User
  cacheKey := "User-" + email
  if item, err := memcache.Gob.Get(ctx, cacheKey, &biboopUser); err == memcache.ErrCacheMiss {
    if biboopUser, err = getOrCreateUserWithoutCache(ctx, email); err != nil {
      return biboopUser, err
    } else {
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

func getOrCreateUserWithoutCache(ctx appengine.Context, email string) (User, error) {
  var biboopUser User
  key := datastore.NewKey(ctx, DatastoreKindUser, email, 0, nil)

  err := datastore.RunInTransaction(ctx, func (ctx appengine.Context) error {
    var err error
    if err = datastore.Get(ctx, key, &biboopUser); err != nil {
      if err == datastore.ErrNoSuchEntity {
        biboopUser.Email = email
        biboopUser.ServerKey = email + "-" + strconv.FormatInt(time.Now().Unix(), 10) + "-" + soggy.UIDString()
        if _, err := datastore.Put(ctx, key, &biboopUser); err != nil {
          return err
        }
      }
    }
    return err
  }, nil)

  return biboopUser, err
}

func GetServerForPollRequest(ctx appengine.Context, pollRequest PollRequest) (Server, error) {
  var server Server
  cacheKey := "Server-" + pollRequest.ServerAPIKey + "-" + pollRequest.ServerID
  if item, err := memcache.Gob.Get(ctx, cacheKey, &server); err == memcache.ErrCacheMiss {
    if server, err = getServerForPollRequestWithoutCache(ctx, pollRequest); err != nil {
      return server, err
    } else {
      item = &memcache.Item{
        Key: cacheKey,
        Object: server,
      }
      memcache.Gob.Set(ctx, item)
    }
  } else if err != nil {
    return server, nil
  }

  return server, nil
}

func getServerForPollRequestWithoutCache(ctx appengine.Context, pollRequest PollRequest) (Server, error) {
  var server Server
  serverKey := datastore.NewKey(ctx, DatastoreKindServer, pollRequest.ServerAPIKey + "-" + pollRequest.ServerID, 0, nil)
  if err := datastore.Get(ctx, serverKey, &server); err != nil {
    return server, err
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
