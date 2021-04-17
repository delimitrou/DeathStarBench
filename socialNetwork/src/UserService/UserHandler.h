#ifndef SOCIAL_NETWORK_MICROSERVICES_USERHANDLER_H
#define SOCIAL_NETWORK_MICROSERVICES_USERHANDLER_H

#include <bson/bson.h>
#include <libmemcached/memcached.h>
#include <libmemcached/util.h>
#include <mongoc.h>

#include <iomanip>
#include <iostream>
#include <jwt/jwt.hpp>
#include <nlohmann/json.hpp>
#include <random>
#include <string>

#include "../../gen-cpp/SocialGraphService.h"
#include "../../gen-cpp/UserService.h"
#include "../../gen-cpp/social_network_types.h"
#include "../../third_party/PicoSHA2/picosha2.h"
#include "../ClientPool.h"
#include "../ThriftClient.h"
#include "../logger.h"
#include "../tracing.h"

// Custom Epoch (January 1, 2018 Midnight GMT = 2018-01-01T00:00:00Z)
#define CUSTOM_EPOCH 1514764800000
#define MONGODB_TIMEOUT_MS 100

namespace social_network {

using std::chrono::duration_cast;
using std::chrono::milliseconds;
using std::chrono::seconds;
using std::chrono::system_clock;
using namespace jwt::params;

static int64_t current_timestamp = -1;
static int counter = 0;

static int GetCounter(int64_t timestamp) {
  if (current_timestamp > timestamp) {
    LOG(fatal) << "Timestamps are not incremental.";
    exit(EXIT_FAILURE);
  }
  if (current_timestamp == timestamp) {
    return counter++;
  } else {
    current_timestamp = timestamp;
    counter = 0;
    return counter++;
  }
}

std::string GenRandomString(const int len) {
  static const std::string alphanum =
      "0123456789"
      "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
      "abcdefghijklmnopqrstuvwxyz";
  std::random_device rd;
  std::mt19937 gen(rd());
  std::uniform_int_distribution<int> dist(
      0, static_cast<int>(alphanum.length() - 1));
  std::string s;
  for (int i = 0; i < len; ++i) {
    s += alphanum[dist(gen)];
  }
  return s;
}

class UserHandler : public UserServiceIf {
 public:
  UserHandler(std::mutex *, const std::string &, const std::string &,
              memcached_pool_st *, mongoc_client_pool_t *,
              ClientPool<ThriftClient<SocialGraphServiceClient>> *);
  ~UserHandler() override = default;
  void RegisterUser(int64_t, const std::string &, const std::string &,
                    const std::string &, const std::string &,
                    const std::map<std::string, std::string> &) override;
  void RegisterUserWithId(int64_t, const std::string &, const std::string &,
                          const std::string &, const std::string &, int64_t,
                          const std::map<std::string, std::string> &) override;

  void ComposeCreatorWithUserId(
      Creator &, int64_t, int64_t, const std::string &,
      const std::map<std::string, std::string> &) override;
  void ComposeCreatorWithUsername(
      Creator &, int64_t, const std::string &,
      const std::map<std::string, std::string> &) override;
  void Login(std::string &, int64_t, const std::string &, const std::string &,
             const std::map<std::string, std::string> &) override;
  int64_t GetUserId(int64_t, const std::string &,
                    const std::map<std::string, std::string> &) override;

 private:
  std::string _machine_id;
  std::string _secret;
  std::mutex *_thread_lock;
  memcached_pool_st *_memcached_client_pool;
  mongoc_client_pool_t *_mongodb_client_pool;
  ClientPool<ThriftClient<SocialGraphServiceClient>> *_social_graph_client_pool;
};

UserHandler::UserHandler(std::mutex *thread_lock, const std::string &machine_id,
                         const std::string &secret,
                         memcached_pool_st *memcached_client_pool,
                         mongoc_client_pool_t *mongodb_client_pool,
                         ClientPool<ThriftClient<SocialGraphServiceClient>>
                             *social_graph_client_pool) {
  _thread_lock = thread_lock;
  _machine_id = machine_id;
  _memcached_client_pool = memcached_client_pool;
  _mongodb_client_pool = mongodb_client_pool;
  _secret = secret;
  _social_graph_client_pool = social_graph_client_pool;
}

void UserHandler::RegisterUserWithId(
    const int64_t req_id, const std::string &first_name,
    const std::string &last_name, const std::string &username,
    const std::string &password, const int64_t user_id,
    const std::map<std::string, std::string> &carrier) {
  // Initialize a span
  TextMapReader reader(carrier);
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "register_user_withid_server",
      {opentracing::ChildOf(parent_span->get())});
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  // Store user info into mongodb
  mongoc_client_t *mongodb_client =
      mongoc_client_pool_pop(_mongodb_client_pool);
  if (!mongodb_client) {
    ServiceException se;
    se.errorCode = ErrorCode::SE_MONGODB_ERROR;
    se.message = "Failed to pop a client from MongoDB pool";
    throw se;
  }
  auto collection =
      mongoc_client_get_collection(mongodb_client, "user", "user");

  // Check if the username has existed in the database
  bson_t *query = bson_new();
  BSON_APPEND_UTF8(query, "username", username.c_str());
  mongoc_cursor_t *cursor =
      mongoc_collection_find_with_opts(collection, query, nullptr, nullptr);
  const bson_t *doc;
  bson_error_t error;
  bool found = mongoc_cursor_next(cursor, &doc);
  if (mongoc_cursor_error(cursor, &error)) {
    LOG(warning) << error.message;
    bson_destroy(query);
    mongoc_cursor_destroy(cursor);
    mongoc_collection_destroy(collection);
    mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
    ServiceException se;
    se.errorCode = ErrorCode::SE_MONGODB_ERROR;
    se.message = error.message;
    throw se;
  } else if (found) {
    LOG(warning) << "User " << username << " already existed.";
    ServiceException se;
    se.errorCode = ErrorCode::SE_THRIFT_HANDLER_ERROR;
    se.message = "User " + username + " already existed";
    bson_destroy(query);
    mongoc_cursor_destroy(cursor);
    mongoc_collection_destroy(collection);
    mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
    throw se;
  } else {
    bson_t *new_doc = bson_new();
    BSON_APPEND_INT64(new_doc, "user_id", user_id);
    BSON_APPEND_UTF8(new_doc, "first_name", first_name.c_str());
    BSON_APPEND_UTF8(new_doc, "last_name", last_name.c_str());
    BSON_APPEND_UTF8(new_doc, "username", username.c_str());
    std::string salt = GenRandomString(32);
    BSON_APPEND_UTF8(new_doc, "salt", salt.c_str());
    std::string password_hashed = picosha2::hash256_hex_string(password + salt);
    BSON_APPEND_UTF8(new_doc, "password", password_hashed.c_str());

    bson_error_t error;
    auto user_insert_span = opentracing::Tracer::Global()->StartSpan(
        "user_mongo_insert_cilent", {opentracing::ChildOf(&span->context())});
    if (!mongoc_collection_insert_one(collection, new_doc, nullptr, nullptr,
                                      &error)) {
      LOG(error) << "Failed to insert user " << username
                 << " to MongoDB: " << error.message;
      ServiceException se;
      se.errorCode = ErrorCode::SE_THRIFT_HANDLER_ERROR;
      se.message =
          "Failed to insert user " + username + " to MongoDB: " + error.message;
      bson_destroy(query);
      mongoc_cursor_destroy(cursor);
      mongoc_collection_destroy(collection);
      mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
      throw se;
    } else {
      LOG(debug) << "User: " << username << " registered";
    }
    user_insert_span->Finish();
    bson_destroy(new_doc);
  }
  bson_destroy(query);
  mongoc_cursor_destroy(cursor);
  mongoc_collection_destroy(collection);
  mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);

  if (!found) {
    auto social_graph_client_wrapper = _social_graph_client_pool->Pop();
    if (!social_graph_client_wrapper) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_THRIFT_CONN_ERROR;
      se.message = "Failed to connect to social-graph-service";
      throw se;
    }
    auto social_graph_client = social_graph_client_wrapper->GetClient();
    try {
      social_graph_client->InsertUser(req_id, user_id, writer_text_map);
    } catch (...) {
      _social_graph_client_pool->Remove(social_graph_client_wrapper);
      LOG(error) << "Failed to insert user to social-graph-client";
      throw;
    }
    _social_graph_client_pool->Keepalive(social_graph_client_wrapper);
  }

  span->Finish();
}

void UserHandler::RegisterUser(
    const int64_t req_id, const std::string &first_name,
    const std::string &last_name, const std::string &username,
    const std::string &password,
    const std::map<std::string, std::string> &carrier) {
  // Initialize a span
  TextMapReader reader(carrier);
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "register_user_server", {opentracing::ChildOf(parent_span->get())});
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  // Compose user_id
  _thread_lock->lock();
  int64_t timestamp =
      duration_cast<milliseconds>(system_clock::now().time_since_epoch())
          .count() -
      CUSTOM_EPOCH;
  int idx = GetCounter(timestamp);
  _thread_lock->unlock();

  std::stringstream sstream;
  sstream << std::hex << timestamp;
  std::string timestamp_hex(sstream.str());
  if (timestamp_hex.size() > 10) {
    timestamp_hex.erase(0, timestamp_hex.size() - 10);
  } else if (timestamp_hex.size() < 10) {
    timestamp_hex = std::string(10 - timestamp_hex.size(), '0') + timestamp_hex;
  }
  // Empty the sstream buffer.
  sstream.clear();
  sstream.str(std::string());

  sstream << std::hex << idx;
  std::string counter_hex(sstream.str());

  if (counter_hex.size() > 3) {
    counter_hex.erase(0, counter_hex.size() - 3);
  } else if (counter_hex.size() < 3) {
    counter_hex = std::string(3 - counter_hex.size(), '0') + counter_hex;
  }
  std::string user_id_str = _machine_id + timestamp_hex + counter_hex;
  int64_t user_id = stoul(user_id_str, nullptr, 16) & 0x7FFFFFFFFFFFFFFF;
  ;
  LOG(debug) << "The user_id of the request " << req_id << " is " << user_id;

  // Store user info into mongodb
  mongoc_client_t *mongodb_client =
      mongoc_client_pool_pop(_mongodb_client_pool);
  if (!mongodb_client) {
    ServiceException se;
    se.errorCode = ErrorCode::SE_MONGODB_ERROR;
    se.message = "Failed to pop a client from MongoDB pool";
    throw se;
  }
  auto collection =
      mongoc_client_get_collection(mongodb_client, "user", "user");

  // Check if the username has existed in the database
  bson_t *query = bson_new();
  BSON_APPEND_UTF8(query, "username", username.c_str());
  mongoc_cursor_t *cursor =
      mongoc_collection_find_with_opts(collection, query, nullptr, nullptr);
  const bson_t *doc;
  bson_error_t error;
  bool found = mongoc_cursor_next(cursor, &doc);
  if (mongoc_cursor_error(cursor, &error)) {
    LOG(error) << error.message;
    bson_destroy(query);
    mongoc_cursor_destroy(cursor);
    mongoc_collection_destroy(collection);
    mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
    ServiceException se;
    se.errorCode = ErrorCode::SE_MONGODB_ERROR;
    se.message = error.message;
    throw se;
  } else if (found) {
    LOG(warning) << "User " << username << " already existed.";
    ServiceException se;
    se.errorCode = ErrorCode::SE_THRIFT_HANDLER_ERROR;
    se.message = "User " + username + " already existed";
    bson_destroy(query);
    mongoc_cursor_destroy(cursor);
    mongoc_collection_destroy(collection);
    mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
    throw se;
  } else {
    bson_t *new_doc = bson_new();
    BSON_APPEND_INT64(new_doc, "user_id", user_id);
    BSON_APPEND_UTF8(new_doc, "first_name", first_name.c_str());
    BSON_APPEND_UTF8(new_doc, "last_name", last_name.c_str());
    BSON_APPEND_UTF8(new_doc, "username", username.c_str());
    std::string salt = GenRandomString(32);
    BSON_APPEND_UTF8(new_doc, "salt", salt.c_str());
    std::string password_hashed = picosha2::hash256_hex_string(password + salt);
    BSON_APPEND_UTF8(new_doc, "password", password_hashed.c_str());

    auto user_insert_span = opentracing::Tracer::Global()->StartSpan(
        "user_mongo_insert_client", {opentracing::ChildOf(&span->context())});
    if (!mongoc_collection_insert_one(collection, new_doc, nullptr, nullptr,
                                      &error)) {
      LOG(error) << "Failed to insert user " << username
                 << " to MongoDB: " << error.message;
      ServiceException se;
      se.errorCode = ErrorCode::SE_THRIFT_HANDLER_ERROR;
      se.message =
          "Failed to insert user " + username + " to MongoDB: " + error.message;
      bson_destroy(query);
      mongoc_cursor_destroy(cursor);
      mongoc_collection_destroy(collection);
      mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
      throw se;
    } else {
      LOG(debug) << "User: " << username << " registered";
    }
    user_insert_span->Finish();
    bson_destroy(new_doc);
  }
  bson_destroy(query);
  mongoc_cursor_destroy(cursor);
  mongoc_collection_destroy(collection);
  mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);

  if (!found) {
    auto social_graph_client_wrapper = _social_graph_client_pool->Pop();
    if (!social_graph_client_wrapper) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_THRIFT_CONN_ERROR;
      se.message = "Failed to connect to social-graph-service";
      throw se;
    }
    auto social_graph_client = social_graph_client_wrapper->GetClient();
    try {
      social_graph_client->InsertUser(req_id, user_id, writer_text_map);
    } catch (...) {
      _social_graph_client_pool->Remove(social_graph_client_wrapper);
      LOG(error) << "Failed to insert user to social-graph-service";
      throw;
    }

    _social_graph_client_pool->Keepalive(social_graph_client_wrapper);
  }

  span->Finish();
}

void UserHandler::ComposeCreatorWithUsername(
    Creator &_return, const int64_t req_id, const std::string &username,
    const std::map<std::string, std::string> &carrier) {
  TextMapReader reader(carrier);
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "compose_creator_server", {opentracing::ChildOf(parent_span->get())});
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  size_t user_id_size;
  uint32_t memcached_flags;

  memcached_return_t memcached_rc;
  memcached_st *memcached_client =
      memcached_pool_pop(_memcached_client_pool, true, &memcached_rc);
  char *user_id_mmc;
  if (memcached_client) {
    auto id_get_span = opentracing::Tracer::Global()->StartSpan(
        "user_mmc_get_client", {opentracing::ChildOf(&span->context())});
    user_id_mmc =
        memcached_get(memcached_client, (username + ":user_id").c_str(),
                      (username + ":user_id").length(), &user_id_size,
                      &memcached_flags, &memcached_rc);
    id_get_span->Finish();
    if (!user_id_mmc && memcached_rc != MEMCACHED_NOTFOUND) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
      se.message = memcached_strerror(memcached_client, memcached_rc);
      memcached_pool_push(_memcached_client_pool, memcached_client);
      throw se;
    }
    memcached_pool_push(_memcached_client_pool, memcached_client);
  } else {
    LOG(warning) << "Failed to pop a client from memcached pool";
  }

  int64_t user_id = -1;
  bool cached = false;
  if (user_id_mmc) {
    cached = true;
    LOG(debug) << "Found user_id of username :" << username << " in Memcached";
    user_id = std::stoul(user_id_mmc);
    free(user_id_mmc);
  }

  // If not cached in memcached
  else {
    LOG(debug) << "user_id not cached in Memcached";
    mongoc_client_t *mongodb_client =
        mongoc_client_pool_pop(_mongodb_client_pool);
    if (!mongodb_client) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_MONGODB_ERROR;
      se.message = "Failed to pop a client from MongoDB pool";
      throw se;
    }
    auto collection =
        mongoc_client_get_collection(mongodb_client, "user", "user");
    if (!collection) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_MONGODB_ERROR;
      se.message = "Failed to create collection user from DB user";
      throw se;
    }
    bson_t *query = bson_new();
    BSON_APPEND_UTF8(query, "username", username.c_str());

    auto find_span = opentracing::Tracer::Global()->StartSpan(
        "user_mongo_find_client", {opentracing::ChildOf(&span->context())});
    mongoc_cursor_t *cursor =
        mongoc_collection_find_with_opts(collection, query, nullptr, nullptr);
    const bson_t *doc;
    bool found = mongoc_cursor_next(cursor, &doc);
    find_span->Finish();
    if (!found) {
      bson_error_t error;
      if (mongoc_cursor_error(cursor, &error)) {
        LOG(error) << error.message;
        bson_destroy(query);
        mongoc_cursor_destroy(cursor);
        mongoc_collection_destroy(collection);
        mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
        ServiceException se;
        se.errorCode = ErrorCode::SE_MONGODB_ERROR;
        se.message = error.message;
        throw se;
      } else {
        LOG(warning) << "User: " << username << " doesn't exist in MongoDB";
        bson_destroy(query);
        mongoc_cursor_destroy(cursor);
        mongoc_collection_destroy(collection);
        mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
        ServiceException se;
        se.errorCode = ErrorCode::SE_THRIFT_HANDLER_ERROR;
        se.message = "User: " + username + " is not registered";
        throw se;
      }
    } else {
      LOG(debug) << "User: " << username << " found in MongoDB";
      bson_iter_t iter;
      if (bson_iter_init_find(&iter, doc, "user_id")) {
        user_id = bson_iter_value(&iter)->value.v_int64;
      } else {
        LOG(error) << "user_id attribute of user " << username
                   << " was not found in the User object";
        bson_destroy(query);
        mongoc_cursor_destroy(cursor);
        mongoc_collection_destroy(collection);
        mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
        ServiceException se;
        se.errorCode = ErrorCode::SE_THRIFT_HANDLER_ERROR;
        se.message = "user_id attribute of user: " + username +
                     " was not found in the User object";
        throw se;
      }
    }
    bson_destroy(query);
    mongoc_cursor_destroy(cursor);
    mongoc_collection_destroy(collection);
    mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
  }

  Creator creator;
  creator.username = username;
  creator.user_id = user_id;

  if (user_id != -1) {
    _return = creator;
  }

  memcached_client =
      memcached_pool_pop(_memcached_client_pool, true, &memcached_rc);
  if (memcached_client) {
    if (user_id != -1 && !cached) {
      auto id_set_span = opentracing::Tracer::Global()->StartSpan(
          "user_mmc_set_cilent", {opentracing::ChildOf(&span->context())});
      std::string user_id_str = std::to_string(user_id);
      memcached_rc =
          memcached_set(memcached_client, (username + ":user_id").c_str(),
                        (username + ":user_id").length(), user_id_str.c_str(),
                        user_id_str.length(), static_cast<time_t>(0),
                        static_cast<uint32_t>(0));
      id_set_span->Finish();
      if (memcached_rc != MEMCACHED_SUCCESS) {
        LOG(warning) << "Failed to set the user_id of user " << username
                     << " to Memcached: "
                     << memcached_strerror(memcached_client, memcached_rc);
      }
    }
    memcached_pool_push(_memcached_client_pool, memcached_client);
  } else {
    LOG(warning) << "Failed to pop a client from memcached pool";
  }
  span->Finish();
}

void UserHandler::ComposeCreatorWithUserId(
    Creator &_return, int64_t req_id, int64_t user_id,
    const std::string &username,
    const std::map<std::string, std::string> &carrier) {
  TextMapReader reader(carrier);
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "compose_creator_server", {opentracing::ChildOf(parent_span->get())});
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  Creator creator;
  creator.username = username;
  creator.user_id = user_id;

  _return = creator;

  span->Finish();
}

void UserHandler::Login(std::string &_return, int64_t req_id,
                        const std::string &username,
                        const std::string &password,
                        const std::map<std::string, std::string> &carrier) {
  TextMapReader reader(carrier);
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "login_server", {opentracing::ChildOf(parent_span->get())});
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  size_t login_size;
  uint32_t memcached_flags;

  memcached_return_t memcached_rc;
  memcached_st *memcached_client =
      memcached_pool_pop(_memcached_client_pool, true, &memcached_rc);
  char *login_mmc;
  if (!memcached_client) {
    LOG(warning) << "Failed to pop a client from memcached pool";
  } else {
    auto get_login_span = opentracing::Tracer::Global()->StartSpan(
        "user_mmc_get_client", {opentracing::ChildOf(&span->context())});
    login_mmc = memcached_get(memcached_client, (username + ":login").c_str(),
                              (username + ":login").length(), &login_size,
                              &memcached_flags, &memcached_rc);
    get_login_span->Finish();
    if (!login_mmc && memcached_rc != MEMCACHED_NOTFOUND) {
      LOG(warning) << "Memcached error: "
                   << memcached_strerror(memcached_client, memcached_rc);
    }
    memcached_pool_push(_memcached_client_pool, memcached_client);
  }

  std::string password_stored;
  std::string salt_stored;
  int64_t user_id_stored = -1;
  bool cached = false;
  json login_json;

  if (login_mmc) {
    // If not cached in memcached
    LOG(debug) << "Found username: " << username << " in Memcached";
    login_json = json::parse(std::string(login_mmc, login_size));
    password_stored = login_json["password"];
    salt_stored = login_json["salt"];
    user_id_stored = login_json["user_id"];
    cached = true;
    free(login_mmc);
  }

  else {
    // If not cached in memcached
    LOG(debug) << "Username: " << username << " NOT cached in Memcached";

    mongoc_client_t *mongodb_client =
        mongoc_client_pool_pop(_mongodb_client_pool);
    if (!mongodb_client) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_MONGODB_ERROR;
      se.message = "Failed to pop a client from MongoDB pool";
      throw se;
    }
    auto collection =
        mongoc_client_get_collection(mongodb_client, "user", "user");
    if (!collection) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_MONGODB_ERROR;
      se.message = "Failed to create collection user from DB user";
      throw se;
    }
    bson_t *query = bson_new();
    BSON_APPEND_UTF8(query, "username", username.c_str());

    auto find_span = opentracing::Tracer::Global()->StartSpan(
        "user_mongo_find_client", {opentracing::ChildOf(&span->context())});
    mongoc_cursor_t *cursor =
        mongoc_collection_find_with_opts(collection, query, nullptr, nullptr);
    const bson_t *doc;
    bool found = mongoc_cursor_next(cursor, &doc);
    find_span->Finish();

    bson_error_t error;
    if (mongoc_cursor_error(cursor, &error)) {
      LOG(error) << error.message;
      bson_destroy(query);
      mongoc_cursor_destroy(cursor);
      mongoc_collection_destroy(collection);
      mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
      ServiceException se;
      se.errorCode = ErrorCode::SE_MONGODB_ERROR;
      se.message = error.message;
      throw se;
    }

    if (!found) {
      LOG(warning) << "User: " << username << " doesn't exist in MongoDB";
      bson_destroy(query);
      mongoc_cursor_destroy(cursor);
      mongoc_collection_destroy(collection);
      mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
      ServiceException se;
      se.errorCode = ErrorCode::SE_UNAUTHORIZED;
      se.message = "User: " + username + " is not registered";
      throw se;
    } else {
      LOG(debug) << "Username: " << username << " found in MongoDB";
      bson_iter_t iter_password;
      bson_iter_t iter_salt;
      bson_iter_t iter_user_id;
      if (bson_iter_init_find(&iter_password, doc, "password") &&
          bson_iter_init_find(&iter_salt, doc, "salt") &&
          bson_iter_init_find(&iter_user_id, doc, "user_id")) {
        password_stored = bson_iter_value(&iter_password)->value.v_utf8.str;
        salt_stored = bson_iter_value(&iter_salt)->value.v_utf8.str;
        user_id_stored = bson_iter_value(&iter_user_id)->value.v_int64;
        login_json["password"] = password_stored;
        login_json["salt"] = salt_stored;
        login_json["user_id"] = user_id_stored;
      } else {
        LOG(error) << "user: " << username << " entry is NOT complete";
        bson_destroy(query);
        mongoc_cursor_destroy(cursor);
        mongoc_collection_destroy(collection);
        mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
        ServiceException se;
        se.errorCode = ErrorCode::SE_THRIFT_HANDLER_ERROR;
        se.message = "user: " + username + " entry is NOT complete";
        throw se;
      }
      bson_destroy(query);
      mongoc_cursor_destroy(cursor);
      mongoc_collection_destroy(collection);
      mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
    }
  }

  if (user_id_stored != -1 && !salt_stored.empty() &&
      !password_stored.empty()) {
    bool auth =
        picosha2::hash256_hex_string(password + salt_stored) == password_stored;
    if (auth) {
      auto user_id_str = std::to_string(user_id_stored);
      auto timestamp_str = std::to_string(
          duration_cast<seconds>(system_clock::now().time_since_epoch())
              .count());
      jwt::jwt_object obj{algorithm("HS256"), secret(_secret),
                          payload({{"user_id", user_id_str},
                                   {"username", username},
                                   {"timestamp", timestamp_str},
                                   {"ttl", "3600"}})};
      _return = obj.signature();
    } else {
      ServiceException se;
      se.errorCode = ErrorCode::SE_UNAUTHORIZED;
      se.message = "Incorrect username or password";
      throw se;
    }
  } else {
    ServiceException se;
    se.errorCode = ErrorCode::SE_THRIFT_HANDLER_ERROR;
    se.message = "Username: " + username + " incomplete login information.";
    throw se;
  }

  if (!cached) {
    memcached_client =
        memcached_pool_pop(_memcached_client_pool, true, &memcached_rc);
    if (!memcached_client) {
      LOG(warning) << "Failed to pop a client from memcached pool";
    } else {
      auto set_login_span = opentracing::Tracer::Global()->StartSpan(
          "user_mmc_set_client", {opentracing::ChildOf(&span->context())});
      std::string login_str = login_json.dump();
      memcached_rc =
          memcached_set(memcached_client, (username + ":login").c_str(),
                        (username + ":login").length(), login_str.c_str(),
                        login_str.length(), 0, 0);
      set_login_span->Finish();
      if (memcached_rc != MEMCACHED_SUCCESS) {
        LOG(warning) << "Failed to set the login info of user " << username
                     << " to Memcached: "
                     << memcached_strerror(memcached_client, memcached_rc);
      }
      memcached_pool_push(_memcached_client_pool, memcached_client);
    }
  }
  span->Finish();
}
int64_t UserHandler::GetUserId(
    int64_t req_id, const std::string &username,
    const std::map<std::string, std::string> &carrier) {
  TextMapReader reader(carrier);
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "get_user_id_server", {opentracing::ChildOf(parent_span->get())});
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  size_t user_id_size;
  uint32_t memcached_flags;

  memcached_return_t memcached_rc;
  memcached_st *memcached_client =
      memcached_pool_pop(_memcached_client_pool, true, &memcached_rc);
  char *user_id_mmc;
  if (memcached_client) {
    auto id_get_span = opentracing::Tracer::Global()->StartSpan(
        "user_mmc_get_user_id_client",
        {opentracing::ChildOf(&span->context())});
    user_id_mmc =
        memcached_get(memcached_client, (username + ":user_id").c_str(),
                      (username + ":user_id").length(), &user_id_size,
                      &memcached_flags, &memcached_rc);
    id_get_span->Finish();
    if (!user_id_mmc && memcached_rc != MEMCACHED_NOTFOUND) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
      se.message = memcached_strerror(memcached_client, memcached_rc);
      memcached_pool_push(_memcached_client_pool, memcached_client);
      throw se;
    }
    memcached_pool_push(_memcached_client_pool, memcached_client);
  } else {
    LOG(warning) << "Failed to pop a client from memcached pool";
  }

  int64_t user_id = -1;
  bool cached = false;
  if (user_id_mmc) {
    cached = true;
    LOG(debug) << "Found user_id of username :" << username << " in Memcached";
    user_id = std::stoul(user_id_mmc);
    free(user_id_mmc);
  } else {
    // If not cached in memcached
    LOG(debug) << "user_id not cached in Memcached";
    mongoc_client_t *mongodb_client =
        mongoc_client_pool_pop(_mongodb_client_pool);
    if (!mongodb_client) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_MONGODB_ERROR;
      se.message = "Failed to pop a client from MongoDB pool";
      throw se;
    }
    auto collection =
        mongoc_client_get_collection(mongodb_client, "user", "user");
    if (!collection) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_MONGODB_ERROR;
      se.message = "Failed to create collection user from DB user";
      throw se;
    }
    bson_t *query = bson_new();
    BSON_APPEND_UTF8(query, "username", username.c_str());

    auto find_span = opentracing::Tracer::Global()->StartSpan(
        "user_mongo_find_client", {opentracing::ChildOf(&span->context())});
    mongoc_cursor_t *cursor =
        mongoc_collection_find_with_opts(collection, query, nullptr, nullptr);
    const bson_t *doc;
    bool found = mongoc_cursor_next(cursor, &doc);
    find_span->Finish();
    if (!found) {
      bson_error_t error;
      if (mongoc_cursor_error(cursor, &error)) {
        LOG(error) << error.message;
        bson_destroy(query);
        mongoc_cursor_destroy(cursor);
        mongoc_collection_destroy(collection);
        mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
        ServiceException se;
        se.errorCode = ErrorCode::SE_MONGODB_ERROR;
        se.message = error.message;
        throw se;
      } else {
        LOG(warning) << "User: " << username << " doesn't exist in MongoDB";
        bson_destroy(query);
        mongoc_cursor_destroy(cursor);
        mongoc_collection_destroy(collection);
        mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
        ServiceException se;
        se.errorCode = ErrorCode::SE_THRIFT_HANDLER_ERROR;
        se.message = "User: " + username + " is not registered";
        throw se;
      }
    } else {
      LOG(debug) << "User: " << username << " found in MongoDB";
      bson_iter_t iter;
      if (bson_iter_init_find(&iter, doc, "user_id")) {
        user_id = bson_iter_value(&iter)->value.v_int64;
      } else {
        LOG(error) << "user_id attribute of user " << username
                   << " was not found in the User object";
        bson_destroy(query);
        mongoc_cursor_destroy(cursor);
        mongoc_collection_destroy(collection);
        mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
        ServiceException se;
        se.errorCode = ErrorCode::SE_THRIFT_HANDLER_ERROR;
        se.message = "user_id attribute of user: " + username +
                     " was not found in the User object";
        throw se;
      }
    }
    bson_destroy(query);
    mongoc_cursor_destroy(cursor);
    mongoc_collection_destroy(collection);
    mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
  }

  if (!cached) {
    memcached_client =
        memcached_pool_pop(_memcached_client_pool, true, &memcached_rc);
    if (!memcached_client) {
      LOG(warning) << "Failed to pop a client from memcached pool";
    } else {
      std::string user_id_str = std::to_string(user_id);
      auto set_login_span = opentracing::Tracer::Global()->StartSpan(
          "user_mmc_set_client", {opentracing::ChildOf(&span->context())});
      memcached_rc =
          memcached_set(memcached_client, (username + ":user_id").c_str(),
                        (username + ":user_id").length(), user_id_str.c_str(),
                        user_id_str.length(), 0, 0);
      set_login_span->Finish();
      if (memcached_rc != MEMCACHED_SUCCESS) {
        LOG(warning) << "Failed to set the login info of user " << username
                     << " to Memcached: "
                     << memcached_strerror(memcached_client, memcached_rc);
      }
      memcached_pool_push(_memcached_client_pool, memcached_client);
    }
  }

  span->Finish();
  return user_id;
}

/*
 * The following code which obtaines machine ID from machine's MAC address was
 * inspired from https://stackoverflow.com/a/16859693.
 *
 * MAC address is obtained from /sys/class/net/<netif>/address
 */
u_int16_t HashMacAddressPid(const std::string &mac) {
  u_int16_t hash = 0;
  std::string mac_pid = mac + std::to_string(getpid());
  for (unsigned int i = 0; i < mac_pid.size(); i++) {
    hash += (mac[i] << ((i & 1) * 8));
  }
  return hash;
}

std::string GetMachineId(std::string &netif) {
  std::string mac_hash;

  std::string mac_addr_filename = "/sys/class/net/" + netif + "/address";
  std::ifstream mac_addr_file;
  mac_addr_file.open(mac_addr_filename);
  if (!mac_addr_file) {
    LOG(fatal) << "Cannot read MAC address from net interface " << netif;
    return "";
  }
  std::string mac;
  mac_addr_file >> mac;
  if (mac == "") {
    LOG(fatal) << "Cannot read MAC address from net interface " << netif;
    return "";
  }
  mac_addr_file.close();

  LOG(info) << "MAC address = " << mac;

  std::stringstream stream;
  stream << std::hex << HashMacAddressPid(mac);
  mac_hash = stream.str();

  if (mac_hash.size() > 3) {
    mac_hash.erase(0, mac_hash.size() - 3);
  } else if (mac_hash.size() < 3) {
    mac_hash = std::string(3 - mac_hash.size(), '0') + mac_hash;
  }
  return mac_hash;
}
}  // namespace social_network

#endif  // SOCIAL_NETWORK_MICROSERVICES_USERHANDLER_H
