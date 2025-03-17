#!/usr/bin/env tarantool

box.cfg {
    listen = 3301
}

box.once('init', function()
    box.schema.space.create('users')
end)