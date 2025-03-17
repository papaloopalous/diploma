#!/usr/bin/env tarantool

local vshard = require('vshard')
local log = require('log')
local fiber = require('fiber')

box.cfg{
    listen = 3301
}

-- Настраиваем роутер
vshard.router.cfg({
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
    },
    bucket_count = 100
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
