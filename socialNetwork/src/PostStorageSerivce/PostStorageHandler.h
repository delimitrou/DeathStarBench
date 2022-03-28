#ifndef SOCIAL_NETWORK_MICROSERVICES_POSTSTORAGEHANDLER_H
#define SOCIAL_NETWORK_MICROSERVICES_POSTSTORAGEHANDLER_H

#include <bson/bson.h>
#include <libmemcached/memcached.h>
#include <libmemcached/util.h>
#include <mongoc.h>

#include <future>
#include <iostream>
#include <nlohmann/json.hpp>
#include <string>

#include "../../gen-cpp/PostStorageService.h"
#include "../logger.h"
#include "../tracing.h"

namespace social_network {
using json = nlohmann::json;

class PostStorageHandler : public PostStorageServiceIf {
 public:
  PostStorageHandler(memcached_pool_st *, mongoc_client_pool_t *);
  ~PostStorageHandler() override = default;

  void StorePost(int64_t req_id, const Post &post,
                 const std::map<std::string, std::string> &carrier) override;

  void ReadPost(Post &_return, int64_t req_id, int64_t post_id,
                const std::map<std::string, std::string> &carrier) override;

  void ReadPosts(std::vector<Post> &_return, int64_t req_id,
                 const std::vector<int64_t> &post_ids,
                 const std::map<std::string, std::string> &carrier) override;

 private:
  memcached_pool_st *_memcached_client_pool;
  mongoc_client_pool_t *_mongodb_client_pool;
};

PostStorageHandler::PostStorageHandler(
    memcached_pool_st *memcached_client_pool,
    mongoc_client_pool_t *mongodb_client_pool) {
  _memcached_client_pool = memcached_client_pool;
  _mongodb_client_pool = mongodb_client_pool;
}

void PostStorageHandler::StorePost(
    int64_t req_id, const social_network::Post &post,
    const std::map<std::string, std::string> &carrier) {
  // Initialize a span
  TextMapReader reader(carrier);
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "store_post_server", {opentracing::ChildOf(parent_span->get())});
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  mongoc_client_t *mongodb_client =
      mongoc_client_pool_pop(_mongodb_client_pool);
  if (!mongodb_client) {
    ServiceException se;
    se.errorCode = ErrorCode::SE_MONGODB_ERROR;
    se.message = "Failed to pop a client from MongoDB pool";
    throw se;
  }

  auto collection =
      mongoc_client_get_collection(mongodb_client, "post", "post");
  if (!collection) {
    ServiceException se;
    se.errorCode = ErrorCode::SE_MONGODB_ERROR;
    se.message = "Failed to create collection user from DB user";
    mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
    throw se;
  }

  bson_t *new_doc = bson_new();
  BSON_APPEND_INT64(new_doc, "post_id", post.post_id);
  BSON_APPEND_INT64(new_doc, "timestamp", post.timestamp);
  BSON_APPEND_UTF8(new_doc, "text", post.text.c_str());
  BSON_APPEND_INT64(new_doc, "req_id", post.req_id);
  BSON_APPEND_INT32(new_doc, "post_type", post.post_type);

  bson_t creator_doc;
  BSON_APPEND_DOCUMENT_BEGIN(new_doc, "creator", &creator_doc);
  BSON_APPEND_INT64(&creator_doc, "user_id", post.creator.user_id);
  BSON_APPEND_UTF8(&creator_doc, "username", post.creator.username.c_str());
  bson_append_document_end(new_doc, &creator_doc);

  const char *key;
  int idx = 0;
  char buf[16];

  bson_t url_list;
  BSON_APPEND_ARRAY_BEGIN(new_doc, "urls", &url_list);
  for (auto &url : post.urls) {
    bson_uint32_to_string(idx, &key, buf, sizeof buf);
    bson_t url_doc;
    BSON_APPEND_DOCUMENT_BEGIN(&url_list, key, &url_doc);
    BSON_APPEND_UTF8(&url_doc, "shortened_url", url.shortened_url.c_str());
    BSON_APPEND_UTF8(&url_doc, "expanded_url", url.expanded_url.c_str());
    bson_append_document_end(&url_list, &url_doc);
    idx++;
  }
  bson_append_array_end(new_doc, &url_list);

  bson_t user_mention_list;
  idx = 0;
  BSON_APPEND_ARRAY_BEGIN(new_doc, "user_mentions", &user_mention_list);
  for (auto &user_mention : post.user_mentions) {
    bson_uint32_to_string(idx, &key, buf, sizeof buf);
    bson_t user_mention_doc;
    BSON_APPEND_DOCUMENT_BEGIN(&user_mention_list, key, &user_mention_doc);
    BSON_APPEND_INT64(&user_mention_doc, "user_id", user_mention.user_id);
    BSON_APPEND_UTF8(&user_mention_doc, "username",
                     user_mention.username.c_str());
    bson_append_document_end(&user_mention_list, &user_mention_doc);
    idx++;
  }
  bson_append_array_end(new_doc, &user_mention_list);

  bson_t media_list;
  idx = 0;
  BSON_APPEND_ARRAY_BEGIN(new_doc, "media", &media_list);
  for (auto &media : post.media) {
    bson_uint32_to_string(idx, &key, buf, sizeof buf);
    bson_t media_doc;
    BSON_APPEND_DOCUMENT_BEGIN(&media_list, key, &media_doc);
    BSON_APPEND_INT64(&media_doc, "media_id", media.media_id);
    BSON_APPEND_UTF8(&media_doc, "media_type", media.media_type.c_str());
    bson_append_document_end(&media_list, &media_doc);
    idx++;
  }
  bson_append_array_end(new_doc, &media_list);

  bson_error_t error;
  auto insert_span = opentracing::Tracer::Global()->StartSpan(
      "post_storage_mongo_insert_client",
      {opentracing::ChildOf(&span->context())});
  bool inserted = mongoc_collection_insert_one(collection, new_doc, nullptr,
                                               nullptr, &error);
  insert_span->Finish();

  if (!inserted) {
    LOG(error) << "Error: Failed to insert post to MongoDB: " << error.message;
    ServiceException se;
    se.errorCode = ErrorCode::SE_MONGODB_ERROR;
    se.message = error.message;
    bson_destroy(new_doc);
    mongoc_collection_destroy(collection);
    mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
    throw se;
  }

  bson_destroy(new_doc);
  mongoc_collection_destroy(collection);
  mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);

  span->Finish();
}

void PostStorageHandler::ReadPost(
    Post &_return, int64_t req_id, int64_t post_id,
    const std::map<std::string, std::string> &carrier) {
  // Initialize a span
  TextMapReader reader(carrier);
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "read_post_server", {opentracing::ChildOf(parent_span->get())});
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  std::string post_id_str = std::to_string(post_id);

  memcached_return_t memcached_rc;
  memcached_st *memcached_client =
      memcached_pool_pop(_memcached_client_pool, true, &memcached_rc);
  if (!memcached_client) {
    ServiceException se;
    se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
    se.message = "Failed to pop a client from memcached pool";
    throw se;
  }

  size_t post_mmc_size;
  uint32_t memcached_flags;
  auto get_span = opentracing::Tracer::Global()->StartSpan(
      "post_storage_mmc_get_client", {opentracing::ChildOf(&span->context())});
  char *post_mmc =
      memcached_get(memcached_client, post_id_str.c_str(), post_id_str.length(),
                    &post_mmc_size, &memcached_flags, &memcached_rc);
  if (!post_mmc && memcached_rc != MEMCACHED_NOTFOUND) {
    ServiceException se;
    se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
    se.message = memcached_strerror(memcached_client, memcached_rc);
    memcached_pool_push(_memcached_client_pool, memcached_client);
    throw se;
  }
  memcached_pool_push(_memcached_client_pool, memcached_client);
  get_span->Finish();

  if (post_mmc) {
    LOG(debug) << "Get post " << post_id << " cache hit from Memcached";
    json post_json =
        json::parse(std::string(post_mmc, post_mmc + post_mmc_size));
    _return.req_id = post_json["req_id"];
    _return.timestamp = post_json["timestamp"];
    _return.post_id = post_json["post_id"];
    _return.creator.user_id = post_json["creator"]["user_id"];
    _return.creator.username = post_json["creator"]["username"];
    _return.post_type = post_json["post_type"];
    _return.text = post_json["text"];
    for (auto &item : post_json["media"]) {
      Media media;
      media.media_id = item["media_id"];
      media.media_type = item["media_type"];
      _return.media.emplace_back(media);
    }
    for (auto &item : post_json["user_mentions"]) {
      UserMention user_mention;
      user_mention.username = item["username"];
      user_mention.user_id = item["user_id"];
      _return.user_mentions.emplace_back(user_mention);
    }
    for (auto &item : post_json["urls"]) {
      Url url;
      url.shortened_url = item["shortened_url"];
      url.expanded_url = item["expanded_url"];
      _return.urls.emplace_back(url);
    }
    free(post_mmc);
  } else {
    // If not cached in memcached
    mongoc_client_t *mongodb_client =
        mongoc_client_pool_pop(_mongodb_client_pool);
    if (!mongodb_client) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_MONGODB_ERROR;
      se.message = "Failed to pop a client from MongoDB pool";
      throw se;
    }

    auto collection =
        mongoc_client_get_collection(mongodb_client, "post", "post");
    if (!collection) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_MONGODB_ERROR;
      se.message = "Failed to create collection user from DB user";
      mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
      throw se;
    }

    bson_t *query = bson_new();
    BSON_APPEND_INT64(query, "post_id", post_id);
    auto find_span = opentracing::Tracer::Global()->StartSpan(
        "post_storage_mongo_find_client",
        {opentracing::ChildOf(&span->context())});
    mongoc_cursor_t *cursor =
        mongoc_collection_find_with_opts(collection, query, nullptr, nullptr);
    const bson_t *doc;
    bool found = mongoc_cursor_next(cursor, &doc);
    find_span->Finish();
    if (!found) {
      bson_error_t error;
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
      } else {
        LOG(warning) << "Post_id: " << post_id << " doesn't exist in MongoDB";
        bson_destroy(query);
        mongoc_cursor_destroy(cursor);
        mongoc_collection_destroy(collection);
        mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
        ServiceException se;
        se.errorCode = ErrorCode::SE_THRIFT_HANDLER_ERROR;
        se.message =
            "Post_id: " + std::to_string(post_id) + " doesn't exist in MongoDB";
        throw se;
      }
    } else {
      LOG(debug) << "Post_id: " << post_id << " found in MongoDB";
      auto post_json_char = bson_as_json(doc, nullptr);
      json post_json = json::parse(post_json_char);
      _return.req_id = post_json["req_id"];
      _return.timestamp = post_json["timestamp"];
      _return.post_id = post_json["post_id"];
      _return.creator.user_id = post_json["creator"]["user_id"];
      _return.creator.username = post_json["creator"]["username"];
      _return.post_type = post_json["post_type"];
      _return.text = post_json["text"];
      for (auto &item : post_json["media"]) {
        Media media;
        media.media_id = item["media_id"];
        media.media_type = item["media_type"];
        _return.media.emplace_back(media);
      }
      for (auto &item : post_json["user_mentions"]) {
        UserMention user_mention;
        user_mention.username = item["username"];
        user_mention.user_id = item["user_id"];
        _return.user_mentions.emplace_back(user_mention);
      }
      for (auto &item : post_json["urls"]) {
        Url url;
        url.shortened_url = item["shortened_url"];
        url.expanded_url = item["expanded_url"];
        _return.urls.emplace_back(url);
      }
      bson_destroy(query);
      mongoc_cursor_destroy(cursor);
      mongoc_collection_destroy(collection);
      mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);

      // upload post to memcached
      memcached_client =
          memcached_pool_pop(_memcached_client_pool, true, &memcached_rc);
      if (!memcached_client) {
        ServiceException se;
        se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
        se.message = "Failed to pop a client from memcached pool";
        throw se;
      }
      auto set_span = opentracing::Tracer::Global()->StartSpan(
          "post_storage_mmc_set_client",
          {opentracing::ChildOf(&span->context())});

      memcached_rc = memcached_set(
          memcached_client, post_id_str.c_str(), post_id_str.length(),
          post_json_char, std::strlen(post_json_char), static_cast<time_t>(0),
          static_cast<uint32_t>(0));
      if (memcached_rc != MEMCACHED_SUCCESS) {
        LOG(warning) << "Failed to set post to Memcached: "
                     << memcached_strerror(memcached_client, memcached_rc);
      }
      set_span->Finish();
      bson_free(post_json_char);
      memcached_pool_push(_memcached_client_pool, memcached_client);
    }
  }

  span->Finish();
}
void PostStorageHandler::ReadPosts(
    std::vector<Post> &_return, int64_t req_id,
    const std::vector<int64_t> &post_ids,
    const std::map<std::string, std::string> &carrier) {
  // Initialize a span
  TextMapReader reader(carrier);
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "post_storage_read_posts_server",
      {opentracing::ChildOf(parent_span->get())});
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  if (post_ids.empty()) {
    return;
  }

  std::set<int64_t> post_ids_not_cached(post_ids.begin(), post_ids.end());
  if (post_ids_not_cached.size() != post_ids.size()) {
    LOG(error)<< "Post_ids are duplicated";
    ServiceException se;
    se.errorCode = ErrorCode::SE_THRIFT_HANDLER_ERROR;
    se.message = "Post_ids are duplicated";
    throw se;
  }
  std::map<int64_t, Post> return_map;
  memcached_return_t memcached_rc;
  auto memcached_client =
      memcached_pool_pop(_memcached_client_pool, true, &memcached_rc);
  if (!memcached_client) {
    ServiceException se;
    se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
    se.message = "Failed to pop a client from memcached pool";
    throw se;
  }

  char **keys;
  size_t *key_sizes;
  keys = new char *[post_ids.size()];
  key_sizes = new size_t[post_ids.size()];
  int idx = 0;
  for (auto &post_id : post_ids) {
    std::string key_str = std::to_string(post_id);
    keys[idx] = new char[key_str.length() + 1];
    strcpy(keys[idx], key_str.c_str());
    key_sizes[idx] = key_str.length();
    idx++;
  }
  memcached_rc =
      memcached_mget(memcached_client, keys, key_sizes, post_ids.size());
  if (memcached_rc != MEMCACHED_SUCCESS) {
    LOG(error) << "Cannot get post_ids of request " << req_id << ": "
               << memcached_strerror(memcached_client, memcached_rc);
    ServiceException se;
    se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
    se.message = memcached_strerror(memcached_client, memcached_rc);
    memcached_pool_push(_memcached_client_pool, memcached_client);
    throw se;
  }

  char return_key[MEMCACHED_MAX_KEY];
  size_t return_key_length;
  char *return_value;
  size_t return_value_length;
  uint32_t flags;
  auto get_span = opentracing::Tracer::Global()->StartSpan(
      "post_storage_mmc_mget_client", {opentracing::ChildOf(&span->context())});

  while (true) {
    return_value =
        memcached_fetch(memcached_client, return_key, &return_key_length,
                        &return_value_length, &flags, &memcached_rc);
    if (return_value == nullptr) {
      LOG(debug) << "Memcached mget finished";
      break;
    }
    if (memcached_rc != MEMCACHED_SUCCESS) {
      free(return_value);
      memcached_quit(memcached_client);
      memcached_pool_push(_memcached_client_pool, memcached_client);
      LOG(error) << "Cannot get posts of request " << req_id;
      ServiceException se;
      se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
      se.message = "Cannot get posts of request " + std::to_string(req_id);
      throw se;
    }
    Post new_post;
    json post_json = json::parse(
        std::string(return_value, return_value + return_value_length));
    new_post.req_id = post_json["req_id"];
    new_post.timestamp = post_json["timestamp"];
    new_post.post_id = post_json["post_id"];
    new_post.creator.user_id = post_json["creator"]["user_id"];
    new_post.creator.username = post_json["creator"]["username"];
    new_post.post_type = post_json["post_type"];
    new_post.text = post_json["text"];
    for (auto &item : post_json["media"]) {
      Media media;
      media.media_id = item["media_id"];
      media.media_type = item["media_type"];
      new_post.media.emplace_back(media);
    }
    for (auto &item : post_json["user_mentions"]) {
      UserMention user_mention;
      user_mention.username = item["username"];
      user_mention.user_id = item["user_id"];
      new_post.user_mentions.emplace_back(user_mention);
    }
    for (auto &item : post_json["urls"]) {
      Url url;
      url.shortened_url = item["shortened_url"];
      url.expanded_url = item["expanded_url"];
      new_post.urls.emplace_back(url);
    }
    return_map.insert(std::make_pair(new_post.post_id, new_post));
    post_ids_not_cached.erase(new_post.post_id);
    free(return_value);
  }
  get_span->Finish();
  memcached_quit(memcached_client);
  memcached_pool_push(_memcached_client_pool, memcached_client);
  for (int i = 0; i < post_ids.size(); ++i) {
    delete keys[i];
  }
  delete[] keys;
  delete[] key_sizes;

  std::vector<std::future<void>> set_futures;
  std::map<int64_t, std::string> post_json_map;

  // Find the rest in MongoDB
  if (!post_ids_not_cached.empty()) {
    mongoc_client_t *mongodb_client =
        mongoc_client_pool_pop(_mongodb_client_pool);
    if (!mongodb_client) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_MONGODB_ERROR;
      se.message = "Failed to pop a client from MongoDB pool";
      throw se;
    }
    auto collection =
        mongoc_client_get_collection(mongodb_client, "post", "post");
    if (!collection) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_MONGODB_ERROR;
      se.message = "Failed to create collection user from DB user";
      mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
      throw se;
    }
    bson_t *query = bson_new();
    bson_t query_child;
    bson_t query_post_id_list;
    const char *key;
    idx = 0;
    char buf[16];

    BSON_APPEND_DOCUMENT_BEGIN(query, "post_id", &query_child);
    BSON_APPEND_ARRAY_BEGIN(&query_child, "$in", &query_post_id_list);
    for (auto &item : post_ids_not_cached) {
      bson_uint32_to_string(idx, &key, buf, sizeof buf);
      BSON_APPEND_INT64(&query_post_id_list, key, item);
      idx++;
    }
    bson_append_array_end(&query_child, &query_post_id_list);
    bson_append_document_end(query, &query_child);
    mongoc_cursor_t *cursor =
        mongoc_collection_find_with_opts(collection, query, nullptr, nullptr);
    const bson_t *doc;

    auto find_span = opentracing::Tracer::Global()->StartSpan(
        "mongo_find_client", {opentracing::ChildOf(&span->context())});
    while (true) {
      bool found = mongoc_cursor_next(cursor, &doc);
      if (!found) {
        break;
      }
      Post new_post;
      char *post_json_char = bson_as_json(doc, nullptr);
      json post_json = json::parse(post_json_char);
      new_post.req_id = post_json["req_id"];
      new_post.timestamp = post_json["timestamp"];
      new_post.post_id = post_json["post_id"];
      new_post.creator.user_id = post_json["creator"]["user_id"];
      new_post.creator.username = post_json["creator"]["username"];
      new_post.post_type = post_json["post_type"];
      new_post.text = post_json["text"];
      for (auto &item : post_json["media"]) {
        Media media;
        media.media_id = item["media_id"];
        media.media_type = item["media_type"];
        new_post.media.emplace_back(media);
      }
      for (auto &item : post_json["user_mentions"]) {
        UserMention user_mention;
        user_mention.username = item["username"];
        user_mention.user_id = item["user_id"];
        new_post.user_mentions.emplace_back(user_mention);
      }
      for (auto &item : post_json["urls"]) {
        Url url;
        url.shortened_url = item["shortened_url"];
        url.expanded_url = item["expanded_url"];
        new_post.urls.emplace_back(url);
      }
      post_json_map.insert({new_post.post_id, std::string(post_json_char)});
      return_map.insert({new_post.post_id, new_post});
      bson_free(post_json_char);
    }
    find_span->Finish();
    bson_error_t error;
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
    }
    bson_destroy(query);
    mongoc_cursor_destroy(cursor);
    mongoc_collection_destroy(collection);
    mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);

    // upload posts to memcached
    set_futures.emplace_back(std::async(std::launch::async, [&]() {
      memcached_return_t _rc;
      auto _memcached_client =
          memcached_pool_pop(_memcached_client_pool, true, &_rc);
      if (!_memcached_client) {
        LOG(error) << "Failed to pop a client from memcached pool";
        ServiceException se;
        se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
        se.message = "Failed to pop a client from memcached pool";
        throw se;
      }
      auto set_span = opentracing::Tracer::Global()->StartSpan(
          "mmc_set_client", {opentracing::ChildOf(&span->context())});
      for (auto &it : post_json_map) {
        std::string id_str = std::to_string(it.first);
        _rc = memcached_set(_memcached_client, id_str.c_str(), id_str.length(),
                            it.second.c_str(), it.second.length(),
                            static_cast<time_t>(0), static_cast<uint32_t>(0));
      }
      memcached_pool_push(_memcached_client_pool, _memcached_client);
      set_span->Finish();
    }));
  }

  if (return_map.size() != post_ids.size()) {
    try {
      for (auto &it : set_futures) {
        it.get();
      }
    } catch (...) {
      LOG(warning) << "Failed to set posts to memcached";
    }
    LOG(error) << "Return set incomplete";
    ServiceException se;
    se.errorCode = ErrorCode::SE_THRIFT_HANDLER_ERROR;
    se.message = "Return set incomplete";
    throw se;
  }

  for (auto &post_id : post_ids) {
    _return.emplace_back(return_map[post_id]);
  }

  try {
    for (auto &it : set_futures) {
      it.get();
    }
  } catch (...) {
    LOG(warning) << "Failed to set posts to memcached";
  }
}

}  // namespace social_network

#endif  // SOCIAL_NETWORK_MICROSERVICES_POSTSTORAGEHANDLER_H
