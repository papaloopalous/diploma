local M = {}

function M.init_sessions_space()
    box.once('sessions_space_init', function()
        local sessions = box.schema.space.create('sessions', {
            format = {
                {name = 'session_id', type = 'string'},
                {name = 'user_id',   type = 'string'},
                {name = 'role',      type = 'string'},
                {name = 'expires_at',type = 'number'}
            },
            if_not_exists = true
        })

        sessions:create_index('primary', {
            parts = {{field = 'session_id', type = 'string'}},
            if_not_exists = true
        })

        sessions:create_index('expires', {
            parts = {{field = 'expires_at', type = 'number'}},
            if_not_exists = true
        })

        box.schema.user.grant('guest', 'read,write',
                               'space', 'sessions', nil,
                               {if_not_exists = true})
    end)
end

return M
