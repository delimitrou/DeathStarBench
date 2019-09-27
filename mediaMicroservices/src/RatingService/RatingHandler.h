#ifndef MEDIA_MICROSERVICES_RATINGHANDLER_H
#define MEDIA_MICROSERVICES_RATINGHANDLER_H

#include <iostream>
#include <string>
#include <future>


#include "../../gen-cpp/RatingService.h"
#include "../../gen-cpp/ComposeReviewService.h"
#include "../ClientPool.h"
#include "../ThriftClient.h"
#include "../RedisClient.h"
#include "../logger.h"
#include "../tracing.h"


namespace media_service {
class RatingHandler : public RatingServiceIf {
 public:
  RatingHandler(
      ClientPool<ThriftClient<ComposeReviewServiceClient>> *,
      ClientPool<RedisClient> *);
  ~RatingHandler() override = default;
  void UploadRating(int64_t, const std::string &, int32_t,
      const std::map<std::string, std::string> &) override;

 private:
  ClientPool<ThriftClient<ComposeReviewServiceClient>> *_compose_client_pool;
  ClientPool<RedisClient> *_redis_client_pool;
};

RatingHandler::RatingHandler(
    ClientPool<ThriftClient<ComposeReviewServiceClient>> *compose_client_pool,
    ClientPool<RedisClient> *redis_client_pool) {
  _compose_client_pool = compose_client_pool;
  _redis_client_pool = redis_client_pool;
}
void RatingHandler::UploadRating(
    int64_t req_id,
    const std::string &movie_id,
    int32_t rating,
    const std::map<std::string, std::string> & carrier) {

  // Initialize a span
  TextMapReader reader(carrier);
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "UploadRating",
      { opentracing::ChildOf(parent_span->get()) });
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  std::future<void> upload_future;
  std::future<void> redis_future;

  upload_future = std::async(std::launch::async, [&](){
    auto compose_client_wrapper = _compose_client_pool->Pop();
    if (!compose_client_wrapper) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_THRIFT_CONN_ERROR;
      se.message = "Failed to connected to compose-review-service";
      throw se;
    }
    auto compose_client = compose_client_wrapper->GetClient();
    try {
      compose_client->UploadRating(req_id, rating, writer_text_map);
    } catch (...) {
      _compose_client_pool->Push(compose_client_wrapper);
      LOG(error) << "Failed to upload rating to compose-review-service";
      throw;
    }
    _compose_client_pool->Push(compose_client_wrapper);
  });

  redis_future = std::async(std::launch::async, [&](){
    auto redis_client_wrapper = _redis_client_pool->Pop();
    if (!redis_client_wrapper) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_REDIS_ERROR;
      se.message = "Cannot connected to Redis server";
      throw se;
    }
    auto redis_client = redis_client_wrapper->GetClient();
    auto redis_span = opentracing::Tracer::Global()->StartSpan(
        "RedisInsert", {opentracing::ChildOf(&span->context())});
    redis_client->incrby(movie_id + ":uncommit_sum", rating);
    redis_client->incr(movie_id + ":uncommit_num");
    redis_client->sync_commit();
    redis_span->Finish();
    _redis_client_pool->Push(redis_client_wrapper);
  });

  try {
    upload_future.get();
  } catch (...) {
    LOG(error) << "Failed to upload rating to compose-review-service";
    throw;
  }

  try {
    redis_future.get();
  } catch (...) {
    LOG(error) << "Failed to update rating to rating-redis";
    throw;
  }
  span->Finish();
}


} // namespace media_service

#endif //MEDIA_MICROSERVICES_RATINGHANDLER_H
