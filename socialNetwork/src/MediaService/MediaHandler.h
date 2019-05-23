#ifndef SOCIAL_NETWORK_MICROSERVICES_SRC_MEDIASERVICE_MEDIAHANDLER_H_
#define SOCIAL_NETWORK_MICROSERVICES_SRC_MEDIASERVICE_MEDIAHANDLER_H_

#include <iostream>
#include <string>
#include <chrono>
#include <mutex>
#include <sstream>
#include <iomanip>
#include <arpa/inet.h>
#include <net/if.h>
#include <sys/ioctl.h>
#include <sys/socket.h>

#include "../../gen-cpp/MediaService.h"
#include "../../gen-cpp/ComposePostService.h"
#include "../ClientPool.h"
#include "../ThriftClient.h"
#include "../logger.h"
#include "../tracing.h"

#define CUSTOM_EPOCH 1514764800000

namespace social_network {

using std::chrono::milliseconds;
using std::chrono::duration_cast;
using std::chrono::system_clock;


class MediaHandler : public MediaServiceIf {
 public:
  explicit MediaHandler(ClientPool<ThriftClient<ComposePostServiceClient>> *);
  ~MediaHandler() override = default;

  void UploadMedia(int64_t, const std::vector<std::string> &,
      const std::vector<int64_t> &, const std::map<std::string,
      std::string> &) override;

 private:
  ClientPool<ThriftClient<ComposePostServiceClient>> *_compose_client_pool;
};

MediaHandler::MediaHandler(
    ClientPool<ThriftClient<ComposePostServiceClient>> *compose_client_pool) {
  _compose_client_pool = compose_client_pool;
}

void MediaHandler::UploadMedia(
    int64_t req_id,
    const std::vector<std::string> &media_types,
    const std::vector<int64_t> &media_ids,
    const std::map<std::string, std::string> &carrier) {

  // Initialize a span
  TextMapReader reader(carrier);
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "UploadMedia",
      { opentracing::ChildOf(parent_span->get()) });
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  if (media_types.size() != media_ids.size()) {
    ServiceException se;
    se.errorCode = ErrorCode::SE_THRIFT_HANDLER_ERROR;
    se.message = "The lengths of media_id list and media_type list are not equal";
    throw se;
  }

  std::vector<Media> media;
  for (int i = 0; i < media_ids.size(); ++i) {
    Media new_media;
    new_media.media_id = media_ids[i];
    new_media.media_type = media_types[i];
    media.emplace_back(new_media);
  }

  // Upload to compose post service
  auto compose_post_client_wrapper = _compose_client_pool->Pop();
  if (!compose_post_client_wrapper) {
    ServiceException se;
    se.errorCode = ErrorCode::SE_THRIFT_CONN_ERROR;
    se.message = "Failed to connect to compose-post-service";
    throw se;
  }
  auto compose_post_client = compose_post_client_wrapper->GetClient();
  try {
    compose_post_client->UploadMedia(req_id, media, writer_text_map);
  } catch (...) {
    _compose_client_pool->Push(compose_post_client_wrapper);
    LOG(error) << "Failed to upload media to compose-post-service";
    throw;
  }
  _compose_client_pool->Push(compose_post_client_wrapper);
  span->Finish();

}

} //namespace social_network

#endif //SOCIAL_NETWORK_MICROSERVICES_SRC_MEDIASERVICE_MEDIAHANDLER_H_
