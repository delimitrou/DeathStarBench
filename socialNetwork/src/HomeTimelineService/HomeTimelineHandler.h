#ifndef SOCIAL_NETWORK_MICROSERVICES_SRC_HOMETIMELINESERVICE_HOMETIMELINEHANDLER_H_
#define SOCIAL_NETWORK_MICROSERVICES_SRC_HOMETIMELINESERVICE_HOMETIMELINEHANDLER_H_

#include <iostream>
#include <string>
#include <future>

#include <cpp_redis/cpp_redis>

#include "../../gen-cpp/HomeTimelineService.h"
#include "../../gen-cpp/PostStorageService.h"
#include "../../gen-cpp/SocialGraphService.h"
#include "../logger.h"
#include "../tracing.h"
#include "../ClientPool.h"
#include "../RedisClient.h"
#include "../ThriftClient.h"

namespace social_network {

class HomeTimelineHandler : public HomeTimelineServiceIf {
 public:
  explicit HomeTimelineHandler(
      ClientPool<RedisClient> *,
      ClientPool<ThriftClient<PostStorageServiceClient>> *,
      ClientPool<ThriftClient<SocialGraphServiceClient>> *);
  ~HomeTimelineHandler() override = default;

  void ReadHomeTimeline(std::vector<Post> &, int64_t, int64_t, int, int,
      const std::map<std::string, std::string> &) override ;

  void WriteHomeTimeline(int64_t, int64_t, int64_t, int64_t, 
      const std::vector<int64_t> &, 
      const std::map<std::string, std::string> &) override ;

 private:
  ClientPool<RedisClient> *_redis_client_pool;
  ClientPool<ThriftClient<PostStorageServiceClient>> *_post_client_pool;
  ClientPool<ThriftClient<SocialGraphServiceClient>> *_social_graph_client_pool;
};


HomeTimelineHandler::HomeTimelineHandler(
    ClientPool<RedisClient> *redis_pool,
    ClientPool<ThriftClient<PostStorageServiceClient>> *post_client_pool,
    ClientPool<ThriftClient<SocialGraphServiceClient>> *social_graph_client_pool) {
  _redis_client_pool = redis_pool;
  _post_client_pool = post_client_pool;
  _social_graph_client_pool = social_graph_client_pool;
}

void HomeTimelineHandler::WriteHomeTimeline(
    int64_t req_id, 
    int64_t post_id,
    int64_t user_id, 
    int64_t timestamp, 
    const std::vector<int64_t> &user_mentions_id,
    const std::map<std::string, std::string> &carrier) {
  
  // Initialize a span
  TextMapReader reader(carrier);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "write_home_timeline_server",
      { opentracing::ChildOf(parent_span->get()) });

  // Find followers of the user
  auto followers_span = opentracing::Tracer::Global()->StartSpan(
      "get_followers_client", {opentracing::ChildOf(&span->context())});
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  opentracing::Tracer::Global()->Inject(followers_span->context(), writer);
  
  auto social_graph_client_wrapper = _social_graph_client_pool->Pop();
  if (!social_graph_client_wrapper)
  {
    ServiceException se;
    se.errorCode = ErrorCode::SE_THRIFT_CONN_ERROR;
    se.message = "Failed to connect to social-graph-service";
    throw se;
  }
  auto social_graph_client = social_graph_client_wrapper->GetClient();
  std::vector<int64_t> followers_id;
  try
  {
    social_graph_client->GetFollowers(followers_id, req_id, user_id,
                                      writer_text_map);
  }
  catch (...)
  {
    LOG(error) << "Failed to get followers from social-network-service";
    _social_graph_client_pool->Remove(social_graph_client_wrapper);
    throw;
  }
  _social_graph_client_pool->Keepalive(social_graph_client_wrapper);
  followers_span->Finish();

  std::set<int64_t> followers_id_set(followers_id.begin(),
                                      followers_id.end());
  followers_id_set.insert(user_mentions_id.begin(), user_mentions_id.end());

  // Update Redis ZSet
  auto redis_span = opentracing::Tracer::Global()->StartSpan(
      "write_home_timeline_redis_update_client", {opentracing::ChildOf(&span->context())});
  auto redis_client_wrapper = _redis_client_pool->Pop();
  if (!redis_client_wrapper)
  {
    ServiceException se;
    se.errorCode = ErrorCode::SE_REDIS_ERROR;
    se.message = "Cannot connect to Redis server";
    throw se;
  }
  auto redis_client = redis_client_wrapper->GetClient();
  std::vector<std::string> options{"NX"};
  std::string post_id_str = std::to_string(post_id);
  std::string timestamp_str = std::to_string(timestamp);
  std::multimap<std::string, std::string> value =
      {{timestamp_str, post_id_str}};

  std::vector<std::future<cpp_redis::reply>> zadd_reply_futures;

  for (auto &follower_id : followers_id_set)
  {
    zadd_reply_futures.emplace_back(redis_client->zadd(std::to_string(follower_id), options, value));
  }

  redis_client->sync_commit();
  for (auto &zadd_reply_future: zadd_reply_futures) {
    auto reply = zadd_reply_future.get();
    if (!reply.ok()) {
      LOG(error) << "Home-timeline Redis zadd error:" << reply.error();
      _redis_client_pool->Remove(redis_client_wrapper);
      redis_span->Finish();
      return;
    }
  }
  _redis_client_pool->Keepalive(redis_client_wrapper);
  redis_span->Finish();
}

void HomeTimelineHandler::ReadHomeTimeline(
    std::vector<Post> & _return,
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
      "read_home_timeline_server",
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
      "home_timeline_redis_find_client", {opentracing::ChildOf(&span->context())});
  auto post_ids_future = redis_client->zrevrange(
      std::to_string(user_id), start, stop - 1);
  redis_client->sync_commit();
  cpp_redis::reply post_ids_reply;
  try {
    post_ids_reply = post_ids_future.get();
  } catch (...) {
    _redis_client_pool->Remove(redis_client_wrapper);
    LOG(error) << "Failed to read post_ids from home-timeline-redis";
    throw;
  }
  _redis_client_pool->Keepalive(redis_client_wrapper);
  redis_span->Finish();


  std::vector<int64_t> post_ids;
  if (post_ids_reply.is_error()) {
    LOG(error) << "Failed to read post_ids from home-timeline-redis:" << post_ids_reply.error();
  }
  auto post_ids_reply_array = post_ids_reply.as_array();
  for (auto &post_id_reply : post_ids_reply_array) {
    post_ids.emplace_back(std::stoul(post_id_reply.as_string()));
  }

  auto post_client_wrapper = _post_client_pool->Pop();
  if (!post_client_wrapper) {
    ServiceException se;
    se.errorCode = ErrorCode::SE_THRIFT_CONN_ERROR;
    se.message = "Failed to connect to post-storage-service";
    throw se;
  }
  auto post_client = post_client_wrapper->GetClient();
  try {
    post_client->ReadPosts(_return, req_id, post_ids, writer_text_map);
  } catch (...) {
    _post_client_pool->Remove(post_client_wrapper);
    LOG(error) << "Failed to read posts from post-storage-service";
    throw;
  }
  _post_client_pool->Keepalive(post_client_wrapper);
  span->Finish();
}

} // namespace social_network

#endif //SOCIAL_NETWORK_MICROSERVICES_SRC_HOMETIMELINESERVICE_HOMETIMELINEHANDLER_H_
