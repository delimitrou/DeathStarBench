#ifndef SOCIAL_NETWORK_MICROSERVICES_SRC_USERTIMELINESERVICE_USERTIMELINEHANDLER_H_
#define SOCIAL_NETWORK_MICROSERVICES_SRC_USERTIMELINESERVICE_USERTIMELINEHANDLER_H_

#include <iostream>
#include <string>

#include <mongoc.h>
#include <bson/bson.h>

#include "../../gen-cpp/UserTimelineService.h"
#include "../../gen-cpp/PostStorageService.h"
#include "../logger.h"
#include "../tracing.h"
#include "../ClientPool.h"
#include "../RedisClient.h"
#include "../ThriftClient.h"

namespace social_network {

class UserTimelineHandler : public UserTimelineServiceIf {
 public:
  UserTimelineHandler(
      ClientPool<RedisClient> *,
      mongoc_client_pool_t *,
      ClientPool<ThriftClient<PostStorageServiceClient>> *);
  ~UserTimelineHandler() override = default;

  void WriteUserTimeline(int64_t req_id, int64_t post_id, int64_t user_id,
      int64_t timestamp, const std::map<std::string, std::string> &carrier)
      override;

  void ReadUserTimeline(std::vector<Post> &, int64_t, int64_t, int, int,
                        const std::map<std::string, std::string> &) override ;

 private:
  ClientPool<RedisClient> *_redis_client_pool;
  mongoc_client_pool_t *_mongodb_client_pool;
  ClientPool<ThriftClient<PostStorageServiceClient>> *_post_client_pool;
};

UserTimelineHandler::UserTimelineHandler(
    ClientPool<RedisClient> *redis_pool,
    mongoc_client_pool_t *mongodb_pool,
    ClientPool<ThriftClient<PostStorageServiceClient>> *post_client_pool) {
  _redis_client_pool = redis_pool;
  _mongodb_client_pool = mongodb_pool;
  _post_client_pool = post_client_pool;
}

void UserTimelineHandler::WriteUserTimeline(
    int64_t req_id,
    int64_t post_id,
    int64_t user_id,
    int64_t timestamp,
    const std::map<std::string, std::string> &carrier) {

  // Initialize a span
  TextMapReader reader(carrier);
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "WriteUserTimeline",
      { opentracing::ChildOf(parent_span->get()) });
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  mongoc_client_t *mongodb_client = mongoc_client_pool_pop(
      _mongodb_client_pool);
  if (!mongodb_client) {
    ServiceException se;
    se.errorCode = ErrorCode::SE_MONGODB_ERROR;
    se.message = "Failed to pop a client from MongoDB pool";
    throw se;
  }
  auto collection = mongoc_client_get_collection(
      mongodb_client, "user-timeline", "user-timeline");
  if (!collection) {
    ServiceException se;
    se.errorCode = ErrorCode::SE_MONGODB_ERROR;
    se.message = "Failed to create collection user-timeline from MongoDB";
    mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
    throw se;
  }
  bson_t *query = bson_new();

  BSON_APPEND_INT64(query, "user_id", user_id);

  bson_t *update = BCON_NEW(
      "$push", "{",
          "posts", "{",
              "$each", "[", "{",
                  "post_id", BCON_INT64(post_id),
                  "timestamp", BCON_INT64(timestamp),
              "}", "]",
              "$position", BCON_INT32(0),
          "}",
      "}"
  );
  bson_error_t error;
  bson_t reply;
  auto update_span = opentracing::Tracer::Global()->StartSpan(
      "MongoInsert", {opentracing::ChildOf(&span->context())});
  // If no document matches the query criteria,
  // insert a single document (upsert: true)
  bool updated = mongoc_collection_find_and_modify(
      collection, query, nullptr, update, nullptr, false, true,
      true, &reply, &error);
  update_span->Finish();

  if (!updated) {
    // update the newly inserted document (upsert: false)
    updated = mongoc_collection_find_and_modify(
      collection, query, nullptr, update, nullptr, false, false,
      true, &reply, &error);
    if (!updated) {
      LOG(error) << "Failed to update user-timeline for user " << user_id
                  << " to MongoDB: " << error.message;
      ServiceException se;
      se.errorCode = ErrorCode::SE_MONGODB_ERROR;
      se.message = error.message;
      bson_destroy(update);
      bson_destroy(query);
      bson_destroy(&reply);
      mongoc_collection_destroy(collection);
      mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
      throw se;
    }
  }

  bson_destroy(update);
  bson_destroy(&reply);
  bson_destroy(query);
  mongoc_collection_destroy(collection);
  mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);

  auto redis_client_wrapper = _redis_client_pool->Pop();
  if (!redis_client_wrapper) {
    ServiceException se;
    se.errorCode = ErrorCode::SE_REDIS_ERROR;
    se.message = "Cannot connect to Redis server";
    throw se;
  }
  auto redis_client = redis_client_wrapper->GetClient();
  auto redis_span = opentracing::Tracer::Global()->StartSpan(
      "RedisUpdate", {opentracing::ChildOf(&span->context())});
  auto num_posts = redis_client->zcard(std::to_string(user_id));
  redis_client->sync_commit();
  auto num_posts_reply = num_posts.get();
  std::vector<std::string> options{"NX"};
  if (num_posts_reply.ok() && num_posts_reply.as_integer()) {
    std::string key = std::to_string(user_id);
    std::multimap<std::string, std::string> value = {{
      std::to_string(timestamp), std::to_string(post_id)}};
    redis_client->zadd(key, options, value);
    redis_client->sync_commit();
  }
  _redis_client_pool->Push(redis_client_wrapper);
  redis_span->Finish();
  span->Finish();

}
void UserTimelineHandler::ReadUserTimeline(
    std::vector<Post> &_return,
    int64_t req_id,
    int64_t user_id,
    int start,
    int stop,
    const std::map<std::string, std::string> &carrier) {

  // Initialize a span
  TextMapReader reader(carrier);
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "ReadUserTimeline",
      { opentracing::ChildOf(parent_span->get()) });
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  if (stop <= start || start < 0) {
    return;
  }

  auto redis_client_wrapper = _redis_client_pool->Pop();
  if (!redis_client_wrapper) {
    ServiceException se;
    se.errorCode = ErrorCode::SE_REDIS_ERROR;
    se.message = "Cannot connect to Redis server";
    throw se;
  }
  auto redis_client = redis_client_wrapper->GetClient();
  auto redis_span = opentracing::Tracer::Global()->StartSpan(
      "RedisFind", {opentracing::ChildOf(&span->context())});
  auto post_ids_future = redis_client->zrevrange(
      std::to_string(user_id), start, stop - 1);
  redis_client->commit();
  redis_span->Finish();

  cpp_redis::reply post_ids_reply;
  try {
    post_ids_reply = post_ids_future.get();
  } catch (...) {
    LOG(error) << "Failed to read post_ids from user-timeline-redis";
    _redis_client_pool->Push(redis_client_wrapper);
    throw;
  }
  _redis_client_pool->Push(redis_client_wrapper);
  
  std::vector<int64_t> post_ids;
  auto post_ids_reply_array = post_ids_reply.as_array();
  for (auto &post_id_reply : post_ids_reply_array) {
    post_ids.emplace_back(std::stoul(post_id_reply.as_string()));
  }

  int mongo_start = start + post_ids.size();
  std::multimap<std::string, std::string> redis_update_map;
  if (mongo_start < stop) {
    // Instead find post_ids from mongodb
    mongoc_client_t *mongodb_client = mongoc_client_pool_pop(
        _mongodb_client_pool);
    if (!mongodb_client) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_MONGODB_ERROR;
      se.message = "Failed to pop a client from MongoDB pool";
      throw se;
    }
    auto collection = mongoc_client_get_collection(
        mongodb_client, "user-timeline", "user-timeline");
    if (!collection) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_MONGODB_ERROR;
      se.message = "Failed to create collection user-timeline from MongoDB";
      throw se;
    }

    bson_t *query = BCON_NEW("user_id", BCON_INT64(user_id));
    bson_t *opts = BCON_NEW(
        "projection", "{",
            "posts", "{",
                "$slice", "[",
                    BCON_INT32(0), BCON_INT32(stop),
                "]",
            "}",
        "}");

    auto find_span = opentracing::Tracer::Global()->StartSpan(
        "MongoFindUserTimeline", { opentracing::ChildOf(&span->context()) });
    mongoc_cursor_t *cursor = mongoc_collection_find_with_opts(
        collection, query, opts, nullptr);
    find_span->Finish();
    const bson_t *doc;
    bool found = mongoc_cursor_next(cursor, &doc);
    if (found) {
      bson_iter_t iter_0;
      bson_iter_t iter_1;
      bson_iter_t post_id_child;
      bson_iter_t timestamp_child;
      int idx = 0;
      bson_iter_init(&iter_0, doc);
      bson_iter_init(&iter_1, doc);
      while (bson_iter_find_descendant(&iter_0,
              ("posts." + std::to_string(idx) + ".post_id").c_str(),
              &post_id_child)
          && BSON_ITER_HOLDS_INT64(&post_id_child)
          && bson_iter_find_descendant(&iter_1,
              ("posts." + std::to_string(idx) + ".timestamp").c_str(),
              &timestamp_child)
          && BSON_ITER_HOLDS_INT64(&timestamp_child)) {
        auto curr_post_id = bson_iter_int64(&post_id_child);
        auto curr_timestamp = bson_iter_int64(&timestamp_child);
        if (idx >= mongo_start) {
          post_ids.emplace_back(curr_post_id);
        }
        redis_update_map.insert(std::make_pair(std::to_string(curr_timestamp),
            std::to_string(curr_post_id)));
        bson_iter_init(&iter_0, doc);
        bson_iter_init(&iter_1, doc);
        idx++;
      }
    }
    bson_destroy(opts);
    bson_destroy(query);
    mongoc_cursor_destroy(cursor);
    mongoc_collection_destroy(collection);
    mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
  }

  std::future<std::vector<Post>> post_future = std::async(
      std::launch::async, [&]() {
        auto post_client_wrapper = _post_client_pool->Pop();
        if (!post_client_wrapper) {
          ServiceException se;
          se.errorCode = ErrorCode::SE_THRIFT_CONN_ERROR;
          se.message = "Failed to connect to post-storage-service";
          throw se;
        }
        std::vector<Post> _return_posts;
          auto post_client = post_client_wrapper->GetClient();
          try {
            post_client->ReadPosts(
                _return_posts, req_id, post_ids, writer_text_map);
          } catch (...) {
            _post_client_pool->Push(post_client_wrapper);
            LOG(error) << "Failed to read posts from post-storage-service";
            throw;
          }
          _post_client_pool->Push(post_client_wrapper);
          return _return_posts;
      });

  std::future<cpp_redis::reply> zadd_reply_future;
  if (!redis_update_map.empty()) {
    // Update Redis
    redis_client_wrapper = _redis_client_pool->Pop();
    if (!redis_client_wrapper) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_REDIS_ERROR;
      se.message = "Cannot connect to Redis server";
      throw se;
    }
    redis_client = redis_client_wrapper->GetClient();
    auto redis_update_span = opentracing::Tracer::Global()->StartSpan(
        "RedisUpdate", {opentracing::ChildOf(&span->context())});
    std::string user_id_str = std::to_string(user_id);
    redis_client->del(std::vector<std::string>{user_id_str});
    std::vector<std::string> options{"NX"};
    zadd_reply_future = redis_client->zadd(
        user_id_str, options, redis_update_map);
    redis_client->commit();
    redis_update_span->Finish();
  }

  try {
    _return = post_future.get();
  } catch (...) {
    LOG(error) << "Failed to get post from post-storage-service";
    if (!redis_update_map.empty()) {
      try {
        zadd_reply_future.get();
      } catch (...) {
        LOG(error) << "Failed to Update Redis Server";
      }
      _redis_client_pool->Push(redis_client_wrapper);
    }
    throw;
  }


  if (!redis_update_map.empty()) {
    try {
      zadd_reply_future.get();
    } catch (...) {
      LOG(error) << "Failed to Update Redis Server";
      _redis_client_pool->Push(redis_client_wrapper);
      throw;
    }
    _redis_client_pool->Push(redis_client_wrapper);
  }

  span->Finish();

}

} // namespace social_network


#endif //SOCIAL_NETWORK_MICROSERVICES_SRC_USERTIMELINESERVICE_USERTIMELINEHANDLER_H_
