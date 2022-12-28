#ifndef SOCIAL_NETWORK_MICROSERVICES_SOCIALGRAPHHANDLER_H
#define SOCIAL_NETWORK_MICROSERVICES_SOCIALGRAPHHANDLER_H

#include <bson/bson.h>
#include <mongoc.h>
#include <sw/redis++/redis++.h>

#include <chrono>
#include <future>
#include <iostream>
#include <string>
#include <thread>
#include <vector>

#include "../../gen-cpp/SocialGraphService.h"
#include "../../gen-cpp/UserService.h"
#include "../ClientPool.h"
#include "../ThriftClient.h"
#include "../logger.h"
#include "../tracing.h"

using namespace sw::redis;

namespace social_network {

using std::chrono::duration_cast;
using std::chrono::milliseconds;
using std::chrono::system_clock;

class SocialGraphHandler : public SocialGraphServiceIf {
 public:
  SocialGraphHandler(mongoc_client_pool_t *, Redis *,
                     ClientPool<ThriftClient<UserServiceClient>> *);
  SocialGraphHandler(mongoc_client_pool_t *, Redis *, Redis *,
      ClientPool<ThriftClient<UserServiceClient>>*);
  SocialGraphHandler(mongoc_client_pool_t *, RedisCluster *,
                     ClientPool<ThriftClient<UserServiceClient>> *);
  ~SocialGraphHandler() override = default;
  bool IsRedisReplicationEnabled();
  void GetFollowers(std::vector<int64_t> &, int64_t, int64_t,
                    const std::map<std::string, std::string> &) override;
  void GetFollowees(std::vector<int64_t> &, int64_t, int64_t,
                    const std::map<std::string, std::string> &) override;
  void Follow(int64_t, int64_t, int64_t,
              const std::map<std::string, std::string> &) override;
  void Unfollow(int64_t, int64_t, int64_t,
                const std::map<std::string, std::string> &) override;
  void FollowWithUsername(int64_t, const std::string &, const std::string &,
                          const std::map<std::string, std::string> &) override;
  void UnfollowWithUsername(
      int64_t, const std::string &, const std::string &,
      const std::map<std::string, std::string> &) override;
  void InsertUser(int64_t, int64_t,
                  const std::map<std::string, std::string> &) override;

 private:
  mongoc_client_pool_t *_mongodb_client_pool;
  Redis *_redis_client_pool;
  Redis *_redis_replica_client_pool;
  Redis *_redis_primary_client_pool;
  RedisCluster *_redis_cluster_client_pool;
  ClientPool<ThriftClient<UserServiceClient>> *_user_service_client_pool;
};

SocialGraphHandler::SocialGraphHandler(
    mongoc_client_pool_t *mongodb_client_pool, Redis *redis_client_pool,
    ClientPool<ThriftClient<UserServiceClient>> *user_service_client_pool) {
  _mongodb_client_pool = mongodb_client_pool;
  _redis_client_pool = redis_client_pool;
  _redis_replica_client_pool = nullptr;
  _redis_primary_client_pool = nullptr;
  _redis_cluster_client_pool = nullptr;
  _user_service_client_pool = user_service_client_pool;
}

SocialGraphHandler::SocialGraphHandler(
    mongoc_client_pool_t* mongodb_client_pool, Redis* redis_replica_client_pool, Redis* redis_primary_client_pool,
    ClientPool<ThriftClient<UserServiceClient>>* user_service_client_pool) {
    _mongodb_client_pool = mongodb_client_pool;
    _redis_client_pool = nullptr;
    _redis_replica_client_pool = redis_replica_client_pool;
    _redis_primary_client_pool = redis_primary_client_pool;
    _redis_cluster_client_pool = nullptr;
    _user_service_client_pool = user_service_client_pool;
}

SocialGraphHandler::SocialGraphHandler(
    mongoc_client_pool_t *mongodb_client_pool,
    RedisCluster *redis_cluster_client_pool,
    ClientPool<ThriftClient<UserServiceClient>> *user_service_client_pool) {
  _mongodb_client_pool = mongodb_client_pool;
  _redis_client_pool = nullptr;
  _redis_replica_client_pool = nullptr;
  _redis_primary_client_pool = nullptr;
  _redis_cluster_client_pool = redis_cluster_client_pool;
  _user_service_client_pool = user_service_client_pool;
}

bool SocialGraphHandler::IsRedisReplicationEnabled() {
    return (_redis_primary_client_pool || _redis_replica_client_pool);
}

void SocialGraphHandler::Follow(
    int64_t req_id, int64_t user_id, int64_t followee_id,
    const std::map<std::string, std::string> &carrier) {
  // Initialize a span
  TextMapReader reader(carrier);
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "follow_server", {opentracing::ChildOf(parent_span->get())});
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  int64_t timestamp =
      duration_cast<milliseconds>(system_clock::now().time_since_epoch())
          .count();

  std::future<void> mongo_update_follower_future =
      std::async(std::launch::async, [&]() {
        mongoc_client_t *mongodb_client =
            mongoc_client_pool_pop(_mongodb_client_pool);
        if (!mongodb_client) {
          ServiceException se;
          se.errorCode = ErrorCode::SE_MONGODB_ERROR;
          se.message = "Failed to pop a client from MongoDB pool";
          throw se;
        }
        auto collection = mongoc_client_get_collection(
            mongodb_client, "social-graph", "social-graph");
        if (!collection) {
          ServiceException se;
          se.errorCode = ErrorCode::SE_MONGODB_ERROR;
          se.message = "Failed to create collection social_graph from MongoDB";
          mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
          throw se;
        }

        // Update follower->followee edges
        const bson_t *doc;
        bson_t *search_not_exist = BCON_NEW(
            "$and", "[", "{", "user_id", BCON_INT64(user_id), "}", "{",
            "followees", "{", "$not", "{", "$elemMatch", "{", "user_id",
            BCON_INT64(followee_id), "}", "}", "}", "}", "]");
        bson_t *update = BCON_NEW("$push", "{", "followees", "{", "user_id",
                                  BCON_INT64(followee_id), "timestamp",
                                  BCON_INT64(timestamp), "}", "}");
        bson_error_t error;
        bson_t reply;
        auto update_span = opentracing::Tracer::Global()->StartSpan(
            "mongo_update_client", {opentracing::ChildOf(&span->context())});
        bool updated = mongoc_collection_find_and_modify(
            collection, search_not_exist, nullptr, update, nullptr, false,
            false, true, &reply, &error);
        if (!updated) {
          LOG(error) << "Failed to update social graph for user " << user_id
                     << " to MongoDB: " << error.message;
          ServiceException se;
          se.errorCode = ErrorCode::SE_MONGODB_ERROR;
          se.message = error.message;
          bson_destroy(&reply);
          bson_destroy(update);
          bson_destroy(search_not_exist);
          mongoc_collection_destroy(collection);
          mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
          throw se;
        }
        update_span->Finish();
        bson_destroy(&reply);
        bson_destroy(update);
        bson_destroy(search_not_exist);
        mongoc_collection_destroy(collection);
        mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
      });

  std::future<void> mongo_update_followee_future =
      std::async(std::launch::async, [&]() {
        mongoc_client_t *mongodb_client =
            mongoc_client_pool_pop(_mongodb_client_pool);
        if (!mongodb_client) {
          ServiceException se;
          se.errorCode = ErrorCode::SE_MONGODB_ERROR;
          se.message = "Failed to pop a client from MongoDB pool";
          throw se;
        }
        auto collection = mongoc_client_get_collection(
            mongodb_client, "social-graph", "social-graph");
        if (!collection) {
          ServiceException se;
          se.errorCode = ErrorCode::SE_MONGODB_ERROR;
          se.message = "Failed to create collection social_graph from MongoDB";
          mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
          throw se;
        }

        // Update followee->follower edges
        bson_t *search_not_exist =
            BCON_NEW("$and", "[", "{", "user_id", BCON_INT64(followee_id), "}",
                     "{", "followers", "{", "$not", "{", "$elemMatch", "{",
                     "user_id", BCON_INT64(user_id), "}", "}", "}", "}", "]");
        bson_t *update = BCON_NEW("$push", "{", "followers", "{", "user_id",
                                  BCON_INT64(user_id), "timestamp",
                                  BCON_INT64(timestamp), "}", "}");
        bson_error_t error;
        auto update_span = opentracing::Tracer::Global()->StartSpan(
            "social_graph_mongo_update_client",
            {opentracing::ChildOf(&span->context())});
        bson_t reply;
        bool updated = mongoc_collection_find_and_modify(
            collection, search_not_exist, nullptr, update, nullptr, false,
            false, true, &reply, &error);
        if (!updated) {
          LOG(error) << "Failed to update social graph for user " << followee_id
                     << " to MongoDB: " << error.message;
          ServiceException se;
          se.errorCode = ErrorCode::SE_MONGODB_ERROR;
          se.message = error.message;
          bson_destroy(update);
          bson_destroy(&reply);
          bson_destroy(search_not_exist);
          mongoc_collection_destroy(collection);
          mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
          throw se;
        }
        update_span->Finish();
        bson_destroy(update);
        bson_destroy(&reply);
        bson_destroy(search_not_exist);
        mongoc_collection_destroy(collection);
        mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
      });

  std::future<void> redis_update_future = std::async(std::launch::async, [&]() {
    auto redis_span = opentracing::Tracer::Global()->StartSpan(
        "social_graph_redis_update_client",
        {opentracing::ChildOf(&span->context())});

    {
      if (_redis_client_pool) {
        auto pipe = _redis_client_pool->pipeline(false);
        pipe.zadd(std::to_string(user_id) + ":followees",
                  std::to_string(followee_id), timestamp, UpdateType::NOT_EXIST)
            .zadd(std::to_string(followee_id) + ":followers",
                  std::to_string(user_id), timestamp, UpdateType::NOT_EXIST);
        try {
          auto replies = pipe.exec();
        } catch (const Error &err) {
          LOG(error) << err.what();
          throw err;
        }
      }
      else if (IsRedisReplicationEnabled()) {
          auto pipe = _redis_primary_client_pool->pipeline(false);
          pipe.zadd(std::to_string(user_id) + ":followees",
              std::to_string(followee_id), timestamp, UpdateType::NOT_EXIST)
              .zadd(std::to_string(followee_id) + ":followers",
                  std::to_string(user_id), timestamp, UpdateType::NOT_EXIST);
          try {
              auto replies = pipe.exec();
          }
          catch (const Error& err) {
              LOG(error) << err.what();
              throw err;
          }
      }
      else {
        // TODO: Redis++ currently does not support pipeline with multiple
        //       hashtags in cluster mode.
        //       Currently, we send one request for each follower, which may
        //       incur some performance overhead. We are following the updates
        //       of Redis++ clients:
        //       https://github.com/sewenew/redis-plus-plus/issues/212
        try {
          _redis_cluster_client_pool->zadd(
              std::to_string(user_id) + ":followees",
              std::to_string(followee_id), timestamp, UpdateType::NOT_EXIST);
          _redis_cluster_client_pool->zadd(
              std::to_string(followee_id) + ":followers",
              std::to_string(user_id), timestamp, UpdateType::NOT_EXIST);
        } catch (const Error &err) {
          LOG(error) << err.what();
          throw err;
        }
      }
    }
    redis_span->Finish();
  });

  try {
    redis_update_future.get();
    mongo_update_follower_future.get();
    mongo_update_followee_future.get();
  } catch (const std::exception &e) {
    LOG(warning) << e.what();
    throw;
  } catch (...) {
    throw;
  }

  span->Finish();
}

void SocialGraphHandler::Unfollow(
    int64_t req_id, int64_t user_id, int64_t followee_id,
    const std::map<std::string, std::string> &carrier) {
  // Initialize a span
  TextMapReader reader(carrier);
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "unfollow_server", {opentracing::ChildOf(parent_span->get())});
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  std::future<void> mongo_update_follower_future =
      std::async(std::launch::async, [&]() {
        mongoc_client_t *mongodb_client =
            mongoc_client_pool_pop(_mongodb_client_pool);
        if (!mongodb_client) {
          ServiceException se;
          se.errorCode = ErrorCode::SE_MONGODB_ERROR;
          se.message = "Failed to pop a client from MongoDB pool";
          throw se;
        }
        auto collection = mongoc_client_get_collection(
            mongodb_client, "social-graph", "social-graph");
        if (!collection) {
          ServiceException se;
          se.errorCode = ErrorCode::SE_MONGODB_ERROR;
          se.message = "Failed to create collection social_graph from MongoDB";
          mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
          throw se;
        }
        bson_t *query = bson_new();

        // Update follower->followee edges
        BSON_APPEND_INT64(query, "user_id", user_id);
        bson_t *update = BCON_NEW("$pull", "{", "followees", "{", "user_id",
                                  BCON_INT64(followee_id), "}", "}");
        bson_t reply;
        bson_error_t error;
        auto update_span = opentracing::Tracer::Global()->StartSpan(
            "social_graph_mongo_delete_client",
            {opentracing::ChildOf(&span->context())});
        bool updated = mongoc_collection_find_and_modify(
            collection, query, nullptr, update, nullptr, false, false, true,
            &reply, &error);
        if (!updated) {
          LOG(error) << "Failed to delete social graph for user " << user_id
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
        update_span->Finish();
        bson_destroy(update);
        bson_destroy(query);
        bson_destroy(&reply);
        mongoc_collection_destroy(collection);
        mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
      });

  std::future<void> mongo_update_followee_future =
      std::async(std::launch::async, [&]() {
        mongoc_client_t *mongodb_client =
            mongoc_client_pool_pop(_mongodb_client_pool);
        if (!mongodb_client) {
          ServiceException se;
          se.errorCode = ErrorCode::SE_MONGODB_ERROR;
          se.message = "Failed to pop a client from MongoDB pool";
          throw se;
        }
        auto collection = mongoc_client_get_collection(
            mongodb_client, "social-graph", "social-graph");
        if (!collection) {
          ServiceException se;
          se.errorCode = ErrorCode::SE_MONGODB_ERROR;
          se.message = "Failed to create collection social_graph from MongoDB";
          mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
          throw se;
        }
        bson_t *query = bson_new();

        // Update followee->follower edges
        BSON_APPEND_INT64(query, "user_id", followee_id);
        bson_t *update = BCON_NEW("$pull", "{", "followers", "{", "user_id",
                                  BCON_INT64(user_id), "}", "}");
        bson_t reply;
        bson_error_t error;
        auto update_span = opentracing::Tracer::Global()->StartSpan(
            "social_graph_mongo_delete_client",
            {opentracing::ChildOf(&span->context())});
        bool updated = mongoc_collection_find_and_modify(
            collection, query, nullptr, update, nullptr, false, false, true,
            &reply, &error);
        if (!updated) {
          LOG(error) << "Failed to delete social graph for user " << followee_id
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
        update_span->Finish();
        bson_destroy(update);
        bson_destroy(query);
        bson_destroy(&reply);
        mongoc_collection_destroy(collection);
        mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
      });

  std::future<void> redis_update_future = std::async(std::launch::async, [&]() {
    auto redis_span = opentracing::Tracer::Global()->StartSpan(
        "social_graph_redis_update_client",
        {opentracing::ChildOf(&span->context())});
    {
      if (_redis_client_pool) {
        auto pipe = _redis_client_pool->pipeline(false);
        std::string followee_key = std::to_string(user_id) + ":followees";
        std::string follower_key = std::to_string(followee_id) + ":followers";
        pipe.zrem(followee_key, std::to_string(followee_id))
            .zrem(follower_key, std::to_string(user_id));

        try {
          auto replies = pipe.exec();
        } catch (const Error &err) {
          LOG(error) << err.what();
          throw err;
        }
      } 
      else if (IsRedisReplicationEnabled()) {
          auto pipe = _redis_primary_client_pool->pipeline(false);
          std::string followee_key = std::to_string(user_id) + ":followees";
          std::string follower_key = std::to_string(followee_id) + ":followers";
          pipe.zrem(followee_key, std::to_string(followee_id))
              .zrem(follower_key, std::to_string(user_id));

          try {
              auto replies = pipe.exec();
          }
          catch (const Error& err) {
              LOG(error) << err.what();
              throw err;
          }
      }
      else {
        std::string followee_key = std::to_string(user_id) + ":followees";
        std::string follower_key = std::to_string(followee_id) + ":followers";
        try {
          _redis_cluster_client_pool->zrem(followee_key,
                                           std::to_string(followee_id));
          _redis_cluster_client_pool->zrem(follower_key,
                                           std::to_string(user_id));
        } catch (const Error &err) {
          LOG(error) << err.what();
          throw err;
        }
      }
    }
    redis_span->Finish();
  });

  try {
    redis_update_future.get();
    mongo_update_follower_future.get();
    mongo_update_followee_future.get();
  } catch (...) {
    throw;
  }

  span->Finish();
}

void SocialGraphHandler::GetFollowers(
    std::vector<int64_t> &_return, const int64_t req_id, const int64_t user_id,
    const std::map<std::string, std::string> &carrier) {
  // Initialize a span
  TextMapReader reader(carrier);
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "get_followers_server", {opentracing::ChildOf(parent_span->get())});
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  auto redis_span = opentracing::Tracer::Global()->StartSpan(
      "social_graph_redis_get_client",
      {opentracing::ChildOf(&span->context())});

  std::vector<std::string> followers_str;
  std::string key = std::to_string(user_id) + ":followers";
  try {
    if (_redis_client_pool) {
      _redis_client_pool->zrange(key, 0, -1, std::back_inserter(followers_str));
    } 
    else if (IsRedisReplicationEnabled()) {
        _redis_replica_client_pool->zrange(key, 0, -1, std::back_inserter(followers_str));
    }
    else {
      _redis_cluster_client_pool->zrange(key, 0, -1,
                                         std::back_inserter(followers_str));
    }
  } catch (const Error &err) {
    LOG(error) << err.what();
    throw err;
  }
  redis_span->Finish();

  // If user_id in the sodical graph Redis server, read from Redis
  if (followers_str.size() > 0) {
    for (auto const &follower_str : followers_str) {
      _return.emplace_back(std::stoul(follower_str));
    }
  }
  // If user_id in the sodical graph Redis server, read from MongoDB and
  // update Redis.
  else {
    mongoc_client_t *mongodb_client =
        mongoc_client_pool_pop(_mongodb_client_pool);
    if (!mongodb_client) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_MONGODB_ERROR;
      se.message = "Failed to pop a client from MongoDB pool";
      throw se;
    }
    auto collection = mongoc_client_get_collection(
        mongodb_client, "social-graph", "social-graph");
    if (!collection) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_MONGODB_ERROR;
      se.message = "Failed to create collection social_graph from MongoDB";
      mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
      throw se;
    }
    bson_t *query = bson_new();
    BSON_APPEND_INT64(query, "user_id", user_id);
    auto find_span = opentracing::Tracer::Global()->StartSpan(
        "social_graph_mongo_find_client",
        {opentracing::ChildOf(&span->context())});
    mongoc_cursor_t *cursor =
        mongoc_collection_find_with_opts(collection, query, nullptr, nullptr);
    const bson_t *doc;
    bool found = mongoc_cursor_next(cursor, &doc);
    if (found) {
      bson_iter_t iter_0;
      bson_iter_t iter_1;
      bson_iter_t user_id_child;
      bson_iter_t timestamp_child;
      int index = 0;
      std::unordered_map<std::string, double> redis_zset;
      bson_iter_init(&iter_0, doc);
      bson_iter_init(&iter_1, doc);

      while (bson_iter_find_descendant(
                 &iter_0,
                 ("followers." + std::to_string(index) + ".user_id").c_str(),
                 &user_id_child) &&
             BSON_ITER_HOLDS_INT64(&user_id_child) &&
             bson_iter_find_descendant(
                 &iter_1,
                 ("followers." + std::to_string(index) + ".timestamp").c_str(),
                 &timestamp_child) &&
             BSON_ITER_HOLDS_INT64(&timestamp_child)) {
        auto iter_user_id = bson_iter_int64(&user_id_child);
        auto iter_timestamp = bson_iter_int64(&timestamp_child);
        _return.emplace_back(iter_user_id);
        redis_zset.emplace(std::pair<std::string, double>(
            std::to_string(iter_user_id), (double)iter_timestamp));
        bson_iter_init(&iter_0, doc);
        bson_iter_init(&iter_1, doc);
        index++;
      }
      find_span->Finish();
      bson_destroy(query);
      mongoc_cursor_destroy(cursor);
      mongoc_collection_destroy(collection);
      mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);

      // Update Redis
      std::string key = std::to_string(user_id) + ":followers";
      auto redis_insert_span = opentracing::Tracer::Global()->StartSpan(
          "social_graph_redis_insert_client",
          {opentracing::ChildOf(&span->context())});
      try {
        if (_redis_client_pool) {
          _redis_client_pool->zadd(key, redis_zset.begin(), redis_zset.end());
        } 
        else if (IsRedisReplicationEnabled()) {
            _redis_primary_client_pool->zadd(key, redis_zset.begin(), redis_zset.end());
        }
        else {
          _redis_cluster_client_pool->zadd(key, redis_zset.begin(),
                                           redis_zset.end());
        }
      } catch (const Error &err) {
        LOG(error) << err.what();
        throw err;
      }
      redis_span->Finish();
    } else {
      LOG(warning) << "user_id: " << user_id << " not found";
      find_span->Finish();
      bson_destroy(query);
      mongoc_cursor_destroy(cursor);
      mongoc_collection_destroy(collection);
      mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
    }
  }
  span->Finish();
}

void SocialGraphHandler::GetFollowees(
    std::vector<int64_t> &_return, const int64_t req_id, const int64_t user_id,
    const std::map<std::string, std::string> &carrier) {
  // Initialize a span
  TextMapReader reader(carrier);
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "get_followees_server", {opentracing::ChildOf(parent_span->get())});
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  auto redis_span = opentracing::Tracer::Global()->StartSpan(
      "social_graph_redis_get_client",
      {opentracing::ChildOf(&span->context())});

  std::vector<std::string> followees_str;
  std::string key = std::to_string(user_id) + ":followees";
  try {
    if (_redis_client_pool) {
      _redis_client_pool->zrange(key, 0, -1, std::back_inserter(followees_str));
    }
    else if (IsRedisReplicationEnabled()) {
        _redis_replica_client_pool->zrange(key, 0, -1, std::back_inserter(followees_str));
    }
    else {
      _redis_cluster_client_pool->zrange(key, 0, -1,
                                         std::back_inserter(followees_str));
    }
  } catch (const Error &err) {
    LOG(error) << err.what();
    throw err;
  }
  redis_span->Finish();

  // If user_id in the sodical graph Redis server, read from Redis
  if (followees_str.size() > 0) {
    for (auto const &followee_str : followees_str) {
      _return.emplace_back(std::stoul(followee_str));
    }
  }
  // If user_id in the sodical graph Redis server, read from MongoDB and
  // update Redis.
  else {
    redis_span->Finish();
    mongoc_client_t *mongodb_client =
        mongoc_client_pool_pop(_mongodb_client_pool);
    if (!mongodb_client) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_MONGODB_ERROR;
      se.message = "Failed to pop a client from MongoDB pool";
      throw se;
    }
    auto collection = mongoc_client_get_collection(
        mongodb_client, "social-graph", "social-graph");
    if (!collection) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_MONGODB_ERROR;
      se.message = "Failed to create collection social_graph from MongoDB";
      mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
      throw se;
    }
    bson_t *query = bson_new();
    BSON_APPEND_INT64(query, "user_id", user_id);
    auto find_span = opentracing::Tracer::Global()->StartSpan(
        "social_graph_mongo_find_client",
        {opentracing::ChildOf(&span->context())});
    mongoc_cursor_t *cursor =
        mongoc_collection_find_with_opts(collection, query, nullptr, nullptr);
    const bson_t *doc;
    bool found = mongoc_cursor_next(cursor, &doc);
    if (!found) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_THRIFT_HANDLER_ERROR;
      se.message = "Cannot find user_id in MongoDB.";
      bson_destroy(query);
      mongoc_cursor_destroy(cursor);
      mongoc_collection_destroy(collection);
      mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
      throw se;
    } else {
      bson_iter_t iter_0;
      bson_iter_t iter_1;
      bson_iter_t user_id_child;
      bson_iter_t timestamp_child;
      int index = 0;

      bson_iter_init(&iter_0, doc);
      bson_iter_init(&iter_1, doc);

      std::multimap<std::string, double> redis_zset;

      while (bson_iter_find_descendant(
                 &iter_0,
                 ("followees." + std::to_string(index) + ".user_id").c_str(),
                 &user_id_child) &&
             BSON_ITER_HOLDS_INT64(&user_id_child) &&
             bson_iter_find_descendant(
                 &iter_1,
                 ("followees." + std::to_string(index) + ".timestamp").c_str(),
                 &timestamp_child) &&
             BSON_ITER_HOLDS_INT64(&timestamp_child)) {
        auto iter_user_id = bson_iter_int64(&user_id_child);
        auto iter_timestamp = bson_iter_int64(&timestamp_child);
        _return.emplace_back(iter_user_id);

        redis_zset.emplace(std::pair<std::string, double>(
            std::to_string(iter_user_id), (double)iter_timestamp));
        bson_iter_init(&iter_0, doc);
        bson_iter_init(&iter_1, doc);
        index++;
      }

      find_span->Finish();
      bson_destroy(query);
      mongoc_cursor_destroy(cursor);
      mongoc_collection_destroy(collection);
      mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);

      // Update redis
      std::string key = std::to_string(user_id) + ":followees";
      auto redis_insert_span = opentracing::Tracer::Global()->StartSpan(
          "social_graph_redis_insert_client",
          {opentracing::ChildOf(&span->context())});
      try {
        if (_redis_client_pool) {
          _redis_client_pool->zadd(key, redis_zset.begin(), redis_zset.end());
        } 
        else if (IsRedisReplicationEnabled()) {
            _redis_primary_client_pool->zadd(key, redis_zset.begin(), redis_zset.end());
        }
        else {
          _redis_cluster_client_pool->zadd(key, redis_zset.begin(),
                                           redis_zset.end());
        }
      } catch (const Error &err) {
        LOG(error) << err.what();
        throw err;
      }
      redis_span->Finish();
    }
  }
  span->Finish();
}

void SocialGraphHandler::InsertUser(
    int64_t req_id, int64_t user_id,
    const std::map<std::string, std::string> &carrier) {
  // Initialize a span
  TextMapReader reader(carrier);
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "insert_user_server", {opentracing::ChildOf(parent_span->get())});
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  mongoc_client_t *mongodb_client =
      mongoc_client_pool_pop(_mongodb_client_pool);
  if (!mongodb_client) {
    ServiceException se;
    se.errorCode = ErrorCode::SE_MONGODB_ERROR;
    se.message = "Failed to pop a client from MongoDB pool";
    throw se;
  }
  auto collection = mongoc_client_get_collection(mongodb_client, "social-graph",
                                                 "social-graph");
  if (!collection) {
    ServiceException se;
    se.errorCode = ErrorCode::SE_MONGODB_ERROR;
    se.message = "Failed to create collection social_graph from MongoDB";
    mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
    throw se;
  }

  bson_t *new_doc = BCON_NEW("user_id", BCON_INT64(user_id), "followers", "[",
                             "]", "followees", "[", "]");
  bson_error_t error;
  auto insert_span = opentracing::Tracer::Global()->StartSpan(
      "social_graph_mongo_insert_client",
      {opentracing::ChildOf(&span->context())});
  bool inserted = mongoc_collection_insert_one(collection, new_doc, nullptr,
                                               nullptr, &error);
  insert_span->Finish();
  if (!inserted) {
    LOG(error) << "Failed to insert social graph for user " << user_id
               << " to MongoDB: " << error.message;
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

void SocialGraphHandler::FollowWithUsername(
    int64_t req_id, const std::string &user_name,
    const std::string &followee_name,
    const std::map<std::string, std::string> &carrier) {
  // Initialize a span
  TextMapReader reader(carrier);
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "follow_with_username_server",
      {opentracing::ChildOf(parent_span->get())});
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  std::future<int64_t> user_id_future = std::async(std::launch::async, [&]() {
    auto user_client_wrapper = _user_service_client_pool->Pop();
    if (!user_client_wrapper) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_THRIFT_CONN_ERROR;
      se.message = "Failed to connect to social-graph-service";
      throw se;
    }
    auto user_client = user_client_wrapper->GetClient();
    int64_t _return;
    try {
      _return = user_client->GetUserId(req_id, user_name, writer_text_map);
    } catch (...) {
      _user_service_client_pool->Remove(user_client_wrapper);
      LOG(error) << "Failed to get user_id from user-service";
      throw;
    }
    _user_service_client_pool->Keepalive(user_client_wrapper);
    return _return;
  });

  std::future<int64_t> followee_id_future =
      std::async(std::launch::async, [&]() {
        auto user_client_wrapper = _user_service_client_pool->Pop();
        if (!user_client_wrapper) {
          ServiceException se;
          se.errorCode = ErrorCode::SE_THRIFT_CONN_ERROR;
          se.message = "Failed to connect to social-graph-service";
          throw se;
        }
        auto user_client = user_client_wrapper->GetClient();
        int64_t _return;
        try {
          _return =
              user_client->GetUserId(req_id, followee_name, writer_text_map);
        } catch (...) {
          _user_service_client_pool->Remove(user_client_wrapper);
          LOG(error) << "Failed to get user_id from user-service";
          throw;
        }
        _user_service_client_pool->Keepalive(user_client_wrapper);
        return _return;
      });

  int64_t user_id;
  int64_t followee_id;
  try {
    user_id = user_id_future.get();
    followee_id = followee_id_future.get();
  } catch (const std::exception &e) {
    LOG(warning) << e.what();
    throw;
  }

  if (user_id >= 0 && followee_id >= 0) {
    Follow(req_id, user_id, followee_id, writer_text_map);
  }
  span->Finish();
}

void SocialGraphHandler::UnfollowWithUsername(
    int64_t req_id, const std::string &user_name,
    const std::string &followee_name,
    const std::map<std::string, std::string> &carrier) {
  // Initialize a span
  TextMapReader reader(carrier);
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "unfollow_with_username_server",
      {opentracing::ChildOf(parent_span->get())});
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  std::future<int64_t> user_id_future = std::async(std::launch::async, [&]() {
    auto user_client_wrapper = _user_service_client_pool->Pop();
    if (!user_client_wrapper) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_THRIFT_CONN_ERROR;
      se.message = "Failed to connect to social-graph-service";
      throw se;
    }
    auto user_client = user_client_wrapper->GetClient();
    int64_t _return;
    try {
      _return = user_client->GetUserId(req_id, user_name, writer_text_map);
    } catch (...) {
      _user_service_client_pool->Remove(user_client_wrapper);
      LOG(error) << "Failed to get user_id from user-service";
      throw;
    }
    _user_service_client_pool->Keepalive(user_client_wrapper);
    return _return;
  });

  std::future<int64_t> followee_id_future =
      std::async(std::launch::async, [&]() {
        auto user_client_wrapper = _user_service_client_pool->Pop();
        if (!user_client_wrapper) {
          ServiceException se;
          se.errorCode = ErrorCode::SE_THRIFT_CONN_ERROR;
          se.message = "Failed to connect to social-graph-service";
          throw se;
        }
        auto user_client = user_client_wrapper->GetClient();
        int64_t _return;
        try {
          _return =
              user_client->GetUserId(req_id, followee_name, writer_text_map);
        } catch (...) {
          _user_service_client_pool->Remove(user_client_wrapper);
          LOG(error) << "Failed to get user_id from user-service";
          throw;
        }
        _user_service_client_pool->Keepalive(user_client_wrapper);
        return _return;
      });

  int64_t user_id;
  int64_t followee_id;
  try {
    user_id = user_id_future.get();
    followee_id = followee_id_future.get();
  } catch (...) {
    throw;
  }

  if (user_id >= 0 && followee_id >= 0) {
    try {
      Unfollow(req_id, user_id, followee_id, writer_text_map);
    } catch (...) {
      throw;
    }
  }
  span->Finish();
}

}  // namespace social_network

#endif  // SOCIAL_NETWORK_MICROSERVICES_SOCIALGRAPHHANDLER_H
