#ifndef SOCIAL_NETWORK_MICROSERVICES_SRC_HOMETIMELINESERVICE_HOMETIMELINEHANDLER_H_
#define SOCIAL_NETWORK_MICROSERVICES_SRC_HOMETIMELINESERVICE_HOMETIMELINEHANDLER_H_

#include <sw/redis++/redis++.h>

#include <future>
#include <iostream>
#include <string>

#include "../../gen-cpp/HomeTimelineService.h"
#include "../../gen-cpp/PostStorageService.h"
#include "../../gen-cpp/SocialGraphService.h"
#include "../ClientPool.h"
#include "../ThriftClient.h"
#include "../logger.h"
#include "../tracing.h"

using namespace sw::redis;
namespace social_network {
class HomeTimelineHandler : public HomeTimelineServiceIf {
 public:
  explicit HomeTimelineHandler(
      Redis *, ClientPool<ThriftClient<PostStorageServiceClient>> *,
      ClientPool<ThriftClient<SocialGraphServiceClient>> *);
  ~HomeTimelineHandler() override = default;

  void ReadHomeTimeline(std::vector<Post> &, int64_t, int64_t, int, int,
                        const std::map<std::string, std::string> &) override;

  void WriteHomeTimeline(int64_t, int64_t, int64_t, int64_t,
                         const std::vector<int64_t> &,
                         const std::map<std::string, std::string> &) override;

 private:
  Redis *_redis_client_pool;
  ClientPool<ThriftClient<PostStorageServiceClient>> *_post_client_pool;
  ClientPool<ThriftClient<SocialGraphServiceClient>> *_social_graph_client_pool;
};

HomeTimelineHandler::HomeTimelineHandler(
    Redis *redis_pool,
    ClientPool<ThriftClient<PostStorageServiceClient>> *post_client_pool,
    ClientPool<ThriftClient<SocialGraphServiceClient>>
        *social_graph_client_pool) {
  _redis_client_pool = redis_pool;
  _post_client_pool = post_client_pool;
  _social_graph_client_pool = social_graph_client_pool;
}

void HomeTimelineHandler::WriteHomeTimeline(
    int64_t req_id, int64_t post_id, int64_t user_id, int64_t timestamp,
    const std::vector<int64_t> &user_mentions_id,
    const std::map<std::string, std::string> &carrier) {
  // Initialize a span
  TextMapReader reader(carrier);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "write_home_timeline_server", {opentracing::ChildOf(parent_span->get())});

  // Find followers of the user
  auto followers_span = opentracing::Tracer::Global()->StartSpan(
      "get_followers_client", {opentracing::ChildOf(&span->context())});
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  opentracing::Tracer::Global()->Inject(followers_span->context(), writer);

  auto social_graph_client_wrapper = _social_graph_client_pool->Pop();
  if (!social_graph_client_wrapper) {
    ServiceException se;
    se.errorCode = ErrorCode::SE_THRIFT_CONN_ERROR;
    se.message = "Failed to connect to social-graph-service";
    throw se;
  }
  auto social_graph_client = social_graph_client_wrapper->GetClient();
  std::vector<int64_t> followers_id;
  try {
    social_graph_client->GetFollowers(followers_id, req_id, user_id,
                                      writer_text_map);
  } catch (...) {
    LOG(error) << "Failed to get followers from social-network-service";
    _social_graph_client_pool->Remove(social_graph_client_wrapper);
    throw;
  }
  _social_graph_client_pool->Keepalive(social_graph_client_wrapper);
  followers_span->Finish();

  std::set<int64_t> followers_id_set(followers_id.begin(), followers_id.end());
  followers_id_set.insert(user_mentions_id.begin(), user_mentions_id.end());

  // Update Redis ZSet
  // Zset key: follower_id, Zset value: post_id_str, Zset score: timestamp_str
  auto redis_span = opentracing::Tracer::Global()->StartSpan(
      "write_home_timeline_redis_update_client",
      {opentracing::ChildOf(&span->context())});
  std::string post_id_str = std::to_string(post_id);

  {
    auto pipe = _redis_client_pool->pipeline(false);
    for (auto &follower_id : followers_id_set) {
      pipe.zadd(std::to_string(follower_id), post_id_str, timestamp,
                UpdateType::NOT_EXIST);
    }
    try {
      auto replies = pipe.exec();
    } catch (const Error &err) {
      LOG(error) << err.what();
      throw err;
    }
  }
  redis_span->Finish();
}

void HomeTimelineHandler::ReadHomeTimeline(
    std::vector<Post> &_return, int64_t req_id, int64_t user_id, int start,
    int stop, const std::map<std::string, std::string> &carrier) {
  // Initialize a span
  TextMapReader reader(carrier);
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "read_home_timeline_server", {opentracing::ChildOf(parent_span->get())});
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  if (stop <= start || start < 0) {
    return;
  }

  auto redis_span = opentracing::Tracer::Global()->StartSpan(
      "read_home_timeline_redis_find_client",
      {opentracing::ChildOf(&span->context())});

  std::vector<std::string> post_ids_str;
  try {
    _redis_client_pool->zrevrange(std::to_string(user_id), start, stop - 1,
                                  std::back_inserter(post_ids_str));
  } catch (const Error &err) {
    LOG(error) << err.what();
    throw err;
  }
  redis_span->Finish();

  std::vector<int64_t> post_ids;
  for (auto &post_id_str : post_ids_str) {
    post_ids.emplace_back(std::stoul(post_id_str));
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

}  // namespace social_network

#endif  // SOCIAL_NETWORK_MICROSERVICES_SRC_HOMETIMELINESERVICE_HOMETIMELINEHANDLER_H_
