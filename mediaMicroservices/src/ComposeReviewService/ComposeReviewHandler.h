#ifndef MEDIA_MICROSERVICES_COMPOSEREVIEWHANDLER_H
#define MEDIA_MICROSERVICES_COMPOSEREVIEWHANDLER_H

#include <iostream>
#include <string>
#include <chrono>
#include <future>

#include <libmemcached/memcached.h>
#include <libmemcached/util.h>

#include "../../gen-cpp/ComposeReviewService.h"
#include "../../gen-cpp/media_service_types.h"
#include "../../gen-cpp/ReviewStorageService.h"
#include "../../gen-cpp/UserReviewService.h"
#include "../../gen-cpp/MovieReviewService.h"
#include "../ClientPool.h"
#include "../ThriftClient.h"
#include "../logger.h"
#include "../tracing.h"

namespace media_service {
#define NUM_COMPONENTS 5
#define MMC_EXP_TIME 10

using std::chrono::milliseconds;
using std::chrono::duration_cast;
using std::chrono::system_clock;

class ComposeReviewHandler : public ComposeReviewServiceIf {
 public:
  ComposeReviewHandler(
      memcached_pool_st *,
      ClientPool<ThriftClient<ReviewStorageServiceClient>> *,
      ClientPool<ThriftClient<UserReviewServiceClient>> *,
      ClientPool<ThriftClient<MovieReviewServiceClient>> *);
  ~ComposeReviewHandler() override = default;

  void UploadText(int64_t, const std::string &,
      const std::map<std::string, std::string> &) override;
  void UploadRating(int64_t, int32_t,
      const std::map<std::string, std::string> &) override;
  void UploadUniqueId(int64_t, int64_t,
      const std::map<std::string, std::string> &) override;
  void UploadMovieId(int64_t, const std::string &,
                     const std::map<std::string, std::string> &) override;
  void UploadUserId(int64_t, int64_t,
      const std::map<std::string, std::string> &) override;


 private:
  memcached_pool_st *_memcached_client_pool;
  ClientPool<ThriftClient<ReviewStorageServiceClient>>
      *_review_storage_client_pool;
  ClientPool<ThriftClient<UserReviewServiceClient>>
      *_user_review_client_pool;
  ClientPool<ThriftClient<MovieReviewServiceClient>>
      *_movie_review_client_pool;
  void _ComposeAndUpload(int64_t, const std::map<std::string, std::string> &);
};

ComposeReviewHandler::ComposeReviewHandler(
    memcached_pool_st *memcached_client_pool,
    ClientPool<ThriftClient<ReviewStorageServiceClient>> 
        *review_storage_client_pool,
    ClientPool<ThriftClient<UserReviewServiceClient>>
        *user_review_client_pool,
    ClientPool<ThriftClient<MovieReviewServiceClient>>
        *movie_review_client_pool ) {
  _memcached_client_pool = memcached_client_pool;
  _review_storage_client_pool = review_storage_client_pool;
  _user_review_client_pool = user_review_client_pool;
  _movie_review_client_pool = movie_review_client_pool;
}

void ComposeReviewHandler::_ComposeAndUpload(
    int64_t req_id, const std::map<std::string, std::string> &writer_text_map) {

  std::string key_unique_id = std::to_string(req_id) + ":review_id";
  std::string key_movie_id = std::to_string(req_id) + ":movie_id";
  std::string key_user_id = std::to_string(req_id) + ":user_id";
  std::string key_text = std::to_string(req_id) + ":text";
  std::string key_rating = std::to_string(req_id) + ":rating";

  const char* keys[NUM_COMPONENTS] = {
      key_unique_id.c_str(),
      key_movie_id.c_str(),
      key_user_id.c_str(),
      key_text.c_str(),
      key_rating.c_str()
  };

  size_t key_sizes[NUM_COMPONENTS] = {
      key_unique_id.size(),
      key_movie_id.size(),
      key_user_id.size(),
      key_text.size(),
      key_rating.size()
  };

  // Compose a review from the components obtained from memcached
  memcached_return_t rc;
  auto client = memcached_pool_pop(_memcached_client_pool, true, &rc);
  if (!client) {
    ServiceException se;
    se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
    se.message = "Failed to pop a client from memcached pool";
    throw se;
  }

  Review new_review;
  rc = memcached_mget(client, keys, key_sizes, NUM_COMPONENTS);
  if (rc != MEMCACHED_SUCCESS) {
    LOG(error) << "Cannot get components of request " << req_id << ": "
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
      se.message =  "Cannot get components of request " + std::to_string(req_id);
      throw se;
    }
    std::string key_str(return_key, return_key + return_key_length);
    std::string value_str(return_value, return_value + return_value_length);
    if (key_str == key_unique_id) {
      new_review.review_id = std::stoul(value_str);
    } else if (key_str == key_movie_id) {
      new_review.movie_id = value_str;
    } else if (key_str == key_text) {
      new_review.text = value_str;
    } else if (key_str == key_user_id) {
      new_review.user_id = std::stoul(value_str);
    } else if (key_str == key_rating) {
      new_review.rating = std::stoi(value_str);
    } else {
      LOG(error) << "Unexpected memcached fetched data of request " << req_id
                   << " key: " << key_str
                   << " value: " << value_str;
      ServiceException se;
      se.errorCode = ErrorCode::SE_THRIFT_HANDLER_ERROR;
      se.message = "Unexpected memcached fetched data of request " +
          std::to_string(req_id);
      free(return_value);
      memcached_quit(client);
      memcached_pool_push(_memcached_client_pool, client);
      throw se;
    }
    free(return_value);
  }

  memcached_quit(client);
  memcached_pool_push(_memcached_client_pool, client);

  new_review.timestamp = duration_cast<milliseconds>(
      system_clock::now().time_since_epoch()).count();
  new_review.req_id = req_id;

  std::future<void> review_future;
  std::future<void> user_review_future;
  std::future<void> movie_review_future;
  
  review_future = std::async(std::launch::async, [&](){
    auto review_storage_client_wrapper = _review_storage_client_pool->Pop();
    if (!review_storage_client_wrapper) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_THRIFT_CONN_ERROR;
      se.message = "Failed to connected to review-storage-service";
      throw se;
    }
    auto review_storage_client = review_storage_client_wrapper->GetClient();
    try {
      review_storage_client->StoreReview(req_id, new_review, writer_text_map);
    } catch (...) {
      _review_storage_client_pool->Push(review_storage_client_wrapper);
      LOG(error) << "Failed to upload review to review-storage-service";
      throw;
    }
    _review_storage_client_pool->Push(review_storage_client_wrapper);
  });

  user_review_future = std::async(std::launch::async, [&](){
    auto user_review_client_wrapper = _user_review_client_pool->Pop();
    if (!user_review_client_wrapper) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_THRIFT_CONN_ERROR;
      se.message = "Failed to connected to user-review-service";
      throw se;
    }
    auto user_review_client = user_review_client_wrapper->GetClient();
    try {
      user_review_client->UploadUserReview(req_id, new_review.user_id,
          new_review.review_id, new_review.timestamp, writer_text_map);
    } catch (...) {
      _user_review_client_pool->Push(user_review_client_wrapper);
      LOG(error) << "Failed to upload review to user-review-service";
      throw;
    }
    _user_review_client_pool->Push(user_review_client_wrapper);
  });

  movie_review_future = std::async(std::launch::async, [&](){
    auto movie_review_client_wrapper = _movie_review_client_pool->Pop();
    if (!movie_review_client_wrapper) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_THRIFT_CONN_ERROR;
      se.message = "Failed to connected to movie-review-service";
      throw se;
    }
    auto movie_review_client = movie_review_client_wrapper->GetClient();
    try {
      movie_review_client->UploadMovieReview(req_id, new_review.movie_id,
          new_review.review_id, new_review.timestamp, writer_text_map);
    } catch (...) {
      _movie_review_client_pool->Push(movie_review_client_wrapper);
      LOG(error) << "Failed to upload review to movie-review-service";
      throw;
    }
    _movie_review_client_pool->Push(movie_review_client_wrapper);
  });
  
  try {
    review_future.get();
    user_review_future.get();
    movie_review_future.get();
  } catch (...) {
    throw;
  }
}

void ComposeReviewHandler::UploadMovieId(
    int64_t req_id,
    const std::string &movie_id,
    const std::map<std::string, std::string> & carrier) {

  // Initialize a span
  TextMapReader reader(carrier);
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "UploadMovieId",
      { opentracing::ChildOf(parent_span->get()) });
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  memcached_return_t memcached_rc;
  std::string key_counter = std::to_string(req_id) + ":counter";
  memcached_st *memcached_client = memcached_pool_pop(
      _memcached_client_pool, true, &memcached_rc);

  // Initialize the counter to 0 if there it is not in the memcached
  memcached_rc = memcached_add(
      memcached_client,
      key_counter.c_str(),
      key_counter.size(),
      "0", 1, MMC_EXP_TIME, 0);

  // error if it cannot be stored
  if (memcached_rc != MEMCACHED_SUCCESS &&
      memcached_rc != MEMCACHED_DATA_EXISTS) {
    LOG(error) << "Failed to initilize the counter for request " << req_id
        << " Error code: "
        << memcached_strerror(memcached_client, memcached_rc);
    ServiceException se;
    se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
    se.message = memcached_strerror(memcached_client, memcached_rc);
    memcached_pool_push(_memcached_client_pool, memcached_client);
    throw se;
  }

  // Store movie_id to memcached
  uint64_t counter_value;
  std::string key_movie_id = std::to_string(req_id) + ":movie_id";
  memcached_rc = memcached_add(
      memcached_client,
      key_movie_id.c_str(),
      key_movie_id.size(),
      movie_id.c_str(),
      movie_id.size(),
      MMC_EXP_TIME, 0);
  if (memcached_rc == MEMCACHED_DATA_EXISTS) {
    // Another thread has uploaded movie_id, which is an unexpected behaviour.
    LOG(warning) << "movie_id of request " << req_id
                 << " has already been stored";
    size_t value_size;
    char *counter_value_str = memcached_get(
        memcached_client,
        key_counter.c_str(),
        key_counter.size(),
        &value_size,
        nullptr,
        &memcached_rc);
    if (memcached_rc != MEMCACHED_SUCCESS) {
      LOG(error) << "Cannot get the counter of request " << req_id;
      ServiceException se;
      se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
      se.message = memcached_strerror(memcached_client, memcached_rc);
      memcached_pool_push(_memcached_client_pool, memcached_client);
      throw se;
    }
    counter_value = std::stoul(counter_value_str);
    free(counter_value_str);
  } else if (memcached_rc != MEMCACHED_SUCCESS) {
    LOG(error) << "Cannot store movie_id of request " << req_id;
    ServiceException se;
    se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
    se.message = memcached_strerror(memcached_client, memcached_rc);
    memcached_pool_push(_memcached_client_pool, memcached_client);
    throw se;
  } else {
    // Atomically increment and get the counter value
    memcached_increment(
        memcached_client,
        key_counter.c_str(),
        key_counter.size(),
        1, &counter_value);
    if (memcached_rc != MEMCACHED_SUCCESS) {
      LOG(error) << "Cannot increment and get the counter of request "
          << req_id;
      ServiceException se;
      se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
      se.message = memcached_strerror(memcached_client, memcached_rc);
      memcached_pool_push(_memcached_client_pool, memcached_client);
      throw se;
    }
  }
  LOG(debug) << "req_id " << req_id
      << " caching movie_id to Memcached finished";
  memcached_pool_push(_memcached_client_pool, memcached_client);

  // If this thread is the last one uploading the review components,
  // it is in charge of compose the request and upload to the microservices in
  // the next tier.
  if (counter_value == NUM_COMPONENTS) {
    _ComposeAndUpload(req_id, writer_text_map);
  }
  span->Finish();
}

void ComposeReviewHandler::UploadUserId(
    int64_t req_id, int64_t user_id,
    const std::map<std::string, std::string> & carrier) {

  // Initialize a span
  TextMapReader reader(carrier);
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "UploadUserId",
      { opentracing::ChildOf(parent_span->get()) });
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  memcached_return_t memcached_rc;
  std::string key_counter = std::to_string(req_id) + ":counter";
  memcached_st *memcached_client = memcached_pool_pop(
      _memcached_client_pool, true, &memcached_rc);

  // Initialize the counter to 0 if there it is not in the memcached
  memcached_rc = memcached_add(
      memcached_client,
      key_counter.c_str(),
      key_counter.size(),
      "0", 1, MMC_EXP_TIME, 0);

  // error if it cannot be stored
  if (memcached_rc != MEMCACHED_SUCCESS &&
      memcached_rc != MEMCACHED_DATA_EXISTS) {
    LOG(error) << "Failed to initilize the counter for request " << req_id
        << " Error code: "
        << memcached_strerror(memcached_client, memcached_rc);
    ServiceException se;
    se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
    se.message = memcached_strerror(memcached_client, memcached_rc);
    memcached_pool_push(_memcached_client_pool, memcached_client);
    throw se;
  }

  // Store user_id to memcached
  uint64_t counter_value;
  std::string key_user_id = std::to_string(req_id) + ":user_id";
  std::string user_id_str = std::to_string(user_id);
  memcached_rc = memcached_add(
      memcached_client,
      key_user_id.c_str(),
      key_user_id.size(),
      user_id_str.c_str(),
      user_id_str.size(),
      MMC_EXP_TIME, 0);
  if (memcached_rc == MEMCACHED_DATA_EXISTS) {
    // Another thread has uploaded user_id, which is an unexpected behaviour.
    LOG(warning) << "user_id of request " << req_id
                 << " has already been stored";
    size_t value_size;
    char *counter_value_str = memcached_get(
        memcached_client,
        key_counter.c_str(),
        key_counter.size(),
        &value_size,
        0,
        &memcached_rc);
    if (memcached_rc != MEMCACHED_SUCCESS) {
      LOG(error) << "Cannot get the counter of request " << req_id;
      ServiceException se;
      se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
      se.message = memcached_strerror(memcached_client, memcached_rc);
      memcached_pool_push(_memcached_client_pool, memcached_client);
      throw se;
    }
    counter_value = std::stoul(counter_value_str);
    free(counter_value_str);
  } else if (memcached_rc != MEMCACHED_SUCCESS) {
    LOG(error) << "Cannot store user_id of request " << req_id
               << " Error code: "
               << memcached_strerror(memcached_client, memcached_rc);
    ServiceException se;
    se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
    se.message = memcached_strerror(memcached_client, memcached_rc);
    memcached_pool_push(_memcached_client_pool, memcached_client);
    throw se;
  } else {
    // Atomically increment and get the counter value
    memcached_increment(
        memcached_client,
        key_counter.c_str(),
        key_counter.size(),
        1, &counter_value);
    if (memcached_rc != MEMCACHED_SUCCESS) {
      LOG(error) << "Cannot increment and get the counter of request "
                 << req_id;
      ServiceException se;
      se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
      se.message = memcached_strerror(memcached_client, memcached_rc);
      memcached_pool_push(_memcached_client_pool, memcached_client);
      throw se;
    }
  }
  LOG(debug) << "req_id " << req_id << "caching user to Memcached finished";
  memcached_pool_push(_memcached_client_pool, memcached_client);

  // If this thread is the last one uploading the review components,
  // it is in charge of compose the request and upload to the microservices in
  // the next tier.
  if (counter_value == NUM_COMPONENTS) {
    _ComposeAndUpload(req_id, writer_text_map);
  }
  span->Finish();
}

void ComposeReviewHandler::UploadUniqueId(
    int64_t req_id, int64_t review_id,
    const std::map<std::string, std::string> & carrier) {

  // Initialize a span
  TextMapReader reader(carrier);
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "UploadUniqueId",
      { opentracing::ChildOf(parent_span->get()) });
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  memcached_return_t memcached_rc;
  std::string key_counter = std::to_string(req_id) + ":counter";
  memcached_st *memcached_client = memcached_pool_pop(
      _memcached_client_pool, true, &memcached_rc);

  // Initialize the counter to 0 if there it is not in the memcached
  memcached_rc = memcached_add(
      memcached_client,
      key_counter.c_str(),
      key_counter.size(),
      "0", 1, MMC_EXP_TIME, 0);

  // error if it cannot be stored
  if (memcached_rc != MEMCACHED_SUCCESS &&
      memcached_rc != MEMCACHED_DATA_EXISTS) {
    LOG(error) << "Failed to initilize the counter for request " << req_id
               << " Error code: "
               << memcached_strerror(memcached_client, memcached_rc);
    ServiceException se;
    se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
    se.message = memcached_strerror(memcached_client, memcached_rc);
    memcached_pool_push(_memcached_client_pool, memcached_client);
    throw se;
  }

  // Store review_id to memcached
  uint64_t counter_value;
  std::string key_unique_id = std::to_string(req_id) + ":review_id";
  std::string unique_id_str = std::to_string(review_id);
  memcached_rc = memcached_add(
      memcached_client,
      key_unique_id.c_str(),
      key_unique_id.size(),
      unique_id_str.c_str(),
      unique_id_str.size(),
      MMC_EXP_TIME, 0);
  if (memcached_rc == MEMCACHED_DATA_EXISTS) {
    // Another thread has uploaded review_id, which is an unexpected behaviour.
    LOG(warning) << "review_id of request " << req_id
                 << " has already been stored";
    size_t value_size;
    char *counter_value_str = memcached_get(
        memcached_client,
        key_counter.c_str(),
        key_counter.size(),
        &value_size,
        nullptr,
        &memcached_rc);

    if (memcached_rc != MEMCACHED_SUCCESS) {
      LOG(error) << "Cannot get the counter of request " << req_id;
      ServiceException se;
      se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
      se.message = memcached_strerror(memcached_client, memcached_rc);
      memcached_pool_push(_memcached_client_pool, memcached_client);
      throw se;
    }
    counter_value = std::stoul(counter_value_str);
    free(counter_value_str);
  } else if (memcached_rc != MEMCACHED_SUCCESS) {
    LOG(error) << "Cannot store review_id of request " << req_id;
    ServiceException se;
    se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
    se.message = memcached_strerror(memcached_client, memcached_rc);
    memcached_pool_push(_memcached_client_pool, memcached_client);
    throw se;
  } else {
    // Atomically increment and get the counter value
    memcached_increment(
        memcached_client,
        key_counter.c_str(),
        key_counter.size(),
        1, &counter_value);
    if (memcached_rc != MEMCACHED_SUCCESS) {
      LOG(error) << "Cannot increment and get the counter of request "
                 << req_id;
      ServiceException se;
      se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
      se.message = memcached_strerror(memcached_client, memcached_rc);
      memcached_pool_push(_memcached_client_pool, memcached_client);
      throw se;
    }
  }
  LOG(debug) << "req_id " << req_id
             << " caching review_id to Memcached finished";

  memcached_pool_push(_memcached_client_pool, memcached_client);

  // If this thread is the last one uploading the review components,
  // it is in charge of compose the request and upload to the microservices in
  // the next tier.
  if (counter_value == NUM_COMPONENTS) {
    _ComposeAndUpload(req_id, writer_text_map);
  }
  span->Finish();
}

void ComposeReviewHandler::UploadText(
    int64_t req_id,
    const std::string &text,
    const std::map<std::string, std::string> & carrier) {

  // Initialize a span
  TextMapReader reader(carrier);
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "UploadText",
      { opentracing::ChildOf(parent_span->get()) });
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  memcached_return_t memcached_rc;
  std::string key_counter = std::to_string(req_id) + ":counter";
  memcached_st *memcached_client = memcached_pool_pop(
      _memcached_client_pool, true, &memcached_rc);

  // Initialize the counter to 0 if there it is not in the memcached
  memcached_rc = memcached_add(
      memcached_client,
      key_counter.c_str(),
      key_counter.size(),
      "0", 1, MMC_EXP_TIME, 0);

  // error if it cannot be stored
  if (memcached_rc != MEMCACHED_SUCCESS &&
      memcached_rc != MEMCACHED_DATA_EXISTS) {
    LOG(error) << "Failed to initilize the counter for request " << req_id
               << " Error code: "
               << memcached_strerror(memcached_client, memcached_rc);
    ServiceException se;
    se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
    se.message = memcached_strerror(memcached_client, memcached_rc);
    memcached_pool_push(_memcached_client_pool, memcached_client);
    throw se;
  }

  // Store text to memcached
  uint64_t counter_value;
  std::string key_text = std::to_string(req_id) + ":text";
  memcached_rc = memcached_add(
      memcached_client,
      key_text.c_str(),
      key_text.size(),
      text.c_str(),
      text.size(),
      MMC_EXP_TIME, 0);
  if (memcached_rc == MEMCACHED_DATA_EXISTS) {
    // Another thread has uploaded text, which is an unexpected behaviour.
    LOG(warning) << "text of request " << req_id
                 << " has already been stored";
    size_t value_size;
    char *counter_value_str = memcached_get(
        memcached_client,
        key_counter.c_str(),
        key_counter.size(),
        &value_size,
        nullptr,
        &memcached_rc);
    if (memcached_rc != MEMCACHED_SUCCESS) {
      LOG(error) << "Cannot get the counter of request " << req_id;
      ServiceException se;
      se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
      se.message = memcached_strerror(memcached_client, memcached_rc);
      memcached_pool_push(_memcached_client_pool, memcached_client);
      throw se;
    }
    counter_value = std::stoul(counter_value_str);
    free(counter_value_str);
  } else if (memcached_rc != MEMCACHED_SUCCESS) {
    LOG(error) << "Cannot store text of request " << req_id;
    ServiceException se;
    se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
    se.message = memcached_strerror(memcached_client, memcached_rc);
    memcached_pool_push(_memcached_client_pool, memcached_client);
    throw se;
  } else {
    // Atomically increment and get the counter value
    memcached_increment(
        memcached_client,
        key_counter.c_str(),
        key_counter.size(),
        1, &counter_value);
    if (memcached_rc != MEMCACHED_SUCCESS) {
      LOG(error) << "Cannot increment and get the counter of request "
                 << req_id;
      ServiceException se;
      se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
      se.message = memcached_strerror(memcached_client, memcached_rc);
      memcached_pool_push(_memcached_client_pool, memcached_client);
      throw se;
    }
  }
  LOG(debug) << "req_id " << req_id << "caching text to Memcached finished";
  memcached_pool_push(_memcached_client_pool, memcached_client);

  // If this thread is the last one uploading the review components,
  // it is in charge of compose the request and upload to the microservices in
  // the next tier.
  if (counter_value == NUM_COMPONENTS) {
    _ComposeAndUpload(req_id, writer_text_map);
  }
  span->Finish();
}

void ComposeReviewHandler::UploadRating(
    int64_t req_id, int32_t rating, const std::map<std::string, std::string> & carrier) {

  // Initialize a span
  TextMapReader reader(carrier);
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "UploadRating",
      { opentracing::ChildOf(parent_span->get()) });
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  memcached_return_t memcached_rc;
  std::string key_counter = std::to_string(req_id) + ":counter";
  memcached_st *memcached_client = memcached_pool_pop(
      _memcached_client_pool, true, &memcached_rc);

  // Initialize the counter to 0 if there it is not in the memcached
  memcached_rc = memcached_add(
      memcached_client,
      key_counter.c_str(),
      key_counter.size(),
      "0", 1, MMC_EXP_TIME, 0);

  // error if it cannot be stored
  if (memcached_rc != MEMCACHED_SUCCESS &&
      memcached_rc != MEMCACHED_DATA_EXISTS) {
    LOG(error) << "Failed to initilize the counter for request " << req_id
               << " Error code: " << memcached_strerror(memcached_client, memcached_rc);
    ServiceException se;
    se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
    se.message = memcached_strerror(memcached_client, memcached_rc);
    memcached_pool_push(_memcached_client_pool, memcached_client);
    throw se;
  }

  // Store rating to memcached
  uint64_t counter_value;
  std::string key_rating = std::to_string(req_id) + ":rating";
  std::string rating_str = std::to_string(rating);
  memcached_rc = memcached_add(
      memcached_client,
      key_rating.c_str(),
      key_rating.size(),
      rating_str.c_str(),
      rating_str.size(),
      MMC_EXP_TIME, 0);
  if (memcached_rc == MEMCACHED_DATA_EXISTS) {
    // Another thread has uploaded rating, which is an unexpected behaviour.
    LOG(warning) << "rating of request " << req_id
                 << " has already been stored";
    size_t value_size;
    char *counter_value_str = memcached_get(
        memcached_client,
        key_counter.c_str(),
        key_counter.size(),
        &value_size,
        0,
        &memcached_rc);
    counter_value = std::stoul(counter_value_str);
    free(counter_value_str);
    if (memcached_rc != MEMCACHED_SUCCESS) {
      LOG(error) << "Cannot get the counter of request " << req_id;
      ServiceException se;
      se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
      se.message = memcached_strerror(memcached_client, memcached_rc);
      memcached_pool_push(_memcached_client_pool, memcached_client);
      throw se;
    }
  } else if (memcached_rc != MEMCACHED_SUCCESS) {
    LOG(error) << "Cannot store rating of request " << req_id;
    ServiceException se;
    se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
    se.message = memcached_strerror(memcached_client, memcached_rc);
    memcached_pool_push(_memcached_client_pool, memcached_client);
    throw se;
  } else {
    // Atomically increment and get the counter value
    memcached_increment(
        memcached_client,
        key_counter.c_str(),
        key_counter.size(),
        1, &counter_value);
    if (memcached_rc != MEMCACHED_SUCCESS) {
      LOG(error) << "Cannot increment and get the counter of request "
                 << req_id;
      ServiceException se;
      se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
      se.message = memcached_strerror(memcached_client, memcached_rc);
      memcached_pool_push(_memcached_client_pool, memcached_client);
      throw se;
    }
  }
  LOG(debug) << "req_id " << req_id << " caching rating to Memcached finished";
  memcached_pool_push(_memcached_client_pool, memcached_client);

  // If this thread is the last one uploading the review components,
  // it is in charge of compose the request and upload to the microservices in
  // the next tier.
  if (counter_value == NUM_COMPONENTS) {
    _ComposeAndUpload(req_id, writer_text_map);
  }
  span->Finish();
}

} // namespace media_service


#endif //MEDIA_MICROSERVICES_COMPOSEREVIEWHANDLER_H
