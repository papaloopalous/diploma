# Используем Ubuntu, чтобы установить tarantool + tarantool-dev
FROM ubuntu:latest

RUN apt-get update && apt-get install -y \
    curl \
    tarantool \
    tarantool-dev \
    lua5.1 \
    lua5.1-dev \
    luarocks \
 && rm -rf /var/lib/apt/lists/*

# Установка vshard через tarantoolctl
RUN tarantoolctl rocks install vshard

WORKDIR /opt/tarantool

CMD ["tarantool"]
