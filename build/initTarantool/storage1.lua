#!/usr/bin/env tarantool

local vshard = require('vshard')
local log = require('log')

-- Настраиваем Tarantool
box.cfg{
    listen = 3302,
    memtx_memory = 512 * 1024 * 1024,
}

-- Инициализируем vshard на этом узле (storage)
vshard.storage.cfg({
    sharding = {
        ["shard1"] = {
            replicas = {
                -- НАЗВАНИЕ КЛЮЧА (uuid или любое) не так важно, главное - master = true
                ["storage1"] = {
                    uri = "storage1:3302",
                    name = "storage1",
                    master = true
                },
                ["storage2"] = {
                    uri = "storage2:3303",
                    name = "storage2"
                }
            }
        }
    },
    bucket_count = 100
}, box.info.uuid)

-- Создаем spaces и индексы. Выполняется только один раз.
box.once("init_storage1", function()
    -- ОБЯЗАТЕЛЬНЫЙ СИСТЕМНЫЙ СПЕЙС для vshard
    box.schema.space.create('_bucket', { if_not_exists = true })
    box.space._bucket:create_index('pk', { parts = {1, 'unsigned'}, if_not_exists = true })

    -- Пользовательский спейс (пример)
    box.schema.space.create('users', { if_not_exists = true })
    box.space.users:create_index('primary', {
        parts = {1, 'unsigned'},
        if_not_exists = true
    })

    -- Создаём бакеты на мастере (только на одном узле!)
    vshard.storage.bucket_force_create(1, 100)

    log.info("Storage1 (master) initialized")
end)

log.info("Storage1 (master) is ready")
return vshard.storage
