#ifndef MEDIA_MICROSERVICES_CASTINFOHANDLER_H
#define MEDIA_MICROSERVICES_CASTINFOHANDLER_H

#include <iostream>
#include <string>
#include <future>

#include <mongoc.h>
#include <libmemcached/memcached.h>
#include <libmemcached/util.h>
#include <bson/bson.h>

#include "../../gen-cpp/CastInfoService.h"
#include "../ClientPool.h"
#include "../ThriftClient.h"
#include "../logger.h"
#include "../tracing.h"

namespace media_service {

class CastInfoHandler : public CastInfoServiceIf {
 public:
  CastInfoHandler(
      memcached_pool_st *,
      mongoc_client_pool_t *);
  ~CastInfoHandler() override = default;

  void WriteCastInfo(int64_t req_id, int64_t cast_info_id,
      const std::string &name, bool gender, const std::string &intro,
      const std::map<std::string, std::string>& carrier) override;

  void ReadCastInfo(std::vector<CastInfo>& _return, int64_t req_id,
      const std::vector<int64_t> & cast_info_ids,
      const std::map<std::string, std::string>& carrier) override;

 private:
  memcached_pool_st *_memcached_client_pool;
  mongoc_client_pool_t *_mongodb_client_pool;
};

CastInfoHandler::CastInfoHandler(
    memcached_pool_st *memcached_client_pool,
    mongoc_client_pool_t *mongodb_client_pool) {
  _memcached_client_pool = memcached_client_pool;
  _mongodb_client_pool = mongodb_client_pool;
}
void CastInfoHandler::WriteCastInfo(
    int64_t req_id,
    int64_t cast_info_id,
    const std::string &name,
    bool gender,
    const std::string &intro,
    const std::map<std::string, std::string> &carrier) {
  // Initialize a span
  TextMapReader reader(carrier);
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "WriteCastInfo",
      { opentracing::ChildOf(parent_span->get()) });
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  bson_t *new_doc = bson_new();
  BSON_APPEND_INT64(new_doc, "cast_info_id", cast_info_id);
  BSON_APPEND_UTF8(new_doc, "name", name.c_str());
  BSON_APPEND_BOOL(new_doc, "gender", gender);
  BSON_APPEND_UTF8(new_doc, "intro", intro.c_str());

  mongoc_client_t *mongodb_client = mongoc_client_pool_pop(
      _mongodb_client_pool);
  if (!mongodb_client) {
    ServiceException se;
    se.errorCode = ErrorCode::SE_MONGODB_ERROR;
    se.message = "Failed to pop a client from MongoDB pool";
    throw se;
  }
  auto collection = mongoc_client_get_collection(
      mongodb_client, "cast-info", "cast-info");
  if (!collection) {
    ServiceException se;
    se.errorCode = ErrorCode::SE_MONGODB_ERROR;
    se.message = "Failed to create collection cast-info from DB cast-info";
    mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
    throw se;
  }

  bson_error_t error;
  auto insert_span = opentracing::Tracer::Global()->StartSpan(
      "MongoInsertCastInfo", { opentracing::ChildOf(&span->context()) });
  bool plotinsert = mongoc_collection_insert_one (
      collection, new_doc, nullptr, nullptr, &error);
  insert_span->Finish();
  if (!plotinsert) {
    LOG(error) << "Error: Failed to insert cast-info to MongoDB: "
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

void CastInfoHandler::ReadCastInfo(
    std::vector<CastInfo> &_return,
    int64_t req_id,
    const std::vector<int64_t> &cast_info_ids,
    const std::map<std::string, std::string> &carrier) {

  // Initialize a span
  TextMapReader reader(carrier);
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "ReadCastInfo",
      { opentracing::ChildOf(parent_span->get()) });
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  if (cast_info_ids.empty()) {
    return;
  }

  std::set<int64_t> cast_info_ids_not_cached(cast_info_ids.begin(), cast_info_ids.end());
  if (cast_info_ids_not_cached.size() != cast_info_ids.size()) {
    ServiceException se;
    se.errorCode = ErrorCode::SE_THRIFT_HANDLER_ERROR;
    se.message = "cast_info_ids are duplicated";
    throw se;
  }

  std::map<int64_t, CastInfo> return_map;
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
  keys = new char* [cast_info_ids.size()];
  key_sizes = new size_t [cast_info_ids.size()];
  int idx = 0;
  for (auto &cast_info_id : cast_info_ids) {
    std::string key_str = std::to_string(cast_info_id);
    keys[idx] = new char [key_str.length() + 1];
    strcpy(keys[idx], key_str.c_str());
    key_sizes[idx] = key_str.length();
    idx++;
  }
  memcached_rc = memcached_mget(memcached_client, keys, key_sizes, cast_info_ids.size());
  if (memcached_rc != MEMCACHED_SUCCESS) {
    LOG(error) << "Cannot get cast_info_ids of request " << req_id << ": "
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
      "MmcMgetCastInfo", { opentracing::ChildOf(&span->context()) });
  while (true) {
    return_value = memcached_fetch(memcached_client, return_key,
        &return_key_length, &return_value_length, &flags, &memcached_rc);
    if (return_value == nullptr) {
      LOG(debug) << "Memcached mget finished";
      break;
    }
    if (memcached_rc != MEMCACHED_SUCCESS) {
      free(return_value);
      memcached_quit(memcached_client);
      memcached_pool_push(_memcached_client_pool, memcached_client);
      LOG(error) << "Cannot get components of request " << req_id;
      ServiceException se;
      se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
      se.message =  "Cannot get usernames of request " + std::to_string(req_id);
      throw se;
    }
    CastInfo new_cast_info;
    json cast_info_json = json::parse(std::string(
        return_value, return_value + return_value_length));
    new_cast_info.cast_info_id = cast_info_json["cast_info_id"];
    new_cast_info.gender = cast_info_json["gender"];
    new_cast_info.name = cast_info_json["name"];
    new_cast_info.intro = cast_info_json["intro"];
    return_map.insert(std::make_pair(new_cast_info.cast_info_id, new_cast_info));
    cast_info_ids_not_cached.erase(new_cast_info.cast_info_id);
    free(return_value);
  }
  get_span->Finish();
  memcached_quit(memcached_client);
  memcached_pool_push(_memcached_client_pool, memcached_client);
  for (int i = 0; i < cast_info_ids.size(); ++i) {
    delete keys[i];
  }
  delete[] keys;
  delete[] key_sizes;

  std::vector<std::future<void>> set_futures;
  std::map<int64_t, std::string> cast_info_json_map;

  // Find the rest in MongoDB
  if (!cast_info_ids_not_cached.empty()) {
    bson_t *query = bson_new();
    bson_t query_child;
    bson_t query_cast_info_id_list;
    const char *key;
    idx = 0;
    char buf[16];
    BSON_APPEND_DOCUMENT_BEGIN(query, "cast_info_id", &query_child);
    BSON_APPEND_ARRAY_BEGIN(&query_child, "$in", &query_cast_info_id_list);
    for (auto &item : cast_info_ids_not_cached) {
      bson_uint32_to_string(idx, &key, buf, sizeof buf);
      BSON_APPEND_INT64(&query_cast_info_id_list, key, item);
      idx++;
    }
    bson_append_array_end(&query_child, &query_cast_info_id_list);
    bson_append_document_end(query, &query_child);

    mongoc_client_t *mongodb_client = mongoc_client_pool_pop(
        _mongodb_client_pool);
    if (!mongodb_client) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_MONGODB_ERROR;
      se.message = "Failed to pop a client from MongoDB pool";
      throw se;
    }
    auto collection = mongoc_client_get_collection(
        mongodb_client, "cast-info", "cast-info");
    if (!collection) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_MONGODB_ERROR;
      se.message = "Failed to create collection user from DB user";
      mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
      throw se;
    }

    mongoc_cursor_t *cursor = mongoc_collection_find_with_opts(
        collection, query, nullptr, nullptr);
    const bson_t *doc;

    auto find_span = opentracing::Tracer::Global()->StartSpan(
        "MongoFindCastInfo", {opentracing::ChildOf(&span->context())});

    while (true) {
      bool found = mongoc_cursor_next(cursor, &doc);
      if (!found) {
        break;
      }
      bson_iter_t iter;
      CastInfo new_cast_info;
      char *cast_info_json_char = bson_as_json(doc, nullptr);
      json cast_info_json = json::parse(cast_info_json_char);
      new_cast_info.cast_info_id = cast_info_json["cast_info_id"];
      new_cast_info.gender = cast_info_json["gender"];
      new_cast_info.name = cast_info_json["name"];
      new_cast_info.intro = cast_info_json["intro"];
      cast_info_json_map.insert({
        new_cast_info.cast_info_id, std::string(cast_info_json_char)});
      return_map.insert({new_cast_info.cast_info_id, new_cast_info});
      bson_free(cast_info_json_char);
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

    // Upload cast-info to memcached
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
          "MmcSetCastInfo", {opentracing::ChildOf(&span->context())});
      for (auto & it : cast_info_json_map) {
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

  if (return_map.size() != cast_info_ids.size()) {
    try {
      for (auto &it : set_futures) { it.get(); }
    } catch (...) {
      LOG(warning) << "Failed to set cast-info to memcached";
    }
    LOG(error) << "cast-info-service return set incomplete";
    ServiceException se;
    se.errorCode = ErrorCode::SE_THRIFT_HANDLER_ERROR;
    se.message = "cast-info-service return set incomplete";
    throw se;
  }

  for (auto &cast_info_id : cast_info_ids) {
    _return.emplace_back(return_map[cast_info_id]);
  }

  try {
    for (auto &it : set_futures) { it.get(); }
  } catch (...) {
    LOG(warning) << "Failed to set cast-info to memcached";
  }
}

} // namespace media_service

#endif //MEDIA_MICROSERVICES_CASTINFOHANDLER_H
