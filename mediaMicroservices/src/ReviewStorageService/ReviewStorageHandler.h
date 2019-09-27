#ifndef MEDIA_MICROSERVICES_REVIEWSTOREHANDLER_H
#define MEDIA_MICROSERVICES_REVIEWSTOREHANDLER_H

#include <iostream>
#include <string>
#include <future>

#include <mongoc.h>
#include <libmemcached/memcached.h>
#include <libmemcached/util.h>
#include <bson/bson.h>

#include "../../gen-cpp/ReviewStorageService.h"
#include "../logger.h"
#include "../tracing.h"

namespace media_service {

class ReviewStorageHandler : public ReviewStorageServiceIf{
 public:
  ReviewStorageHandler(memcached_pool_st *, mongoc_client_pool_t *);
  ~ReviewStorageHandler() override = default;
  void StoreReview(int64_t, const Review &, 
      const std::map<std::string, std::string> &) override;
  void ReadReviews(std::vector<Review> &, int64_t, const std::vector<int64_t> &,
                   const std::map<std::string, std::string> &) override;
  
 private:
  memcached_pool_st *_memcached_client_pool;
  mongoc_client_pool_t *_mongodb_client_pool;
};

ReviewStorageHandler::ReviewStorageHandler(
    memcached_pool_st *memcached_pool,
    mongoc_client_pool_t *mongodb_pool) {
  _memcached_client_pool = memcached_pool;
  _mongodb_client_pool = mongodb_pool;
}

void ReviewStorageHandler::StoreReview(
    int64_t req_id, 
    const Review &review,
    const std::map<std::string, std::string> & carrier) {

  // Initialize a span
  TextMapReader reader(carrier);
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "StoreReview",
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
      mongodb_client, "review", "review");
  if (!collection) {
    ServiceException se;
    se.errorCode = ErrorCode::SE_MONGODB_ERROR;
    se.message = "Failed to create collection user from DB user";
    mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
    throw se;
  }

  bson_t *new_doc = bson_new();
  BSON_APPEND_INT64(new_doc, "review_id", review.review_id);
  BSON_APPEND_INT64(new_doc, "timestamp", review.timestamp);
  BSON_APPEND_INT64(new_doc, "user_id", review.user_id);
  BSON_APPEND_UTF8(new_doc, "movie_id", review.movie_id.c_str());
  BSON_APPEND_UTF8(new_doc, "text", review.text.c_str());
  BSON_APPEND_INT32(new_doc, "rating", review.rating);
  BSON_APPEND_INT64(new_doc, "req_id", review.req_id);
  bson_error_t error;

  auto insert_span = opentracing::Tracer::Global()->StartSpan(
      "MongoInsertReview", { opentracing::ChildOf(&span->context()) });
  bool plotinsert = mongoc_collection_insert_one (
      collection, new_doc, nullptr, nullptr, &error);
  insert_span->Finish();

  if (!plotinsert) {
    LOG(error) << "Error: Failed to insert review to MongoDB: "
        << error.message;
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
void ReviewStorageHandler::ReadReviews(
    std::vector<Review> & _return,
    int64_t req_id,
    const std::vector<int64_t> &review_ids,
    const std::map<std::string, std::string> &carrier) {

  // Initialize a span
  TextMapReader reader(carrier);
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "ReadReviews",
      { opentracing::ChildOf(parent_span->get()) });
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  if (review_ids.empty()) {
    return;
  }

  std::set<int64_t> review_ids_not_cached(review_ids.begin(), review_ids.end());
  if (review_ids_not_cached.size() != review_ids.size()) {
    ServiceException se;
    se.errorCode = ErrorCode::SE_THRIFT_HANDLER_ERROR;
    se.message = "Post_ids are duplicated";
    throw se;
  }
  std::map<int64_t, Review> return_map;
  memcached_return_t memcached_rc;
  auto memcached_client = memcached_pool_pop(
      _memcached_client_pool, true, &memcached_rc);
  if (!memcached_client) {
    ServiceException se;
    se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
    se.message = "Failed to pop a client from memcached pool";
    throw se;
  }

  char** keys;
  size_t *key_sizes;
  keys = new char* [review_ids.size()];
  key_sizes = new size_t [review_ids.size()];
  int idx = 0;
  for (auto &review_id : review_ids) {
    std::string key_str = std::to_string(review_id);
    keys[idx] = new char [key_str.length() + 1];
    strcpy(keys[idx], key_str.c_str());
    key_sizes[idx] = key_str.length();
    idx++;
  }
  memcached_rc = memcached_mget(
      memcached_client, keys, key_sizes, review_ids.size());
  if (memcached_rc != MEMCACHED_SUCCESS) {
    LOG(error) << "Cannot get review-ids of request " << req_id << ": "
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
      "MemcachedMget", { opentracing::ChildOf(&span->context()) });

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
      LOG(error) << "Cannot get reviews of request " << req_id;
      ServiceException se;
      se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
      se.message = "Cannot get reviews of request " + std::to_string(req_id);
      throw se;
    }
    Review new_review;
    json review_json = json::parse(std::string(
        return_value, return_value + return_value_length));
    new_review.req_id = review_json["req_id"];
    new_review.user_id = review_json["user_id"];
    new_review.movie_id = review_json["movie_id"];
    new_review.text = review_json["text"];
    new_review.rating = review_json["rating"];
    new_review.timestamp = review_json["timestamp"];
    new_review.review_id = review_json["review_id"];
    return_map.insert(std::make_pair(new_review.review_id, new_review));
    review_ids_not_cached.erase(new_review.review_id);
    free(return_value);
    LOG(debug) << "Review: " << new_review.review_id << " found in memcached";
  }
  get_span->Finish();
  memcached_quit(memcached_client);
  memcached_pool_push(_memcached_client_pool, memcached_client);
  for (int i = 0; i < review_ids.size(); ++i) {
    delete keys[i];
  }
  delete[] keys;
  delete[] key_sizes;

  std::vector<std::future<void>> set_futures;
  std::map<int64_t, std::string> review_json_map;
  
  // Find the rest in MongoDB
  if (!review_ids_not_cached.empty()) {
    mongoc_client_t *mongodb_client = mongoc_client_pool_pop(
        _mongodb_client_pool);
    if (!mongodb_client) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_MONGODB_ERROR;
      se.message = "Failed to pop a client from MongoDB pool";
      throw se;
    }
    auto collection = mongoc_client_get_collection(
        mongodb_client, "review", "review");
    if (!collection) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_MONGODB_ERROR;
      se.message = "Failed to create collection user from DB user";
      mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
      throw se;
    }
    bson_t *query = bson_new();
    bson_t query_child;
    bson_t query_review_id_list;
    const char *key;
    idx = 0;
    char buf[16];
    BSON_APPEND_DOCUMENT_BEGIN(query, "review_id", &query_child);
    BSON_APPEND_ARRAY_BEGIN(&query_child, "$in", &query_review_id_list);
    for (auto &item : review_ids_not_cached) {
      bson_uint32_to_string(idx, &key, buf, sizeof buf);
      BSON_APPEND_INT64(&query_review_id_list, key, item);
      idx++;
    }
    bson_append_array_end(&query_child, &query_review_id_list);
    bson_append_document_end(query, &query_child);
    mongoc_cursor_t *cursor = mongoc_collection_find_with_opts(
        collection, query, nullptr, nullptr);
    const bson_t *doc;
    auto find_span = opentracing::Tracer::Global()->StartSpan(
        "MongoFindPosts", {opentracing::ChildOf(&span->context())});
    while (true) {
      bool found = mongoc_cursor_next(cursor, &doc);
      if (!found) {
        break;
      }
      Review new_review;
      char *review_json_char = bson_as_json(doc, nullptr);
      json review_json = json::parse(review_json_char);
      new_review.req_id = review_json["req_id"];
      new_review.user_id = review_json["user_id"];
      new_review.movie_id = review_json["movie_id"];
      new_review.text = review_json["text"];
      new_review.rating = review_json["rating"];
      new_review.timestamp = review_json["timestamp"];
      new_review.review_id = review_json["review_id"];
      review_json_map.insert({new_review.review_id, std::string(review_json_char)});
      return_map.insert({new_review.review_id, new_review});
      bson_free(review_json_char);
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

    // upload reviews to memcached
    set_futures.emplace_back(std::async(std::launch::async, [&]() {
      memcached_return_t _rc;
      auto _memcached_client = memcached_pool_pop(
          _memcached_client_pool, true, &_rc);
      if (!_memcached_client) {
        LOG(error) << "Failed to pop a client from memcached pool";
        ServiceException se;
        se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
        se.message = "Failed to pop a client from memcached pool";
        throw se;
      }
      auto set_span = opentracing::Tracer::Global()->StartSpan(
          "MmcSetPost", {opentracing::ChildOf(&span->context())});
      for (auto & it : review_json_map) {
        std::string id_str = std::to_string(it.first);
        _rc = memcached_set(
            _memcached_client,
            id_str.c_str(),
            id_str.length(),
            it.second.c_str(),
            it.second.length(),
            static_cast<time_t>(0),
            static_cast<uint32_t>(0));
      }
      memcached_pool_push(_memcached_client_pool, _memcached_client);
      set_span->Finish();
    }));
  }

  if (return_map.size() != review_ids.size()) {
    try {
      for (auto &it : set_futures) { it.get(); }
    } catch (...) {
      LOG(warning) << "Failed to set reviews to memcached";
    }
    LOG(error) << "review storage service: return set incomplete";
    ServiceException se;
    se.errorCode = ErrorCode::SE_THRIFT_HANDLER_ERROR;
    se.message = "review storage service: return set incomplete";
    throw se;
  }

  for (auto &review_id : review_ids) {
    _return.emplace_back(return_map[review_id]);
  }

  try {
    for (auto &it : set_futures) { it.get(); }
  } catch (...) {
    LOG(warning) << "Failed to set reviews to memcached";
  }
  
}

} // namespace media_service


#endif //MEDIA_MICROSERVICES_REVIEWSTOREHANDLER_H
