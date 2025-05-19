#!/usr/bin/env tarantool

local vshard = require('vshard')
local log = require('log')

box.cfg {
    listen = 3302,
    memtx_memory = 128 * 1024 * 1024,
    vinyl_memory = 128 * 1024 * 1024,
    replication = {'storage2:3303'},
    work_dir = '/var/lib/tarantool'
}

local schemas = require('schemas')
schemas.init_sessions_space()

vshard.storage.cfg({
    bucket_count = 100,
    sharding = {
        ['shard1'] = {
            replicas = {
                ['storage1'] = {
                    uri = 'storage1:3302',
                    name = 'storage1',
                    master = true
                },
                ['storage2'] = {
                    uri = 'storage2:3303',
                    name = 'storage2'
                }
            }
        }
    }
}, 'storage1')

log.info("Storage1 started at port 3302")
