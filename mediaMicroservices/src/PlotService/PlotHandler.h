#ifndef MEDIA_MICROSERVICES_PLOTHANDLER_H
#define MEDIA_MICROSERVICES_PLOTHANDLER_H

#include <iostream>
#include <string>

#include <libmemcached/memcached.h>
#include <libmemcached/util.h>
#include <mongoc.h>
#include <bson/bson.h>

#include "../../gen-cpp/PlotService.h"
#include "../logger.h"
#include "../tracing.h"

namespace media_service {

class PlotHandler : public PlotServiceIf {
 public:
  PlotHandler(
      memcached_pool_st *,
      mongoc_client_pool_t *);
  ~PlotHandler() override = default;

  void WritePlot(int64_t req_id, int64_t plot_id, const std::string& plot,
      const std::map<std::string, std::string> & carrier) override;
  void ReadPlot(std::string& _return, int64_t req_id, int64_t plot_id,
      const std::map<std::string, std::string> & carrier) override;

 private:
  memcached_pool_st *_memcached_client_pool;
  mongoc_client_pool_t *_mongodb_client_pool;
};

PlotHandler::PlotHandler(
    memcached_pool_st *memcached_client_pool,
    mongoc_client_pool_t *mongodb_client_pool) {
  _memcached_client_pool = memcached_client_pool;
  _mongodb_client_pool = mongodb_client_pool;
}

void PlotHandler::ReadPlot(
    std::string &_return,
    int64_t req_id,
    int64_t plot_id,
    const std::map<std::string, std::string> & carrier) {

  // Initialize a span
  TextMapReader reader(carrier);
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "ReadPlot",
      { opentracing::ChildOf(parent_span->get()) });
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  memcached_return_t memcached_rc;
  memcached_st *memcached_client = memcached_pool_pop(
      _memcached_client_pool, true, &memcached_rc);
  if (!memcached_client) {
    ServiceException se;
    se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
    se.message = "Failed to pop a client from memcached pool";
    throw se;
  }

  size_t plot_size;
  uint32_t memcached_flags;

  // Look for the movie id from memcached
  auto get_span = opentracing::Tracer::Global()->StartSpan(
      "MmcGetPlot", { opentracing::ChildOf(&span->context()) });
  auto plot_id_str = std::to_string(plot_id);

  char* plot_mmc = memcached_get(
      memcached_client,
      plot_id_str.c_str(),
      plot_id_str.length(),
      &plot_size,
      &memcached_flags,
      &memcached_rc);
  if (!plot_mmc && memcached_rc != MEMCACHED_NOTFOUND) {
    ServiceException se;
    se.errorCode = ErrorCode::SE_MEMCACHED_ERROR;
    se.message = memcached_strerror(memcached_client, memcached_rc);
    memcached_pool_push(_memcached_client_pool, memcached_client);
    throw se;
  }
  get_span->Finish();
  memcached_pool_push(_memcached_client_pool, memcached_client);

  // If cached in memcached
  if (plot_mmc) {
    LOG(debug) << "Get plot " << plot_mmc
        << " cache hit from Memcached";
    _return = std::string(plot_mmc);
    free(plot_mmc);
  } else {
    // If not cached in memcached
    mongoc_client_t *mongodb_client = mongoc_client_pool_pop(
        _mongodb_client_pool);
    if (!mongodb_client) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_MONGODB_ERROR;
      se.message = "Failed to pop a client from MongoDB pool";
      free(plot_mmc);
      throw se;
    }
    auto collection = mongoc_client_get_collection(
        mongodb_client, "plot", "plot");
    if (!collection) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_MONGODB_ERROR;
      se.message = "Failed to create collection plot from DB plot";
      free(plot_mmc);
      mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
      throw se;
    }

    bson_t *query = bson_new();
    BSON_APPEND_INT64(query, "plot_id", plot_id);

    auto find_span = opentracing::Tracer::Global()->StartSpan(
        "MongoFindPlot", { opentracing::ChildOf(&span->context()) });
    mongoc_cursor_t *cursor = mongoc_collection_find_with_opts(
        collection, query, nullptr, nullptr);
    const bson_t *doc;
    bool found = mongoc_cursor_next(cursor, &doc);
    find_span->Finish();

    if (found) {
      bson_iter_t iter;
      if (bson_iter_init_find(&iter, doc, "plot")) {
        char *plot_mongo_char = bson_iter_value(&iter)->value.v_utf8.str;
        size_t plot_mongo_len = bson_iter_value(&iter)->value.v_utf8.len;
        LOG(debug) << "Find plot " << plot_id << " cache miss";
        _return = std::string(plot_mongo_char, plot_mongo_char + plot_mongo_len);
        bson_destroy(query);
        mongoc_cursor_destroy(cursor);
        mongoc_collection_destroy(collection);
        mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
        memcached_client = memcached_pool_pop(
            _memcached_client_pool, true, &memcached_rc);

        // Upload the plot to memcached
        auto set_span = opentracing::Tracer::Global()->StartSpan(
            "MmcSetPlot", { opentracing::ChildOf(&span->context()) });
        memcached_rc = memcached_set(
            memcached_client,
            plot_id_str.c_str(),
            plot_id_str.length(),
            _return.c_str(),
            _return.length(),
            static_cast<time_t>(0),
            static_cast<uint32_t>(0)
        );
        set_span->Finish();

        if (memcached_rc != MEMCACHED_SUCCESS) {
          LOG(warning) << "Failed to set plot to Memcached: "
              << memcached_strerror(memcached_client, memcached_rc);
        }
        memcached_pool_push(_memcached_client_pool, memcached_client);
      } else {
        LOG(error) << "Attribute plot is not find in MongoDB";
        bson_destroy(query);
        mongoc_cursor_destroy(cursor);
        mongoc_collection_destroy(collection);
        mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
        ServiceException se;
        se.errorCode = ErrorCode::SE_THRIFT_HANDLER_ERROR;
        se.message = "Attribute plot is not find in MongoDB";
        free(plot_mmc);
        throw se;
      }
    } else {
      LOG(error) << "Plot_id " << plot_id << " is not found in MongoDB";
      bson_destroy(query);
      mongoc_cursor_destroy(cursor);
      mongoc_collection_destroy(collection);
      mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
      ServiceException se;
      se.errorCode = ErrorCode::SE_THRIFT_HANDLER_ERROR;
      se.message = "Plot_id " + plot_id_str + " is not found in MongoDB";
      free(plot_mmc);
      throw se;
    }
  }
  span->Finish();
}

void PlotHandler::WritePlot(
    int64_t req_id,
    int64_t plot_id,
    const std::string &plot,
    const std::map<std::string, std::string> &carrier) {
  // Initialize a span
  TextMapReader reader(carrier);
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "WritePlot",
      { opentracing::ChildOf(parent_span->get()) });
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  bson_t *new_doc = bson_new();
  BSON_APPEND_INT64(new_doc, "plot_id", plot_id);
  BSON_APPEND_UTF8(new_doc, "plot", plot.c_str());

  mongoc_client_t *mongodb_client = mongoc_client_pool_pop(
      _mongodb_client_pool);
  if (!mongodb_client) {
    ServiceException se;
    se.errorCode = ErrorCode::SE_MONGODB_ERROR;
    se.message = "Failed to pop a client from MongoDB pool";
    throw se;
  }
  auto collection = mongoc_client_get_collection(
      mongodb_client, "plot", "plot");
  if (!collection) {
    ServiceException se;
    se.errorCode = ErrorCode::SE_MONGODB_ERROR;
    se.message = "Failed to create collection plot from DB plot";
    mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
    throw se;
  }
  bson_error_t error;
  auto insert_span = opentracing::Tracer::Global()->StartSpan(
      "MongoInsertPlot", { opentracing::ChildOf(&span->context()) });
  bool plotinsert = mongoc_collection_insert_one (
      collection, new_doc, nullptr, nullptr, &error);
  insert_span->Finish();
  if (!plotinsert) {
    LOG(error) << "Error: Failed to insert plot to MongoDB: "
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

} // namespace media_service

#endif //MEDIA_MICROSERVICES_PLOTHANDLER_H
