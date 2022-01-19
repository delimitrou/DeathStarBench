ARG RESTY_IMAGE_BASE="ubuntu"
ARG RESTY_IMAGE_TAG="xenial"

FROM ${RESTY_IMAGE_BASE}:${RESTY_IMAGE_TAG}

LABEL maintainer="Yu Gan <yg397@cornell.edu>"

# Docker Build Arguments
ARG RESTY_VERSION="1.15.8.1rc1"
ARG RESTY_LUAROCKS_VERSION="3.5.0"
ARG RESTY_OPENSSL_VERSION="1.1.0j"
ARG RESTY_PCRE_VERSION="8.42"
ARG JAEGER_TRACING_VERSION="0.4.2"
ARG NGINX_OPENTRACING_VERSION="0.8.0"
ARG OPENTRACING_CPP_VERSION="1.5.1"
ARG RESTY_J="1"
ARG RESTY_CONFIG_OPTIONS="\
    --with-file-aio \
    --with-http_addition_module \
    --with-http_auth_request_module \
    --with-http_dav_module \
    --with-http_flv_module \
    --with-http_geoip_module=dynamic \
    --with-http_gunzip_module \
    --with-http_gzip_static_module \
    --with-http_image_filter_module=dynamic \
    --with-http_mp4_module \
    --with-http_random_index_module \
    --with-http_realip_module \
    --with-http_secure_link_module \
    --with-http_slice_module \
    --with-http_ssl_module \
    --with-http_stub_status_module \
    --with-http_sub_module \
    --with-http_v2_module \
    --with-http_xslt_module=dynamic \
    --with-ipv6 \
    --with-mail \
    --with-mail_ssl_module \
    --with-md5-asm \
    --with-pcre-jit \
    --with-sha1-asm \
    --with-stream \
    --with-stream_ssl_module \
    --with-threads \
    --add-dynamic-module=/usr/local/nginx-opentracing-${NGINX_OPENTRACING_VERSION}/opentracing \
    "
ARG RESTY_CONFIG_OPTIONS_MORE=""
ARG RESTY_ADD_PACKAGE_BUILDDEPS=""
ARG RESTY_ADD_PACKAGE_RUNDEPS=""
ARG RESTY_EVAL_PRE_CONFIGURE=""
ARG RESTY_EVAL_POST_MAKE=""

LABEL resty_version="${RESTY_VERSION}"
LABEL resty_luarocks_version="${RESTY_LUAROCKS_VERSION}"
LABEL resty_openssl_version="${RESTY_OPENSSL_VERSION}"
LABEL resty_pcre_version="${RESTY_PCRE_VERSION}"
LABEL resty_config_options="${RESTY_CONFIG_OPTIONS}"
LABEL resty_config_options_more="${RESTY_CONFIG_OPTIONS_MORE}"
LABEL resty_add_package_builddeps="${RESTY_ADD_PACKAGE_BUILDDEPS}"
LABEL resty_add_package_rundeps="${RESTY_ADD_PACKAGE_RUNDEPS}"
LABEL resty_eval_pre_configure="${RESTY_EVAL_PRE_CONFIGURE}"
LABEL resty_eval_post_make="${RESTY_EVAL_POST_MAKE}"

# These are not intended to be user-specified
ARG _RESTY_CONFIG_DEPS="--with-openssl=/tmp/openssl-${RESTY_OPENSSL_VERSION} --with-pcre=/tmp/pcre-${RESTY_PCRE_VERSION}"


# 1) Install apt dependencies
# 2) Download and untar OpenSSL, PCRE, and OpenResty
# 3) Build OpenResty
# 4) Cleanup

RUN DEBIAN_FRONTEND=noninteractive apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
        build-essential \
        ca-certificates \
        curl \
        git \
        gettext-base \
        libgd-dev \
        libgeoip-dev \
        libncurses5-dev \
        libperl-dev \
        libreadline-dev \
        libxslt1-dev \
        make \
        perl \
        unzip \
        zlib1g-dev \
        cmake \
        ${RESTY_ADD_PACKAGE_BUILDDEPS} \
        ${RESTY_ADD_PACKAGE_RUNDEPS} \
    && cd /tmp \
    && if [ -n "${RESTY_EVAL_PRE_CONFIGURE}" ]; then eval $(echo ${RESTY_EVAL_PRE_CONFIGURE}); fi \
    && curl -fSL https://www.openssl.org/source/openssl-${RESTY_OPENSSL_VERSION}.tar.gz -o openssl-${RESTY_OPENSSL_VERSION}.tar.gz \
    && tar xzf openssl-${RESTY_OPENSSL_VERSION}.tar.gz \
    && curl -fSL http://ftp.cs.stanford.edu/pub/exim/pcre/pcre-${RESTY_PCRE_VERSION}.tar.gz -o pcre-${RESTY_PCRE_VERSION}.tar.gz \
    && tar xzf pcre-${RESTY_PCRE_VERSION}.tar.gz \
  ### Build opentracing-cpp
    && cd /tmp \
    && curl -fSL https://github.com/opentracing/opentracing-cpp/archive/v${OPENTRACING_CPP_VERSION}.tar.gz -o opentracing-cpp-${OPENTRACING_CPP_VERSION}.tar.gz\
    && tar -zxf opentracing-cpp-${OPENTRACING_CPP_VERSION}.tar.gz \
    && cd opentracing-cpp-${OPENTRACING_CPP_VERSION} \
    && mkdir -p cmake-build \
    && cd cmake-build \
    && cmake -DCMAKE_BUILD_TYPE=Release \
             -DBUILD_MOCKTRACER=OFF \
             -DBUILD_STATIC_LIBS=OFF \
             -DBUILD_TESTING=OFF \
       .. \
    && make -j$(nproc) \
    && make -j$(nproc) install \
    && cd /tmp \
    && rm -rf opentracing-cpp-${OPENTRACING_CPP_VERSION}.tar.gz \
              opentracing-cpp-${OPENTRACING_CPP_VERSION} \
  ### Build lua-resty-hmac
    && cd /tmp \
    && git clone https://github.com/jkeys089/lua-resty-hmac.git \
    && cd lua-resty-hmac \
    && make \
    && make install \
    && cd /tmp \
    && rm -rf lua-resty-hmac \

  ### Add Jaeger plugin
    && curl -fSL https://github.com/jaegertracing/jaeger-client-cpp/archive/refs/tags/v${JAEGER_TRACING_VERSION}.tar.gz -o v${JAEGER_TRACING_VERSION}.tar.gz \
    && tar xf v${JAEGER_TRACING_VERSION}.tar.gz \
    && cd jaeger-client-cpp-${JAEGER_TRACING_VERSION} \
    && mkdir build && cd build \
    && cmake CXXFLAGS="-Wno-error=deprecated-copy" -DCMAKE_INSTALL_PREFIX=/opt -DJAEGERTRACING_PLUGIN=ON -DHUNTER_CONFIGURATION_TYPES=Release -DHUNTER_BUILD_SHARED_LIBS=OFF .. \
    && make && make install \
    && cp libjaegertracing_plugin.so /usr/local/lib/ \
    && cd ../.. && rm -rf v${JAEGER_TRACING_VERSION}.tar.gz \
    && rm -rf jaeger-client-cpp-${JAEGER_TRACING_VERSION} && rm -rf /root/.hunter

RUN cd /usr/local \
    && curl -fSL https://github.com/opentracing-contrib/nginx-opentracing/archive/v${NGINX_OPENTRACING_VERSION}.tar.gz -o nginx-opentracing-${NGINX_OPENTRACING_VERSION}.tar.gz \
    && tar -zxf nginx-opentracing-${NGINX_OPENTRACING_VERSION}.tar.gz \
    && rm nginx-opentracing-${NGINX_OPENTRACING_VERSION}.tar.gz \
    && cd /tmp \
    && curl -fSL https://openresty.org/download/openresty-${RESTY_VERSION}.tar.gz -o openresty-${RESTY_VERSION}.tar.gz \
    && tar xzf openresty-${RESTY_VERSION}.tar.gz \
    && cd /tmp/openresty-${RESTY_VERSION} \
    && ./configure -j${RESTY_J} ${_RESTY_CONFIG_DEPS} ${RESTY_CONFIG_OPTIONS} ${RESTY_CONFIG_OPTIONS_MORE} \
    && make -j${RESTY_J} \
    && make -j${RESTY_J} install \
    && cd /tmp \
    && rm -rf \
        openssl-${RESTY_OPENSSL_VERSION} \
        openssl-${RESTY_OPENSSL_VERSION}.tar.gz \
        openresty-${RESTY_VERSION}.tar.gz openresty-${RESTY_VERSION} \
        pcre-${RESTY_PCRE_VERSION}.tar.gz pcre-${RESTY_PCRE_VERSION}

RUN cd /tmp \
    && curl -fSL  https://luarocks.github.io/luarocks/releases/luarocks-${RESTY_LUAROCKS_VERSION}.tar.gz -o luarocks-${RESTY_LUAROCKS_VERSION}.tar.gz \
    && tar xzf luarocks-${RESTY_LUAROCKS_VERSION}.tar.gz \
    && cd luarocks-${RESTY_LUAROCKS_VERSION} \
    && ./configure \
        --prefix=/usr/local/openresty/luajit \
        --with-lua=/usr/local/openresty/luajit \
        --lua-suffix=jit-2.1.0-beta3 \
        --with-lua-include=/usr/local/openresty/luajit/include/luajit-2.1 \
    && make build \
    && make install \
    && cd /tmp \
    && if [ -n "${RESTY_EVAL_POST_MAKE}" ]; then eval $(echo ${RESTY_EVAL_POST_MAKE}); fi \
    && rm -rf luarocks-${RESTY_LUAROCKS_VERSION} luarocks-${RESTY_LUAROCKS_VERSION}.tar.gz \
    && if [ -n "${RESTY_ADD_PACKAGE_BUILDDEPS}" ]; then DEBIAN_FRONTEND=noninteractive apt-get remove --purge "${RESTY_ADD_PACKAGE_BUILDDEPS}" ; fi \
    && DEBIAN_FRONTEND=noninteractive apt-get autoremove -y \
    && ln -sf /dev/stdout /usr/local/openresty/nginx/logs/access.log \
    && ln -sf /dev/stderr /usr/local/openresty/nginx/logs/error.log

# Add additional binaries into PATH for convenience
ENV PATH=$PATH:/usr/local/openresty/luajit/bin:/usr/local/openresty/nginx/sbin:/usr/local/openresty/bin

# Add LuaRocks paths
# If OpenResty changes, these may need updating:
#    /usr/local/openresty/bin/resty -e 'print(package.path)'
#    /usr/local/openresty/bin/resty -e 'print(package.cpath)'
ENV LUA_PATH="/usr/local/openresty/site/lualib/?.ljbc;/usr/local/openresty/site/lualib/?/init.ljbc;/usr/local/openresty/lualib/?.ljbc;/usr/local/openresty/lualib/?/init.ljbc;/usr/local/openresty/site/lualib/?.lua;/usr/local/openresty/site/lualib/?/init.lua;/usr/local/openresty/lualib/?.lua;/usr/local/openresty/lualib/?/init.lua;./?.lua;/usr/local/openresty/luajit/share/luajit-2.1.0-beta3/?.lua;/usr/local/share/lua/5.1/?.lua;/usr/local/share/lua/5.1/?/init.lua;/usr/local/openresty/luajit/share/lua/5.1/?.lua;/usr/local/openresty/luajit/share/lua/5.1/?/init.lua;/usr/local/openresty/lualib/?/?.lua;/usr/local/openresty/lualib/thrift/?.lua"
ENV LUA_CPATH="/usr/local/openresty/site/lualib/?.so;/usr/local/openresty/lualib/?.so;./?.so;/usr/local/lib/lua/5.1/?.so;/usr/local/openresty/luajit/lib/lua/5.1/?.so;/usr/local/lib/lua/5.1/loadall.so;/usr/local/openresty/luajit/lib/lua/5.1/?.so;/usr/local/openresty/lualib/?/?.so"


COPY lua-thrift /usr/local/openresty/lualib/thrift
COPY lua-bridge-tracer /tmp/lua-bridge-tracer
COPY lua-json/json.lua /usr/local/openresty/lualib/json/json.lua

RUN cd /usr/local/openresty/lualib/thrift/src \
    && make \
    && make install \
  ### Build bridge tracer
    && cd /tmp/lua-bridge-tracer \
    && mkdir -p cmake-build \
    && cd cmake-build \
    && cmake -DCMAKE_BUILD_TYPE=Release \
       .. \
    && make -j$(nproc) \
    && make -j$(nproc) install
#    && cd /tmp
#    && rm -rf lua-bridge-tracer

ENV LUA_PATH=$LUA_PATH;/gen-lua/?.lua

RUN luarocks install long \
    && luarocks install lua-resty-jwt \
    && ldconfig


# Copy nginx configuration files
COPY nginx.conf /usr/local/openresty/nginx/conf/nginx.conf
COPY nginx.vh.default.conf /etc/nginx/conf.d/default.conf

CMD ["/usr/local/openresty/bin/openresty", "-g", "daemon off;"]

# Use SIGQUIT instead of default SIGTERM to cleanly drain requests
# See https://github.com/openresty/docker-openresty/blob/master/README.md#tips--pitfalls
STOPSIGNAL SIGQUIT
