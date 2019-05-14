#ifndef SOCIAL_NETWORK_MICROSERVICES_SRC_USERMENTIONSERVICE_USERMENTIONHANDLER_H_
#define SOCIAL_NETWORK_MICROSERVICES_SRC_USERMENTIONSERVICE_USERMENTIONHANDLER_H_

#include <mongoc.h>
#include <bson.h>
#include <libmemcached/memcached.h>
#include <libmemcached/util.h>

#include "../../gen-cpp/UserMentionService.h"
#include "../../gen-cpp/ComposePostService.h"
#include "../../gen-cpp/social_network_types.h"
#include "../ClientPool.h"
#include "../ThriftClient.h"
#include "../logger.h"
#include "../tracing.h"
#include "../utils.h"

namespace social_network {

class UserMentionHandler : public UserMentionServiceIf {
 public:
  UserMentionHandler(memcached_pool_st *,
                     mongoc_client_pool_t *,
                     ClientPool<ThriftClient<ComposePostServiceClient>> *);
  ~UserMentionHandler() override = default;

  void UploadUserMentions(int64_t, const std::vector<std::string> &,
      const std::map<std::string, std::string> &) override ;

 private:
  memcached_pool_st *_memcached_client_pool;
  mongoc_client_pool_t *_mongodb_client_pool;
  ClientPool<ThriftClient<ComposePostServiceClient>> *_compose_client_pool;
};

UserMentionHandler::UserMentionHandler(
    memcached_pool_st *memcached_client_pool,
    mongoc_client_pool_t *mongodb_client_pool,
    ClientPool<ThriftClient<ComposePostServiceClient>> *compose_client_pool) {
  _memcached_client_pool = memcached_client_pool;
  _mongodb_client_pool = mongodb_client_pool;
  _compose_client_pool = compose_client_pool;
}

void UserMentionHandler::UploadUserMentions(
    int64_t req_id,
    const std::vector<std::string> &usernames,
    const std::map<std::string, std::string> &carrier) {

  // Initialize a span
  TextMapReader reader(carrier);
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "UploadUserMentions",
      { opentracing::ChildOf(parent_span->get()) });
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  std::vector<UserMention> user_mentions;
  if (!usernames.empty()) {
    std::map<std::string, bool> usernames_not_cached;

    for (auto &username : usernames) {
      usernames_not_cached.emplace(std::make_pair(username, false));
    }

    // Find in Memcached
    memcached_return_t rc;
    auto client = memcached_pool_pop(_memcached_client_pool, true, &rc);
    if (!client) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
      se.message = "Failed to pop a client from memcached pool";
      throw se;
    }

    char** keys;
    size_t *key_sizes;
    keys = new char* [usernames.size()];
    key_sizes = new size_t [usernames.size()];
    int idx = 0;
    for (auto &username : usernames) {
      std::string key_str = username + ":user_id";
      keys[idx] = new char [key_str.length() + 1];
      strcpy(keys[idx], key_str.c_str());
      key_sizes[idx] = key_str.length();
      idx++;
    }

    rc = memcached_mget(client, keys, key_sizes, usernames.size());
    if (rc != MEMCACHED_SUCCESS) {
      LOG(error) << "Cannot get usernames of request " << req_id << ": "
                 << memcached_strerror(client, rc);
      ServiceException se;
      se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
      se.message = memcached_strerror(client, rc);
      memcached_pool_push(_memcached_client_pool, client);
      throw se;
    }

    char return_key[MEMCACHED_MAX_KEY];
    size_t return_key_length;
    char *return_value;
    size_t return_value_length;
    uint32_t flags;

    while (true) {
      return_value = memcached_fetch(client, return_key, &return_key_length,
                                     &return_value_length, &flags, &rc);
      if (return_value == nullptr) {
        LOG(debug) << "Memcached mget finished "
                   << memcached_strerror(client, rc);
        break;
      }
      if (rc != MEMCACHED_SUCCESS) {
        free(return_value);
        memcached_quit(client);
        memcached_pool_push(_memcached_client_pool, client);
        LOG(error) << "Cannot get components of request " << req_id;
        ServiceException se;
        se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
        se.message =  "Cannot get usernames of request " + std::to_string(req_id);
        throw se;
      }
      UserMention new_user_mention;
      std::string username(return_key, return_key + return_key_length);
      username = username.substr(0, username.length() - std::strlen(":user_id"));
      new_user_mention.username = username;
      new_user_mention.user_id = std::stoul(
          std::string(return_value, return_value + return_value_length));
      user_mentions.emplace_back(new_user_mention);
      usernames_not_cached.erase(username);
      free(return_value);
    }
    memcached_quit(client);
    memcached_pool_push(_memcached_client_pool, client);
    for (int i = 0; i < usernames.size(); ++i) {
      delete keys[i];
    }
    delete[] keys;
    delete[] key_sizes;

    // Find the rest in MongoDB
    if (!usernames_not_cached.empty()) {
      mongoc_client_t *mongodb_client = mongoc_client_pool_pop(
          _mongodb_client_pool);
      if (!mongodb_client) {
        ServiceException se;
        se.errorCode = ErrorCode::SE_MONGODB_ERROR;
        se.message = "Failed to pop a client from MongoDB pool";
        throw se;
      }

      auto collection = mongoc_client_get_collection(
          mongodb_client, "user", "user");
      if (!collection) {
        ServiceException se;
        se.errorCode = ErrorCode::SE_MONGODB_ERROR;
        se.message = "Failed to create collection user from DB user";
        mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
        throw se;
      }

      bson_t *query = bson_new();
      bson_t query_child_0;
      bson_t query_username_list;
      const char *key;
      idx = 0;
      char buf[16];

      BSON_APPEND_DOCUMENT_BEGIN(query, "username", &query_child_0);
      BSON_APPEND_ARRAY_BEGIN(&query_child_0, "$in", &query_username_list);
      for (auto &item : usernames_not_cached) {
        bson_uint32_to_string(idx, &key, buf, sizeof buf);
        BSON_APPEND_UTF8(&query_username_list, key, item.first.c_str());
        idx++;
      }
      bson_append_array_end(&query_child_0, &query_username_list);
      bson_append_document_end(query, &query_child_0);

      mongoc_cursor_t *cursor = mongoc_collection_find_with_opts(
          collection, query, nullptr, nullptr);
      const bson_t *doc;

      while (mongoc_cursor_next(cursor, &doc)) {
        bson_iter_t iter;
        UserMention new_user_mention;
        if (bson_iter_init_find(&iter, doc, "user_id")) {
          new_user_mention.user_id = bson_iter_value(&iter)->value.v_int64;
        } else {
          ServiceException se;
          se.errorCode = ErrorCode::SE_MONGODB_ERROR;
          se.message = "Attribute of MongoDB item is not complete";
          bson_destroy(query);
          mongoc_cursor_destroy(cursor);
          mongoc_collection_destroy(collection);
          mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
          throw se;
        }
        if (bson_iter_init_find(&iter, doc, "username")) {
          new_user_mention.username = bson_iter_value(&iter)->value.v_utf8.str;
        } else {
          ServiceException se;
          se.errorCode = ErrorCode::SE_MONGODB_ERROR;
          se.message = "Attribute of MongoDB item is not complete";
          bson_destroy(query);
          mongoc_cursor_destroy(cursor);
          mongoc_collection_destroy(collection);
          mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
          throw se;
        }
        user_mentions.emplace_back(new_user_mention);
      }
      bson_destroy(query);
      mongoc_cursor_destroy(cursor);
      mongoc_collection_destroy(collection);
      mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
    }
  }

  // Upload to compose post service
  auto compose_post_client_wrapper = _compose_client_pool->Pop();
  if (!compose_post_client_wrapper) {
    ServiceException se;
    se.errorCode = ErrorCode::SE_THRIFT_CONN_ERROR;
    se.message = "Failed to connected to compose-post-service";
    throw se;
  }
  auto compose_post_client = compose_post_client_wrapper->GetClient();
  try {
    compose_post_client->UploadUserMentions(req_id, user_mentions,
                                            writer_text_map);
  } catch (...) {
    _compose_client_pool->Push(compose_post_client_wrapper);
    LOG(error) << "Failed to upload user_mentions to user-mention-service";
    throw;
  }  
  _compose_client_pool->Push(compose_post_client_wrapper);
  span->Finish();
}

}

#endif //SOCIAL_NETWORK_MICROSERVICES_SRC_USERMENTIONSERVICE_USERMENTIONHANDLER_H_
