#ifndef SOCIAL_NETWORK_MICROSERVICES_SOCIALGRAPHHANDLER_H
#define SOCIAL_NETWORK_MICROSERVICES_SOCIALGRAPHHANDLER_H

#include <iostream>
#include <string>
#include <chrono>
#include <thread>
#include <vector>

#include <mongoc.h>
#include <bson/bson.h>
#include <cpp_redis/cpp_redis>

#include "../../gen-cpp/SocialGraphService.h"
#include "../../gen-cpp/UserService.h"
#include "../ClientPool.h"
#include "../logger.h"
#include "../tracing.h"
#include "../RedisClient.h"
#include "../ThriftClient.h"

namespace social_network {

using std::chrono::milliseconds;
using std::chrono::duration_cast;
using std::chrono::system_clock;

class SocialGraphHandler : public SocialGraphServiceIf {
 public:
  SocialGraphHandler(
      mongoc_client_pool_t *,
      ClientPool<RedisClient> *,
      ClientPool<ThriftClient<UserServiceClient>> *);
  ~SocialGraphHandler() override = default;
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
  void UnfollowWithUsername(int64_t, const std::string &, const std::string &,
                const std::map<std::string, std::string> &) override;
  void InsertUser(int64_t, int64_t,
                  const std::map<std::string, std::string> &) override;


 private:
  mongoc_client_pool_t *_mongodb_client_pool;
  ClientPool<RedisClient> *_redis_client_pool;
  ClientPool<ThriftClient<UserServiceClient>> *_user_service_client_pool;
};

SocialGraphHandler::SocialGraphHandler(
    mongoc_client_pool_t *mongodb_client_pool,
    ClientPool<RedisClient> *redis_client_pool,
    ClientPool<ThriftClient<UserServiceClient>> *user_service_client_pool) {
  _mongodb_client_pool = mongodb_client_pool;
  _redis_client_pool = redis_client_pool;
  _user_service_client_pool = user_service_client_pool;
}

void SocialGraphHandler::Follow(
    int64_t req_id,
    int64_t user_id,
    int64_t followee_id,
    const std::map<std::string, std::string> &carrier) {

  // Initialize a span
  TextMapReader reader(carrier);
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "Follow",
      {opentracing::ChildOf(parent_span->get())});
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  int64_t timestamp = duration_cast<milliseconds>(
      system_clock::now().time_since_epoch()).count();

  std::future<void> mongo_update_follower_future = std::async(
      std::launch::async, [&]() {
        mongoc_client_t *mongodb_client = mongoc_client_pool_pop(
            _mongodb_client_pool);
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
            "$and", "[",
            "{", "user_id", BCON_INT64(user_id), "}", "{",
            "followees", "{", "$not", "{", "$elemMatch", "{",
            "user_id", BCON_INT64(followee_id), "}", "}", "}", "}", "]"
        );
        bson_t *update = BCON_NEW(
            "$push",
            "{",
            "followees",
            "{",
            "user_id",
            BCON_INT64(followee_id),
            "timestamp",
            BCON_INT64(timestamp),
            "}",
            "}"
        );
        bson_error_t error;
        bson_t reply;
        auto update_span = opentracing::Tracer::Global()->StartSpan(
            "MongoUpdateFollower", {opentracing::ChildOf(&span->context())});
        bool updated = mongoc_collection_find_and_modify(
            collection,
            search_not_exist,
            nullptr,
            update,
            nullptr,
            false,
            false,
            true,
            &reply,
            &error);
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

  std::future<void> mongo_update_followee_future = std::async(
      std::launch::async, [&]() {
        mongoc_client_t *mongodb_client = mongoc_client_pool_pop(
            _mongodb_client_pool);
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
        bson_t *search_not_exist = BCON_NEW(
            "$and", "[", "{", "user_id", BCON_INT64(followee_id), "}", "{",
            "followers", "{", "$not", "{", "$elemMatch", "{",
            "user_id", BCON_INT64(user_id), "}", "}", "}", "}", "]"
        );
        bson_t *update = BCON_NEW(
            "$push", "{", "followers", "{", "user_id", BCON_INT64(user_id),
            "timestamp", BCON_INT64(timestamp), "}", "}"
        );
        bson_error_t error;
        auto update_span = opentracing::Tracer::Global()->StartSpan(
            "MongoUpdateFollowee", {opentracing::ChildOf(&span->context())});
        bson_t reply;
        bool updated = mongoc_collection_find_and_modify(
            collection, search_not_exist, nullptr, update, nullptr, false,
            false, true, &reply, &error);
        if (!updated) {
          LOG(error) << "Failed to update social graph for user "
                     << followee_id << " to MongoDB: " << error.message;
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

  std::future<void> redis_update_future = std::async(
      std::launch::async, [&]() {
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
        auto num_followee = redis_client->zcard(
            std::to_string(user_id) + ":followees");
        auto num_follower = redis_client->zcard(
            std::to_string(followee_id) + ":followers");
        redis_client->sync_commit();
        auto num_followee_reply = num_followee.get();
        auto num_follower_reply = num_follower.get();

        std::vector<std::string> options{"NX"};
        if (num_followee_reply.ok() && num_followee_reply.as_integer()) {
          std::string key = std::to_string(user_id) + ":followees";
          std::multimap<std::string, std::string> value = {{
            std::to_string(timestamp), std::to_string(followee_id)}};
          redis_client->zadd(key, options, value);
        }
        if (num_follower_reply.ok() && num_follower_reply.as_integer()) {
          std::string key = std::to_string(followee_id) + ":followers";
          std::multimap<std::string, std::string> value = {
              {std::to_string(timestamp), std::to_string(user_id)}};
          redis_client->zadd(key, options, value);
        }
        redis_client->sync_commit();
        _redis_client_pool->Push(redis_client_wrapper);
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

void SocialGraphHandler::Unfollow(
    int64_t req_id,
    int64_t user_id,
    int64_t followee_id,
    const std::map<std::string, std::string> &carrier) {
  // Initialize a span
  TextMapReader reader(carrier);
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "Unfollow",
      {opentracing::ChildOf(parent_span->get())});
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  std::future<void> mongo_update_follower_future = std::async(
      std::launch::async, [&]() {
        mongoc_client_t *mongodb_client = mongoc_client_pool_pop(
            _mongodb_client_pool);
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
        bson_t *update = BCON_NEW(
            "$pull", "{", "followees", "{",
            "user_id", BCON_INT64(followee_id), "}", "}"
        );
        bson_t reply;
        bson_error_t error;
        auto update_span = opentracing::Tracer::Global()->StartSpan(
            "MongoDeleteFollowee", {opentracing::ChildOf(&span->context())});
        bool updated = mongoc_collection_find_and_modify(
            collection, query, nullptr, update, nullptr, false, false,
            true, &reply, &error);
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

  std::future<void> mongo_update_followee_future = std::async(
      std::launch::async, [&]() {
        mongoc_client_t *mongodb_client = mongoc_client_pool_pop(
            _mongodb_client_pool);
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
        bson_t *update = BCON_NEW(
            "$pull", "{", "followers", "{",
            "user_id", BCON_INT64(user_id), "}", "}"
        );
        bson_t reply;
        bson_error_t error;
        auto update_span = opentracing::Tracer::Global()->StartSpan(
            "MongoDeleteFollower", {opentracing::ChildOf(&span->context())});
        bool updated = mongoc_collection_find_and_modify(
            collection, query, nullptr, update, nullptr, false, false,
            true, &reply, &error);
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

  std::future<void> redis_update_future = std::async(
      std::launch::async, [&]() {
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
        auto num_followee = redis_client->zcard(
            std::to_string(user_id) + ":followees");
        auto num_follower = redis_client->zcard(
            std::to_string(followee_id) + ":followers");
        redis_client->sync_commit();
        auto num_followee_reply = num_followee.get();
        auto num_follower_reply = num_follower.get();

        if (num_followee_reply.ok() && num_followee_reply.as_integer()) {
          std::string key = std::to_string(user_id) + ":followees";
          std::vector<std::string> members{std::to_string(followee_id)};
          redis_client->zrem(key, members);
        }
        if (num_follower_reply.ok() && num_follower_reply.as_integer()) {
          std::string key = std::to_string(followee_id) + ":followers";
          std::vector<std::string> members{std::to_string(user_id)};
          redis_client->zrem(key, members);
        }
        redis_client->sync_commit();
        _redis_client_pool->Push(redis_client_wrapper);
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
      "GetFollowers",
      {opentracing::ChildOf(parent_span->get())});
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  auto redis_client_wrapper = _redis_client_pool->Pop();
  if (!redis_client_wrapper) {
    ServiceException se;
    se.errorCode = ErrorCode::SE_REDIS_ERROR;
    se.message = "Cannot connect to Redis server";
    throw se;
  }
  auto redis_client = redis_client_wrapper->GetClient();

  auto redis_span = opentracing::Tracer::Global()->StartSpan(
      "RedisGet", {opentracing::ChildOf(&span->context())});
  auto num_follower = redis_client->zcard(
      std::to_string(user_id) + ":followers");
  redis_client->sync_commit();
  auto num_follower_reply = num_follower.get();

  if (num_follower_reply.ok() && num_follower_reply.as_integer()) {
    std::string key = std::to_string(user_id) + ":followers";
    auto redis_followers = redis_client->zrange(key, 0, -1, false);
    redis_client->sync_commit();
    redis_span->Finish();
    auto redis_followers_reply = redis_followers.get();
    if (redis_followers_reply.ok()) {
      auto followers_str = redis_followers_reply.as_array();
      for (auto const &item : followers_str) {
        _return.emplace_back(std::stoul(item.as_string()));
      }
      _redis_client_pool->Push(redis_client_wrapper);
      return;
    } else {
      ServiceException se;
      se.message = "Failed to get followers from Redis";
      se.errorCode = ErrorCode::SE_REDIS_ERROR;
      _redis_client_pool->Push(redis_client_wrapper);
      throw se;
    }
  } else {
    redis_span->Finish();
    _redis_client_pool->Push(redis_client_wrapper);
    mongoc_client_t *mongodb_client = mongoc_client_pool_pop(
        _mongodb_client_pool);
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
        "MongoFindUser", {opentracing::ChildOf(&span->context())});
    mongoc_cursor_t *cursor = mongoc_collection_find_with_opts(
        collection, query, nullptr, nullptr);
    const bson_t *doc;
    bool found = mongoc_cursor_next(cursor, &doc);
    if (found) {
      bson_iter_t iter_0;
      bson_iter_t iter_1;
      bson_iter_t user_id_child;
      bson_iter_t timestamp_child;
      int index = 0;
      std::multimap<std::string, std::string> redis_zset;
      bson_iter_init(&iter_0, doc);
      bson_iter_init(&iter_1, doc);

      while (
          bson_iter_find_descendant(
              &iter_0,
              ("followers." + std::to_string(index) + ".user_id").c_str(),
              &user_id_child) &&
              BSON_ITER_HOLDS_INT64 (&user_id_child) &&
              bson_iter_find_descendant(
                  &iter_1,
                  ("followers." + std::to_string(index) + ".timestamp").c_str(),
                  &timestamp_child)
              && BSON_ITER_HOLDS_INT64 (&timestamp_child)) {

        auto iter_user_id = bson_iter_int64(&user_id_child);
        auto iter_timestamp = bson_iter_int64(&timestamp_child);
        _return.emplace_back(iter_user_id);
        redis_zset.emplace(std::pair<std::string, std::string>(
            std::to_string(iter_timestamp), std::to_string(iter_user_id)));
        bson_iter_init(&iter_0, doc);
        bson_iter_init(&iter_1, doc);
        index++;
      }
      find_span->Finish();
      bson_destroy(query);
      mongoc_cursor_destroy(cursor);
      mongoc_collection_destroy(collection);
      mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);

      redis_client_wrapper = _redis_client_pool->Pop();
      redis_client = redis_client_wrapper->GetClient();
      auto redis_insert_span = opentracing::Tracer::Global()->StartSpan(
          "RedisInsert", {opentracing::ChildOf(&span->context())});
      std::string key = std::to_string(user_id) + ":followers";
      std::vector<std::string> options{"NX"};
      redis_client->zadd(key, options, redis_zset);
      redis_client->sync_commit();
      redis_insert_span->Finish();
      _redis_client_pool->Push(redis_client_wrapper);
    } else {
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
      "GetFollowees",
      {opentracing::ChildOf(parent_span->get())});
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  auto redis_client_wrapper = _redis_client_pool->Pop();
  if (!redis_client_wrapper) {
    ServiceException se;
    se.errorCode = ErrorCode::SE_REDIS_ERROR;
    se.message = "Cannot connect to Redis server";
    throw se;
  }
  auto redis_client = redis_client_wrapper->GetClient();

  auto redis_span = opentracing::Tracer::Global()->StartSpan(
      "RedisGet", {opentracing::ChildOf(&span->context())});
  auto num_followees = redis_client->zcard(
      std::to_string(user_id) + ":followees");
  redis_client->sync_commit();
  auto num_followees_reply = num_followees.get();

  if (num_followees_reply.ok() && num_followees_reply.as_integer()) {
    std::string key = std::to_string(user_id) + ":followees";
    auto redis_followees = redis_client->zrange(key, 0, -1, false);
    redis_client->sync_commit();
    redis_span->Finish();
    auto redis_followees_reply = redis_followees.get();
    if (redis_followees_reply.ok()) {
      auto followees_str = redis_followees_reply.as_array();
      for (auto const &item : followees_str) {
        _return.emplace_back(std::stoul(item.as_string()));
      }
      _redis_client_pool->Push(redis_client_wrapper);
      return;
    } else {
      ServiceException se;
      se.message = "Failed to get followees from Redis";
      se.errorCode = ErrorCode::SE_REDIS_ERROR;
      _redis_client_pool->Push(redis_client_wrapper);
      throw se;
    }
  } else {
    redis_span->Finish();
    mongoc_client_t *mongodb_client = mongoc_client_pool_pop(
        _mongodb_client_pool);
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
        "MongoFindUser", {opentracing::ChildOf(&span->context())});
    mongoc_cursor_t *cursor = mongoc_collection_find_with_opts(
        collection, query, nullptr, nullptr);
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
      std::multimap<std::string, std::string> redis_zset;
      bson_iter_init(&iter_0, doc);
      bson_iter_init(&iter_1, doc);

      while (
          bson_iter_find_descendant(
              &iter_0,
              ("followees." + std::to_string(index) + ".user_id").c_str(),
              &user_id_child) &&
              BSON_ITER_HOLDS_INT64 (&user_id_child) &&
              bson_iter_find_descendant(
                  &iter_1,
                  ("followees." + std::to_string(index) + ".timestamp").c_str(),
                  &timestamp_child)
              && BSON_ITER_HOLDS_INT64 (&timestamp_child)) {

        auto iter_user_id = bson_iter_int64(&user_id_child);
        auto iter_timestamp = bson_iter_int64(&timestamp_child);
        _return.emplace_back(iter_user_id);
        redis_zset.emplace(std::pair<std::string, std::string>(
            std::to_string(iter_timestamp), std::to_string(iter_user_id)));
        bson_iter_init(&iter_0, doc);
        bson_iter_init(&iter_1, doc);
        index++;
      }
      find_span->Finish();
      bson_destroy(query);
      mongoc_cursor_destroy(cursor);
      mongoc_collection_destroy(collection);
      mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
      redis_client_wrapper = _redis_client_pool->Pop();
      redis_client = redis_client_wrapper->GetClient();
      auto redis_insert_span = opentracing::Tracer::Global()->StartSpan(
          "RedisInsert", {opentracing::ChildOf(&span->context())});
      std::string key = std::to_string(user_id) + ":followees";
      std::vector<std::string> options{"NX"};
      redis_client->zadd(key, options, redis_zset);
      redis_client->sync_commit();
      redis_insert_span->Finish();
      _redis_client_pool->Push(redis_client_wrapper);
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
      "InsertUser",
      {opentracing::ChildOf(parent_span->get())});
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
      mongodb_client, "social-graph", "social-graph");
  if (!collection) {
    ServiceException se;
    se.errorCode = ErrorCode::SE_MONGODB_ERROR;
    se.message = "Failed to create collection social_graph from MongoDB";
    mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
    throw se;
  }

  bson_t *new_doc = BCON_NEW(
      "user_id", BCON_INT64(user_id),
      "followers", "[", "]",
      "followees", "[", "]"
  );
  bson_error_t error;
  auto insert_span = opentracing::Tracer::Global()->StartSpan(
      "MongoInsertUser", {opentracing::ChildOf(&span->context())});
  bool inserted = mongoc_collection_insert_one(
      collection, new_doc, nullptr, nullptr, &error);
  insert_span->Finish();
  if (!inserted) {
    LOG(error) << "Failed to insert social graph for user "
               << user_id << " to MongoDB: " << error.message;
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
    int64_t req_id,
    const std::string &user_name,
    const std::string &followee_name,
    const std::map<std::string, std::string> &carrier) {

  // Initialize a span
  TextMapReader reader(carrier);
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "FollowWithUsername",
      {opentracing::ChildOf(parent_span->get())});
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  std::future<int64_t> user_id_future = std::async(
      std::launch::async,[&]() {
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
          _user_service_client_pool->Push(user_client_wrapper);
          LOG(error) << "Failed to get user_id from user-service";
          throw;
        }        
        _user_service_client_pool->Push(user_client_wrapper);
        return _return;
      });

  std::future<int64_t> followee_id_future = std::async(
      std::launch::async,[&]() {
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
          _return = user_client->GetUserId(req_id, followee_name, writer_text_map);
        } catch (...) {
          _user_service_client_pool->Push(user_client_wrapper);
          LOG(error) << "Failed to get user_id from user-service";
          throw;
        }
        _user_service_client_pool->Push(user_client_wrapper);
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
      Follow(req_id, user_id, followee_id, writer_text_map);
    } catch (...) {
      throw;
    }
  }
  span->Finish();
}

void SocialGraphHandler::UnfollowWithUsername(
    int64_t req_id,
    const std::string &user_name,
    const std::string &followee_name,
    const std::map<std::string, std::string> &carrier) {
// Initialize a span
  TextMapReader reader(carrier);
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "UnfollowWithUsername",
      {opentracing::ChildOf(parent_span->get())});
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  std::future<int64_t> user_id_future = std::async(
      std::launch::async,[&]() {
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
          _user_service_client_pool->Push(user_client_wrapper);
          LOG(error) << "Failed to get user_id from user-service";
          throw;
        }
        _user_service_client_pool->Push(user_client_wrapper);
        return _return;
      });

  std::future<int64_t> followee_id_future = std::async(
      std::launch::async,[&]() {
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
          _return = user_client->GetUserId(req_id, followee_name, writer_text_map);
        } catch (...) {
          _user_service_client_pool->Push(user_client_wrapper);
          LOG(error) << "Failed to get user_id from user-service";
          throw;
        }        
        _user_service_client_pool->Push(user_client_wrapper);
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

} // namespace social_network

#endif //SOCIAL_NETWORK_MICROSERVICES_SOCIALGRAPHHANDLER_H
