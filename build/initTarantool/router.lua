#!/usr/bin/env tarantool

local vshard = require('vshard')
local log = require('log')
local fiber = require('fiber')

box.cfg {
    listen = 3301,
    memtx_memory = 128 * 1024 * 1024,
    work_dir = '/var/lib/tarantool'
}

-- Create spaces before configuring vshard
box.once('init', function()
    -- Create session space
    local sessions = box.schema.space.create('sessions', {
        format = {
            {name = 'session_id', type = 'string'},
            {name = 'user_id', type = 'string'},
            {name = 'role', type = 'string'}
        },
        if_not_exists = true
    })

    -- Create primary index
    sessions:create_index('primary', {
        parts = {{field = 'session_id', type = 'string'}},
        if_not_exists = true
    })

    -- Grant permissions
    box.schema.user.grant('guest', 'read,write', 'space', 'sessions', nil, {if_not_exists = true})
end)

-- Настраиваем роутер
vshard.router.cfg({
    bucket_count = 100,
    sharding = {
        ["shard1"] = {
            replicas = {
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
    }
})

log.info("Starting router configuration")

-- Автоматический bootstrap: запускаем в отдельной fiber, с ретраями
box.once("bootstrap", function()
    fiber.create(function()
        local ok, err
        repeat
            log.info("Attempting cluster bootstrap...")
            ok, err = vshard.router.bootstrap()
            if not ok then
                log.warn("Bootstrap failed: %s. Retrying in 1 sec...", err)
                fiber.sleep(1)
            end
        until ok
        log.info("Cluster bootstrap completed successfully.")
    end)
end)

log.info("Router started at port 3301")

return vshard.router
