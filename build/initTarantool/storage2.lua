#!/usr/bin/env tarantool

local vshard = require('vshard')
local log = require('log')

box.cfg {
    listen = 3303,
    memtx_memory = 128 * 1024 * 1024,
    vinyl_memory = 128 * 1024 * 1024,
    replication = {'storage1:3302'},
    work_dir = '/var/lib/tarantool'
}

-- Create spaces and indexes
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

-- Configure storage
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
}, 'storage2')

log.info("Storage2 started at port 3303")
